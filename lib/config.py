import os
import sys
from typing import Any, TypeVar, Union

import msgspec
import yarl

T = TypeVar("T")


def __convert(v: Any, t: type[T]) -> T:
    msgspec.convert(v, type=t, strict=False)


debug: bool = __convert(os.environ.get("DEBUG", sys.platform == "win32"), t=bool)

TELEGRAM_BOT_TOKEN: str = os.environ["TELEGRAM_BOT_TOKEN"]

BANGUMI_APP_ID: str = os.environ["BANGUMI_APP_ID"]
BANGUMI_APP_SECRET: str = os.environ["BANGUMI_APP_SECRET"]

HTTP_PORT = __convert(os.environ.get("HTTP_PORT"), t=Union[int, None])

QUEUE_SIZE = __convert(os.environ.get("QUEUE_SIZE") or 1, t=int)

REDIS_DSN = yarl.URL(os.environ["REDIS_DSN"])

PG_DSN = yarl.URL(os.environ["PG_DSN"])
MYSQL_DSN = yarl.URL(os.environ["MYSQL_DSN"])

KAFKA_BROKER = os.environ["KAFKA_BROKER"]

EXTERNAL_HTTP_ADDRESS = yarl.URL(
    os.environ.get("EXTERNAL_HTTP_ADDRESS", "http://127.0.0.1:4562")
)
