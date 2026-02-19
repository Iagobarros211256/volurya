package models

type Product struct {
	ID          int     `json:"id_product"`
	UserID      int     `json:"user_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}
