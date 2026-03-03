package telegram

import (
	"context"
	"errors"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"scrapbot/internal/timeview"
)

type Runner struct {
	api         *tgbotapi.BotAPI
	timeService *timeview.Service
}

func NewRunner(api *tgbotapi.BotAPI, timeService *timeview.Service) *Runner {
	return &Runner{
		api:         api,
		timeService: timeService,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if r == nil {
		return errors.New("runner is nil")
	}
	if r.api == nil {
		return errors.New("telegram api is nil")
	}
	if r.timeService == nil {
		return errors.New("time service is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := r.api.GetUpdatesChan(updateConfig)
	defer r.api.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}

			reply, shouldReply := r.handleMessage(update.Message)
			if !shouldReply {
				continue
			}

			message := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			message.ReplyToMessageID = update.Message.MessageID

			if _, err := r.api.Send(message); err != nil {
				log.Printf("send reply: %v", err)
			}
		}
	}
}

func (r *Runner) handleMessage(msg *tgbotapi.Message) (string, bool) {
	if msg == nil || msg.Text == "" || !msg.IsCommand() {
		return "", false
	}

	switch commandName(msg) {
	case "time":
		if r == nil || r.timeService == nil {
			return "", false
		}
		return r.timeService.FormatCurrentTimes(), true
	case "":
		return "", false
	default:
		return "Supported command: /time", true
	}
}

func commandName(msg *tgbotapi.Message) string {
	if msg == nil || !msg.IsCommand() {
		return ""
	}

	command := strings.Fields(msg.Text)
	if len(command) == 0 {
		return ""
	}

	name := strings.TrimPrefix(command[0], "/")
	if atIndex := strings.IndexByte(name, '@'); atIndex >= 0 {
		name = name[:atIndex]
	}

	return strings.ToLower(name)
}
