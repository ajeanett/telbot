package services

import (
	"fmt"
	"log"
)

// MultiBarcodeDetector — пробует несколько детекторов по порядку
type MultiBarcodeDetector struct {
	detectors []BarcodeDetector
}

func NewMultiBarcodeDetector(detectors ...BarcodeDetector) *MultiBarcodeDetector {
	return &MultiBarcodeDetector{
		detectors: detectors,
	}
}

func (m *MultiBarcodeDetector) DetectFromImage(imageData []byte) (string, error) {
	var lastErr error

	for i, detector := range m.detectors {
		log.Printf("Попытка распознавания с детектором #%d", i+1)
		barcode, err := detector.DetectFromImage(imageData)
		if err == nil {
			log.Printf("✅ Успешно распознан штрих-код: %s", barcode)
			return barcode, nil
		}
		log.Printf("❌ Детектор #%d не смог распознать: %v", i+1, err)
		lastErr = err
	}

	return "", fmt.Errorf("все детекторы не смогли распознать штрих-код: %w", lastErr)
}
