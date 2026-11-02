package controller

import (
	"api/config"
	"api/models"
	"api/usecase"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
}

func NewPaymentController(paymentUsecase *usecase.PaymentUsecase) *PaymentController {
	return &PaymentController{paymentUsecase: paymentUsecase}
}

func (pc *PaymentController) Checkout(c *gin.Context) {
	// HTTPS verificado pelo proxy — não revalidar aqui

	// userID vem do middleware JWT — não revalidar token
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	// Criar ordem
	orderDetail, err := pc.paymentUsecase.CreateOrderForCheckout(userID, req.Items)
	if err != nil {
		slog.Error("failed to create order", "error", err, "user_id", userID)
		if errors.Is(err, usecase.ErrInsufficientStock) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "insufficient stock"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// Converter para centavos com arredondamento correto
	amountInCents := int64(math.Round(orderDetail.TotalPrice * 100))

	params := &stripe.PaymentIntentParams{
		Amount:                    stripe.Int64(amountInCents),
		Currency:                  stripe.String("brl"),
		Description:               stripe.String("Volurya Order #" + strconv.Itoa(orderDetail.ID)),
		StatementDescriptorSuffix: stripe.String("VOLURYA"),
	}
	// Adicionar email apenas se fornecido
	if req.UserEmail != "" {
		params.ReceiptEmail = stripe.String(req.UserEmail)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		slog.Error("failed to create payment intent", "error", err, "order_id", orderDetail.ID)
		// Cancelar ordem órfã
		if cancelErr := pc.paymentUsecase.CancelOrder(orderDetail.ID); cancelErr != nil {
			slog.Error("failed to cancel orphan order", "error", cancelErr, "order_id", orderDetail.ID)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment intent"})
		return
	}

	// Persistir vínculo e registro financeiro na mesma transação.
	if err := pc.paymentUsecase.RegisterPaymentIntentAtomically(
		orderDetail.ID,
		pi.ID,
		int(amountInCents),
		"BRL",
		req.UserEmail,
	); err != nil {
		slog.Error("failed to register payment intent atomically",
			"error", err,
			"order_id", orderDetail.ID,
			"payment_intent", pi.ID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process order",
		})
		return
	}

	slog.Info("checkout created",
		"order_id", orderDetail.ID,
		"payment_intent", pi.ID,
		"amount_cents", amountInCents,
		"user_id", userID,
	)

	c.JSON(http.StatusCreated, models.CheckoutResponse{
		OrderID:         orderDetail.ID,
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
		Amount:          int(amountInCents),
		Currency:        "BRL",
		// PublishableKey removida — deve ser configurada estaticamente no frontend
	})
}

func (pc *PaymentController) Webhook(c *gin.Context) {
	const maxBodySize = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("failed to read webhook body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	signatureHeader := c.GetHeader("Stripe-Signature")
	if signatureHeader == "" {
		slog.Warn("missing Stripe-Signature header")
		c.JSON(http.StatusForbidden, gin.H{"error": "missing signature"})
		return
	}

	webhookSecret := config.GetStripeWebhookSecret()
	if webhookSecret == "" {
		slog.Error("STRIPE_WEBHOOK_SECRET not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook not configured"})
		return
	}

	// Validação real da assinatura do Stripe
	event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, webhookSecret, webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}

	slog.Info("webhook received", "type", event.Type, "id", event.ID)

	switch event.Type {
	case "payment_intent.succeeded":
		piID, err := extractPaymentIntentID(event.Data.Raw)
		if err != nil {
			slog.Error("failed to extract payment intent ID", "error", err, "event_id", event.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment intent payload"})
			return
		}
		if err := pc.paymentUsecase.HandlePaymentSuccess(piID); err != nil {
			slog.Error("failed to handle payment success", "error", err, "pi_id", piID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook processing failed"})
			return
		}

	case "payment_intent.payment_failed":
		piID, err := extractPaymentIntentID(event.Data.Raw)
		if err != nil {
			slog.Error("failed to extract payment intent ID", "error", err, "event_id", event.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment intent payload"})
			return
		}
		errorMsg := extractPaymentErrorMessage(event.Data.Raw)
		if err := pc.paymentUsecase.HandlePaymentFailed(piID, errorMsg); err != nil {
			slog.Error("failed to handle payment failure", "error", err, "pi_id", piID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook processing failed"})
			return
		}

	case "charge.dispute.created":
		slog.Warn("charge dispute created", "event_id", event.ID)

	default:
		slog.Info("unhandled webhook event", "type", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// extractPaymentIntentID extrai o ID do payment intent do payload raw do webhook
func extractPaymentIntentID(raw []byte) (string, error) {
	var payload struct {
		ID     string `json:"id"`
		Object string `json:"object"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}

	if payload.Object != "payment_intent" {
		return "", errors.New("invalid payment intent object")
	}

	if payload.ID == "" {
		return "", errors.New("payment intent ID not found")
	}

	return payload.ID, nil
}

// extractPaymentErrorMessage extrai a mensagem de erro do payload raw
func extractPaymentErrorMessage(raw []byte) string {
	var payload struct {
		Object           string `json:"object"`
		LastPaymentError *struct {
			Message string `json:"message"`
		} `json:"last_payment_error"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}

	if payload.Object != "payment_intent" {
		return ""
	}

	if payload.LastPaymentError == nil {
		return ""
	}

	return payload.LastPaymentError.Message
}
