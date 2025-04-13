package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"go-simpler.org/env"
	"golang.org/x/sync/errgroup"
)

type handler struct {
	config      Config
	mysql       *sqlx.DB
	pg          *sqlx.DB
	botUser     *telego.User
	bot         *telego.Bot
	redis       rueidis.Client
	client      *resty.Client
	redirectURL string
}

func main() {
	var cfg Config
	lo.Must0(env.Load(&cfg, nil))

	// Get Bot token from environment variables

	pg := sqlx.MustConnect("postgres", cfg.PG_DSN)
	mysql := sqlx.MustConnect("mysql", cfg.MYSQL_DSN)

	redisDSN := lo.Must(url.Parse(cfg.REDIS_DSN))
	redisPassword, _ := redisDSN.User.Password()
	redis := lo.Must(rueidis.NewClient(rueidis.ClientOption{
		Password:    redisPassword,
		InitAddress: []string{redisDSN.Host}},
	))

	pg.MustExec(`CREATE TABLE IF NOT EXISTS telegram_notify_chat (
								chat_id bigint,
								user_id bigint,
								disabled int2,
								primary key (chat_id, user_id)
			);`)

	// Create bot and enable debugging info
	// Note: Please keep in mind that default logger may expose sensitive information,
	// use in development only
	// (more on configuration in examples/configuration/main.go)
	bot := lo.Must(telego.NewBot(cfg.BOT_TOKEN,
		telego.WithDefaultDebugLogger(),
		telego.WithHTTPClient(http.DefaultClient)),
	)

	currentBot := lo.Must(bot.GetMe(context.Background()))

	h := &handler{
		config:      cfg,
		botUser:     currentBot,
		pg:          pg,
		mysql:       mysql,
		bot:         bot,
		client:      resty.New(),
		redis:       redis,
		redirectURL: strings.TrimRight(cfg.EXTERNAL_HTTP_ADDRESS, "/") + "/callback",
	}

	var eg errgroup.Group

	eg.Go(func() error {
		updates, err := bot.UpdatesViaLongPolling(context.Background(), nil)
		if err != nil {
			return err
		}

		bh := lo.Must(th.NewBotHandler(bot, updates))
		// Stop handling updates
		defer func() { _ = bh.Stop() }()

		bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
			state := uuid.Must(uuid.NewV4()).String()

			ctx, cancel := ctx.WithTimeout(time.Second * 5)
			defer cancel()

			err := redis.Do(ctx, redis.B().Set().Key("tg-bot-oauth:"+state).Value(rueidis.JSON(RedisOAuthState{
				ChatID: message.Chat.ID,
			})).ExSeconds(60).Build()).Error()
			if err != nil {
				return err
			}

			_, _ = bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID), "请在 60s 内进行认证").
				WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(tu.InlineKeyboardButton("认证 bangumi 账号").
						WithURL(fmt.Sprintf("%s/redirect?state=%s", cfg.EXTERNAL_HTTP_ADDRESS, state)))),
				))
			return nil
		}, th.CommandEqual("start"))

		return bh.Start()
	})

	eg.Go(func() error {
		return h.ListenAndServe()
	})

	eg.Go(func() error {
		k := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.KAFKA_BROKER},
			GroupID: "tg-notify-bot",
			GroupTopics: []string{
				"debezium.chii.bangumi.chii_pms",
				"debezium.chii.bangumi.chii_notify",
			},
		})

		for {
			msg, err := k.ReadMessage(context.Background())
			if err != nil {
				log.Err(err).Msg("failed to read kafka message")
				continue
			}
			if len(msg.Value) == 0 {
				continue
			}

			switch msg.Topic {
			case "debezium.chii.bangumi.chii_pms":
				h.handlePM(msg)
			case "debezium.chii.bangumi.chii_notify":
				h.handleNotify(msg)
			}
		}
	})

	err := eg.Wait()
	if err != nil {
		panic(err)
	}
}

func (h *handler) handlePM(msg kafka.Message) {

}

func (h *handler) handleNotify(msg kafka.Message) {

}

func (h *handler) sendNotification(ctx context.Context, chatID int64, text string, parseMode string) error {
	// Create message sending parameters
	params := &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	}

	// Add parse mode if it's not empty
	if parseMode != "" {
		params.ParseMode = parseMode
	}

	// Send message
	_, err := h.bot.SendMessage(ctx, params)
	if err == nil {
		return nil
	}

	// Check if the error is because the user is deactivated
	if strings.Contains(err.Error(), "Forbidden: user is deactivated") {
		return h.disableChat(ctx, chatID)
	}

	return err
}

func (h *handler) disableChat(ctx context.Context, chatID int64) error {
	_, dbErr := h.pg.ExecContext(ctx, "UPDATE telegram_notify_chat SET disabled = 1 WHERE chat_id = $1", chatID)
	if dbErr != nil {
		log.Err(dbErr).Int64("chat_id", chatID).Msg("failed to disable chat")
	}

	return dbErr
}
