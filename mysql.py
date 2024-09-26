from __future__ import annotations

import dataclasses

import asyncmy

from lib import config
from lib.debezium import ChiiNotifyField


@dataclasses.dataclass
class User:
    uid: int
    username: str
    nickname: str


async def create_mysql_client() -> MySql:
    db = MySql(
        pool=await asyncmy.create_pool(
            host=config.MYSQL_DSN.host,
            port=config.MYSQL_DSN.port,
            user=config.MYSQL_DSN.user,
            password=config.MYSQL_DSN.password,
            db=config.MYSQL_DSN.path.lstrip("/"),
            autocommit=False,
            pool_recycle=3600 * 7,
        ),
    )
    return db


class MySql:
    __pool: asyncmy.Pool

    def __init__(self, pool: asyncmy.Pool):
        self.__pool = pool
        self.pool = pool

    async def get_notify_field(self, ntf_id: int) -> ChiiNotifyField:
        conn: asyncmy.Connection
        async with self.__pool.acquire() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT ntf_id,ntf_hash,ntf_rid,ntf_title from chii_notify_field where ntf_id = %s",
                ntf_id,
            )
            r = await cur.fetchone()
            return ChiiNotifyField(
                ntf_id=r[0],
                ntf_hash=r[1],
                ntf_rid=r[2],
                ntf_title=r[3],
            )

    async def get_user(self, uid: int) -> User:
        conn: asyncmy.Connection
        async with self.__pool.acquire() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT uid, username, nickname from chii_members where uid = %s",
                uid,
            )
            uid, username, nickname = await cur.fetchone()
            return User(uid=uid, username=username, nickname=nickname)
