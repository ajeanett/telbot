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
	// Запускаем health-check сервер
	services.StartHealthServer()

	// Загрузка конфигурации
	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Инициализация провайдеров продуктов.
	// Порядок важен: Роскачество первым (лучшие данные по РФ), OFF — фолбэк.
	roskachestvo := services.NewRoskachestvoProvider(cfg.RskrfAPI)
	openFoodFacts := services.NewOpenFoodFactsProvider(cfg.OpenFoodFactsAPI)
	productProvider := services.NewMultiProductProvider(roskachestvo, openFoodFacts)

	// Инициализация детекторов штрих-кодов
	gozxingDetector := services.NewGozxingBarcodeDetector()

	// Мульти-детектор
	barcodeDetector := services.NewMultiBarcodeDetector(gozxingDetector)

	// Инициализация анализатора
	analyzer := services.NewAnalyzer()

	// Создание бота
	b, err := bot.NewBot(cfg.TelegramToken, productProvider, analyzer, barcodeDetector)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}
	defer b.Close()

	log.Printf("🤖 Бот авторизован: %s", b.Api().Self.UserName)

	// --- Graceful shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("🚀 Бот запущен. Ожидание сообщений...")

	go b.Start()

	<-stop
	log.Println("🛑 Получен сигнал остановки...")
	log.Println("👋 Завершение работы...")
}
