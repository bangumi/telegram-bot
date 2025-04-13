package main

import (
	"context"

	"github.com/rs/zerolog/log"
)

type User struct {
	Username string
	Nickname string
	UserID   int64
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
