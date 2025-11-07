// services/vision.go
package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
)

var barCodeRegExp = regexp.MustCompile(`\b\d{8,13}\b`)

type VisionService struct {
	// в Google Vision API бесплатно только первые 1000 запросов в месяц
	client *vision.ImageAnnotatorClient
}

func NewVisionService(ctx context.Context) (*VisionService, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}
	return &VisionService{client: client}, nil
}

func (s *VisionService) DetectBarcodeViaText(imageData []byte) (string, error) {
	ctx := context.Background()

	img := &visionpb.Image{
		Content: imageData,
	}

	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: img,
				Features: []*visionpb.Feature{
					{
						Type:       visionpb.Feature_DOCUMENT_TEXT_DETECTION, // Лучше для штрих-кодов
						MaxResults: 1,
					},
				},
			},
		},
	}

	resp, err := s.client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return "", fmt.Errorf("Vision API error: %w", err)
	}

	if len(resp.Responses) == 0 {
		return "", fmt.Errorf("пустой ответ от Vision API")
	}

	response := resp.Responses[0]
	if response.Error != nil {
		return "", fmt.Errorf("ошибка API: %s", response.Error.Message)
	}

	// Извлекаем весь распознанный текст
	var detectedText string
	if response.FullTextAnnotation != nil {
		detectedText = response.FullTextAnnotation.GetText()
	}
	log.Printf("Распознанный текст: %s", detectedText)

	// Ищем штрих-код в тексте
	barcode := extractBarcodeFromText(detectedText)
	if barcode == "" {
		return "", fmt.Errorf("штрих-код не найден в распознанном тексте")
	}

	return barcode, nil
}

func extractBarcodeFromText(text string) string {
	if text == "" {
		return ""
	}

	// Ищем последовательности из 8-13 цифр
	matches := barCodeRegExp.FindAllString(text, -1)

	for _, match := range matches {
		if IsValidBarcode(match) {
			return match
		}
	}
	return ""
}

// isValidBarcode проверяет валидность штрих-кода
func IsValidBarcode(barcode string) bool {
	// Проверяем базовые условия
	if len(barcode) < 8 || len(barcode) > 13 {
		return false
	}

	// Проверяем что это только цифры
	matched, _ := regexp.MatchString(`^\d+$`, barcode)
	if !matched {
		return false
	}

	// Дополнительная проверка контрольной суммы для EAN-13
	if len(barcode) == 13 {
		return validateEAN13(barcode)
	}

	// TODO: для других форматов можно добавить дополнительные проверки
	return true
}

// validateEAN13 проверяет контрольную сумму EAN-13
func validateEAN13(barcode string) bool {
	if len(barcode) != 13 {
		return false
	}

	sum := 0
	for i, char := range barcode[:12] {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}

	checkDigit, err := strconv.Atoi(string(barcode[12]))
	if err != nil {
		return false
	}
	calculatedCheckDigit := (10 - (sum % 10)) % 10

	return checkDigit == calculatedCheckDigit
}

func (s *VisionService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
