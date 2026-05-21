package controller

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetProducts_ValidLimit(t *testing.T) {
	tests := []struct {
		limit    string
		expected int
	}{
		{"10", http.StatusOK},
		{"1", http.StatusOK},
		{"100", http.StatusOK},
		{"0", http.StatusBadRequest},
		{"101", http.StatusBadRequest},
		{"abc", http.StatusBadRequest},
		{"", http.StatusOK},
	}

	for _, test := range tests {
		t.Run("limit="+test.limit, func(t *testing.T) {
			query := url.Values{}
			if test.limit != "" {
				query.Set("limit", test.limit)
			}
			expected := test.expected
			if expected == 0 {
				t.Skip("unable to test without full router setup")
			}
		})
	}
}

//Esse arquivo é basicamente vazio de valor.Esse arquivo é basicamente vazio de valor. Análise direta:

//🔴 Nenhum teste é executado de fato
//goif expected == 0 {
//    t.Skip("unable to test without full router setup")
//}
//expected nunca é 0 — os valores são http.StatusOK (200) ou http.StatusBadRequest (400). Então o t.Skip nunca é chamado. Mas o teste também nunca faz nada além de construir um url.Values que não vai a lugar nenhum. Não há request, não há controller, não há assertion real.

//🔴 url.Values construído mas nunca usado
//goquery := url.Values{}
//if test.limit != "" {
//    query.Set("limit", test.limit)
//}
// fim do teste
//A query string é montada e descartada. Nenhuma requisição é feita.

//O que o teste deveria fazer:
//gofunc TestGetProducts_ValidLimit(t *testing.T) {
//    gin.SetMode(gin.TestMode)

//    mockUsecase := &mockProductUsecase{}
//    controller := NewProductController(mockUsecase, ...)

//    tests := []struct {
//        limit          string
//        expectedStatus int
//    }{
//        {"10",  http.StatusOK},
//        {"1",   http.StatusOK},
//        {"100", http.StatusOK},
//        {"0",   http.StatusBadRequest},
//        {"101", http.StatusBadRequest},
//        {"abc", http.StatusBadRequest},
//        {"",    http.StatusOK},
//    }

//    for _, tt := range tests {
//        t.Run("limit="+tt.limit, func(t *testing.T) {
//            w := httptest.NewRecorder()
//            c, _ := gin.CreateTestContext(w)
//            c.Request = httptest.NewRequest(http.MethodGet, "/products?limit="+tt.limit, nil)
//            // setar user_id no contexto se necessário
//            c.Set("user_id", 1)

//            controller.GetProducts(c)

//            if w.Code != tt.expectedStatus {
//                t.Errorf("limit=%s: expected %d, got %d", tt.limit, tt.expectedStatus, w.Code)
//            }
//        })
//    }
//}
