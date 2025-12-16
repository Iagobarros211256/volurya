package usecase

import (
	"api/models"
	"api/repository"
	"errors"
)

type ProductUsecase struct {
	repository repository.ProductRepository
}

func NewProductUseCase(repo repository.ProductRepository) ProductUsecase {
	return ProductUsecase{
		repository: repo,
	}
}

func (pu *ProductUsecase) GetProducts() ([]models.Product, error) {
	return pu.repository.GetProducts()
}

func (pu *ProductUsecase) CreateProduct(product models.Product) (models.Product, error) {

	productId, err := pu.repository.CreateProduct(product)
	if err != nil {
		return models.Product{}, err
	}

	product.ID = productId

	return product, nil
}

func (pu *ProductUsecase) GetProductById(id_product int) (*models.Product, error) {

	product, err := pu.repository.GetProductById(id_product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (pu *ProductUsecase) UpdateProduct(id_product int, Name string, Description string, Price float64, Stock int) (*models.Product, error) {

	// 🔒 Regra de negócio
	if Price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	product, err := pu.repository.UpdateProduct(id_product, Name, Description, Price, Stock)
	if err != nil {
		return nil, err
	}

	return product, nil
}
