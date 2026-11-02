package controller

import (
	"api/config"
	"api/repository"
	"api/usecase"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	_ "github.com/lib/pq"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestWebhookWithValidSignature(t *testing.T) {
	const secret = "whsec_volurya_hmac_local_only"

	original := config.StripeWebhookSecret
	config.StripeWebhookSecret = secret
	t.Cleanup(func() {
		config.StripeWebhookSecret = original
	})

	router := gin.New()
	router.POST("/webhook", NewPaymentController(nil).Webhook)

	cases := []struct {
		name     string
		payload  string
		expected int
	}{
		{
			name: "evento ignorado validamente",
			payload: `{
				"id": "evt_ignored",
				"object": "event",
				"type": "local.test",
				"data": {
					"object": {
						"id": "pi_local",
						"object": "payment_intent"
					}
				}
			}`,
			expected: http.StatusOK,
		},
		{
			name: "payment intent sem identificador",
			payload: `{
				"id": "evt_missing_id",
				"object": "event",
				"type": "payment_intent.succeeded",
				"data": {
					"object": {
						"object": "payment_intent"
					}
				}
			}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			timestamp := strconv.FormatInt(
				time.Now().Unix(), 10,
			)

			message := timestamp + "." + tc.payload

			mac := hmac.New(
				sha256.New,
				[]byte(secret),
			)
			_, _ = mac.Write([]byte(message))

			signature := "t=" + timestamp +
				",v1=" + hex.EncodeToString(mac.Sum(nil))

			req := httptest.NewRequest(
				http.MethodPost,
				"/webhook",
				strings.NewReader(tc.payload),
			)
			req.Header.Set(
				"Stripe-Signature", signature,
			)
			req.Header.Set(
				"Content-Type", "application/json",
			)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.expected {
				t.Errorf(
					"HTTP recebido=%d, esperado=%d, body=%s",
					rec.Code,
					tc.expected,
					rec.Body.String(),
				)
			}
		})
	}
}

func TestWebhookReturns500OnProcessingFailure(t *testing.T) {
	const secret = "whsec_volurya_http500_local_only"

	original := config.StripeWebhookSecret
	config.StripeWebhookSecret = secret
	t.Cleanup(func() {
		config.StripeWebhookSecret = original
	})

	// Criamos e fechamos o handle SQL para simular
	// uma falha interna, sem conexão de rede.
	testDB, err := sql.Open(
		"postgres",
		"host=127.0.0.1 port=1 user=test password=test "+
			"dbname=volurya_payment_test sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := testDB.Close(); err != nil {
		t.Fatal(err)
	}

	paymentRepo := repository.NewPaymentRepository(testDB)
	paymentUsecase := usecase.NewPaymentUsecase(
		nil, paymentRepo, nil, nil,
	)

	router := gin.New()
	router.POST(
		"/webhook",
		NewPaymentController(paymentUsecase).Webhook,
	)

	payload := `{
		"id": "evt_local_db_failure",
		"object": "event",
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_local_db_failure",
				"object": "payment_intent"
			}
		}
	}`

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + payload))

	signature := "t=" + timestamp +
		",v1=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(
		http.MethodPost,
		"/webhook",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", signature)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(
			"HTTP recebido=%d, esperado=500, body=%s",
			rec.Code,
			rec.Body.String(),
		)
	}
}
