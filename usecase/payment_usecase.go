package usecase

import (
	"api/email"
	"api/models"
	"api/repository"
	"fmt"
	"log/slog"
	"math"
	"time"
)

type PaymentUsecase struct {
	orderRepo    *repository.OrderRepository
	paymentRepo  *repository.PaymentRepository
	productRepo  *repository.ProductRepository
	emailService *email.EmailService
}

func NewPaymentUsecase(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	productRepo *repository.ProductRepository,
	emailService *email.EmailService,
) *PaymentUsecase {
	return &PaymentUsecase{
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		productRepo:  productRepo,
		emailService: emailService,
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

		if product.SalesType != models.SalesTypeStandard {
			return nil, ErrPreorderCheckout
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

	// Pedido, itens e estoque são persistidos na mesma transação.
	order := models.Order{
		UserID: userID,
		Total:  totalPrice,
		Status: string(models.OrderStatusPending),
	}

	orderID, err := pu.orderRepo.CreateCheckoutOrder(order, orderItems)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout order: %w", err)
	}

	for i := range orderItems {
		orderItems[i].OrderID = orderID
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
	return pu.orderRepo.CancelCheckoutOrderAndRestoreStock(orderID)
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

	// O repositório atualiza pagamento e pedido na mesma transação.
	// A verificação de idempotência ocorre sob bloqueio das linhas.
	updated, err := pu.paymentRepo.ConfirmPaymentSucceeded(paymentIntentID)
	if err != nil {
		return fmt.Errorf("failed to confirm payment atomically: %w", err)
	}

	if !updated {
		slog.Info("payment already processed, skipping",
			"payment_intent", paymentIntentID,
			"order_id", payment.OrderID,
		)
		return nil
	}

	slog.Info("payment succeeded",
		"order_id", payment.OrderID,
		"payment_intent", paymentIntentID,
	)

	// Envia email de confirmação — falha não bloqueia o fluxo
	if pu.emailService != nil && payment.UserEmail != "" {
		go func() {
			if err := pu.emailService.SendOrderConfirmation(
				payment.UserEmail,
				payment.OrderID,
				payment.Amount,
				nil,
			); err != nil {
				slog.Error("failed to send order confirmation email",
					"error", err,
					"order_id", payment.OrderID,
				)
			}
		}()
	}

	return nil
}

func (pu *PaymentUsecase) HandlePaymentFailed(
	paymentIntentID, errorMessage string,
) error {
	if paymentIntentID == "" {
		return fmt.Errorf("payment intent ID is required")
	}

	updated, err := pu.paymentRepo.RecordPaymentFailure(
		paymentIntentID, errorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to record payment failure: %w", err)
	}

	if !updated {
		slog.Info("payment failure ignored or already processed",
			"payment_intent", paymentIntentID,
		)
		return nil
	}

	slog.Info("payment attempt failed",
		"payment_intent", paymentIntentID,
	)

	return nil
}

func (pu *PaymentUsecase) CreatePaymentRecord(
	orderID int,
	paymentIntentID string,
	amount int,
	currency string,
	userEmail string,
) (*models.PaymentRecord, error) {
	if orderID <= 0 || paymentIntentID == "" || amount <= 0 {
		return nil, fmt.Errorf("invalid payment parameters")
	}

	payment := models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: paymentIntentID,
		Amount:          amount,
		Currency:        currency,
		UserEmail:       userEmail,
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
