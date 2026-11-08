package usecase

import (
	"api/models"
	"api/repository"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type checkoutFingerprintItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type checkoutFingerprintPayload struct {
	UserEmail string                    `json:"user_email"`
	Items     []checkoutFingerprintItem `json:"items"`
}

// CreateOrderForCheckoutIdempotent cria ou recupera um pedido
// identificado por uma chave estável da tentativa de checkout.
func (pu *PaymentUsecase) CreateOrderForCheckoutIdempotent(
	userID int,
	items []models.CheckoutItemInput,
	key string,
	userEmail string,
) (*models.OrderDetail, bool, error) {
	if userID <= 0 {
		return nil, false, fmt.Errorf("invalid user")
	}

	if key == "" || len(key) > 128 ||
		key != strings.TrimSpace(key) {
		return nil, false, fmt.Errorf("invalid checkout idempotency key")
	}

	if len(items) == 0 {
		return nil, false, fmt.Errorf("cart is empty")
	}

	canonicalItems := make([]checkoutFingerprintItem, 0, len(items))

	for _, item := range items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return nil, false, fmt.Errorf("invalid checkout item")
		}

		canonicalItems = append(canonicalItems, checkoutFingerprintItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	sort.Slice(canonicalItems, func(i, j int) bool {
		if canonicalItems[i].ProductID != canonicalItems[j].ProductID {
			return canonicalItems[i].ProductID < canonicalItems[j].ProductID
		}

		return canonicalItems[i].Quantity < canonicalItems[j].Quantity
	})

	payload := checkoutFingerprintPayload{
		UserEmail: strings.ToLower(strings.TrimSpace(userEmail)),
		Items:     canonicalItems,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, false, fmt.Errorf(
			"encode checkout fingerprint: %w", err,
		)
	}

	hash := sha256.Sum256(encoded)
	fingerprint := hex.EncodeToString(hash[:])

	detail, reused, err := pu.orderRepo.CreateCheckoutOrderIdempotent(
		userID,
		items,
		key,
		fingerprint,
	)
	if errors.Is(err, repository.ErrCheckoutPreorder) {
		return nil, false, ErrPreorderCheckout
	}
	return detail, reused, err
}
