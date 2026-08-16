package services

import (
	"strings"

	"github.com/ajeanett/telbot/internal/models"
	"github.com/ajeanett/telbot/internal/utils"
)

// Analyzer анализирует состав продукта по словарям (analyzer_data.go).
// Внешних вызовов не делает.
type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) AnalyzeProduct(product *models.Product) *models.AnalysisResult {
	result := &models.AnalysisResult{
		Product: product,
	}

	// Анализируем состав из ingredients_text
	if product.Composition != "" {
		a.analyzeComposition(strings.ToLower(product.Composition), result)
	}

	// Анализируем список ингредиентов
	if len(product.Ingredients) > 0 {
		a.analyzeIngredientsList(product.Ingredients, result)
	}

	// Анализируем пищевые добавки (E-шки)
	if len(product.Additives) > 0 {
		a.analyzeAdditives(product.Additives, result)
	}

	// Ищем аллергены
	a.analyzeAllergens(product, result)

	// Формируем итоговые рекомендации
	a.generateRecommendations(result)

	return result
}

func (a *Analyzer) analyzeComposition(composition string, result *models.AnalysisResult) {
	for code, description := range dangerousIngredients {
		if strings.Contains(composition, code) {
			result.Dangerous = utils.AppendIfNotExists(result.Dangerous, description)
		}
	}
	for ingredient, description := range suspiciousIngredients {
		if strings.Contains(composition, ingredient) {
			result.Warnings = utils.AppendIfNotExists(result.Warnings, description)
		}
	}
}

func (a *Analyzer) analyzeIngredientsList(ingredients []models.Ingredient, result *models.AnalysisResult) {
	for _, ingredient := range ingredients {
		text := strings.ToLower(ingredient.Text)
		for ing, description := range suspiciousIngredients {
			if strings.Contains(text, ing) {
				result.Warnings = utils.AppendIfNotExists(result.Warnings, description)
			}
		}
		for code, description := range dangerousIngredients {
			if strings.Contains(text, code) {
				result.Dangerous = utils.AppendIfNotExists(result.Dangerous, description)
			}
		}
	}
}

func (a *Analyzer) analyzeAdditives(additives []string, result *models.AnalysisResult) {
	for _, additive := range additives {
		// Добавки приходят в формате "en:e471" - извлекаем код
		code := strings.TrimPrefix(additive, "en:")
		if description, exists := additiveDescriptions[code]; exists {
			result.Warnings = utils.AppendIfNotExists(result.Warnings, "Добавка "+code+": "+description)
		}
	}
}

// analyzeAllergens ищет аллергены EU-14 по ключевым словам в составе
// и по тегам аллергенов OFF ("en:gluten,en:milk").
func (a *Analyzer) analyzeAllergens(product *models.Product, result *models.AnalysisResult) {
	text := strings.ToLower(product.Composition)
	for _, ing := range product.Ingredients {
		text += " " + strings.ToLower(ing.Text)
	}

	for key, group := range allergens {
		if strings.Contains(text, key) {
			result.Allergens = utils.AppendIfNotExists(result.Allergens, group)
		}
	}

	for _, tag := range strings.Split(product.Allergens, ",") {
		if group, exists := offAllergenTags[strings.TrimSpace(tag)]; exists {
			result.Allergens = utils.AppendIfNotExists(result.Allergens, group)
		}
	}
}

func (a *Analyzer) generateRecommendations(result *models.AnalysisResult) {
	if len(result.Dangerous) > 0 {
		result.Healthy = false
		result.Recommendations = append(result.Recommendations,
			"🚫 Продукт содержит потенциально опасные ингредиенты")
	} else if len(result.Warnings) > 0 {
		result.Recommendations = append(result.Recommendations,
			"⚠️ Продукт содержит сомнительные ингредиенты")
	} else {
		result.Healthy = true
		result.Recommendations = append(result.Recommendations,
			"✅ Продукт выглядит безопасным")
	}

	if len(result.Allergens) > 0 {
		result.Recommendations = append(result.Recommendations,
			"ℹ️ Обратите внимание на аллергены, если у вас есть пищевые аллергии")
	}

	if len(result.Warnings) > 0 {
		result.Recommendations = append(result.Recommendations,
			"💡 Обратите внимание на пищевые добавки в составе")
	}
}
