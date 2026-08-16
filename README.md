# Telegram Barcode Checker Bot

## Overview

This is a Telegram bot written in Go that helps users check product ingredients by scanning barcodes. The bot looks the product up across multiple open data sources, shows its composition and country of origin, and analyzes the ingredients for potentially dangerous or suspicious components, food additives, and allergens.

**Bot Username:** [@insidecode_bot](https://t.me/insidecode_bot)

## Features

- Barcode scanning from photos using local image recognition (gozxing, no external vision services)
- Manual barcode input (8–13 digits, EAN-13 checksum validation)
- Product lookup across two data sources with automatic fallback:
  1. **Roskachestvo** (rskrf.ru official API) — best coverage of Russian products: composition, country of origin, quality rating (1–5), Quality Mark
  2. **Open Food Facts** — international products: composition (RU/EN), structured ingredients, E-additives, allergen tags
- Ingredient analysis:
  - Dangerous ingredients (aspartame E951, MSG E621, sodium nitrite E250, …)
  - Suspicious ingredients (palm oil, GMO, trans fats, preservatives, artificial colors, flavor enhancers)
  - Food additive descriptions (E-codes)
  - **Allergen detection** against the standard EU-14 list (gluten, milk, eggs, fish, peanuts, soy, nuts, crustaceans, molluscs, celery, mustard, sesame, sulphites, lupin) — by RU/EN keywords in the composition and by Open Food Facts allergen tags
- Health recommendations based on the analysis
- Health-check HTTP server for hosting platforms

## Project Architecture

### Structure

```
cmd/bot/            - Main application entry point (composition root)
internal/
  ├── bot/          - Telegram bot handlers and message rendering
  ├── config/       - Configuration from environment variables
  ├── models/       - Domain models (Product, AnalysisResult)
  ├── services/     - Business logic services
  │   ├── roskachestvo.go     - Roskachestvo API provider (primary)
  │   ├── barcode.go          - Open Food Facts provider (fallback)
  │   ├── multi_provider.go   - Provider chain with fallback
  │   ├── analyzer.go         - Ingredient/allergen analysis
  │   ├── analyzer_data.go    - Ingredient, additive and allergen dictionaries
  │   ├── gozxing_detector.go - Barcode detection from images
  │   └── vision.go           - Google Cloud Vision (available, not wired up)
  └── utils/        - Helper functions (barcode validation)
```

### Technology Stack

- **Language:** Go 1.24
- **Bot Framework:** go-telegram-bot-api/telegram-bot-api/v5
- **Barcode Detection:** makiuchi-d/gozxing (local image processing)
- **Alternative detector:** Google Cloud Vision API (in the codebase, requires GCP credentials; not wired up)
- **Data sources:** Roskachestvo public API (rskrf.ru), Open Food Facts API

Providers and detectors are interface-based with fallback chains — to add a new data source, implement the `ProductProvider` interface and register it in `cmd/bot/main.go`. The full architecture lives in [docs/spec/architecture.md](docs/spec/architecture.md).

## Configuration

### Required Environment Variables

- `TELEGRAM_BOT_TOKEN` — Telegram Bot API token

### Optional Environment Variables

- `OPEN_FOOD_FACTS_API` — Open Food Facts API URL (default: `https://world.openfoodfacts.org/api/v0`)
- `RSKRF_API` — Roskachestvo API URL (default: `https://rskrf.ru/rest/1`)
- `PORT` — health-check server port (default: `8080`)

## Running the Bot

```bash
go run ./cmd/bot/main.go
```

Build and test:

```bash
go build ./...
go vet ./...
go test ./...        # 18 unit tests on real API response fixtures
```

## How Users Interact with the Bot

1. Start a conversation with [@insidecode_bot](https://t.me/insidecode_bot) on Telegram
2. Send `/start` to see the welcome message
3. Either:
   - Send a photo of a barcode
   - Type the barcode digits manually (8–13 digits)
4. Receive a card showing:
   - Product name, brand, country of origin, barcode
   - Roskachestvo rating and Quality Mark (when found there)
   - Full composition
   - Possible allergens
   - Dangerous and suspicious ingredients (if any)
   - Health recommendations and a link to the source

## Documentation

- [docs/spec/](docs/spec/) — source of truth: product behaviour and architecture (updated with every change)
- [docs/changes/](docs/changes/) — journal of changes: design specs and implementation plans
- [CLAUDE.md](CLAUDE.md) — conventions for AI-assisted development

## Notes

- This is a backend bot service (no frontend/web UI)
- Barcode detection runs locally (gozxing); Google Cloud credentials are not required
- Data comes from free public APIs; per Roskachestvo's API terms, every reply includes a link to the source portal
- The bot handles graceful shutdown via signal handling
