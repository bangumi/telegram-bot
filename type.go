package main

import (
	"encoding/json"
	"time"
)

type Source struct {
	TsMs int64 `json:"ts_ms"`
}

func (s *Source) Timestamp() time.Time {
	return time.Unix(s.TsMs/1000, 0)
}

type ChiiNotify struct {
	Uid       int64 `json:"nt_uid"`
	FromUid   int64 `json:"nt_from_uid"`
	Status    int   `json:"nt_status"`
	Type      int   `json:"nt_type"`
	Mid       int64 `json:"nt_mid"`        // ID of notify_field
	RelatedId int64 `json:"nt_related_id"` // id of post
	Timestamp int   `json:"nt_dateline"`
}

type ChiiNotifyField struct {
	NtfId    int    `db:"ntf_id"`
	NtfRid   int    `db:"ntf_rid"`
	NtfTitle string `db:"ntf_title"`
	NtfHash  int    `db:"ntf_hash"`
}

type ChiiPm struct {
	MsgId      int64  `json:"msg_id"`
	MsgSid     int64  `json:"msg_sid"` // sender user id
	MsgRid     int64  `json:"msg_rid"` // receiver user id
	MsgNew     int    `json:"msg_new"`
	MsgTitle   string `json:"msg_title"`
	MsgMessage string `json:"msg_message"`
	Timestamp  int    `json:"msg_dateline"`
}

type DebeziumValue struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Op     string          `json:"op"` // 'r', 'c', 'd' ...
	Source Source          `json:"source"`
}

const OpCreate = "c"
