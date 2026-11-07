package models

// PublicProduct contém somente os campos destinados
// à exibição pública no catálogo da Volurya.
type PublicProduct struct {
	SalesType   string           `json:"sales_type"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	ID          int              `json:"id_product"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Price       float64          `json:"price"`
	Stock       int              `json:"stock"`
	ImageURL    string           `json:"image_url,omitempty"`
}

func ToPublicProduct(product Product) PublicProduct {
	salesType := product.SalesType
	if salesType == "" {
		salesType = SalesTypeStandard
	}

	return PublicProduct{
		SalesType:   salesType,
		Variants:    product.Variants,
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
	}
}
