package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCartPageGuardHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		salesMode    string
		payments     string
		wantStatus   int
		wantLocation string
		wantHandler  bool
	}{
		{
			name:         "whatsapp redireciona",
			salesMode:    "whatsapp",
			payments:     "false",
			wantStatus:   http.StatusFound,
			wantLocation: "/store",
			wantHandler:  false,
		},
		{
			name:         "whatsapp bloqueia mesmo com pagamentos ativos",
			salesMode:    "whatsapp",
			payments:     "true",
			wantStatus:   http.StatusFound,
			wantLocation: "/store",
			wantHandler:  false,
		},
		{
			name:         "pagamentos desabilitados redirecionam",
			salesMode:    "",
			payments:     "false",
			wantStatus:   http.StatusFound,
			wantLocation: "/store",
			wantHandler:  false,
		},
		{
			name:        "loja tradicional permite carrinho",
			salesMode:   "",
			payments:    "true",
			wantStatus:  http.StatusOK,
			wantHandler: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VOLURYA_SALES_MODE", tt.salesMode)
			t.Setenv("PAYMENTS_ENABLED", tt.payments)

			handlerExecuted := false

			router := gin.New()
			router.GET("/cart", CartPageGuard, func(c *gin.Context) {
				handlerExecuted = true
				c.String(http.StatusOK, "Carrinho disponível")
			})

			req := httptest.NewRequest(http.MethodGet, "/cart", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"status=%d, esperado=%d",
					rec.Code, tt.wantStatus,
				)
			}

			if got := rec.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf(
					"Location=%q, esperado=%q",
					got, tt.wantLocation,
				)
			}

			if handlerExecuted != tt.wantHandler {
				t.Errorf(
					"handler executado=%v, esperado=%v",
					handlerExecuted, tt.wantHandler,
				)
			}
		})
	}
}
