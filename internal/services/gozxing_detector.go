// services/barcode_detector_simple.go
package services

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"regexp"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

type BarcodeDetector struct{}

func NewBarcodeDetector() *BarcodeDetector {
	return &BarcodeDetector{}
}

func (d *BarcodeDetector) DetectFromImage(imageData []byte) (string, error) {
	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("не удалось декодировать изображение: %w", err)
	}

	// Конвертируем для gozxing
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("ошибка создания bitmap: %w", err)
	}

	// Пробуем разные настройки декодера
	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
		gozxing.DecodeHintType_POSSIBLE_FORMATS: []gozxing.BarcodeFormat{
			gozxing.BarcodeFormat_EAN_13,
			gozxing.BarcodeFormat_EAN_8,
			gozxing.BarcodeFormat_UPC_A,
			gozxing.BarcodeFormat_UPC_E,
			gozxing.BarcodeFormat_CODE_128,
			gozxing.BarcodeFormat_CODE_39,
		},
	}

	// Используем стандартный мультиформатный ридер
	reader := oned.NewMultiFormatUPCEANReader(hints)

	// Пытаемся распознать
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось распознать штрих-код: %w", err)
	}

	barcode := result.GetText()

	// Проверяем валидность
	if !isValidBarcode(barcode) {
		return "", fmt.Errorf("невалидный штрих-код: %s", barcode)
	}

	return barcode, nil
}

func isValidBarcode(barcode string) bool {
	// Убираем все нецифровые символы
	clean := regexp.MustCompile(`\D`).ReplaceAllString(barcode, "")

	// Проверяем длину и что остались только цифры
	if len(clean) < 8 || len(clean) > 13 {
		return false
	}

	return regexp.MustCompile(`^\d+$`).MatchString(clean)
}
