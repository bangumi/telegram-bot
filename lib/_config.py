import os
import sys

import yarl
from pydantic import (
    BaseModel,
    KafkaDsn,
    MySQLDsn,
    PostgresDsn,
    RedisDsn,
    field_validator,
)


class Settings(BaseModel, validate_default=True, arbitrary_types_allowed=True):
    debug: bool = os.environ.get("DEBUG", sys.platform == "win32")  # type: ignore

    TELEGRAM_BOT_TOKEN: str = os.environ["TELEGRAM_BOT_TOKEN"]

    BANGUMI_APP_ID: str = os.environ["BANGUMI_APP_ID"]
    BANGUMI_APP_SECRET: str = os.environ["BANGUMI_APP_SECRET"]

    HTTP_PORT: int | None = os.environ.get("HTTP_PORT")

    QUEUE_SIZE: int = os.environ.get("QUEUE_SIZE") or 1  # type: ignore

    REDIS_DSN: RedisDsn = os.environ["REDIS_DSN"]

    PG_DSN: PostgresDsn = os.environ["PG_DSN"]
    MYSQL_DSN: MySQLDsn = os.environ["MYSQL_DSN"]

    KAFKA_BROKER: KafkaDsn = os.environ["KAFKA_DSN"]

    EXTERNAL_HTTP_ADDRESS: yarl.URL = os.environ.get(
        "EXTERNAL_HTTP_ADDRESS", "http://127.0.0.1:4562"
    )

    @field_validator("EXTERNAL_HTTP_ADDRESS", mode="plain")
    @classmethod
    def __external_http_address(cls, v: str) -> yarl.URL:
        return yarl.URL(v).with_path("")
