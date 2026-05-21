package usecase

import (
	"api/models"
	"api/repository"
	"fmt"
	"log/slog"
	"time"
)

// PaymentUsecase handles payment operations with Stripe
type PaymentUsecase struct {
	orderRepo   *repository.OrderRepository
	paymentRepo *repository.PaymentRepository
	productRepo *repository.ProductRepository
}

// NewPaymentUsecase creates a new payment usecase
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

// CreateOrderForCheckout creates an order and prepares it for payment
func (pu *PaymentUsecase) CreateOrderForCheckout(
	userID int,
	items []models.CheckoutItemInput,
) (*models.OrderDetail, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Calculate total and validate stock
	var totalPrice float64
	var orderItems []models.OrderItem

	for _, item := range items {
		product, err := pu.productRepo.GetProductById(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found: %w", item.ProductID, err)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("product %d has insufficient stock: requested %d, available %d",
				item.ProductID, item.Quantity, product.Stock)
		}

		unitPrice := product.Price
		totalPrice += unitPrice * float64(item.Quantity)

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: unitPrice,
		})
	}

	// Create order in database with pending status
	order := models.Order{
		UserID:   userID,
		Total:    totalPrice,
		Status:   string(models.OrderStatusPending),
	}

	orderID, err := pu.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	slog.Info("Order created for checkout",
		"order_id", orderID,
		"user_id", userID,
		"total", totalPrice,
		"item_count", len(items),
	)

	// Return order detail
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

// UpdateOrderPaymentIntent updates order with payment intent info
func (pu *PaymentUsecase) UpdateOrderPaymentIntent(
	orderID int,
	paymentIntentID string,
	amount int, // in cents
) error {
	if orderID <= 0 || paymentIntentID == "" {
		return fmt.Errorf("invalid order or payment intent")
	}

	// Update order with payment intent ID and status
	err := pu.orderRepo.UpdateOrderPaymentIntent(orderID, paymentIntentID, string(models.OrderStatusPending))
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	slog.Info("Order updated with payment intent",
		"order_id", orderID,
		"payment_intent", paymentIntentID,
		"amount", amount,
	)

	return nil
}

// HandlePaymentSuccess updates order status when payment succeeds
func (pu *PaymentUsecase) HandlePaymentSuccess(paymentIntentID string) error {
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is required")
	}

	// Get payment record to find order
	payment, err := pu.paymentRepo.GetPaymentByIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Update payment status
	err = pu.paymentRepo.UpdatePaymentStatus(payment.ID, string(models.PaymentStatusSucceeded))
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update order status to paid
	err = pu.orderRepo.UpdateOrderStatus(payment.OrderID, string(models.OrderStatusPaid))
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	slog.Info("Payment succeeded",
		"order_id", payment.OrderID,
		"payment_intent", paymentIntentID,
	)

	return nil
}

// HandlePaymentFailed updates order status when payment fails
func (pu *PaymentUsecase) HandlePaymentFailed(paymentIntentID, errorMessage string) error {
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is required")
	}

	// Get payment record to find order
	payment, err := pu.paymentRepo.GetPaymentByIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Update payment status with error message
	err = pu.paymentRepo.UpdatePaymentStatusWithError(payment.ID, string(models.PaymentStatusFailed), errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Update order status to failed
	err = pu.orderRepo.UpdateOrderStatus(payment.OrderID, string(models.OrderStatusFailed))
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	slog.Info("Payment failed",
		"order_id", payment.OrderID,
		"payment_intent", paymentIntentID,
		"error", errorMessage,
	)

	return nil
}

// CreatePaymentRecord creates a new payment record for tracking
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

	slog.Info("Payment record created",
		"id", id,
		"order_id", orderID,
		"amount", amount,
	)

	return &payment, nil
}

// GetOrderWithPayment retrieves an order with its payment information
func (pu *PaymentUsecase) GetOrderWithPayment(orderID int) (*models.OrderDetail, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("invalid order ID")
	}

	// This would fetch order from repository
	// For now, it's a placeholder for future implementation
	return nil, fmt.Errorf("not implemented")
}
