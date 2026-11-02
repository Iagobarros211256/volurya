package controller

import (
	"api/config"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestWebhookRejectsInvalidSignatures(t *testing.T) {
	// Segredo fictício, usado exclusivamente no teste.
	original := config.StripeWebhookSecret
	config.StripeWebhookSecret = "whsec_volurya_local_test_only"

	t.Cleanup(func() {
		config.StripeWebhookSecret = original
	})

	router := gin.New()
	router.POST("/webhook", NewPaymentController(nil).Webhook)

	payload := `{
		"id": "evt_local_test",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_local_test",
				"object": "payment_intent"
			}
		}
	}`

	cases := []struct {
		name      string
		signature string
	}{
		{
			name: "assinatura ausente",
		},
		{
			name: "assinatura inválida",
			signature: "t=" +
				strconv.FormatInt(time.Now().Unix(), 10) +
				",v1=" + strings.Repeat("0", 64),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/webhook",
				strings.NewReader(payload),
			)
			req.Header.Set("Content-Type", "application/json")

			if tc.signature != "" {
				req.Header.Set("Stripe-Signature", tc.signature)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf(
					"HTTP recebido: %d; esperado: 403; body: %s",
					rec.Code,
					rec.Body.String(),
				)
			}
		})
	}
}
