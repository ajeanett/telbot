// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajeanett/telbot/internal/bot"
	"github.com/ajeanett/telbot/internal/config"
	"github.com/ajeanett/telbot/internal/services"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Инициализация сервисов
	barcodeService := services.NewBarcodeService(cfg.OpenFoodFactsAPI)
	analyzer := services.NewAnalyzer()
	barcodeDetector := services.NewBarcodeDetector()

	// Создание бота
	bot, err := bot.NewBot(cfg.TelegramToken, barcodeService, analyzer, barcodeDetector)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}
	defer bot.Close() // Закрываем ресурсы при завершении
	log.Printf("Бот авторизован: %s", bot.Api().Self.UserName)

	// Канал для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем бота в горутине
	go func() {
		log.Printf("Бот запущен: %s", bot.Api().Self.UserName)

		// Запуск бота
		bot.Start()
	}()
	<-stop
	log.Println("🛑 Получен сигнал остановки...")
	log.Println("👋 Завершаем работу бота")
}
