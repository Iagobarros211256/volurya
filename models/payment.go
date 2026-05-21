package models

import "time"

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusRequiresPaymentMethod PaymentStatus = "requires_payment_method"
	PaymentStatusRequiresAction        PaymentStatus = "requires_action"
	PaymentStatusProcessing            PaymentStatus = "processing"
	PaymentStatusSucceeded             PaymentStatus = "succeeded"
	PaymentStatusFailed                PaymentStatus = "failed"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents a single item in an order
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderDetail represents an order with all details
type OrderDetail struct {
	ID                 int          `json:"id"`
	UserID             int          `json:"user_id"`
	Items              []OrderItem  `json:"items"`
	TotalPrice         float64      `json:"total_price"`
	Status             OrderStatus  `json:"status"`
	PaymentIntentID    string       `json:"payment_intent_id,omitempty"`
	PaymentStatus      PaymentStatus `json:"payment_status,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

// PaymentRecord represents a payment transaction in the database
type PaymentRecord struct {
	ID                int           `json:"id"`
	OrderID           int           `json:"order_id"`
	PaymentIntentID   string        `json:"payment_intent_id"`
	Amount            int           `json:"amount"` // in cents (centavos)
	Currency          string        `json:"currency"`
	Status            PaymentStatus `json:"status"`
	StripeCustomerID  string        `json:"stripe_customer_id,omitempty"`
	ErrorMessage      string        `json:"error_message,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// CheckoutRequest represents the request to create a checkout
type CheckoutRequest struct {
	UserEmail  string              `json:"user_email" binding:"required,email"`
	Items      []CheckoutItemInput `json:"items" binding:"required,min=1"`
	SuccessURL string              `json:"success_url,omitempty"`
	CancelURL  string              `json:"cancel_url,omitempty"`
}

// CheckoutItemInput represents an item in the checkout request
type CheckoutItemInput struct {
	ProductID int `json:"product_id" binding:"required,min=1"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CheckoutResponse represents the response from creating a checkout
type CheckoutResponse struct {
	OrderID         int    `json:"order_id"`
	ClientSecret    string `json:"client_secret"`
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          int    `json:"amount"` // in cents
	Currency        string `json:"currency"`
	PublishableKey  string `json:"publishable_key,omitempty"` // Send to frontend for Stripe.js
}

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created int64                  `json:"created"`
}

// PaymentIntentWebhookData represents the data in a payment_intent webhook
type PaymentIntentWebhookData struct {
	Object map[string]interface{} `json:"object"`
}
