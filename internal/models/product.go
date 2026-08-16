package models

// Product — доменная модель продукта. Не зависит от контрактов внешних API:
// маппинг JSON выполняют провайдеры через собственные приватные DTO.
type Product struct {
	Barcode        string
	Name           string
	Brand          string
	Country        string // страна происхождения
	Composition    string // текст состава
	Ingredients    []Ingredient
	Additives      []string // теги OFF вида "en:e471"
	Allergens      string   // теги OFF вида "en:gluten,en:milk"
	ImageURL       string
	SourceName     string // «Роскачество» / «Open Food Facts»
	SourceURL      string // карточка товара на сайте источника
	Rating         int    // 0–5; 0 = рейтинга нет
	HasQualityMark bool
}

// Ingredient — ингредиент из структурированного списка.
type Ingredient struct {
	Text string
}

// AnalysisResult — результат анализа продукта.
type AnalysisResult struct {
	Product         *Product
	Healthy         bool
	Allergens       []string // найденные группы аллергенов
	Warnings        []string
	Dangerous       []string
	Recommendations []string
}
