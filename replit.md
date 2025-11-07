# Telegram Barcode Checker Bot

## Overview
This is a Telegram bot written in Go that helps users check product ingredients by scanning barcodes. The bot analyzes product composition and identifies potentially dangerous or suspicious ingredients.

**Bot Username:** @insidecode_bot

## Features
- Barcode scanning from photos using image recognition
- Manual barcode input (8-13 digits)
- Product information lookup via Open Food Facts API
- Ingredient analysis for dangerous components:
  - Palm oil
  - GMO ingredients
  - Trans fats
  - Preservatives
  - Artificial colors
  - Flavor enhancers
- Health recommendations based on ingredient analysis

## Project Architecture

### Structure
```
cmd/bot/            - Main application entry point
internal/
  ├── bot/          - Telegram bot handlers
  ├── config/       - Configuration management
  ├── models/       - Data models (Product, AnalysisResult)
  ├── services/     - Business logic services
  │   ├── barcode.go        - Open Food Facts API integration
  │   ├── analyzer.go       - Ingredient analysis
  │   └── gozxing_detector.go - Barcode detection from images
  └── utils/        - Helper functions
```

### Technology Stack
- **Language:** Go 1.24
- **Bot Framework:** go-telegram-bot-api/telegram-bot-api/v5
- **Barcode Detection:** makiuchi-d/gozxing (local image processing)
- **Alternative:** Google Cloud Vision API (available but not used by default)
- **External API:** Open Food Facts API (product database)

## Configuration

### Required Environment Variables
- `TELEGRAM_BOT_TOKEN` - Telegram Bot API token (required)

### Optional Environment Variables
- `OPEN_FOOD_FACTS_API` - Open Food Facts API URL (default: https://world.openfoodfacts.org/api/v0)
- `REDIS_URL` - Redis connection URL (not currently used, default: localhost:6379)

## Running the Bot

The bot runs as a console application via the workflow:
```bash
go run ./cmd/bot/main.go
```

### Workflow Configuration
- **Name:** telegram-bot
- **Type:** Console application (backend service)
- **Command:** `go run ./cmd/bot/main.go`

## How Users Interact with the Bot

1. Start a conversation with @insidecode_bot on Telegram
2. Send `/start` to see welcome message
3. Either:
   - Send a photo of a barcode
   - Type the barcode digits manually (8-13 digits)
4. Receive analysis showing:
   - Product name and brand
   - Full composition
   - Dangerous ingredients (if any)
   - Suspicious ingredients (if any)
   - Health recommendations

## Dependencies

All Go dependencies are managed via `go.mod`:
- Telegram Bot API client
- Google Cloud Vision API (optional)
- GoZXing barcode detection library
- Standard Go libraries

## Recent Changes (November 7, 2025)
- ✅ Imported from GitHub repository
- ✅ Installed Go 1.24 toolchain
- ✅ Downloaded all Go module dependencies
- ✅ Configured TELEGRAM_BOT_TOKEN secret
- ✅ Set up telegram-bot workflow (console output)
- ✅ Successfully started bot (@insidecode_bot)

## Notes
- This is a backend bot service (no frontend/web UI)
- The bot uses local barcode detection (gozxing) by default, not requiring Google Cloud credentials
- Product data comes from the free Open Food Facts API
- The bot handles graceful shutdown via signal handling
