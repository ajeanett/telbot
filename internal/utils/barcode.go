package utils

import "regexp"

var digitsOnly = regexp.MustCompile(`^\d+$`)

// IsValidBarcode проверяет формат (8–13 цифр); контрольная сумма —
// только для 13-значных кодов (EAN-13), 8–12-значные проходят лишь проверку формата.
func IsValidBarcode(barcode string) bool {
	if len(barcode) < 8 || len(barcode) > 13 {
		return false
	}
	if !digitsOnly.MatchString(barcode) {
		return false
	}
	if len(barcode) == 13 {
		return checkEAN13Sum(barcode)
	}
	return true
}

// checkEAN13Sum сверяет контрольную цифру EAN-13.
func checkEAN13Sum(barcode string) bool {
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(barcode[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	check := int(barcode[12] - '0')
	return check == (10-(sum%10))%10
}
