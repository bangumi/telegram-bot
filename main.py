from __future__ import annotations

import asyncio
import http
import logging
import secrets
import sys
import traceback
from typing import NamedTuple

import aiorwlock
import httpx
import msgspec.json
import redis.asyncio as redis
import starlette
import starlette.applications
import telegram as tg
import telegram.ext
import uvicorn
import yarl
from aiokafka import AIOKafkaConsumer, ConsumerRecord
from loguru import logger
from starlette.exceptions import HTTPException
from starlette.requests import Request
from starlette.responses import PlainTextResponse, RedirectResponse, Response

import mysql
import pg
from cfg import notify_types
from lib import config, debezium
from pg import PG, create_pg_client

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)
logging.getLogger("httpx").setLevel(logging.WARN)


class Item(NamedTuple):
    chat_id: int
    text: str


class RedisOAuthState(msgspec.Struct):
    chat_id: int


class BangumiOAuthResponse(msgspec.Struct):
    user_id: int


def state_to_redis_key(state: str):
    return f"tg-bot-oauth:{state}"


class TelegramApplication:
    app: tg.ext.Application
    bot: tg.Bot

    mysql: mysql.MySql

    __tg: TelegramApplication
    __user_ids: dict[int, set[int]]
    __lock: aiorwlock.RWLock()
    __queue: asyncio.Queue[Item]

    def __init__(
        self, redis_client: redis.Redis, pg_client: pg.PG, mysql_client=mysql.MySql
    ):
        application = tg.ext.Application.builder()

        if sys.platform == "win32":
            proxy_url = "http://127.0.0.1:7890"
            application = application.proxy_url(proxy_url).get_updates_proxy_url(
                proxy_url
            )

        application = application.token(config.TELEGRAM_BOT_TOKEN).build()

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

    async def init(self):
        await self.read_from_db()
        self.start_tasks()

    @logger.catch
    async def start_command(
        self, update: tg.Update, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        """Send a message when the command /help is issued."""
        logger.trace("start command")
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
        self, update: tg.Update, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.trace("help command")
        await update.message.reply_text("use command `/start`")

    @logger.catch
    async def logout_command(
        self, update: tg.Update, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ):
        logger.trace("logout command")
        await self.pg.logout(chat_id=update.effective_chat.id)
        await self.read_from_db()
        await update.message.reply_text("成功登出")

    @logger.catch
    async def debug_command(
        self, update: tg.Update, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.trace("debug command")
        await update.message.reply_text(f"chat_id: {update.effective_chat.id}")

    @logger.catch
    async def error_handler(
        self, update: object, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.error("Exception while handling an update:", exc_info=context.error)
        tb_list = traceback.format_exception(
            None, context.error, context.error.__traceback__
        )
        tb_string = "".join(tb_list)

        update_str = update.to_dict() if isinstance(update, tg.Update) else str(update)
        print(tb_string)
        print(context.user_data)
        print(context.chat_data)
        print(update_str)

    async def send_notification(self, chat_id: int, text: str):
        await self.bot.send_message(chat_id=chat_id, text=text)

    async def is_watched_user_id(self, user_id: int) -> set[int] | None:
        async with self.__lock.reader:
            return self.__user_ids.get(user_id)

    async def read_from_db(self):
        rr = await self.pg.get_watched_users()
        async with self.__lock.writer:
            self.__user_ids = rr

    async def start_kafka_broker(self):
        logger.info("start_kafka_broker")
        consumer = AIOKafkaConsumer(
            # "debezium.chii.bangumi.chii_members",
            "debezium.chii.bangumi.chii_notify",
            bootstrap_servers=f"{config.KAFKA_BROKER.host}:{config.KAFKA_BROKER.port}",
            group_id="tg-notify-bot",
        )
        await consumer.start()
        try:
            msg: ConsumerRecord
            async for msg in consumer:
                match msg.topic:
                    # case "debezium.chii.bangumi.chii_members":
                    #     await self.handle_member(msg)
                    case "debezium.chii.bangumi.chii_notify":
                        await self.handle_notify_change(msg)

        finally:
            await consumer.stop()

    async def handle_notify_change(self, msg: ConsumerRecord):
        with logger.catch(message="unexpected exception when parsing kafka messages"):
            if not msg.value:
                return

            value: debezium.NotifyValue = msgspec.json.decode(
                msg.value, type=debezium.NotifyValue
            )
            if value.payload.op != "c":
                return

            notify = value.payload.after

            char = await self.is_watched_user_id(notify.nt_uid)
            if not char:
                return

            cfg = notify_types.get(notify.nt_type)
            if not cfg:
                return

            field = await self.mysql.get_notify_field(notify.nt_mid)
            user = await self.mysql.get_user(notify.nt_from_uid)

            msg = f"`{user.nickname}` {cfg.prefix} '{field.ntf_title}' {cfg.suffix}\n\n{cfg.url}/{field.ntf_rid}{cfg.anchor}{notify.nt_related_id}"
            for c in char:
                await self.__queue.put(Item(c, msg))

    async def handle_member(self, msg: ConsumerRecord):
        with logger.catch(message="unexpected exception when parsing kafka messages"):
            if not msg.value:
                return

            value: debezium.MemberValue = msgspec.json.decode(
                msg.value, type=debezium.MemberValue
            )
            if value.payload.op != "u":
                return

            if value.payload.after.new_notify <= value.payload.before.new_notify:
                return

            if not value.payload.after.new_notify:
                return

            user_id = value.payload.after.uid
            if char := await self.is_watched_user_id(user_id):
                for c in char:
                    await self.__queue.put(
                        Item(
                            c,
                            f"你有 {value.payload.after.new_notify} 条新通知",
                        )
                    )

    async def start_queue_consumer(self):
        while True:
            chat_id, text = await self.__queue.get()
            await self.send_notification(chat_id, text)
            self.__queue.task_done()

    def start_tasks(self):
        loop = asyncio.get_event_loop()
        loop.create_task(self.start_queue_consumer())
        loop.create_task(self.start_kafka_broker())


class OAuthHTTPServer:
    def __init__(self, redis: redis.Redis, db: PG, bot: TelegramApplication):
        self.app = starlette.applications.Starlette()
        self.app.add_route("/", self.index_path, ["GET"])
        self.app.add_route("/redirect", self.oauth_redirect, ["GET"])
        self.app.add_route("/callback", self.oauth_callback, ["GET"])
        self.redirect_url = str(config.EXTERNAL_HTTP_ADDRESS.with_path("/callback"))

        self.redis = redis

        self.http_client = httpx.AsyncClient()
        self.db = db
        self.tg = bot

    async def index_path(self, request: Request) -> Response:
        return Response("index page")

    async def oauth_redirect(self, request: Request) -> Response:
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
                        "state": request.query_params.get("state"),
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

    async def start(self):
        logger.info("start http server")
        port = config.HTTP_PORT or config.EXTERNAL_HTTP_ADDRESS.port or 4098
        server = uvicorn.Server(
            uvicorn.Config(
                app=self.app,
                host="0.0.0.0",
                port=port,
                log_level=logging.WARN,
                access_log=False,
            )
        )
        logger.info("http server listen on port={}", port)
        await server.serve()


async def start() -> None:
    redis_client = redis.from_url(str(config.REDIS_DSN))

    pg_client = await create_pg_client()

    tg_app = TelegramApplication(
        redis_client=redis_client,
        pg_client=pg_client,
        mysql_client=await mysql.create_mysql_client(),
    )
    await tg_app.init()

    http_server = OAuthHTTPServer(redis=redis_client, db=pg_client, bot=tg_app)
    async with tg_app.app:
        await tg_app.app.initialize()
        await tg_app.app.start()
        await tg_app.app.updater.start_polling(allowed_updates=tg.Update.ALL_TYPES)

        logger.info("telegram bot start")

        await http_server.start()


def main() -> None:
    asyncio.run(start())


if __name__ == "__main__":
    main()
