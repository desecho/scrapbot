# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Scrapbot is a Telegram bot written in Go that responds to `/time`, `/weather`, and `/help` commands, showing data for three hardcoded cities: Astrakhan, Montreal, and Seattle.

## Commands

```bash
go run ./cmd/bot          # Run the bot (requires TELEGRAM_BOT_TOKEN and WEATHER_API_KEY env vars)
go test ./...             # Run all tests
go test ./internal/weather  # Run tests for a single package
```

## Architecture

- **`cmd/bot/main.go`** — Entrypoint. Reads env vars, initializes services, starts the Telegram polling loop.
- **`internal/telegram/`** — `Runner` handles the Telegram update loop and command dispatch. Commands are routed via a switch in `handleMessage()`.
- **`internal/timeview/`** — `Service` formats current times across cities. Uses a `Clock` interface for testability (inject `fixedClock` in tests).
- **`internal/weather/`** — `Service` fetches weather from WeatherAPI (`api.weatherapi.com`). Uses an `HTTPClient` interface for testability (inject `stubHTTPClient` in tests).

## Key Patterns

- Services accept interfaces (`Clock`, `HTTPClient`) in constructors for dependency injection and test stubbing.
- The weather service uses pointer fields in API response structs to detect missing fields (returns "incomplete response" error if any are nil).
- Temperatures are rounded with `math.Round` before formatting with `%.0f`.
- The bot uses long polling (not webhooks). Graceful shutdown via `signal.NotifyContext`.

## Deployment

Docker multi-stage build, deployed to Kubernetes on DigitalOcean via GitHub Actions on push to `main`. Image pushed to `ghcr.io`. Secrets (`TELEGRAM_BOT_TOKEN`, `WEATHER_API_KEY`) stored as a Kubernetes secret.
