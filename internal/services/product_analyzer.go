package services

import "github.com/ajeanett/telbot/internal/models"

// ProductAnalyzer — интерфейс анализа продукта по словарям.
type ProductAnalyzer interface {
	AnalyzeProduct(product *models.Product) *models.AnalysisResult
}
