//go:build integration

package controller

import (
	"api/db"
	"api/models"
	"api/repository"
	"api/usecase"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCheckoutHTTPRejectsPreorderWithoutSideEffects(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(`
		INSERT INTO users (email, password)
		VALUES ($1, $2) RETURNING id
	`, "preorder-http@example.invalid", "test-password").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(`
		INSERT INTO products (user_id, name, price, stock, sales_type)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, userID, "Camiseta Volurya", 50.00, 10,
		models.SalesTypePreorder).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	paymentUC := usecase.NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	pc := NewPaymentController(paymentUC)
	fake := &fakeCheckoutStripe{}
	pc.paymentIntents = fake

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/checkout", func(c *gin.Context) {
		c.Set("user_id", userID)
		pc.Checkout(c)
	})

	body := `{"idempotency_key":"preorder-http-test-001",` +
		`"user_email":"preorder-http@example.invalid",` +
		`"items":[{"product_id":` + strconv.Itoa(productID) +
		`,"quantity":1}]}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/checkout",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"HTTP recebido=%d, esperado=409, resposta=%s",
			rec.Code, rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), "preorder") {
		t.Fatalf("resposta não identifica a pré-venda: %s", rec.Body.String())
	}

	var orders, payments, stock int

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM payment_records",
	).Scan(&payments); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if orders != 0 || payments != 0 || stock != 10 {
		t.Fatalf(
			"estado indevido: pedidos=%d pagamentos=%d estoque=%d",
			orders, payments, stock,
		)
	}

	if fake.creates != 0 || fake.gets != 0 {
		t.Fatalf(
			"Stripe simulado foi acionado: creates=%d gets=%d",
			fake.creates, fake.gets,
		)
	}
}
