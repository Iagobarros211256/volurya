package usecase

import (
	"api/models"
	"api/repository"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type PaymentUsecase struct {
	orderRepo   *repository.OrderRepository
	paymentRepo *repository.PaymentRepository
	productRepo *repository.ProductRepository
}

func NewPaymentUsecase(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	productRepo *repository.ProductRepository,
) *PaymentUsecase {
	return &PaymentUsecase{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		productRepo: productRepo,
	}
}

func (pu *PaymentUsecase) CreateOrderForCheckout(
	userID int,
	items []models.CheckoutItemInput,
) (*models.OrderDetail, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	var totalPrice float64
	var orderItems []models.OrderItem

	// Validar estoque de todos os itens antes de criar qualquer coisa
	for _, item := range items {
		product, err := pu.productRepo.GetProductById(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found: %w", item.ProductID, err)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w: product %d requested %d available %d",
				ErrInsufficientStock, item.ProductID, item.Quantity, product.Stock)
		}

		unitPrice := product.Price
		totalPrice += unitPrice * float64(item.Quantity)

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: unitPrice,
		})
	}

	// Arredondar total para evitar imprecisão de float
	totalPrice = math.Round(totalPrice*100) / 100

	// Criar ordem
	order := models.Order{
		UserID: userID,
		Total:  totalPrice,
		Status: string(models.OrderStatusPending),
	}

	orderID, err := pu.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Persistir order_items no banco
	for i := range orderItems {
		orderItems[i].OrderID = orderID
	}
	if err := pu.orderRepo.CreateOrderItems(orderID, orderItems); err != nil {
		// Compensar — cancelar ordem órfã
		_ = pu.orderRepo.UpdateOrderStatus(orderID, string(models.OrderStatusCancelled))
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	// Decrementar estoque atomicamente
	for _, item := range orderItems {
		if err := pu.productRepo.DecrementStock(item.ProductID, item.Quantity); err != nil {
			// Compensar — cancelar ordem e restaurar itens já decrementados seria ideal
			// Por ora cancela a ordem e loga — melhoria futura: transação completa
			_ = pu.orderRepo.UpdateOrderStatus(orderID, string(models.OrderStatusCancelled))
			slog.Error("failed to decrement stock",
				"error", err,
				"product_id", item.ProductID,
				"order_id", orderID,
			)
			return nil, fmt.Errorf("failed to update stock: %w", err)
		}
	}

	slog.Info("order created for checkout",
		"order_id", orderID,
		"user_id", userID,
		"total", totalPrice,
		"item_count", len(items),
	)

	return &models.OrderDetail{
		ID:         orderID,
		UserID:     userID,
		Items:      orderItems,
		TotalPrice: totalPrice,
		Status:     models.OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// CancelOrder cancela uma ordem órfã
func (pu *PaymentUsecase) CancelOrder(orderID int) error {
	return pu.orderRepo.UpdateOrderStatus(orderID, string(models.OrderStatusCancelled))
}

func (pu *PaymentUsecase) UpdateOrderPaymentIntent(
	orderID int,
	paymentIntentID string,
	amount int,
) error {
	if orderID <= 0 || paymentIntentID == "" {
		return fmt.Errorf("invalid order or payment intent")
	}

	err := pu.orderRepo.UpdateOrderPaymentIntent(orderID, paymentIntentID, string(models.OrderStatusPending))
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	slog.Info("order updated with payment intent",
		"order_id", orderID,
		"payment_intent", paymentIntentID,
		"amount_cents", amount,
	)

	return nil
}

func (pu *PaymentUsecase) HandlePaymentSuccess(paymentIntentID string) error {
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is required")
	}

	payment, err := pu.paymentRepo.GetPaymentByIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Idempotência — evita processar duas vezes
	if payment.Status == models.PaymentStatusSucceeded {
		slog.Info("payment already processed, skipping",
			"payment_intent", paymentIntentID,
			"order_id", payment.OrderID,
		)
		return nil
	}

	if err := pu.paymentRepo.UpdatePaymentStatus(payment.ID, string(models.PaymentStatusSucceeded)); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	if err := pu.orderRepo.UpdateOrderStatus(payment.OrderID, string(models.OrderStatusPaid)); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	slog.Info("payment succeeded",
		"order_id", payment.OrderID,
		"payment_intent", paymentIntentID,
	)

	return nil
}

func (pu *PaymentUsecase) HandlePaymentFailed(paymentIntentID, errorMessage string) error {
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is required")
	}

	payment, err := pu.paymentRepo.GetPaymentByIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Idempotência
	if payment.Status == models.PaymentStatusFailed {
		return nil
	}

	if err := pu.paymentRepo.UpdatePaymentStatusWithError(
		payment.ID,
		string(models.PaymentStatusFailed),
		errorMessage,
	); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if err := pu.orderRepo.UpdateOrderStatus(payment.OrderID, string(models.OrderStatusFailed)); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	slog.Info("payment failed",
		"order_id", payment.OrderID,
		"payment_intent", paymentIntentID,
		"error", errorMessage,
	)

	return nil
}

func (pu *PaymentUsecase) CreatePaymentRecord(
	orderID int,
	paymentIntentID string,
	amount int,
	currency string,
) (*models.PaymentRecord, error) {
	if orderID <= 0 || paymentIntentID == "" || amount <= 0 {
		return nil, fmt.Errorf("invalid payment parameters")
	}

	payment := models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: paymentIntentID,
		Amount:          amount,
		Currency:        currency,
		Status:          models.PaymentStatusRequiresPaymentMethod,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	id, err := pu.paymentRepo.CreatePaymentRecord(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	payment.ID = id

	slog.Info("payment record created",
		"id", id,
		"order_id", orderID,
		"amount_cents", amount,
	)

	return &payment, nil
}
