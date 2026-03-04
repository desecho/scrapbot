package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"scrapbot/internal/timeview"
	"scrapbot/internal/weather"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestHandleMessageTimeCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(context.Background(), commandMessage("/time"))
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

	got, ok := runner.handleMessage(context.Background(), commandMessage("/time@TimeBot"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := runner.timeService.FormatCurrentTimes()
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageWeatherCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(context.Background(), commandMessage("/weather"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := strings.Join([]string{
		"Astrakhan: 6\u00b0 (3\u00b0) / Cloudy",
		"Montreal: -2\u00b0 (-7\u00b0) / Clear",
		"Seattle: 9\u00b0 (8\u00b0) / Light rain",
	}, "\n")

	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageWeatherCommandWithBotName(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(context.Background(), commandMessage("/weather@TimeBot"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := strings.Join([]string{
		"Astrakhan: 6\u00b0 (3\u00b0) / Cloudy",
		"Montreal: -2\u00b0 (-7\u00b0) / Clear",
		"Seattle: 9\u00b0 (8\u00b0) / Light rain",
	}, "\n")

	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageWeatherCommandReturnsErrorMessage(t *testing.T) {
	runner := newTestRunnerWithWeatherClient(t, stubWeatherClient{
		do: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	})

	got, ok := runner.handleMessage(context.Background(), commandMessage("/weather"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	const want = "Unable to fetch weather right now."
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageHelpCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(context.Background(), commandMessage("/help"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	want := strings.Join([]string{
		"/time — Show current time in Astrakhan, Montreal, and Seattle",
		"/weather — Show current weather in Astrakhan, Montreal, and Seattle",
		"/help — Show available commands",
	}, "\n")

	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageUnknownCommand(t *testing.T) {
	runner := newTestRunner(t)

	got, ok := runner.handleMessage(context.Background(), commandMessage("/start"))
	if !ok {
		t.Fatal("handleMessage() ok = false, want true")
	}

	const want = "Unknown command. Try /help to see available commands."
	if got != want {
		t.Fatalf("handleMessage() = %q, want %q", got, want)
	}
}

func TestHandleMessageIgnoresNonCommandText(t *testing.T) {
	runner := newTestRunner(t)

	if got, ok := runner.handleMessage(context.Background(), &tgbotapi.Message{Text: "hello"}); ok || got != "" {
		t.Fatalf("handleMessage() = (%q, %t), want (\"\", false)", got, ok)
	}
}

func TestHandleMessageIgnoresNilOrEmptyMessages(t *testing.T) {
	runner := newTestRunner(t)

	if got, ok := runner.handleMessage(context.Background(), nil); ok || got != "" {
		t.Fatalf("handleMessage(nil) = (%q, %t), want (\"\", false)", got, ok)
	}

	if got, ok := runner.handleMessage(context.Background(), &tgbotapi.Message{}); ok || got != "" {
		t.Fatalf("handleMessage(empty) = (%q, %t), want (\"\", false)", got, ok)
	}
}

func newTestRunner(t *testing.T) *Runner {
	t.Helper()

	return newTestRunnerWithWeatherClient(t, successfulWeatherClient(t))
}

func newTestRunnerWithWeatherClient(t *testing.T, client weather.HTTPClient) *Runner {
	t.Helper()

	timeService, err := timeview.NewService(fixedClock{
		now: time.Date(2026, time.March, 3, 15, 4, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	weatherService, err := weather.NewService(client, "test-key")
	if err != nil {
		t.Fatalf("weather.NewService() error = %v", err)
	}

	return NewRunner(nil, timeService, weatherService)
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

type stubWeatherClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (c stubWeatherClient) Do(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func successfulWeatherClient(t *testing.T) weather.HTTPClient {
	t.Helper()

	return stubWeatherClient{
		do: func(req *http.Request) (*http.Response, error) {
			switch req.URL.Query().Get("q") {
			case "Astrakhan":
				return weatherAPIResponse(http.StatusOK, weatherPayload("Astrakhan", 5.5, 3.2, "Cloudy")), nil
			case "Montreal":
				return weatherAPIResponse(http.StatusOK, weatherPayload("Montreal", -2.0, -7.1, "Clear")), nil
			case "Seattle":
				return weatherAPIResponse(http.StatusOK, weatherPayload("Seattle", 9.3, 7.8, "Light rain")), nil
			default:
				return nil, fmt.Errorf("unexpected city %q", req.URL.Query().Get("q"))
			}
		},
	}
}

func weatherAPIResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func weatherPayload(name string, tempC, feelsLikeC float64, condition string) string {
	return fmt.Sprintf(
		`{"location":{"name":%q},"current":{"temp_c":%.1f,"feelslike_c":%.1f,"condition":{"text":%q}}}`,
		name,
		tempC,
		feelsLikeC,
		condition,
	)
}
