package services

import (
	"errors"
	"testing"

	"github.com/ajeanett/telbot/internal/models"
)

type stubProvider struct {
	product *models.Product
	err     error
	calls   int
}

func (s *stubProvider) GetProductByBarcode(barcode string) (*models.Product, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.product, nil
}

func TestMultiProductProviderFirstSuccessWins(t *testing.T) {
	first := &stubProvider{product: &models.Product{Barcode: "1", Name: "Роскачество"}}
	second := &stubProvider{product: &models.Product{Barcode: "1", Name: "OFF"}}

	m := NewMultiProductProvider(first, second)
	product, err := m.GetProductByBarcode("4604075031981")
	if err != nil {
		t.Fatalf("ожидался успех: %v", err)
	}
	if product.Name != "Роскачество" {
		t.Errorf("выиграл %q, хочу первый провайдер", product.Name)
	}
	if second.calls != 0 {
		t.Error("второй провайдер не должен опрашиваться при успехе первого")
	}
}

func TestMultiProductProviderFallsBackOnError(t *testing.T) {
	failing := &stubProvider{err: errors.New("продукт не найден в Роскачестве")}
	ok := &stubProvider{product: &models.Product{Barcode: "1", Name: "OFF"}}

	m := NewMultiProductProvider(failing, ok)
	product, err := m.GetProductByBarcode("5000159461123")
	if err != nil {
		t.Fatalf("ожидался успех через фолбэк: %v", err)
	}
	if product.Name != "OFF" {
		t.Errorf("получен %q, хочу фолбэк на OFF", product.Name)
	}
}

func TestMultiProductProviderAllFailed(t *testing.T) {
	m := NewMultiProductProvider(&stubProvider{err: errors.New("нет")})
	if _, err := m.GetProductByBarcode("0000000000000"); err == nil {
		t.Fatal("ожидалась ошибка, когда все провайдеры промахнулись")
	}
}
