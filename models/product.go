package models

type Product struct {
	SalesType   string           `json:"sales_type"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	ID          int              `json:"id_product"`
	UserID      int              `json:"user_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Price       float64          `json:"price"`
	Stock       int              `json:"stock"`
	ImageURL    string           `json:"image_url,omitempty"`
}
