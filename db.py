import aiomysql
import pydantic

from lib import config


class Table(pydantic.BaseModel):
    chat_id: int
    user_id: int
    disabled: bool


async def create_db():
    return DB(
        pool=await aiomysql.create_pool(
            host=config.MYSQL_HOST,
            port=config.MYSQL_PORT,
            user=config.MYSQL_USER,
            password=config.MYSQL_PASS,
            db=config.MYSQL_DB,
            autocommit=False,
        ),
    )


class DB:
    def __init__(self, pool: aiomysql.Pool):
        self.pool = pool

    async def insert_chat_bangumi_map(self, *, chat_id: int, user_id: int):
        conn: aiomysql.Connection
        cur: aiomysql.Cursor
        async with self.pool.acquire() as conn, conn.cursor() as cur:
            await cur.execute(
                "REPLACE INTO telegram_notify_chat(chat_id, user_id, disabled) VALUES (%s, %s, 0)",
                (chat_id, user_id),
            )
            await conn.commit()

    async def is_authorized_user(self, *, chat_id: int) -> Table | None:
        conn: aiomysql.Connection
        cur: aiomysql.Cursor
        async with self.pool.acquire() as conn, conn.cursor() as cur:
            rr = await cur.execute(
                "SELECT chat_id, user_id, disabled from telegram_notify_chat where chat_id=%s",
                (chat_id,),
            )
            if rr:
                r = await cur.fetchone()
                return Table(chat_id=r[0], user_id=r[1], disabled=r[2])
            return None
