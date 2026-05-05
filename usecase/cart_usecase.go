package usecase

import (
	"api/models"
	"api/repository"
	"errors"
)

type CartUsecase struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartUsecase(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartUsecase {
	return &CartUsecase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// GetCart retorna o carrinho do usuário com todos os itens
func (uc *CartUsecase) GetCart(userID int) (*models.Cart, error) {
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	items, err := uc.cartRepo.GetCartItems(cart.ID)
	if err != nil {
		return nil, err
	}

	cart.Items = items
	return cart, nil
}

// AddItem adiciona um produto ao carrinho
func (uc *CartUsecase) AddItem(userID, productID, quantity int) (*models.CartItem, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	// Verifica se o produto existe e tem estoque
	product, err := uc.productRepo.GetProductById(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := uc.cartRepo.AddItem(cart.ID, productID, quantity)
	if err != nil {
		return nil, err
	}

	item.Product = product
	return item, nil
}

// UpdateItem atualiza a quantidade de um item do carrinho
func (uc *CartUsecase) UpdateItem(userID, itemID, quantity int) (*models.CartItem, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	// Verifica se o item pertence ao carrinho do usuário
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := uc.cartRepo.GetItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}
	if item.CartID != cart.ID {
		return nil, errors.New("forbidden: item does not belong to your cart")
	}

	// Verifica estoque
	product, err := uc.productRepo.GetProductById(item.ProductID)
	if err != nil {
		return nil, err
	}
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	updated, err := uc.cartRepo.UpdateItemQuantity(itemID, quantity)
	if err != nil {
		return nil, err
	}

	updated.Product = product
	return updated, nil
}

// RemoveItem remove um item do carrinho
func (uc *CartUsecase) RemoveItem(userID, itemID int) error {
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	item, err := uc.cartRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}
	if item.CartID != cart.ID {
		return errors.New("forbidden: item does not belong to your cart")
	}

	return uc.cartRepo.RemoveItem(itemID)
}

// Checkout finaliza o carrinho e cria uma ordem por produto
func (uc *CartUsecase) Checkout(userID int, orderUsecase *OrderUsecase) ([]string, error) {
	cart, err := uc.GetCart(userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	paymentURLs := make([]string, 0)
	for _, item := range cart.Items {
		url, err := orderUsecase.CreateOrder(userID, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		paymentURLs = append(paymentURLs, url)
	}

	// Limpa o carrinho após checkout
	if err := uc.cartRepo.ClearCart(cart.ID); err != nil {
		return nil, err
	}

	return paymentURLs, nil
}
