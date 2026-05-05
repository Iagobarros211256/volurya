package models

import "time"

type Cart struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	Items     []CartItem `json:"items"`
}

type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"added_at"`
	Product   *Product  `json:"product,omitempty"`
}
