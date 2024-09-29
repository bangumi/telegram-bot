import typing
from datetime import datetime
from typing import TypeVar

import msgspec


class Source(msgspec.Struct):
    ts_ms: int

    def timestamp(self) -> datetime:
        return datetime.fromtimestamp(self.ts_ms / 1000).astimezone()


class ChiiNotify(msgspec.Struct):
    """table of chii_notify as json"""

    # nt_id: Any
    nt_uid: int
    nt_from_uid: int
    nt_status: int
    nt_type: int
    nt_mid: int  # ID of notify_field
    nt_related_id: int  # id of post
    timestamp: int = msgspec.field(name="nt_dateline")


class ChiiNotifyField(msgspec.Struct):
    """table of chii_notify_field as json"""

    ntf_id: int
    ntf_rid: int
    ntf_title: str
    ntf_hash: int


class ChiiPm(msgspec.Struct):
    """table of chii_pms"""

    msg_id: int
    msg_sid: int  # sender user id
    msg_rid: int  # receiver user id
    msg_new: bool
    msg_title: str
    msg_message: str
    timestamp: int = msgspec.field(name="msg_dateline")


class NotifyValue(msgspec.Struct):
    after: ChiiNotify | None
    op: str  # 'r', 'c', 'd' ...


class ChiiMember(msgspec.Struct):
    """table of chii_members as json"""

    uid: int
    newpm: int


T = TypeVar("T")


class DebeziumValue(msgspec.Struct, typing.Generic[T]):
    before: T | None
    after: T | None
    op: str  # 'r', 'c', 'd' ...
    source: Source


class MemberValue(DebeziumValue[ChiiMember]):
    pass
