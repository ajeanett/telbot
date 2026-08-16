package services

// BarcodeDetector — интерфейс для распознавания штрих-кода из изображения
type BarcodeDetector interface {
	DetectFromImage(imageData []byte) (string, error)
}
