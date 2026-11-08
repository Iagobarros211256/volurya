package controller

import (
	"api/config"
	"api/models"
	"api/repository"
	"api/usecase"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

// paymentIntentGateway permite testar o checkout sem chamadas reais ao Stripe.
type paymentIntentGateway interface {
	Create(*stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
	Get(string, *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}

type stripePaymentIntentGateway struct{}

func (stripePaymentIntentGateway) Create(
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	return paymentintent.New(params)
}

func (stripePaymentIntentGateway) Get(
	id string,
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(id, params)
}

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
	paymentIntents paymentIntentGateway
}

func NewPaymentController(paymentUsecase *usecase.PaymentUsecase) *PaymentController {
	return &PaymentController{
		paymentUsecase: paymentUsecase,
		paymentIntents: stripePaymentIntentGateway{},
	}
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

	// Mantém a mesma representação do e-mail no fingerprint,
	// no Stripe e no registro financeiro.
	req.UserEmail = strings.ToLower(strings.TrimSpace(req.UserEmail))

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	if strings.TrimSpace(req.IdempotencyKey) != req.IdempotencyKey {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid idempotency key"})
		return
	}

	// Cria ou recupera o pedido usando a chave da tentativa.
	orderDetail, reused, err := pc.paymentUsecase.CreateOrderForCheckoutIdempotent(
		userID, req.Items, req.IdempotencyKey, req.UserEmail,
	)
	if err != nil {
		slog.Error("failed to create order", "error", err, "user_id", userID)
		if errors.Is(err, usecase.ErrPreorderCheckout) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "product available only through assisted preorder",
			})
			return
		}
		if errors.Is(err, repository.ErrCheckoutIdempotencyConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "idempotency key conflict"})
			return
		}
		if errors.Is(err, usecase.ErrInsufficientStock) ||
			errors.Is(err, repository.ErrCheckoutInsufficientStock) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "insufficient stock"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	if reused && orderDetail.Status != models.OrderStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "checkout is no longer pending"})
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

	// A mesma tentativa utiliza a mesma chave no PostgreSQL e no Stripe.
	keyHash := sha256.Sum256([]byte(strconv.Itoa(userID) + ":" + req.IdempotencyKey))
	params.IdempotencyKey = stripe.String(
		"volurya-checkout-" + hex.EncodeToString(keyHash[:]),
	)

	var pi *stripe.PaymentIntent
	if reused && orderDetail.PaymentIntentID != "" {
		// O pagamento já foi vinculado ao pedido.
		pi, err = pc.paymentIntents.Get(orderDetail.PaymentIntentID, nil)
	} else {
		// Repetições sem vínculo persistido reutilizam a chave no Stripe.
		pi, err = pc.paymentIntents.Create(params)
	}
	if err != nil || pi == nil || pi.ID == "" {
		// Não cancelar automaticamente: uma falha de rede pode ocorrer
		// depois de o Stripe aceitar a criação do PaymentIntent.
		slog.Error("failed to obtain payment intent", "error", err,
			"order_id", orderDetail.ID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "payment service unavailable"})
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

	responseStatus := http.StatusCreated
	if reused {
		responseStatus = http.StatusOK
	}

	c.JSON(responseStatus, models.CheckoutResponse{
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
