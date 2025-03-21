package main

import (
	"os"
)

var BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
var BANGUMI_APP_ID = os.Getenv("BANGUMI_APP_ID")
var BANGUMI_APP_SECRET = os.Getenv("BANGUMI_APP_SECRET")

var EXTERNAL_HTTP_ADDRESS = os.Getenv("EXTERNAL_HTTP_ADDRESS")
