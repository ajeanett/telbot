package models

import "encoding/json"

type Product struct {
	Barcode     string       `json:"code"`
	Name        string       `json:"product_name"`
	Brand       string       `json:"brands"`
	Ingredients []Ingredient `json:"ingredients"`
	Composition string       `json:"ingredients_text"`
	ImageURL    string       `json:"image_url"`
	Additives   []string     `json:"additives_tags"`
	Allergens   string       `json:"allergens"`
}

type Ingredient struct {
	ID         string      `json:"id"`
	Text       string      `json:"text"`
	Percent    json.Number `json:"percent"`
	PercentMin json.Number `json:"percent_min"`
	PercentMax json.Number `json:"percent_max"`
	Vegan      string      `json:"vegan"`
	Vegetarian string      `json:"vegetarian"`
}

// Структура для ответа API
type APIResponse struct {
	Status  int     `json:"status"`
	Product Product `json:"product"`
}

// Результат анализа продукта
type AnalysisResult struct {
	Product         *Product
	Healthy         bool
	Warnings        []string
	Dangerous       []string
	Recommendations []string
}
