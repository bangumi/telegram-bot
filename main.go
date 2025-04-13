package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"

	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
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

	pg := sqlx.MustConnect("postgres", cfg.PgDsn)
	mysql := sqlx.MustConnect("mysql", cfg.MysqlDsn)

	redisDSN := lo.Must(url.Parse(cfg.RedisDsn))
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
	bot := lo.Must(telego.NewBot(cfg.BotToken,
		telego.WithDefaultLogger(cfg.Debug, true),
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
		redirectURL: strings.TrimRight(cfg.ExternalHttpAddress, "/") + "/callback",
	}

	var eg errgroup.Group

	eg.Go(func() error {
		return h.processTelegramMessage()
	})

	eg.Go(func() error {
		return h.ListenAndServe()
	})

	eg.Go(func() error {
		return h.processKafkaMessage()
	})

	err := eg.Wait()
	if err != nil {
		panic(err)
	}
}

func (h *handler) disableChat(ctx context.Context, chatID int64) error {
	_, dbErr := h.pg.ExecContext(ctx, "UPDATE telegram_notify_chat SET disabled = 1 WHERE chat_id = $1", chatID)
	if dbErr != nil {
		log.Err(dbErr).Int64("chat_id", chatID).Msg("failed to disable chat")
	}

	return dbErr
}
