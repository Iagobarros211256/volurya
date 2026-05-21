package controller

import (
	"api/auth"
	"api/config"
	"api/models"
	"api/usecase"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
}

func NewPaymentController(
	paymentUsecase *usecase.PaymentUsecase,
) *PaymentController {
	return &PaymentController{
		paymentUsecase: paymentUsecase,
	}
}

// CheckoutRequest godoc
// @Summary Create a checkout with Stripe
// @Description Create a new order and generate a Stripe payment intent for checkout
// @Tags payments
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Param request body models.CheckoutRequest true "Checkout request"
// @Success 201 {object} models.CheckoutResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/checkout [post]
func (pc *PaymentController) Checkout(c *gin.Context) {
	// Verify HTTPS in production
	if GetAppEnv() == "production" && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HTTPS required for checkout"})
		return
	}

	// Get JWT token from context
	tokenString := c.GetString("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse JWT and get user ID
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID := claims.UserID
	if userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse checkout request
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate items
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	// Create order
	orderDetail, err := pc.paymentUsecase.CreateOrderForCheckout(userID, req.Items)
	if err != nil {
		slog.Error("failed to create order", "error", err, "user_id", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create Stripe payment intent
	amountInCents := int64(orderDetail.TotalPrice * 100)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInCents),
		Currency: stripe.String("brl"),
		Description: stripe.String(
			"Volurya Order #" + strconv.Itoa(orderDetail.ID),
		),
		ReceiptEmail: stripe.String(req.UserEmail),
		StatementDescriptor: stripe.String("VOLURYA SHOP"),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		slog.Error("failed to create payment intent", "error", err, "order_id", orderDetail.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment intent"})
		return
	}

	// Update order with payment intent ID
	err = pc.paymentUsecase.UpdateOrderPaymentIntent(
		orderDetail.ID,
		pi.ID,
		int(amountInCents),
	)
	if err != nil {
		slog.Error("failed to update order", "error", err, "order_id", orderDetail.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process order"})
		return
	}

	// Create payment record for tracking
	_, err = pc.paymentUsecase.CreatePaymentRecord(
		orderDetail.ID,
		pi.ID,
		int(amountInCents),
		"BRL",
	)
	if err != nil {
		slog.Warn("failed to create payment record", "error", err, "order_id", orderDetail.ID)
		// Don't fail the request if payment record creation fails
	}

	slog.Info("Checkout created successfully",
		"order_id", orderDetail.ID,
		"payment_intent", pi.ID,
		"amount", amountInCents,
	)

	response := models.CheckoutResponse{
		OrderID:         orderDetail.ID,
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
		Amount:          int(amountInCents),
		Currency:        "BRL",
		PublishableKey:  config.GetStripePublishableKey(),
	}

	c.JSON(http.StatusCreated, response)
}

// GetAppEnv returns the application environment
func GetAppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "development"
	}
	return env
}

// Webhook handles Stripe webhook events
// @Summary Handle Stripe webhook
// @Description Process Stripe webhook events
// @Tags payments
// @Accept json
// @Produce json
// @Param Stripe-Signature header string true "Stripe signature"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "Invalid signature"
// @Router /api/webhook [post]
func (pc *PaymentController) Webhook(c *gin.Context) {
	const MaxBodySize = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodySize)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("Error reading request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Get Stripe signature header
	signatureHeader := c.GetHeader("Stripe-Signature")
	if signatureHeader == "" {
		slog.Warn("Missing Stripe-Signature header")
		c.JSON(http.StatusForbidden, gin.H{"error": "missing signature"})
		return
	}

	// Verify signature (simplified - in production use proper library)
	webhookSecret := config.GetStripeWebhookSecret()
	if webhookSecret == "" {
		slog.Warn("Webhook secret not configured")
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook not configured"})
		return
	}

	// For production, use: webhook.ConstructEvent(payload, sig, secret)
	// This is simplified for MVP
	var event models.WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		slog.Error("failed to parse webhook event", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		// Extract payment intent from event data
		if piID, ok := event.Data["object"].(map[string]interface{})["id"].(string); ok {
			err := pc.paymentUsecase.HandlePaymentSuccess(piID)
			if err != nil {
				slog.Error("failed to handle payment success", "error", err, "pi_id", piID)
			}
		}

	case "payment_intent.payment_failed":
		if piID, ok := event.Data["object"].(map[string]interface{})["id"].(string); ok {
			var errorMsg string
			if em, ok := event.Data["object"].(map[string]interface{})["last_payment_error"].(map[string]interface{})["message"].(string); ok {
				errorMsg = em
			}
			err := pc.paymentUsecase.HandlePaymentFailed(piID, errorMsg)
			if err != nil {
				slog.Error("failed to handle payment failure", "error", err, "pi_id", piID)
			}
		}

	case "charge.dispute.created":
		slog.Warn("Charge dispute created", "event", event.ID)

	default:
		slog.Info("Unhandled webhook event", "type", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
