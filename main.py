from __future__ import annotations

import asyncio
import contextlib
import functools
import html
import http
import logging
import secrets
import sys
import time
from threading import Thread
from typing import Any, NamedTuple

import aiorwlock
import httpx
import msgspec.json
import redis.asyncio as redis
import sslog
import starlette
import starlette.applications
import telegram as tg
import telegram.ext
import uvicorn
import yarl
from sslog import logger
from starlette.exceptions import HTTPException
from starlette.requests import Request
from starlette.responses import PlainTextResponse, RedirectResponse, Response
from telegram._utils.defaultvalue import DEFAULT_NONE, DefaultValue
from telegram.constants import ParseMode

import pg
from cfg import notify_types
from kafka import KafkaConsumer, Msg
from lib import config, debezium
from lib.debezium import ChiiPm
from mysql import MySql, create_mysql_client
from pg import PG, create_pg_client

logging.basicConfig(level=logging.INFO, handlers=[sslog.InterceptHandler()])
logging.getLogger("httpx").setLevel(logging.WARN)


class Item(NamedTuple):
    chat_id: int
    text: str
    parse_mode: str | DefaultValue[None] = DEFAULT_NONE


class RedisOAuthState(msgspec.Struct):
    chat_id: int


class BangumiOAuthResponse(msgspec.Struct):
    user_id: int


def state_to_redis_key(state: str) -> str:
    return f"tg-bot-oauth:{state}"


