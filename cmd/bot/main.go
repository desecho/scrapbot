package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"scrapbot/internal/telegram"
	"scrapbot/internal/timeview"
	"scrapbot/internal/weather"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	weatherAPIKey := os.Getenv("WEATHER_API_KEY")
	if weatherAPIKey == "" {
		log.Fatal("WEATHER_API_KEY is required")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("create telegram bot: %v", err)
	}

	timeService, err := timeview.NewService(timeview.RealClock{})
	if err != nil {
		log.Fatalf("initialize time service: %v", err)
	}

	weatherService, err := weather.NewService(nil, weatherAPIKey)
	if err != nil {
		log.Fatalf("initialize weather service: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := telegram.NewRunner(api, timeService, weatherService)
	if err := runner.Run(ctx); err != nil {
		log.Fatalf("run bot: %v", err)
	}
}
