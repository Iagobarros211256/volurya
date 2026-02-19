package usecase

import (
	"api/models"
	"api/repository"
	"errors"
	"strconv"
)

func NewProductUsecase(repo *repository.ProductRepository) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

type ProductUsecase struct {
	repo *repository.ProductRepository // ← pointer aqui também
}

func (pu *ProductUsecase) GetProducts(limitStr string, cursorStr string) ([]models.Product, *int, bool, error) {

	const (
		defaultLimit = 10
		maxLimit     = 50
	)

	limit := defaultLimit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			return nil, nil, false, errors.New("invalid limit")
		}
		if parsedLimit > maxLimit {
			limit = maxLimit
		} else {
			limit = parsedLimit
		}
	}

	var cursor *int
	if cursorStr != "" {
		parsedCursor, err := strconv.Atoi(cursorStr)
		if err != nil || parsedCursor < 0 {
			return nil, nil, false, errors.New("invalid cursor")
		}
		cursor = &parsedCursor
	}

	products, hasMore, err :=
		pu.repo.GetProducts(limit, cursor) // ← mudou de repository para repo
	if err != nil {
		return nil, nil, false, err
	}

	var nextCursor *int
	if hasMore && len(products) > 0 {
		lastID := products[len(products)-1].ID
		nextCursor = &lastID
	}

	return products, nextCursor, hasMore, nil
}

func (pu *ProductUsecase) CreateProduct(userID int, product models.Product) (models.Product, error) {

	productId, err := pu.repo.CreateProduct(product, userID) // ← mudou para repo
	if err != nil {
		return models.Product{}, err
	}

	product.ID = productId
	product.UserID = userID
	return product, nil
}

func (pu *ProductUsecase) GetProductById(id_product int) (*models.Product, error) {

	product, err := pu.repo.GetProductById(id_product) // ← mudou para repo
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (pu *ProductUsecase) UpdateProduct(id_product int, Name string, Description string, Price float64, Stock int) (*models.Product, error) {
	// essas regras nao serao necessarias agora mas no futuro e num ambiente de producao serao indispensaveis
	if Price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	product, err := pu.repo.UpdateProduct(id_product, Name, Description, Price, Stock) // ← mudou para repo
	if err != nil {
		return nil, err
	}

	return product, nil
}

// delete one
func (pu *ProductUsecase) Delete(userID int, productID int) error {
	product, err := pu.repo.GetProductById(productID)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	// Checagem de ownership
	if product.UserID != userID {
		return errors.New("forbidden: you do not own this product")
	}

	// Regra futura para admin (opcional)
	// if userRole == "admin" { ok }

	return pu.repo.Delete(productID)
}
