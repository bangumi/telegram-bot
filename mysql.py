import asyncio
import dataclasses

import asyncmy

from lib import config
from lib.debezium import ChiiNotifyField


@dataclasses.dataclass
class User:
    uid: int
    username: str
    nickname: str


async def create_mysql_client():
    db = MySql(
        pool=await asyncmy.create_pool(
            host=config.MYSQL_DSN.host,
            port=config.MYSQL_DSN.port,
            user=config.MYSQL_DSN.username,
            password=config.MYSQL_DSN.password,
            db=config.MYSQL_DSN.path.lstrip("/"),
            autocommit=False,
        ),
    )
    return db


class MySql:
    __pool: asyncmy.Pool

    def __init__(self, pool: asyncmy.Pool):
        self.__pool = pool
        self.pool = pool

    async def get_notify_field(self, ntf_id) -> ChiiNotifyField:
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

    async def get_user(self, uid) -> User:
        conn: asyncmy.Connection
        async with self.__pool.acquire() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT uid, username, nickname from chii_members where uid = %s",
                uid,
            )
            uid, username, nickname = await cur.fetchone()
            return User(uid=uid, username=username, nickname=nickname)


async def test():
    pool = await create_mysql_client()

    r = await pool.get_notify_field(5)
    print(r)
    pool.pool.close()
    await pool.pool.wait_closed()


def main():
    asyncio.run(test())
