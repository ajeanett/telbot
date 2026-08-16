package utils

import "testing"

func TestIsValidBarcode(t *testing.T) {
	cases := []struct {
		name    string
		barcode string
		want    bool
	}{
		{"EAN-13 валидный", "4604075031981", true},
		{"EAN-13 битая контрольная цифра", "4604075031982", false},
		{"EAN-8 валидный", "96385074", true},
		{"слишком короткий", "1234567", false},
		{"слишком длинный", "12345678901234", false},
		{"не цифры", "46040750319ab", false},
		{"пустой", "", false},
	}
	for _, c := range cases {
		if got := IsValidBarcode(c.barcode); got != c.want {
			t.Errorf("%s: IsValidBarcode(%q) = %v, хочу %v", c.name, c.barcode, got, c.want)
		}
	}
}
