package config

import (
	"os"
)

type Config struct {
	TelegramToken    string
	RedisURL         string
	OpenFoodFactsAPI string
}

func Load() *Config {
	return &Config{
		TelegramToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		OpenFoodFactsAPI: getEnv("OPEN_FOOD_FACTS_API", "https://world.openfoodfacts.org/api/v0"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
