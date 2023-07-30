import msgspec


class ChiiMember(msgspec.Struct):
    """table of chii_members as json"""

    uid: int
    new_notify: int


class ChiiNotify(msgspec.Struct):
    """table of chii_notify as json"""

    # nt_id: Any
    nt_uid: int
    nt_from_uid: int
    nt_status: int
    nt_type: int
    nt_mid: int
    nt_related_id: int
    timestamp: int = msgspec.field(name="nt_dateline")


class NotifyValuePayload(msgspec.Struct):
    after: ChiiNotify | None
    op: str  # 'r', 'c', 'd' ...


class NotifyValue(msgspec.Struct):
    payload: NotifyValuePayload


class MemberValuePayload(msgspec.Struct):
    after: ChiiMember | None
    op: str  # 'r', 'c', 'd' ...


class MemberValue(msgspec.Struct):
    payload: MemberValuePayload
