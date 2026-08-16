package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ajeanett/telbot/internal/models"
)

// OpenFoodFactsProvider получает продукты из открытой базы Open Food Facts.
type OpenFoodFactsProvider struct {
	apiURL string
	client *http.Client
}

func NewOpenFoodFactsProvider(apiURL string) *OpenFoodFactsProvider {
	return &OpenFoodFactsProvider{
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// DTO контракта OFF — приватные, наружу не выходят.
type offResponse struct {
	Status  int        `json:"status"`
	Product offProduct `json:"product"`
}

type offProduct struct {
	Name        string          `json:"product_name"`
	Brand       string          `json:"brands"`
	Ingredients []offIngredient `json:"ingredients"`
	Composition string          `json:"ingredients_text"`
	ImageURL    string          `json:"image_url"`
	Additives   []string        `json:"additives_tags"`
	Allergens   string          `json:"allergens"`
	Countries   string          `json:"countries"`
}

type offIngredient struct {
	Text string `json:"text"`
}

func (s *OpenFoodFactsProvider) GetProductByBarcode(barcode string) (*models.Product, error) {
	url := fmt.Sprintf("%s/product/%s.json", s.apiURL, barcode)

	resp, err := s.client.Get(url)
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

	var response offResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if response.Status != 1 {
		return nil, fmt.Errorf("продукт не найден в базе")
	}

	return s.toProduct(barcode, &response.Product), nil
}

func (s *OpenFoodFactsProvider) toProduct(barcode string, p *offProduct) *models.Product {
	ingredients := make([]models.Ingredient, 0, len(p.Ingredients))
	for _, ing := range p.Ingredients {
		ingredients = append(ingredients, models.Ingredient{Text: ing.Text})
	}

	return &models.Product{
		Barcode:     barcode,
		Name:        p.Name,
		Brand:       p.Brand,
		Country:     firstCountry(p.Countries),
		Composition: p.Composition,
		Ingredients: ingredients,
		ImageURL:    p.ImageURL,
		Additives:   p.Additives,
		Allergens:   p.Allergens,
		SourceName:  "Open Food Facts",
		SourceURL:   s.productURL(barcode),
	}
}

// firstCountry берёт первую страну из списка ("Франция, Германия" → "Франция").
func firstCountry(countries string) string {
	if idx := strings.Index(countries, ","); idx >= 0 {
		return strings.TrimSpace(countries[:idx])
	}
	return strings.TrimSpace(countries)
}

// productURL — человеческая карточка товара, а не JSON-эндпоинт API.
func (s *OpenFoodFactsProvider) productURL(barcode string) string {
	site := strings.TrimSuffix(s.apiURL, "/api/v0")
	return fmt.Sprintf("%s/product/%s", site, barcode)
}
