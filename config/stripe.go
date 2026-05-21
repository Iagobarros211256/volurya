package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/stripe/stripe-go/v76"
)

var (
	StripeSecretKey     string
	StripePublishableKey string
	StripeWebhookSecret string
)

// InitStripe initializes the Stripe SDK with API keys from environment variables
func InitStripe() error {
	// Load from environment
	StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
	StripePublishableKey = os.Getenv("STRIPE_PUBLISHABLE_KEY")
	StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Validate that we have at least the secret key
	if StripeSecretKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY environment variable is required")
	}

	// Set the Stripe API key globally
	stripe.Key = StripeSecretKey

	slog.Info("Stripe SDK initialized successfully",
		"has_webhook_secret", StripeWebhookSecret != "",
		"publishable_key_prefix", publishableKeyPrefix(StripePublishableKey),
	)

	return nil
}

// GetStripeSecretKey returns the Stripe secret key
func GetStripeSecretKey() string {
	return StripeSecretKey
}

// GetStripePublishableKey returns the Stripe publishable key
func GetStripePublishableKey() string {
	return StripePublishableKey
}

// GetStripeWebhookSecret returns the Stripe webhook signing secret
func GetStripeWebhookSecret() string {
	return StripeWebhookSecret
}

// publishableKeyPrefix returns the first 10 chars of the key for logging
func publishableKeyPrefix(key string) string {
	if key == "" {
		return "not_set"
	}
	if len(key) <= 10 {
		return key
	}
	return key[:10] + "..."
}

// IsStripeEnabled checks if Stripe is properly configured
func IsStripeEnabled() bool {
	return StripeSecretKey != "" && StripePublishableKey != ""
}

// ValidateWebhookSignature validates a Stripe webhook signature
// This is used to ensure webhooks are from Stripe
func ValidateWebhookSignature(payload []byte, signature string) bool {
	if StripeWebhookSecret == "" {
		slog.Warn("webhook signature validation skipped: STRIPE_WEBHOOK_SECRET not configured")
		return false
	}

	// Import the webhook signing package for proper validation
	// For now, this is a placeholder - actual validation done in controller
	return signature != ""
}
