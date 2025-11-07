package services

import (
	"github.com/ajeanett/telbot/internal/models"
	"github.com/ajeanett/telbot/internal/utils"
	"strings"
)

type Analyzer struct {
	dangerousIngredients  map[string]string
	suspiciousIngredients map[string]string
	additives             map[string]string
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		dangerousIngredients: map[string]string{
			"e951": "Аспартам (искусственный подсластитель)",
			"e621": "Глутамат натрия (усилитель вкуса)",
			"e250": "Нитрит натрия (консервант)",
			"e211": "Бензоат натрия (консервант)",
			"e102": "Тартразин (краситель)",
		},
		suspiciousIngredients: map[string]string{
			"пальмовое масло": "Пальмовое масло",
			"palm oil":        "Пальмовое масло",
			"гмо":             "ГМО",
			"gmo":             "ГМО",
			"трансжиры":       "Трансжиры",
			"trans fat":       "Трансжиры",
			"краситель":       "Искусственные красители",
			"консервант":      "Консерванты",
			"ароматизатор":    "Искусственные ароматизаторы",
			"усилитель вкуса": "Усилители вкуса",
		},
		additives: map[string]string{
			"e471":  "Моно- и диглицериды жирных кислот (эмульгатор)",
			"e440":  "Пектин (загуститель)",
			"e965":  "Мальтит (подсластитель)",
			"e422":  "Глицерин (влагоудерживающий агент)",
			"e150a": "Сахарный колер I (краситель)",
			"e306":  "Концентрат смеси токоферолов (антиокислитель)",
		},
	}
}

func (a *Analyzer) AnalyzeProduct(product *models.Product) *models.AnalysisResult {
	result := &models.AnalysisResult{
		Product: product,
	}

	// Анализируем состав из ingredients_text
	if product.Composition != "" {
		composition := strings.ToLower(product.Composition)
		a.analyzeComposition(composition, result)
	}

	// Анализируем список ингредиентов
	if len(product.Ingredients) > 0 {
		a.analyzeIngredientsList(product.Ingredients, result)
	}

	// Анализируем пищевые добавки (E-шки)
	if len(product.Additives) > 0 {
		a.analyzeAdditives(product.Additives, result)
	}

	// Формируем итоговые рекомендации
	a.generateRecommendations(result)

	return result
}

func (a *Analyzer) analyzeComposition(composition string, result *models.AnalysisResult) {
	// Проверяем опасные ингредиенты
	for code, description := range a.dangerousIngredients {
		if strings.Contains(composition, code) {
			result.Dangerous = append(result.Dangerous, description)
		}
	}

	// Проверяем сомнительные ингредиенты
	for ingredient, description := range a.suspiciousIngredients {
		if strings.Contains(composition, ingredient) {
			result.Warnings = append(result.Warnings, description)
		}
	}
}

func (a *Analyzer) analyzeIngredientsList(ingredients []models.Ingredient, result *models.AnalysisResult) {
	for _, ingredient := range ingredients {
		text := strings.ToLower(ingredient.Text)

		// Проверяем каждый ингредиент
		for ing, description := range a.suspiciousIngredients {
			if strings.Contains(text, ing) {
				result.Warnings = utils.AppendIfNotExists(result.Warnings, description)
			}
		}

		for code, description := range a.dangerousIngredients {
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
		if description, exists := a.additives[code]; exists {
			result.Warnings = utils.AppendIfNotExists(result.Warnings, "Добавка "+code+": "+description)
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

	// Добавляем информацию о добавках если есть
	if len(result.Warnings) > 0 {
		result.Recommendations = append(result.Recommendations,
			"💡 Обратите внимание на пищевые добавки в составе")
	}
}

// // Вспомогательная функция чтобы избежать дубликатов
// func appendIfNotExists(slice []string, item string) []string {
//     for _, existing := range slice {
//         if existing == item {
//             return slice
//         }
//     }
//     return append(slice, item)
// }
