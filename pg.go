package main

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (h *handler) disableChat(ctx context.Context, chatID int64) error {
	_, dbErr := h.pg.ExecContext(ctx, "UPDATE telegram_notify_chat SET disabled = 1 WHERE chat_id = $1", chatID)
	if dbErr != nil {
		log.Err(dbErr).Int64("chat_id", chatID).Msg("failed to disable chat")
	}

	return dbErr
}

func (h *handler) getChats(ctx context.Context, userID int64) ([]int64, error) {
	var chatIDs []int64
	err := h.pg.SelectContext(ctx, &chatIDs, "SELECT chat_id FROM telegram_notify_chat WHERE user_id = $1 and disabled = 0", userID)
	if err != nil {
		return nil, err
	}

	return chatIDs, nil
}
