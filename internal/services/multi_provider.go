package services

import (
	"fmt"
	"log"

	"github.com/ajeanett/telbot/internal/models"
)

// MultiProductProvider опрашивает несколько провайдеров последовательно
type MultiProductProvider struct {
	providers []ProductProvider
}

func NewMultiProductProvider(providers ...ProductProvider) *MultiProductProvider {
	return &MultiProductProvider{
		providers: providers,
	}
}

func (m *MultiProductProvider) GetProductByBarcode(barcode string) (*models.Product, error) {
	for i, provider := range m.providers {
		log.Printf("Запрос к провайдеру #%d", i+1)
		product, err := provider.GetProductByBarcode(barcode)
		if err == nil {
			log.Printf("✅ Продукт найден в провайдере #%d", i+1)
			return product, nil
		}
		log.Printf("❌ Провайдер #%d не вернул продукт: %v", i+1, err)
		// TODO: Возможно потребуется ограничение по времени в будущем
	}

	return nil, fmt.Errorf("продукт не найден ни в одном из провайдеров")
}
