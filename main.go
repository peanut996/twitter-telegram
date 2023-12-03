package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

var (
	apiHost     = "https://co.wuk.sh"
	apiJsonPath = "/api/json"

	telegramBotToken = "6764469070:AAHdzd207b651_6QKewWGF6z9MPONnAi7Vk"

	debug = false

	channelName = ""
)

func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	debugMode := os.Getenv("DEBUG")
	channel := os.Getenv("TELEGRAM_CHANNEL_NAME")
	if token != "" {
		telegramBotToken = token
	}

	if telegramBotToken == "" {
		log.Fatal("telegram bot token is empty")
	}
	if debugMode == "true" {
		debug = true
	}
	if channel != "" {
		channelName = channel
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go handleMessage(update, bot)
	}
}
