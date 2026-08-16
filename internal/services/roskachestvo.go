package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/ajeanett/telbot/internal/models"
)

// RoskachestvoProvider получает продукты из открытого API Роскачества.
// Документация: https://rskrf.ru/about/dev/
// Условие использования — активная ссылка на портал (SourceURL) в выдаче.
type RoskachestvoProvider struct {
	apiURL string
	client *http.Client
}

func NewRoskachestvoProvider(apiURL string) *RoskachestvoProvider {
	return &RoskachestvoProvider{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
			// Антибот Роскачества: на запрос без куки — 302 на тот же URL
			// + Set-Cookie (__hash_). Без Jar клиент зацикливается
			// на редиректах («stopped after 10 redirects»).
			Jar: newCookieJar(),
		},
	}
}

// newCookieJar — Jar для прохождения cookie-challenge Роскачества.
// cookiejar.New(nil) ошибку не возвращает (невозможна при nil options).
func newCookieJar() http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("неожиданная ошибка cookiejar: " + err.Error())
	}
	return jar
}

// DTO ответов API Роскачества.
type rskrfSearchResponse struct {
	Response struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"response"`
}

type rskrfProductResponse struct {
	Response rskrfProduct `json:"response"`
}

type rskrfProduct struct {
	ID             int              `json:"id"`
	Title          string           `json:"title"`
	Manufacturer   string           `json:"manufacturer"`
	TotalRating    int              `json:"total_rating"`
	HasQualityMark bool             `json:"has_quality_mark"`
	ProductLink    string           `json:"product_link"`
	Thumbnail      string           `json:"thumbnail"`
	ProductInfo    []rskrfInfoEntry `json:"product_info"`
}

type rskrfInfoEntry struct {
	Name string `json:"name"`
	Info string `json:"info"`
}

func (p *RoskachestvoProvider) GetProductByBarcode(barcode string) (*models.Product, error) {
	// Шаг 1: поиск по штрихкоду
	id, err := p.searchByBarcode(barcode)
	if err != nil {
		return nil, err
	}

	// Шаг 2: карточка товара со составом и страной
	card, err := p.fetchProductCard(id)
	if err != nil {
		return nil, err
	}

	// Массив product_info → map по имени поля
	info := make(map[string]string, len(card.ProductInfo))
	for _, entry := range card.ProductInfo {
		info[entry.Name] = entry.Info
	}

	return &models.Product{
		Barcode:        barcode,
		Name:           card.Title,
		Brand:          card.Manufacturer,
		Country:        info["Страна производства"],
		Composition:    info["Состав"],
		ImageURL:       card.Thumbnail,
		SourceName:     "Роскачество",
		SourceURL:      card.ProductLink,
		Rating:         card.TotalRating,
		HasQualityMark: card.HasQualityMark,
	}, nil
}

// searchByBarcode ищет товар по штрихкоду. Пустой response = не найдено.
func (p *RoskachestvoProvider) searchByBarcode(barcode string) (int, error) {
	url := fmt.Sprintf("%s/search/barcode?barcode=%s", p.apiURL, barcode)

	resp, err := p.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("ошибка запроса к Роскачеству: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Роскачество: неожиданный статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var search rskrfSearchResponse
	if err := json.Unmarshal(body, &search); err != nil {
		return 0, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	if search.Response.ID == 0 {
		return 0, fmt.Errorf("продукт не найден в Роскачестве")
	}

	return search.Response.ID, nil
}

// fetchProductCard получает карточку товара.
func (p *RoskachestvoProvider) fetchProductCard(id int) (*rskrfProduct, error) {
	url := fmt.Sprintf("%s/product/%d/", p.apiURL, id)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса карточки: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("карточка товара: неожиданный статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения карточки: %w", err)
	}

	var product rskrfProductResponse
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("ошибка парсинга карточки: %w", err)
	}

	return &product.Response, nil
}
