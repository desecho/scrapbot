# scrapbot

A small Telegram bot in Go that replies to `/time` with the current local time for Astrakhan, Montreal, and Seattle.

## Requirements

- Go 1.23+
- A Telegram bot token exposed as `TELEGRAM_BOT_TOKEN`

## Run

```bash
export TELEGRAM_BOT_TOKEN=your-token-here
go run ./cmd/bot
```

The bot uses Telegram long polling. Send `/time` to the bot and it will reply with one line per city in this format:

```text
Astrakhan: 2026-03-03 19:04
Montreal: 2026-03-03 10:04
Seattle: 2026-03-03 07:04
```

Unknown commands return:

```text
Supported command: /time
```

## Test

```bash
go test ./...
```
