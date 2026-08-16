package bot

import (
	"fmt"
	"strings"

	"github.com/ajeanett/telbot/internal/models"
)

// formatAnalysisResult собирает Markdown-ответ по результату анализа.
// Пустые блоки скрываются.
func formatAnalysisResult(result *models.AnalysisResult) string {
	var b strings.Builder
	p := result.Product

	b.WriteString(fmt.Sprintf("🏷️ *%s*\n", p.Name))
	if p.Brand != "" {
		b.WriteString(fmt.Sprintf("👨‍💼 *Бренд:* %s\n", p.Brand))
	}
	if p.Country != "" {
		b.WriteString(fmt.Sprintf("🌍 *Страна:* %s\n", p.Country))
	}
	b.WriteString(fmt.Sprintf("📊 *Штрих-код:* %s\n", p.Barcode))

	if p.Rating > 0 {
		b.WriteString(fmt.Sprintf("⭐ *Рейтинг Роскачества:* %d/5", p.Rating))
		if p.HasQualityMark {
			b.WriteString(" ✅ Знак качества")
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString("*Состав:*\n")
	if p.Composition != "" {
		b.WriteString(p.Composition + "\n\n")
	} else {
		b.WriteString("Не указан\n\n")
	}

	writeSection(&b, "🥜 *ВОЗМОЖНЫЕ АЛЛЕРГЕНЫ:*", result.Allergens)
	writeSection(&b, "🚫 *ОПАСНЫЕ ИНГРЕДИЕНТЫ:*", result.Dangerous)
	writeSection(&b, "⚠️ *СОМНИТЕЛЬНЫЕ ИНГРЕДИЕНТЫ:*", result.Warnings)

	b.WriteString("*Рекомендации:*\n")
	for _, rec := range result.Recommendations {
		b.WriteString(rec + "\n")
	}

	if p.SourceName != "" {
		b.WriteString(fmt.Sprintf("\n🔗 *Источник:* %s\n", sourceLine(p)))
	}

	return b.String()
}

// writeSection пишет блок со списком, если список не пуст.
func writeSection(b *strings.Builder, header string, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString(header + "\n")
	for _, item := range items {
		b.WriteString(fmt.Sprintf("• %s\n", item))
	}
	b.WriteString("\n")
}

// sourceLine — ссылка на источник (обязательное условие API Роскачества).
func sourceLine(p *models.Product) string {
	if p.SourceURL != "" {
		return fmt.Sprintf("[%s](%s)", p.SourceName, p.SourceURL)
	}
	return p.SourceName
}
