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
	NtUid       int `json:"nt_uid"`
	NtFromUid   int `json:"nt_from_uid"`
	NtStatus    int `json:"nt_status"`
	NtType      int `json:"nt_type"`
	NtMid       int `json:"nt_mid"`        // ID of notify_field
	NtRelatedId int `json:"nt_related_id"` // id of post
	Timestamp   int `json:"nt_dateline"`
}

type ChiiNotifyField struct {
	NtfId    int    `json:"ntf_id"`
	NtfRid   int    `json:"ntf_rid"`
	NtfTitle string `json:"ntf_title"`
	NtfHash  int    `json:"ntf_hash"`
}

type ChiiPm struct {
	MsgId      int    `json:"msg_id"`
	MsgSid     int    `json:"msg_sid"` // sender user id
	MsgRid     int    `json:"msg_rid"` // receiver user id
	MsgNew     int    `json:"msg_new"`
	MsgTitle   string `json:"msg_title"`
	MsgMessage string `json:"msg_message"`
	Timestamp  int    `json:"msg_dateline"`
}

type NotifyValue struct {
	After *ChiiNotify `json:"after"`
	Op    string      `json:"op"` // 'r', 'c', 'd' ...
}

type DebeziumValue struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Op     string          `json:"op"` // 'r', 'c', 'd' ...
	Source Source          `json:"source"`
}
