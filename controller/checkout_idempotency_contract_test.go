package controller

import (
	"api/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCheckoutRequiresIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/checkout", func(c *gin.Context) {
		var req models.CheckoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusNoContent)
	})

	cases := []struct {
		name string
		body string
		want int
	}{
		{
			name: "sem chave",
			body: `{"items":[{"product_id":1,"quantity":1}]}`,
			want: http.StatusBadRequest,
		},
		{
			name: "com chave",
			body: `{"idempotency_key":"checkout-test-001","items":[{"product_id":1,"quantity":1}]}`,
			want: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost, "/checkout",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("status=%d, esperado=%d", rec.Code, tc.want)
			}
		})
	}
}
