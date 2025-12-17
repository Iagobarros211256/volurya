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
	// essas regras nao serao necessarias agora mas no ffutoro e num ambiente de producao serao indispensaveis
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

//// provavelmete esse sistema VAI crescer.
// se prescisarmos de regras mais robustas isso podera ser adicionando
//func (pu *ProductUsecase) Delete(id int) (*models.Product, error) {

// 1. verificar existência

//product, err := pu.repository.GetProductById(id)
//if err != nil {
//	return nil, err
//}

// verificar se o produto tem estoque. se tiver nao podera ser deletado
//if product.Stock > 0 {
//	return nil, errors.New("cannot delete product with stock")
//}

// 3. deletar de fato
//err = pu.repository.Delete(id)
//if err != nil {
//	return nil, err
//}

// 4. retornar o que foi deletado (opcional)
//return product, nil

// delete one
func (pu *ProductUsecase) Delete(id int) error {
	return pu.repository.Delete(id)
}
