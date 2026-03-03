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
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("create telegram bot: %v", err)
	}

	timeService, err := timeview.NewService(timeview.RealClock{})
	if err != nil {
		log.Fatalf("initialize time service: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := telegram.NewRunner(api, timeService)
	if err := runner.Run(ctx); err != nil {
		log.Fatalf("run bot: %v", err)
	}
}
