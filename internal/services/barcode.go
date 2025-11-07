package services

import (
	"encoding/json"
	"fmt"
	"github.com/ajeanett/telbot/internal/models"
	"io"
	"net/http"
)

type BarcodeService struct {
	apiURL string
}

func NewBarcodeService(apiURL string) *BarcodeService {
	return &BarcodeService{
		apiURL: apiURL,
	}
}

func (s *BarcodeService) GetProductByBarcode(barcode string) (*models.Product, error) {
	url := fmt.Sprintf("%s/product/%s.json", s.apiURL, barcode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("продукт не найден (статус: %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var response models.APIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if response.Status != 1 {
		return nil, fmt.Errorf("продукт не найден в базе")
	}

	return &response.Product, nil
}
