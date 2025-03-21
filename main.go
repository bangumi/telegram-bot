package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Get Bot token from environment variables
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	pg := sqlx.MustConnect("postgres", os.Getenv("PG_DSN"))
	mysql := sqlx.MustConnect("mysql", os.Getenv("MYSQL_DSN"))
	redis := lo.Must(rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{os.Getenv("REDIS_DSN")}}))

	// Create bot and enable debugging info
	// Note: Please keep in mind that default logger may expose sensitive information,
	// use in development only
	// (more on configuration in examples/configuration/main.go)
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	oauth := NewOAuthHTTPServer(pg, redis, bot, 4096)

	// Call method getMe (https://core.telegram.org/bots/api#getme)
	botUser, err := bot.GetMe(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
	}

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
		return oauth.Start()
	})

	eg.Go(func() error {
		k := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{os.Getenv("KAFKA_BROKER")},
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

			fmt.Println(len(msg.Value))

			switch {
			case strings.HasSuffix(msg.Topic, ".chii_pms"):
			case strings.HasSuffix(msg.Topic, ".chii_notify"):
			}
		}
	})

	eg.Wait()
}
