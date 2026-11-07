package models

const (
	SalesTypeStandard = "standard"
	SalesTypePreorder = "preorder"
)

// ProductVariant representa uma opção de um produto,
// como o tamanho de uma camiseta.
type ProductVariant struct {
	ID           int    `json:"id"`
	ProductID    int    `json:"product_id"`
	Label        string `json:"label"`
	PriceCents   int64  `json:"price_cents"`
	DisplayOrder int    `json:"display_order"`
}
