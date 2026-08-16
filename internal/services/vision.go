// services/vision.go
package services

import (
	"context"
	"fmt"
	"log"
	"regexp"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/ajeanett/telbot/internal/utils"
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
		if utils.IsValidBarcode(match) {
			return match
		}
	}
	return ""
}

func (s *VisionService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
