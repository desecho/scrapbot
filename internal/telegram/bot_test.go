package telegram

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"scrapbot/internal/timeview"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestHandleMessageTimeCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(commandMessage("/time"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := runner.timeService.FormatCurrentTimes()
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageTimeCommandWithBotName(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(commandMessage("/time@TimeBot"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := runner.timeService.FormatCurrentTimes()
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageUnknownCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(commandMessage("/start"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	const want = "Supported command: /time"
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageIgnoresNonCommandText(t *testing.T) {
	runner := newTestRunner(t)

	if got, ok := runner.handleMessage(&tgbotapi.Message{Text: "hello"}); ok || got != "" {
		t.Fatalf("handleMessage() = (%q, %t), want (\"\", false)", got, ok)
	}
}

func TestHandleMessageIgnoresNilOrEmptyMessages(t *testing.T) {
	runner := newTestRunner(t)

	if got, ok := runner.handleMessage(nil); ok || got != "" {
		t.Fatalf("handleMessage(nil) = (%q, %t), want (\"\", false)", got, ok)
	}

	if got, ok := runner.handleMessage(&tgbotapi.Message{}); ok || got != "" {
		t.Fatalf("handleMessage(empty) = (%q, %t), want (\"\", false)", got, ok)
	}
}

func newTestRunner(t *testing.T) *Runner {
	t.Helper()

	service, err := timeview.NewService(fixedClock{
		now: time.Date(2026, time.March, 3, 15, 4, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return NewRunner(nil, service)
}

func commandMessage(text string) *tgbotapi.Message {
	return &tgbotapi.Message{
		Text: text,
		Entities: []tgbotapi.MessageEntity{
			{
				Type:   "bot_command",
				Offset: 0,
				Length: len(text),
			},
		},
	}
}
