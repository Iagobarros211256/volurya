package usecase

import (
	"testing"
)

func TestCartValidation_PositivePrice(t *testing.T) {
	cartItems := []struct {
		name     string
		price    float64
		quantity int
		valid    bool
	}{
		{"item with price", 10.50, 1, true},
		{"item with zero price", 0, 1, false},
		{"item with negative price", -5.00, 1, false},
		{"item with zero quantity", 10.50, 0, false},
		{"item with negative quantity", 10.50, -1, false},
	}

	for _, item := range cartItems {
		t.Run(item.name, func(t *testing.T) {
			isValid := item.price > 0 && item.quantity > 0
			if isValid != item.valid {
				t.Fatalf("expected valid=%v, got %v", item.valid, isValid)
			}
		})
	}
}
