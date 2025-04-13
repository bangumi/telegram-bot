package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"go-simpler.org/env"
	"golang.org/x/sync/errgroup"
)

type handler struct {
	config  Config
	mysql   *sqlx.DB
	pg      *sqlx.DB
	botUser *telego.User
	bot     *telego.Bot
	redis   rueidis.Client
	client  *resty.Client
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.MessageFieldName = "msg"
}

func main() {
	log.Info().Msg("app start")

	var cfg Config
	log.Info().Msg("load config")
	lo.Must0(env.Load(&cfg, nil))
	cfg.ExternalHttpAddress = strings.TrimSuffix(cfg.ExternalHttpAddress, "/")

	log.Info().Msg("connect to pg")
	pg := sqlx.MustConnect("postgres", cfg.PgDsn)

	log.Info().Msg("connect to mysql")
	mysql := sqlx.MustConnect("mysql", cfg.MysqlDsn)

	log.Info().Msg("connect to redis")
	redisDSN := lo.Must(url.Parse(cfg.RedisDsn))
	redisPassword, _ := redisDSN.User.Password()
	redis := lo.Must(rueidis.NewClient(rueidis.ClientOption{
		Password:    redisPassword,
		InitAddress: []string{redisDSN.Host}},
	))

	log.Info().Msg("prepare pg table")
	pg.MustExec(`CREATE TABLE IF NOT EXISTS telegram_notify_chat (
								chat_id bigint,
								user_id bigint,
								disabled int2,
								primary key (chat_id, user_id)
			);`)

	log.Info().Msg("create telegram bot")
	bot := lo.Must(telego.NewBot(cfg.BotToken,
		telego.WithDefaultLogger(cfg.Debug, true),
		telego.WithHTTPClient(http.DefaultClient)),
	)

	log.Info().Msg("get bot info")
	currentBot := lo.Must(bot.GetMe(context.Background()))

	h := &handler{
		config:  cfg,
		botUser: currentBot,
		pg:      pg,
		mysql:   mysql,
		bot:     bot,
		client:  resty.New(),
		redis:   redis,
	}

	log.Info().Msg("start main loop")
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
