package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func (h *handler) processTelegramMessage() error {
	updates, ue := h.bot.UpdatesViaLongPolling(context.Background(), nil)
	if ue != nil {
		return ue
	}

	bh := lo.Must(th.NewBotHandler(h.bot, updates))
	defer func() { _ = bh.Stop() }()

	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		state := uuid.Must(uuid.NewV4()).String()

		ctx, cancel := ctx.WithTimeout(time.Second * 5)
		defer cancel()

		err := h.redis.Do(ctx, h.redis.B().Set().Key(redisStateKey(state)).Value(rueidis.JSON(RedisOAuthState{
			ChatID: message.Chat.ID,
		})).ExSeconds(60).Build()).Error()
		if err != nil {
			return err
		}

		_, _ = h.bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID), "请在 60s 内进行认证").
			WithReplyMarkup(tu.InlineKeyboard(
				tu.InlineKeyboardRow(tu.InlineKeyboardButton("认证 bangumi 账号").
					WithURL(fmt.Sprintf("%s/redirect?state=%s", h.config.ExternalHttpAddress, state)))),
			))
		return nil
	}, th.CommandEqual("start"))

	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		ctx, cancel := ctx.WithTimeout(time.Second * 5)
		defer cancel()

		err := h.disableChat(ctx, message.Chat.ID)
		if err != nil {
			log.Err(err).Msg("failed to disable chat")
			return nil
		}

		_, _ = h.bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID), "已禁用"))

		return nil
	}, th.CommandEqual("logout"))

	return bh.Start()
}
