package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"go-simpler.org/env"
	"golang.org/x/sync/errgroup"
)

type handler struct {
	config      Config
	pg          *sqlx.DB
	mysql       *sqlx.DB
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
	redis := lo.Must(rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{cfg.REDIS_DSN}}))

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
	bot := lo.Must(telego.NewBot(cfg.BOT_TOKEN, telego.WithDefaultDebugLogger()))

	h := &handler{
		config:      cfg,
		pg:          pg,
		mysql:       mysql,
		bot:         bot,
		redis:       redis,
		redirectURL: strings.TrimRight(cfg.EXTERNAL_HTTP_ADDRESS, "/") + "/callback",
	}

	// Call method getMe (https://core.telegram.org/bots/api#getme)
	botUser := lo.Must(bot.GetMe(context.Background()))

	var eg errgroup.Group

	eg.Go(func() error {
		// Get updates channel
		// (more on configuration in examples/updates_long_polling/main.go)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		updates, err := bot.UpdatesViaLongPolling(ctx, nil)
		if err != nil {
			return err
		}

		// Loop through all updates when they came
		for update := range updates {
			fmt.Printf("Update: %+v\n", update)
		}

		// Print Bot information
		fmt.Printf("Bot user: %+v\n", botUser)
		return nil
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
