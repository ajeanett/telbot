package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// rskrfServer поднимает фейковый API Роскачества на фикстурах.
func rskrfServer(t *testing.T, miss bool, cardStatus int) *httptest.Server {
	t.Helper()
	searchFixture := "testdata/rskrf_search_hit.json"
	if miss {
		searchFixture = "testdata/rskrf_search_miss.json"
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search/barcode":
			writeFixture(t, w, searchFixture)
		case r.URL.Path == "/product/3514028/":
			if cardStatus != http.StatusOK {
				w.WriteHeader(cardStatus)
				return
			}
			writeFixture(t, w, "testdata/rskrf_product_card.json")
		default:
			http.NotFound(w, r)
		}
	}))
}

func writeFixture(t *testing.T, w http.ResponseWriter, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("не удалось прочитать фикстуру %s: %v", path, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func TestRoskachestvoProviderCookieChallenge(t *testing.T) {
	// Антибот Роскачества: на запрос без куки — 302 на тот же URL + Set-Cookie
	// __hash_; повторный запрос с кукушкой — 200. Без CookieJar клиент
	// зацикливается на редиректах («stopped after 10 redirects»).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search/barcode":
			if _, err := r.Cookie("__hash_"); err != nil {
				http.SetCookie(w, &http.Cookie{Name: "__hash_", Value: "test123"})
				http.Redirect(w, r, r.URL.RequestURI(), http.StatusFound)
				return
			}
			writeFixture(t, w, "testdata/rskrf_search_hit.json")
		case r.URL.Path == "/product/3514028/":
			writeFixture(t, w, "testdata/rskrf_product_card.json")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewRoskachestvoProvider(srv.URL)
	product, err := p.GetProductByBarcode("4604075031981")
	if err != nil {
		t.Fatalf("ожидался успех через cookie-challenge, получена ошибка: %v", err)
	}
	if product.Name == "" {
		t.Error("Name пуст после прохождения cookie-challenge")
	}
}

func TestRoskachestvoProviderMapping(t *testing.T) {
	srv := rskrfServer(t, false, http.StatusOK)
	defer srv.Close()

	p := NewRoskachestvoProvider(srv.URL)
	product, err := p.GetProductByBarcode("4604075031981")
	if err != nil {
		t.Fatalf("ожидался успех, получена ошибка: %v", err)
	}

	assertions := map[string]struct{ got, want string }{
		"Name":       {product.Name, "Майонез оливковый organic, 67%"},
		"Brand":      {product.Brand, "Mr.Ricco"},
		"Country":    {product.Country, "РОССИЯ"},
		"SourceName": {product.SourceName, "Роскачество"},
		"SourceURL":  {product.SourceURL, "https://rskrf.ru/goods/mayonez-olivkovyy-organic-67/"},
	}
	for name, a := range assertions {
		if a.got != a.want {
			t.Errorf("%s = %q, хочу %q", name, a.got, a.want)
		}
	}
	if !strings.HasPrefix(product.Composition, "масло подсолнечное") {
		t.Errorf("Composition = %q, хочу начало «масло подсолнечное»", product.Composition)
	}
	if product.Rating != 5 {
		t.Errorf("Rating = %d, хочу 5", product.Rating)
	}
	if !product.HasQualityMark {
		t.Error("HasQualityMark = false, хочу true")
	}
}

func TestRoskachestvoProviderNotFound(t *testing.T) {
	srv := rskrfServer(t, true, http.StatusOK)
	defer srv.Close()

	p := NewRoskachestvoProvider(srv.URL)
	_, err := p.GetProductByBarcode("1234567890123")
	if err == nil {
		t.Fatal("ожидалась ошибка «продукт не найден в Роскачестве»")
	}
}

func TestRoskachestvoProviderCardFailure(t *testing.T) {
	// Поиск удался, карточка упала — провайдер обязан вернуть ошибку,
	// чтобы MultiProductProvider ушёл в фолбэк на OFF.
	srv := rskrfServer(t, false, http.StatusInternalServerError)
	defer srv.Close()

	p := NewRoskachestvoProvider(srv.URL)
	if _, err := p.GetProductByBarcode("4604075031981"); err == nil {
		t.Fatal("ожидалась ошибка при сбое запроса карточки")
	}
}
