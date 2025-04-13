from __future__ import annotations

import asyncio
import functools
import html
import logging
import sys
import time
from threading import Thread
from typing import Any, NamedTuple

import aiorwlock
import msgspec.json
import redis.asyncio as redis
import sslog
import telegram as tg
import telegram.ext as tg_ext
from async_lru import alru_cache
from sslog import logger
from telegram.constants import ParseMode

import pg
from cfg import notify_types
from kafka import KafkaConsumer, Msg
from lib import config, debezium
from lib.debezium import ChiiPm
from mysql import MySql, User

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
    app: tg_ext.Application[Any, Any, Any, Any, Any, Any]
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
        builder = tg_ext.Application.builder()

        if sys.platform == "win32":
            proxy_url = "http://192.168.1.3:7890"
            builder = builder.proxy(proxy_url).get_updates_proxy(proxy_url)

        application = builder.token(config.TELEGRAM_BOT_TOKEN).build()

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
        await self.app.initialize()
        updater = self.app.updater
        if updater is None:
            logger.fatal("app.updater is None")
            return
        await updater.start_polling(allowed_updates=tg.Update.ALL_TYPES)
        logger.info("telegram bot start")
        await self.app.start()

    @logger.catch
    async def error_handler(
        self, _update: object, context: tg_ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        logger.exception("Exception while handling an update", error=context.error)

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
        user = await self.get_user_info(notify.nt_from_uid)

        url = f"{cfg.url.rstrip('/')}/{field.ntf_rid}"

        if notify.nt_related_id:
            url += f"{cfg.anchor}{notify.nt_related_id}"

        msg = f"<code>{html.escape(user.nickname)}</code>"

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

        value: debezium.DebeziumValue[ChiiPm] = self.pms_decoder.decode(msg.value)

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
        user = await self.get_user_info(after.msg_sid)

        for c in chats:
            await self.__queue.put(
                Item(
                    c,
                    f"有一条来自 <code>{html.escape(user.nickname)}</code> 的新私信",
                    parse_mode=ParseMode.HTML,
                )
            )

    @alru_cache(1024)
    async def get_user_info(self, uid: int) -> User:
        return await self.mysql.get_user(uid)

    async def get_block_list(self, uid: int) -> set[int]:
        return await self.mysql.get_blocklist(uid)

    async def start_queue_consumer(self) -> None:
        while True:
            chat_id, text, parse_mode = await self.__queue.get()
            try:
                await self.send_notification(chat_id, text, parse_mode)
            except Exception:
                logger.exception("failed to send message to chat")
