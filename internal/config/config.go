package config

import (
	"os"
)

type Config struct {
	TelegramToken    string
	OpenFoodFactsAPI string
	RskrfAPI         string
}

func Load() *Config {
	return &Config{
		TelegramToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		OpenFoodFactsAPI: getEnv("OPEN_FOOD_FACTS_API", "https://world.openfoodfacts.org/api/v0"),
		RskrfAPI:         getEnv("RSKRF_API", "https://rskrf.ru/rest/1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