class TelegramApplication:
    app: tg.ext.Application[Any, Any, Any, Any, Any, Any]
    bot: tg.Bot

    mysql: MySql

    __tg: TelegramApplication
    __user_ids: dict[int, set[int]]
    __lock: aiorwlock.RWLock
    __queue: asyncio.Queue[Item]
    __background_tasks: set[Any]

    # queue for kafka message
    __notify_queue: asyncio.Queue[Msg]
    __pm_queue: asyncio.Queue[Msg]

    def __init__(
        self, redis_client: redis.Redis, pg_client: pg.PG, mysql_client: MySql
    ):
        builder = tg.ext.Application.builder()

        if sys.platform == "win32":
            proxy_url = "http://192.168.1.3:7890"
            builder = builder.proxy(proxy_url).get_updates_proxy(proxy_url)

        application = builder.token(config.TELEGRAM_BOT_TOKEN).build()

        # on different commands - answer in Telegram
        application.add_handler(tg.ext.CommandHandler("start", self.start_command))
        application.add_handler(tg.ext.CommandHandler("logout", self.logout_command))
        application.add_handler(tg.ext.CommandHandler("help", self.start_command))
        application.add_handler(tg.ext.CommandHandler("debug", self.debug_command))

        application.add_error_handler(self.error_handler)

        self.app = application
        self.bot = application.bot
        self.redis = redis_client
        self.pg = pg_client
        self.mysql = mysql_client
        self.__lock = aiorwlock.RWLock()
        self.__queue = asyncio.Queue(maxsize=config.QUEUE_SIZE)
        self.__notify_queue = asyncio.Queue(maxsize=config.QUEUE_SIZE)
        self.__pm_queue = asyncio.Queue(maxsize=config.QUEUE_SIZE)
        self.__background_tasks = set()

    async def start(self) -> None:
        await self.read_from_db()
        self.start_tasks()
        await self.app.initialize()
        updater = self.app.updater
        if updater is None:
            logger.fatal("app.updater is None")
            return
        await updater.start_polling(allowed_updates=tg.Update.ALL_TYPES)
        logger.info("telegram bot start")
        await self.app.start()

    @logger.catch
    async def start_command(
        self, update: tg.Update, _context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        """Send a message when the command /help is issued."""
        logger.trace("start command")
        if update.message is None or update.effective_chat is None:
            return

        if user := await self.pg.is_authorized_user(chat_id=update.effective_chat.id):
            await update.message.reply_text(
                f"你已经作为用户 {user.user_id} 成功进行认证"
            )
            return

        token = secrets.token_urlsafe(32)
        await self.redis.set(
            state_to_redis_key(token),
            msgspec.json.encode(RedisOAuthState(chat_id=update.effective_chat.id)),
            ex=60,
        )
        reply_markup = tg.InlineKeyboardMarkup(
            [
                [
                    tg.InlineKeyboardButton(
                        "认证 bangumi 账号",
                        url=f"{config.EXTERNAL_HTTP_ADDRESS}/redirect?state={token}",
                    )
                ]
            ]
        )
        await update.message.reply_text("请在60s内进行认证", reply_markup=reply_markup)

    @logger.catch
    async def help_command(
        self, update: tg.Update, _context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.trace("help command")
        if update.message is None:
            return
        await update.message.reply_text("use command `/start`")

    @logger.catch
    async def logout_command(
        self, update: tg.Update, _context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.trace("logout command")
        if update.effective_chat is None:
            return
        if update.message is None:
            return

        await self.pg.logout(chat_id=update.effective_chat.id)
        await self.read_from_db()
        await update.message.reply_text("成功登出")

    @logger.catch
    async def debug_command(
        self, update: tg.Update, _context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.trace("debug command")

        if update.effective_chat is None:
            return
        if update.message is None:
            return

        await update.message.reply_text(f"chat_id: {update.effective_chat.id}")
        await self.bot.send_message(
            chat_id=update.effective_chat.id, text="debug command"
        )

    @logger.catch
    async def error_handler(
        self, _update: object, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.exception("Exception while handling an update", error=context.error)

    async def send_notification(
        self,
        chat_id: int,
        text: str,
        parse_mode: str | DefaultValue[None] = DEFAULT_NONE,
    ) -> None:
        await self.bot.send_message(chat_id=chat_id, text=text, parse_mode=parse_mode)

    async def get_chats(self, user_id: int) -> set[int] | None:
        async with self.__lock.reader:
            return self.__user_ids.get(user_id)

    async def read_from_db(self) -> None:
        rr = await self.pg.get_watched_users()
        async with self.__lock.writer:
            self.__user_ids = rr

    @sslog.logger.catch
    def __watch_kafka_messages(self, loop: asyncio.AbstractEventLoop) -> None:
        logger.info("start watching kafka message")
        consumer = KafkaConsumer(
            *[
                "debezium.chii.bangumi." + table
                for table in [
                    "chii_pms",
                    "chii_notify",
                ]
            ]
        )

        while True:
            try:
                for msg in consumer:
                    logger.debug("new message", topic=msg.topic, offset=msg.offset)
                    match msg.topic:
                        case "debezium.chii.bangumi.chii_pms":
                            asyncio.run_coroutine_threadsafe(
                                self.__pm_queue.put(msg), loop
                            ).result()
                        case "debezium.chii.bangumi.chii_notify":
                            asyncio.run_coroutine_threadsafe(
                                self.__notify_queue.put(msg), loop
                            ).result()
            except Exception:
                logger.exception("failed to fetch kafka message")

    async def __handle_new_notify(self) -> None:
        while True:
            msg = await self.__notify_queue.get()
            logger.debug(
                "new message from chii_notify", topic=msg.topic, offset=msg.offset
            )
            try:
                await self.handle_notify_change(msg)
            except Exception:
                logger.exception("failed to handle notify change event")

    async def __handle_new_pm(self) -> None:
        while True:
            msg = await self.__pm_queue.get()
            logger.debug(
                "new message from chii_pms", topic=msg.topic, offset=msg.offset
            )
            try:
                await self.handle_pm(msg)
            except Exception:
                logger.exception("failed to handle member change event")

    def watch_kafka_message(self) -> None:
        loop = asyncio.get_running_loop()
        self.__background_tasks.add(loop.create_task(self.__handle_new_notify()))
        self.__background_tasks.add(loop.create_task(self.__handle_new_pm()))
        t = Thread(
            target=functools.partial(self.__watch_kafka_messages, loop), daemon=True
        )
        self.__background_tasks.add(t)
        t.start()

    notify_decoder = msgspec.json.Decoder(debezium.NotifyValue)

    async def handle_notify_change(self, m: Msg) -> None:
        if not m.value:
            return

        value: debezium.NotifyValue = self.notify_decoder.decode(m.value)
        if value.op != "c":
            return

        notify = value.after
        if notify is None:
            return

        if notify.timestamp < time.time() - 60 * 2:
            # skip notification older than 2 min
            return

        char = await self.get_chats(notify.nt_uid)
        if not char:
            return

        cfg = notify_types.get(notify.nt_type)
        if not cfg:
            return

        field = await self.mysql.get_notify_field(notify.nt_mid)
        user = await self.mysql.get_user(notify.nt_from_uid)

        url = f"{cfg.url.rstrip('/')}/{field.ntf_rid}"

        if notify.nt_related_id:
            url += f"{cfg.anchor}{notify.nt_related_id}"

        msg = f"<code>{user.nickname}</code>"

        if cfg.suffix:
            msg += f" {cfg.prefix} <b>{html.escape(field.ntf_title)}</b> {cfg.suffix}"
        else:
            msg += f"{cfg.prefix}"

        msg += f"\n\n{url}"

        logger.info("should send message for pm", user_id=notify.nt_from_uid)

        for c in char:
            await self.__queue.put(Item(c, msg, parse_mode=ParseMode.HTML))

    pms_decoder = msgspec.json.Decoder(debezium.DebeziumValue[ChiiPm])

    async def handle_pm(self, msg: Msg) -> None:
        if not msg.value:
            return

        try:
            value: debezium.DebeziumValue[ChiiPm] = self.pms_decoder.decode(msg.value)
        except msgspec.ValidationError:
            return

        after = value.after
        if after is None:
            return

        if not after.msg_new:
            return

        user_id = after.msg_rid

        chats = await self.get_chats(user_id)
        if not chats:
            return

        blocklist = await self.get_block_list(user_id)
        if after.msg_sid in blocklist:
            return

        logger.info("should send message for pm", user_id=user_id)

        for c in chats:
            await self.__queue.put(Item(c, f"你有一条来自 {after.msg_sid} 的新私信"))

    async def get_block_list(self, uid: int) -> set[int]:
        return await self.mysql.get_blocklist(uid)

    async def start_queue_consumer(self) -> None:
        while True:
            chat_id, text, parse_mode = await self.__queue.get()
            try:
                await self.send_notification(chat_id, text, parse_mode)
            except Exception:
                logger.exception("failed to send message to chat")

    def start_tasks(self) -> None:
        loop = asyncio.get_event_loop()
        task = loop.create_task(self.start_queue_consumer())
        self.watch_kafka_message()
        self.__background_tasks.add(task)


class OAuthHTTPServer:
    def __init__(self, r: redis.Redis, db: PG, bot: TelegramApplication):
        self.app = starlette.applications.Starlette()
        self.app.add_route("/", self.index_path, ["GET"])
        self.app.add_route("/redirect", self.oauth_redirect, ["GET"])
        self.app.add_route("/callback", self.oauth_callback, ["GET"])
        self.redirect_url = str(config.EXTERNAL_HTTP_ADDRESS.with_path("/callback"))

        self.redis = r

        self.http_client = httpx.AsyncClient()
        self.db = db
        self.tg = bot

    async def start(self) -> None:
        port = config.HTTP_PORT or config.EXTERNAL_HTTP_ADDRESS.port or 4098
        server = uvicorn.Server(
            uvicorn.Config(
                app=self.app,
                host="0.0.0.0",
                lifespan="on",
                workers=1,
                port=port,
                log_level=logging.WARN,
                access_log=False,
            )
        )
        logger.info("http server listen on port={}", port)
        await server.serve()

    async def index_path(self, _request: Request) -> Response:
        return Response("index page")

    async def oauth_redirect(self, request: Request) -> Response:
        state = request.query_params.get("state")
        if not state:
            return PlainTextResponse("请求无效，请在 telegram 中重新认证")

        return RedirectResponse(
            str(
                yarl.URL.build(
                    scheme="https",
                    host="bgm.tv",
                    path="/oauth/authorize",
                    query={
                        "client_id": config.BANGUMI_APP_ID,
                        "response_type": "code",
                        "redirect_uri": self.redirect_url,
                        "state": state,
                    },
                )
            )
        )

    async def oauth_callback(self, request: Request) -> Response:
        code = request.query_params.get("code")
        state = request.query_params.get("state")
        if not code or not state:
            raise HTTPException(
                http.HTTPStatus.BAD_REQUEST,
                detail="非法请求，请使用telegram重新获取认证链接",
            )
        redis_state_value_raw = await self.redis.get(state_to_redis_key(state))
        if redis_state_value_raw is None:
            raise HTTPException(
                http.HTTPStatus.BAD_REQUEST,
                detail="非法请求，请使用telegram重新获取认证链接",
            )

        redis_state = msgspec.json.decode(redis_state_value_raw, type=RedisOAuthState)

        resp = await self.http_client.post(
            "https://bgm.tv/oauth/access_token",
            data={
                "client_id": config.BANGUMI_APP_ID,
                "client_secret": config.BANGUMI_APP_SECRET,
                "grant_type": "authorization_code",
                "code": code,
                "redirect_uri": self.redirect_url,
            },
        )
        if resp.status_code >= 300:
            logger.error("bad oauth response", data=resp.json())
            raise HTTPException(http.HTTPStatus.BAD_GATEWAY, "请尝试重新认证")
        data = msgspec.json.decode(resp.text, type=BangumiOAuthResponse)

        await self.db.insert_chat_bangumi_map(
            user_id=data.user_id, chat_id=redis_state.chat_id
        )

        await self.tg.send_notification(
            chat_id=redis_state.chat_id, text=f"已经成功关联用户 {data.user_id}"
        )

        await self.tg.read_from_db()

        return PlainTextResponse("你已经成功认证，请关闭页面返回 telegram")


async def start(loop: asyncio.AbstractEventLoop) -> Any:
    redis_client = redis.client.Redis.from_url(str(config.REDIS_DSN))

    pg_client = await create_pg_client()

    tg_app = TelegramApplication(
        redis_client=redis_client,
        pg_client=pg_client,
        mysql_client=await create_mysql_client(),
    )

    http_server = OAuthHTTPServer(r=redis_client, db=pg_client, bot=tg_app)

    tasks = set()
    tasks.add(loop.create_task(tg_app.start(), name="telegram bot"))
    tasks.add(loop.create_task(http_server.start(), name="exc"))

    with contextlib.suppress(Exception):
        await asyncio.gather(*tasks)

    await loop.shutdown_default_executor()


def main() -> None:
    loop = asyncio.get_event_loop()
    loop.run_until_complete(start(loop))


if __name__ == "__main__":
    main()
