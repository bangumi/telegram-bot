from __future__ import annotations

from collections import defaultdict

import asyncpg
import msgspec

from lib import config


class Table(msgspec.Struct):
    chat_id: int
    user_id: int
    disabled: bool


async def create_pg_client() -> PG:
    pool = asyncpg.create_pool(dsn=str(config.PG_DSN))
    await pool
    db = PG(pool)
    await db.init()
    return db


class PG:
    __pool: asyncpg.Pool[asyncpg.Record]

    def __init__(self, pool: asyncpg.Pool[asyncpg.Record]):
        self.__pool: asyncpg.Pool[asyncpg.Record] = pool

    async def init(self) -> None:
        await self.__pool.execute(
            """
            CREATE TABLE IF NOT EXISTS telegram_notify_chat (
                chat_id bigint,
                user_id bigint,
                disabled int2,
                primary key (chat_id, user_id)
            );
            """
        )

    async def insert_chat_bangumi_map(self, *, chat_id: int, user_id: int) -> None:
        await self.__pool.execute(
            "INSERT INTO telegram_notify_chat(chat_id, user_id, disabled) VALUES ($1, $2, 0)",
            chat_id,
            user_id,
        )

    async def logout(self, *, chat_id: int) -> None:
        await self.__pool.execute(
            "DELETE from telegram_notify_chat where chat_id=$1",
            chat_id,
        )

    async def is_authorized_user(self, *, chat_id: int) -> Table | None:
        rr = await self.__pool.fetchrow(
            "SELECT chat_id, user_id, disabled from telegram_notify_chat where chat_id=$1",
            chat_id,
        )
        if rr:
            return Table(chat_id=rr[0], user_id=rr[1], disabled=rr[2])
        return None

    async def get_watched_users(self) -> dict[int, set[int]]:
        rr = await self.__pool.fetch(
            "SELECT user_id, chat_id from telegram_notify_chat where disabled = 0"
        )
        if rr:
            d = defaultdict(set)
            for user_id, chat_id in rr:
                d[user_id].add(chat_id)
            return d
        return {}

    async def disable_chat(self, chat_id: int) -> None:
        await self.__pool.execute(
            "update telegram_notify_chat set disabled = 1 where chat_id = $1 and disabled = 0",
            chat_id,
        )
