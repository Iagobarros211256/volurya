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
		Amount:              stripe.Int64(amountInCents),
		Currency:            stripe.String("brl"),
		Description:         stripe.String("Volurya Order #" + strconv.Itoa(orderDetail.ID)),
		StatementDescriptor: stripe.String("VOLURYA SHOP"),
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

	// Atualizar ordem com payment intent
	if err := pc.paymentUsecase.UpdateOrderPaymentIntent(orderDetail.ID, pi.ID, int(amountInCents)); err != nil {
		slog.Error("failed to update order", "error", err, "order_id", orderDetail.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process order"})
		return
	}

	// Registrar pagamento para auditoria
	if _, err := pc.paymentUsecase.CreatePaymentRecord(orderDetail.ID, pi.ID, int(amountInCents), "BRL"); err != nil {
		slog.Error("failed to create payment record", "error", err, "order_id", orderDetail.ID)
		// Não falha o request mas loga como erro — não como warn
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
	event, err := webhook.ConstructEvent(payload, signatureHeader, webhookSecret)
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
			break
		}
		if err := pc.paymentUsecase.HandlePaymentSuccess(piID); err != nil {
			slog.Error("failed to handle payment success", "error", err, "pi_id", piID)
		}

	case "payment_intent.payment_failed":
		piID, err := extractPaymentIntentID(event.Data.Raw)
		if err != nil {
			slog.Error("failed to extract payment intent ID", "error", err, "event_id", event.ID)
			break
		}
		errorMsg := extractPaymentErrorMessage(event.Data.Raw)
		if err := pc.paymentUsecase.HandlePaymentFailed(piID, errorMsg); err != nil {
			slog.Error("failed to handle payment failure", "error", err, "pi_id", piID)
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
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", err
	}
	obj, ok := data["object"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid webhook data structure")
	}
	id, ok := obj["id"].(string)
	if !ok {
		return "", errors.New("payment intent ID not found")
	}
	return id, nil
}

// extractPaymentErrorMessage extrai a mensagem de erro do payload raw
func extractPaymentErrorMessage(raw []byte) string {
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return ""
	}
	obj, ok := data["object"].(map[string]interface{})
	if !ok {
		return ""
	}
	errData, ok := obj["last_payment_error"].(map[string]interface{})
	if !ok {
		return ""
	}
	msg, _ := errData["message"].(string)
	return msg
}
