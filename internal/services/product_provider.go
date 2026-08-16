package services

import "github.com/ajeanett/telbot/internal/models"

// ProductProvider — интерфейс для получения продукта по штрих-коду
type ProductProvider interface {
	GetProductByBarcode(barcode string) (*models.Product, error)
}