package main

import (
	"context"

	"github.com/rs/zerolog/log"
)

type User struct {
	Username string `db:"username"`
	Nickname string `db:"nickname"`
	UserID   int64  `db:"uid"`
}

func (h *handler) getUserInfo(ctx context.Context, uid int64) (User, error) {
	var user User
	err := h.mysql.GetContext(ctx, &user,
		`SELECT uid, username, nickname FROM chii_members WHERE uid = ? LIMIT 1`,
		uid)
	if err != nil {
		log.Err(err).Int64("uid", uid).Msg("failed to query user info")
		return User{}, err
	}

	return user, nil

}

func (h *handler) getNotifyField(ctx context.Context, mid int64) (ChiiNotifyField, error) {
	var field ChiiNotifyField
	err := h.mysql.GetContext(ctx, &field, "SELECT ntf_id,ntf_hash,ntf_rid,ntf_title from chii_notify_field where ntf_id = ? limit 1",
		mid)
	return field, err
}
