package services

import (
	"sort"
	"testing"

	"github.com/ajeanett/telbot/internal/models"
)

// Состав майонеза Mr. Ricco из карточки Роскачества — контрольный пример:
// яйца (яичный желток) и горчица (масло горчичное) находятся,
// «масло подсолнечное» НЕ даёт ложного «молока».
const mayonnaiseComposition = "масло подсолнечное рафинированное дезодорированное, вода, " +
	"масло оливковое рафинированное, сахар, яичный желток, соль, пищевые волокна цитрусовые, " +
	"регулятор кислотности уксусная кислота, масло горчичное нерафинированное, " +
	"масло оливковое нерафинированное \"Extra Virgin\", краситель каротин."

func analyze(t *testing.T, product *models.Product) *models.AnalysisResult {
	t.Helper()
	return NewAnalyzer().AnalyzeProduct(product)
}

func TestAnalyzerAllergensFromComposition(t *testing.T) {
	result := analyze(t, &models.Product{Composition: mayonnaiseComposition})

	got := append([]string(nil), result.Allergens...)
	sort.Strings(got)
	want := []string{"Горчица", "Яйца"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Allergens = %v, хочу %v", got, want)
	}
}

func TestAnalyzerAllergensNoFalsePositive(t *testing.T) {
	result := analyze(t, &models.Product{Composition: "масло подсолнечное, вода, соль"})
	for _, a := range result.Allergens {
		if a == "Молоко и лактоза" {
			t.Error("«масло подсолнечное» не должно давать аллерген «Молоко и лактоза»")
		}
	}
}

func TestAnalyzerAllergensNoFalsePositiveOnLookalikes(t *testing.T) {
	// «мускатный орех» — приправа, «ароматизатор крабовый» — без ракообразных,
	// «сахар-сырец» — не сыр. Ни один не должен давать аллерген.
	result := analyze(t, &models.Product{
		Composition: "масло подсолнечное, мускатный орех, ароматизатор крабовый, сахар-сырец",
	})
	if len(result.Allergens) != 0 {
		t.Errorf("Allergens = %v, хочу пустой список", result.Allergens)
	}
}

func TestAnalyzerAllergensRealNutsCheeseShrimp(t *testing.T) {
	result := analyze(t, &models.Product{
		Composition: "грецкий орех, сыр твёрдый, креветки",
	})
	found := map[string]bool{}
	for _, a := range result.Allergens {
		found[a] = true
	}
	if !found["Орехи"] || !found["Молоко и лактоза"] || !found["Ракообразные"] {
		t.Errorf("Allergens = %v, хочу Орехи, Молоко и лактоза, Ракообразные", result.Allergens)
	}
}

func TestAnalyzerAllergensFromOFFTags(t *testing.T) {
	result := analyze(t, &models.Product{Allergens: "en:gluten,en:milk"})

	found := map[string]bool{}
	for _, a := range result.Allergens {
		found[a] = true
	}
	if !found["Глютен (пшеница, рожь, ячмень)"] || !found["Молоко и лактоза"] {
		t.Errorf("Allergens = %v, хочу группы глютена и молока", result.Allergens)
	}
}

func TestAnalyzerDeduplicatesWarnings(t *testing.T) {
	// Composition и Ingredients дублируют друг друга — в Warnings не должно быть повторов
	product := &models.Product{
		Composition: "состав: пальмовое масло, сахар",
		Ingredients: []models.Ingredient{{Text: "пальмовое масло"}, {Text: "сахар"}},
	}
	result := analyze(t, product)

	counts := map[string]int{}
	for _, w := range result.Warnings {
		counts[w]++
	}
	for w, n := range counts {
		if n > 1 {
			t.Errorf("предупреждение %q встречается %d раз", w, n)
		}
	}
}

func TestAnalyzerDangerousAdditive(t *testing.T) {
	result := analyze(t, &models.Product{
		Composition: "вода, e621",
		Additives:   []string{"en:e471"},
	})
	if len(result.Dangerous) == 0 {
		t.Error("добавка e621 должна попасть в Dangerous")
	}
	if !contains(result.Warnings, "Добавка e471") {
		t.Errorf("Warnings = %v, хочу запись о добавке e471", result.Warnings)
	}
}

func contains(list []string, substr string) bool {
	for _, s := range list {
		if len(s) >= len(substr) && s[:len(substr)] == substr {
			return true
		}
	}
	return false
}
