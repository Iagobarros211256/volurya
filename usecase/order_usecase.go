package usecase

import (
	"api/metrics"
	"api/models"
	"api/repository"
	"fmt"
)

type OrderUsecase struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
}

func NewOrderUsecase(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// CreateOrder cria uma ordem simples com um único produto.
// Usado pelo order_controller diretamente.
// Para checkout com múltiplos itens, use PaymentUsecase.CreateOrderForCheckout.
func (ou *OrderUsecase) CreateOrder(userID int, productID int, quantity int) (string, error) {
	if quantity <= 0 {
		return "", ErrInvalidQuantity
	}

	product, err := ou.productRepo.GetProductById(productID)
	if err != nil {
		return "", fmt.Errorf("product not found: %w", err)
	}

	if product.Stock < quantity {
		return "", fmt.Errorf("%w: requested %d, available %d",
			ErrInsufficientStock, quantity, product.Stock)
	}

	order := models.Order{
		UserID: userID,
		Total:  product.Price * float64(quantity),
		Status: string(models.OrderStatusPending),
	}

	orderID, err := ou.orderRepo.CreateOrder(order)
	if err != nil {
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	// Persistir item da ordem
	items := []models.OrderItem{
		{
			OrderID:   orderID,
			ProductID: productID,
			Quantity:  quantity,
			UnitPrice: product.Price,
		},
	}
	if err := ou.orderRepo.CreateOrderItems(orderID, items); err != nil {
		_ = ou.orderRepo.UpdateOrderStatus(orderID, string(models.OrderStatusCancelled))
		return "", fmt.Errorf("failed to create order items: %w", err)
	}

	// Decrementar estoque atomicamente
	if err := ou.productRepo.DecrementStock(productID, quantity); err != nil {
		_ = ou.orderRepo.UpdateOrderStatus(orderID, string(models.OrderStatusCancelled))
		return "", fmt.Errorf("failed to update stock: %w", err)
	}

	metrics.OrdersTotal.Inc()

	// Retorna URL vazia — pagamento é processado via PaymentUsecase/Stripe
	return "", nil
}
