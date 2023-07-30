import os
import sys

import yarl
from pydantic import (
    BaseModel,
    KafkaDsn,
    RedisDsn,
    field_validator,
)


class Settings(BaseModel, validate_default=True, arbitrary_types_allowed=True):
    debug: bool = os.environ.get("DEBUG", sys.platform == "win32")  # type: ignore

    TELEGRAM_BOT_TOKEN: str = os.environ["TELEGRAM_BOT_TOKEN"]

    BANGUMI_APP_ID: str = os.environ["BANGUMI_APP_ID"]
    BANGUMI_APP_SECRET: str = os.environ["BANGUMI_APP_SECRET"]

    HTTP_PORT: str | None = os.environ.get("HTTP_PORT")

    QUEUE_SIZE: int = os.environ.get("QUEUE_SIZE") or 1  # type: ignore

    REDIS_DSN: RedisDsn = os.environ["REDIS_DSN"]

    MYSQL_HOST: str = os.getenv("MYSQL_HOST") or "127.0.0.1"  # type: ignore
    MYSQL_PORT: int = os.getenv("MYSQL_PORT") or 3306  # type: ignore
    MYSQL_USER: str = os.getenv("MYSQL_USER") or "user"  # type: ignore
    MYSQL_PASS: str = os.getenv("MYSQL_PASS") or "password"  # type: ignore
    MYSQL_DB: str = os.getenv("MYSQL_DB") or "bangumi"  # type: ignore

    KAFKA_BROKER: KafkaDsn = os.environ["KAFKA_DSN"]

    EXTERNAL_HTTP_ADDRESS: yarl.URL = os.environ.get(
        "EXTERNAL_HTTP_ADDRESS", "http://127.0.0.1:4098"
    )

    @property
    def MYSQL_SYNC_DSN(self) -> str:
        return "mysql+aiomysql://{}:{}@{}:{}/{}".format(
            self.MYSQL_USER,
            self.MYSQL_PASS,
            self.MYSQL_HOST,
            self.MYSQL_PORT,
            self.MYSQL_DB,
        )

    @field_validator("EXTERNAL_HTTP_ADDRESS", mode="plain")
    @classmethod
    def __external_http_address(cls, v: str) -> yarl.URL:
        return yarl.URL(v).with_path("")
