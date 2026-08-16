package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenFoodFactsProviderMapping(t *testing.T) {
	fixture := `{
		"status": 1,
		"product": {
			"product_name": "Snickers",
			"brands": "Mars",
			"ingredients_text": "Сахар, глюкозный сироп, молоко",
			"image_url": "https://example.com/snickers.jpg",
			"additives_tags": ["en:e471"],
			"allergens": "en:milk,en:peanuts",
			"countries": "Франция, Германия",
			"ingredients": [{"text": "сахар"}, {"text": "молоко"}]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/product/5000159461123.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fixture))
	}))
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.URL)
	product, err := p.GetProductByBarcode("5000159461123")
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	assertions := map[string]struct{ got, want string }{
		"Name":       {product.Name, "Snickers"},
		"Brand":      {product.Brand, "Mars"},
		"Country":    {product.Country, "Франция"},
		"SourceName": {product.SourceName, "Open Food Facts"},
		"SourceURL":  {product.SourceURL, srv.URL + "/product/5000159461123"},
	}
	for name, a := range assertions {
		if a.got != a.want {
			t.Errorf("%s = %q, хочу %q", name, a.got, a.want)
		}
	}
	if len(product.Ingredients) != 2 || product.Ingredients[0].Text != "сахар" {
		t.Errorf("Ingredients = %v, хочу два элемента, первый «сахар»", product.Ingredients)
	}
	if product.Allergens != "en:milk,en:peanuts" {
		t.Errorf("Allergens = %q", product.Allergens)
	}
}

func TestOpenFoodFactsProviderNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": 0}`))
	}))
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.URL)
	if _, err := p.GetProductByBarcode("1234567890123"); err == nil {
		t.Fatal("ожидалась ошибка «продукт не найден»")
	}
}
