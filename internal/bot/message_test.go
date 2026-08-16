package bot

import (
	"strings"
	"testing"

	"github.com/ajeanett/telbot/internal/models"
)

func rskrfResult() *models.AnalysisResult {
	return &models.AnalysisResult{
		Product: &models.Product{
			Barcode:        "4604075031981",
			Name:           "Майонез оливковый organic, 67%",
			Brand:          "Mr.Ricco",
			Country:        "РОССИЯ",
			Composition:    "масло подсолнечное, яичный желток",
			SourceName:     "Роскачество",
			SourceURL:      "https://rskrf.ru/goods/mayonez-olivkovyy-organic-67/",
			Rating:         5,
			HasQualityMark: true,
		},
		Allergens:       []string{"Яйца"},
		Warnings:        []string{"Искусственные красители"},
		Recommendations: []string{"⚠️ Продукт содержит сомнительные ингредиенты"},
	}
}

func TestFormatAnalysisResultFull(t *testing.T) {
	msg := formatAnalysisResult(rskrfResult())

	fragments := []string{
		"Майонез оливковый organic, 67%",
		"*Бренд:* Mr.Ricco",
		"*Страна:* РОССИЯ",
		"*Штрих-код:* 4604075031981",
		"*Рейтинг Роскачества:* 5/5",
		"Знак качества",
		"*Состав:*",
		"ВОЗМОЖНЫЕ АЛЛЕРГЕНЫ",
		"Яйца",
		"СОМНИТЕЛЬНЫЕ ИНГРЕДИЕНТЫ",
		"*Рекомендации:*",
		"[Роскачество](https://rskrf.ru/goods/mayonez-olivkovyy-organic-67/)",
	}
	for _, f := range fragments {
		if !strings.Contains(msg, f) {
			t.Errorf("в сообщении нет фрагмента %q:\n%s", f, msg)
		}
	}
}

func TestFormatAnalysisResultHidesEmptyBlocks(t *testing.T) {
	result := &models.AnalysisResult{
		Product: &models.Product{
			Barcode: "5000159461123",
			Name:    "Snickers",
		},
		Recommendations: []string{"✅ Продукт выглядит безопасным"},
	}
	msg := formatAnalysisResult(result)

	for _, absent := range []string{"Страна:", "Бренд:", "Рейтинг", "АЛЛЕРГЕН", "СОМНИТЕЛЬНЫЕ", "ОПАСНЫЕ", "Источник"} {
		if strings.Contains(msg, absent) {
			t.Errorf("пустой блок %q не должен появляться:\n%s", absent, msg)
		}
	}
	if !strings.Contains(msg, "Не указан") {
		t.Error("при пустом составе должно быть «Не указан»")
	}
}
