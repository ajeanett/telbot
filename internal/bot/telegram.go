package bot

import (
	// "context"
	"fmt"
	"github.com/ajeanett/telbot/internal/models"
	"github.com/ajeanett/telbot/internal/services"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	barcodeService  *services.BarcodeService
	analyzer        *services.Analyzer
	barcodeDetector *services.BarcodeDetector
	httpClient      *http.Client
}

func NewBot(
	token string,
	barcodeService *services.BarcodeService,
	analyzer *services.Analyzer,
	barcodeDetector *services.BarcodeDetector,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // Таймаут на запросы
		Transport: &http.Transport{
			MaxIdleConns:       10,               // Максимум idle соединений
			IdleConnTimeout:    30 * time.Second, // Таймаут idle соединений
			DisableCompression: false,            // Включить gzip сжатие
		},
	}

	return &Bot{
		api:             api,
		barcodeService:  barcodeService,
		analyzer:        analyzer,
		barcodeDetector: barcodeDetector,
		httpClient:      httpClient,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	if message.Photo != nil {
		// Обработка фото со штрих-кодом
		b.handleBarcodePhoto(message)
		return
	}

	text := strings.TrimSpace(message.Text)

	switch {
	case text == "/start":
		b.sendWelcomeMessage(message.Chat.ID)
	case len(text) >= 8 && len(text) <= 13 && isNumeric(text):
		// Предполагаем что это штрих-код
		b.handleBarcodeText(message.Chat.ID, text)
	default:
		b.sendHelpMessage(message.Chat.ID)
	}
}

func (b *Bot) handleBarcodeText(chatID int64, barcode string) {
	msg := tgbotapi.NewMessage(chatID, "🔍 Ищу информацию о продукте...")
	b.api.Send(msg)

	product, err := b.barcodeService.GetProductByBarcode(barcode)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(chatID,
			"❌ Не удалось найти продукт с таким штрих-кодом")
		b.api.Send(errorMsg)
		return
	}

	result := b.analyzer.AnalyzeProduct(product)
	b.sendAnalysisResult(chatID, result)
}

func (b *Bot) sendAnalysisResult(chatID int64, result *models.AnalysisResult) {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("🏷️ *%s*\n", result.Product.Name))
	message.WriteString(fmt.Sprintf("👨‍💼 *Бренд:* %s\n", result.Product.Brand))
	message.WriteString(fmt.Sprintf("📊 *Штрих-код:* %s\n\n", result.Product.Barcode))

	message.WriteString("*Состав:*\n")
	if result.Product.Composition != "" {
		message.WriteString(result.Product.Composition + "\n\n")
	} else {
		message.WriteString("Не указан\n\n")
	}

	if len(result.Dangerous) > 0 {
		message.WriteString("🚫 *ОПАСНЫЕ ИНГРЕДИЕНТЫ:*\n")
		for _, ingredient := range result.Dangerous {
			message.WriteString(fmt.Sprintf("• %s\n", ingredient))
		}
		message.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		message.WriteString("⚠️ *СОМНИТЕЛЬНЫЕ ИНГРЕДИЕНТЫ:*\n")
		for _, ingredient := range result.Warnings {
			message.WriteString(fmt.Sprintf("• %s\n", ingredient))
		}
		message.WriteString("\n")
	}

	message.WriteString("*Рекомендации:*\n")
	for _, rec := range result.Recommendations {
		message.WriteString(fmt.Sprintf("%s\n", rec))
	}

	msg := tgbotapi.NewMessage(chatID, message.String())
	msg.ParseMode = "Markdown"

	// TODO: отправлять фото если есть
	// Если есть изображение продукта
	// if result.Product.ImageURL != "" {
	// 	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(result.Product.ImageURL))
	// 	photo.Caption = message.String()
	// 	photo.ParseMode = "Markdown"
	// 	b.api.Send(photo)
	// } else {
	b.api.Send(msg)
	// }
}

func (b *Bot) sendWelcomeMessage(chatID int64) {
	text := `👋 *Добро пожаловать в FoodCheckerBot!*

Я помогу вам проверить состав продуктов по штрих-коду.

📱 *Как использовать:*
1. Отправьте мне фото штрих-кода
2. Или введите цифры штрих-кода вручную

Я проанализирую состав и выделю потенциально опасные ингредиенты.

🚫 *Проверяю:*
• Пальмовое масло
• ГМО
• Трансжиры  
• Консерванты
• Искусственные красители
• Усилители вкуса

_Данные предоставляются из открытой базы Open Food Facts_`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func (b *Bot) sendHelpMessage(chatID int64) {
	text := `📋 *Помощь*

Просто отправьте мне:
• 📷 Фото штрих-кода
• 🔢 Цифры штрих-кода (8-13 цифр)

Я найду информацию о продукте и проанализирую его состав на наличие опасных ингредиентов.`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func (b *Bot) handleBarcodePhoto(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Отправляем сообщение о начале обработки
	msg := tgbotapi.NewMessage(chatID, "📷 Обрабатываю изображение...")
	b.api.Send(msg)

	// Скачиваем изображение
	// Берем последний элемент, тк это самое качественное изображение
	imageData, err := b.downloadImage(message.Photo[len(message.Photo)-1].FileID)
	if err != nil {
		log.Printf("Ошибка загрузки изображения: %v", err)
		b.sendError(chatID, "Не удалось загрузить изображение. Попробуйте еще раз.")
		return
	}

	// Проверяем что VisionService доступен
	if b.barcodeDetector == nil {
		log.Println("BarcodeDetector не инициализирован")
		b.sendBarcodeDetectorError(chatID)
		return
	}

	// Распознаем штрих-код через BarcodeDetector
	barcode, err := b.barcodeDetector.DetectFromImage(imageData)
	if err != nil {
		log.Printf("Ошибка распознавания штрих-кода: %v", err)
		b.sendBarcodeNotFound(chatID)
		return
	}

	log.Printf("✅ Распознан штрих-код: %s", barcode)

	// Обрабатываем найденный штрих-код
	b.handleBarcodeText(chatID, barcode)
}

// downloadImage скачивает изображение по fileID
func (b *Bot) downloadImage(fileID string) ([]byte, error) {
	fileURL, err := b.api.GetFileDirectURL(fileID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить URL файла: %w", err)
	}

	// Используем общий HTTP клиент
	resp, err := b.httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось скачать файл: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP %d при скачивании файла", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (b *Bot) Api() *tgbotapi.BotAPI {
	return b.api
}

// sendBarcodeNotFound отправляет сообщение если штрих-код не найден
func (b *Bot) sendBarcodeNotFound(chatID int64) {
	text := `❌ Не удалось распознать штрих-код на фото.

Советы для лучшего распознавания:
• 📏 Убедитесь, что штрих-код четкий и не размытый
• 💡 Хорошее освещение без бликов
• 📐 Прямой угол съемки
• 🔍 Штрих-код занимает большую часть фото

Или введите цифры штрих-кода вручную.`

	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

// sendBarcodeDetectorError отправляет сообщение если barcodeDetector недоступен
func (b *Bot) sendBarcodeDetectorError(chatID int64) {
	text := `🔧 Распознавание фото временно недоступно.

Пожалуйста, введите цифры штрих-кода вручную.

Техническая информация: сервис распознавания изображений не настроен.`

	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *Bot) sendError(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+message)
	b.api.Send(msg)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Close закрывает ресурсы бота
func (b *Bot) Close() error {
	if b.httpClient != nil {
		// Закрываем idle соединения
		b.httpClient.CloseIdleConnections()
	}
	return nil
}
