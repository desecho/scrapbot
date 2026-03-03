# scrapbot

A small Telegram bot in Go that replies to `/time` with the current local time and `/weather` with the current weather for Astrakhan, Montreal, and Seattle.

## Requirements

- Go 1.23+
- A Telegram bot token exposed as `TELEGRAM_BOT_TOKEN`
- A WeatherAPI key exposed as `WEATHER_API_KEY`

## Run

```bash
export TELEGRAM_BOT_TOKEN=your-token-here
export WEATHER_API_KEY=your-weatherapi-key-here
go run ./cmd/bot
```

The bot uses Telegram long polling. Send `/time` to the bot and it will reply with one line per city in this format:

```text
Astrakhan: 2026-03-03 19:04
Montreal: 2026-03-03 10:04
Seattle: 2026-03-03 07:04
```

Send `/weather` to the bot and it will reply with one line per city in this format:

```text
Astrakhan: 5.5° (3.2°) / Cloudy
Montreal: -2.0° (-7.1°) / Clear
Seattle: 9.3° (7.8°) / Light rain
```

Unknown commands return:

```text
Supported commands: /time, /weather
```

## Test

```bash
go test ./...
```
