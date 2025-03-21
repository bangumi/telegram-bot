package main

type Config struct {
	BOT_TOKEN string `env:"TELEGRAM_BOT_TOKEN"`

	BANGUMI_APP_ID     string `env:"BANGUMI_APP_ID"`
	BANGUMI_APP_SECRET string `env:"BANGUMI_APP_SECRET"`

	EXTERNAL_HTTP_ADDRESS string `env:"EXTERNAL_HTTP_ADDRESS" default:"http://127.0.0.1:4562"`

	PORT uint16 `env:"PORT" default:"4096"`

	REDIS_DSN string `env:"REDIS_DSN"`

	PG_DSN    string `env:"PG_DSN"`
	MYSQL_DSN string `env:"MYSQL_DSN"`

	KAFKA_BROKER string `env:"KAFKA_BROKER"`
}
