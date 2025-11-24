package helpers

import "strings"

// ConvertAmount converts a major currency amount to its minor unit equivalent.
// Example: NGN → kobo, USD → cents.
func ConvertAmount(amount int, currency string) int64 {
	switch strings.ToUpper(currency) {
	// ×100 currencies
	case "NGN", "USD", "EUR", "GBP", "GHS", "KES":
		return int64(amount * 100)

	// ×1 currencies (no minor units)
	case "JPY", "KRW", "VND":
		return int64(amount)

	default:
		// default to ×100 for unknown currencies,
		return int64(amount * 100)
	}
}
