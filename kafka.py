from typing import Iterator, NamedTuple
from uuid import uuid4

from confluent_kafka import Consumer
from sslog import logger

from lib import config


class Msg(NamedTuple):
    topic: str
    offset: int
    key: bytes | None
    value: bytes


class KafkaConsumer:
    def __init__(self, *topics: str):
        self.c = Consumer(
            {
                "group.id": "tg-notify-bot" + str(uuid4()),
                "bootstrap.servers": f"{config.KAFKA_BROKER.host}:{config.KAFKA_BROKER.port}",
                "auto.offset.reset": "earliest",
            }
        )
        self.c.subscribe(list(topics))

    def __iter__(self) -> Iterator[Msg]:
        while True:
            msg = self.c.poll(30)
            if msg is None:
                continue

            if msg.error():
                logger.error("consumer error", err=msg.error())
                continue

            msg_value = _ensure_binary(msg.value())
            if msg_value is None:
                continue

            yield Msg(
                topic=msg.topic() or "",
                offset=msg.offset() or 0,
                key=_ensure_binary(msg.key()),
                value=msg_value,
            )


def _ensure_binary(s: str | bytes | None) -> bytes | None:
    if isinstance(s, str):
        return s.encode()
    return s
