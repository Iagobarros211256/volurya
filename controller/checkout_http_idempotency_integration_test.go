//go:build integration

package controller

import (
	"api/db"
	"api/models"
	"api/repository"
	"api/usecase"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
)

type fakeCheckoutStripe struct {
	creates int
	gets    int
}

func (f *fakeCheckoutStripe) Create(
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.creates++

	if params == nil || params.IdempotencyKey == nil ||
		*params.IdempotencyKey == "" {
		return nil, errors.New("Stripe sem chave idempotente")
	}

	return &stripe.PaymentIntent{
		ID:           "pi_volurya_fake_checkout",
		ClientSecret: "pi_volurya_fake_checkout_secret_test",
	}, nil
}

func (f *fakeCheckoutStripe) Get(
	id string, _ *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.gets++

	if id != "pi_volurya_fake_checkout" {
		return nil, errors.New("PaymentIntent desconhecido")
	}

	return &stripe.PaymentIntent{
		ID:           id,
		ClientSecret: "pi_volurya_fake_checkout_secret_test",
	}, nil
}

func TestCheckoutHTTPReusesOrderAndPaymentIntent(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"http-idempotency@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Produto normal", 50.00, 5,
	).Scan(&productID)
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

	body := `{"idempotency_key":"http-retry-001",` +
		`"user_email":"http-idempotency@example.invalid",` +
		`"items":[{"product_id":` +
		strconv.Itoa(productID) +
		`,"quantity":2}]}`

	var responses [2]models.CheckoutResponse

	for i, expectedStatus := range []int{
		http.StatusCreated,
		http.StatusOK,
	} {
		attemptBody := body
		if i == 1 {
			attemptBody = strings.Replace(
				body,
				"http-idempotency@example.invalid",
				"HTTP-IDEMPOTENCY@EXAMPLE.INVALID",
				1,
			)
		}

		req := httptest.NewRequest(
			http.MethodPost, "/api/checkout",
			strings.NewReader(attemptBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != expectedStatus {
			t.Fatalf(
				"tentativa %d: HTTP %d, esperado %d: %s",
				i+1, rec.Code, expectedStatus, rec.Body.String(),
			)
		}

		if err := json.Unmarshal(
			rec.Body.Bytes(), &responses[i],
		); err != nil {
			t.Fatal(err)
		}
	}

	// A mesma chave não pode representar um carrinho diferente.
	conflictBody := strings.Replace(
		body, `"quantity":2`, `"quantity":1`, 1,
	)
	if conflictBody == body {
		t.Fatal("não foi possível montar o carrinho divergente")
	}

	conflictReq := httptest.NewRequest(
		http.MethodPost, "/api/checkout",
		strings.NewReader(conflictBody),
	)
	conflictReq.Header.Set("Content-Type", "application/json")

	conflictRec := httptest.NewRecorder()
	router.ServeHTTP(conflictRec, conflictReq)

	if conflictRec.Code != http.StatusConflict {
		t.Fatalf(
			"chave reutilizada com carrinho diferente: HTTP %d, esperado 409: %s",
			conflictRec.Code, conflictRec.Body.String(),
		)
	}

	if fake.creates != 1 || fake.gets != 1 {
		t.Fatalf(
			"conflito acionou o Stripe simulado: creates=%d, gets=%d",
			fake.creates, fake.gets,
		)
	}

	if responses[0].OrderID != responses[1].OrderID ||
		responses[0].PaymentIntentID != responses[1].PaymentIntentID {
		t.Fatal("repetição gerou pedido ou pagamento diferente")
	}

	if fake.creates != 1 || fake.gets != 1 {
		t.Fatalf(
			"Stripe simulado: creates=%d, gets=%d",
			fake.creates, fake.gets,
		)
	}

	var orders, payments, stock int

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		`SELECT COUNT(*) FROM payment_records
		 WHERE order_id = $1`,
		responses[0].OrderID,
	).Scan(&payments); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if orders != 1 || payments != 1 || stock != 3 {
		t.Fatalf(
			"pedidos=%d pagamentos=%d estoque=%d",
			orders, payments, stock,
		)
	}
}

// flakyCheckoutStripe simula uma falha temporária na primeira tentativa.
type flakyCheckoutStripe struct {
	*fakeCheckoutStripe
	attempts int
}

func (f *flakyCheckoutStripe) Create(
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.attempts++

	if f.attempts == 1 {
		return nil, errors.New("falha de rede simulada")
	}

	return f.fakeCheckoutStripe.Create(params)
}

func TestCheckoutHTTPRecoversFromTransientStripeFailure(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"checkout-recovery@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, 50, 5) RETURNING id`,
		userID, "Produto normal",
	).Scan(&productID)
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
	fake := &flakyCheckoutStripe{
		fakeCheckoutStripe: &fakeCheckoutStripe{},
	}
	pc.paymentIntents = fake

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/checkout", func(c *gin.Context) {
		c.Set("user_id", userID)
		pc.Checkout(c)
	})

	body := `{"idempotency_key":"recovery-test-001",` +
		`"user_email":"checkout-recovery@example.invalid",` +
		`"items":[{"product_id":` +
		strconv.Itoa(productID) +
		`,"quantity":2}]}`

	var originalOrderID, originalPaymentIntent string

	for attempt, expectedStatus := range []int{
		http.StatusBadGateway,
		http.StatusOK,
		http.StatusOK,
	} {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/checkout",
			strings.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != expectedStatus {
			t.Fatalf(
				"tentativa %d: HTTP %d, esperado %d: %s",
				attempt+1, rec.Code, expectedStatus,
				rec.Body.String(),
			)
		}

		if attempt == 0 {
			continue
		}

		var response models.CheckoutResponse
		if err := json.Unmarshal(
			rec.Body.Bytes(), &response,
		); err != nil {
			t.Fatal(err)
		}

		if attempt == 1 {
			originalOrderID = strconv.Itoa(response.OrderID)
			originalPaymentIntent = response.PaymentIntentID
		} else if originalOrderID != strconv.Itoa(response.OrderID) ||
			originalPaymentIntent != response.PaymentIntentID {
			t.Fatal("recuperação retornou pedido ou pagamento diferente")
		}
	}

	if fake.attempts != 2 ||
		fake.creates != 1 ||
		fake.gets != 1 {
		t.Fatalf(
			"gateway: tentativas=%d, criações=%d, consultas=%d",
			fake.attempts, fake.creates, fake.gets,
		)
	}

	var orders, payments, stock int

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		`SELECT COUNT(*) FROM payment_records
		 WHERE order_id IN (
		   SELECT id FROM orders WHERE user_id = $1
		 )`,
		userID,
	).Scan(&payments); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if orders != 1 || payments != 1 || stock != 3 {
		t.Fatalf(
			"estado inconsistente: pedidos=%d, pagamentos=%d, estoque=%d",
			orders, payments, stock,
		)
	}
}

// Simula um pagamento aceito pelo Stripe cuja resposta se perdeu.
type lostResponseStripe struct {
	key       string
	intent    *stripe.PaymentIntent
	creates   int
	gets      int
	realCount int
}

func (f *lostResponseStripe) Create(
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.creates++

	if params == nil || params.IdempotencyKey == nil ||
		*params.IdempotencyKey == "" {
		return nil, errors.New("chave idempotente ausente")
	}

	key := *params.IdempotencyKey

	if f.intent == nil {
		f.key = key
		f.realCount++
		f.intent = &stripe.PaymentIntent{
			ID:           "pi_accepted_but_response_lost",
			ClientSecret: "pi_accepted_but_response_lost_secret_test",
		}

		// O provedor aceitou, mas a conexão caiu.
		return nil, errors.New("resposta perdida após criação")
	}

	if key != f.key {
		return nil, errors.New("tentativa usou chave Stripe diferente")
	}

	return f.intent, nil
}

func (f *lostResponseStripe) Get(
	id string, _ *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.gets++

	if f.intent == nil || id != f.intent.ID {
		return nil, errors.New("pagamento desconhecido")
	}

	return f.intent, nil
}

func TestCheckoutRecoversWhenStripeResponseIsLost(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"lost-response@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, 50, 5) RETURNING id`,
		userID, "Produto normal",
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := usecase.NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	gateway := &lostResponseStripe{}
	pc := NewPaymentController(uc)
	pc.paymentIntents = gateway

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/checkout", func(c *gin.Context) {
		c.Set("user_id", userID)
		pc.Checkout(c)
	})

	body := `{"idempotency_key":"lost-response-001",` +
		`"user_email":"lost-response@example.invalid",` +
		`"items":[{"product_id":` +
		strconv.Itoa(productID) +
		`,"quantity":2}]}`

	var successfulOrderID int
	var successfulIntentID string

	for attempt, expected := range []int{
		http.StatusBadGateway,
		http.StatusOK,
		http.StatusOK,
	} {
		req := httptest.NewRequest(
			http.MethodPost, "/api/checkout",
			strings.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != expected {
			t.Fatalf(
				"tentativa %d: HTTP %d, esperado %d: %s",
				attempt+1, rec.Code, expected, rec.Body.String(),
			)
		}

		if attempt == 0 {
			// O primeiro erro não deve cancelar o pedido.
			var orders, payments, stock int

			if err := conn.QueryRow(
				"SELECT COUNT(*) FROM orders WHERE user_id=$1",
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
				"SELECT stock FROM products WHERE id=$1",
				productID,
			).Scan(&stock); err != nil {
				t.Fatal(err)
			}

			if orders != 1 || payments != 0 || stock != 3 {
				t.Fatalf(
					"estado após falha: pedidos=%d pagamentos=%d estoque=%d",
					orders, payments, stock,
				)
			}
			continue
		}

		var response models.CheckoutResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		if attempt == 1 {
			successfulOrderID = response.OrderID
			successfulIntentID = response.PaymentIntentID
		} else if successfulOrderID != response.OrderID ||
			successfulIntentID != response.PaymentIntentID {
			t.Fatal("repetição retornou pedido ou pagamento diferente")
		}
	}

	var orders, payments, stock int

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id=$1",
		userID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM payment_records WHERE order_id=$1",
		successfulOrderID,
	).Scan(&payments); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id=$1",
		productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if gateway.creates != 2 || gateway.gets != 1 ||
		gateway.realCount != 1 {
		t.Fatalf(
			"gateway: chamadas create=%d, get=%d, pagamentos criados=%d",
			gateway.creates, gateway.gets, gateway.realCount,
		)
	}

	if orders != 1 || payments != 1 || stock != 3 {
		t.Fatalf(
			"estado final: pedidos=%d pagamentos=%d estoque=%d",
			orders, payments, stock,
		)
	}
}

type concurrentCheckoutStripe struct {
	mu      sync.Mutex
	intents map[string]*stripe.PaymentIntent
	created int
}

func (f *concurrentCheckoutStripe) Create(
	params *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	if params == nil || params.IdempotencyKey == nil ||
		*params.IdempotencyKey == "" {
		return nil, errors.New("chave Stripe ausente")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.intents == nil {
		f.intents = make(map[string]*stripe.PaymentIntent)
	}

	key := *params.IdempotencyKey
	if existing := f.intents[key]; existing != nil {
		return existing, nil
	}

	intent := &stripe.PaymentIntent{
		ID:           "pi_concurrent_checkout_test",
		ClientSecret: "pi_concurrent_checkout_secret_test",
	}
	f.intents[key] = intent
	f.created++
	return intent, nil
}

func (f *concurrentCheckoutStripe) Get(
	id string, _ *stripe.PaymentIntentParams,
) (*stripe.PaymentIntent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, intent := range f.intents {
		if intent.ID == id {
			return intent, nil
		}
	}
	return nil, errors.New("PaymentIntent inexistente")
}

func TestCheckoutHTTPConcurrentRequestsShareOrderAndPaymentIntent(
	t *testing.T,
) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"concurrent-http@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, 50, 5) RETURNING id`,
		userID, "Produto normal",
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := usecase.NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	gateway := &concurrentCheckoutStripe{}
	pc := NewPaymentController(uc)
	pc.paymentIntents = gateway

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/checkout", func(c *gin.Context) {
		c.Set("user_id", userID)
		pc.Checkout(c)
	})

	body := `{"idempotency_key":"concurrent-http-test-001",` +
		`"user_email":"concurrent-http@example.invalid",` +
		`"items":[{"product_id":` +
		strconv.Itoa(productID) + `,"quantity":2}]}`

	type result struct {
		status int
		body   string
	}

	start := make(chan struct{})
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/checkout",
				strings.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			results <- result{
				status: rec.Code,
				body:   rec.Body.String(),
			}
		}()
	}

	close(start)

	first := <-results
	second := <-results

	statuses := map[int]int{
		first.status:  0,
		second.status: 0,
	}
	statuses[first.status]++
	statuses[second.status]++

	if statuses[http.StatusCreated] != 1 ||
		statuses[http.StatusOK] != 1 {
		t.Fatalf(
			"respostas inesperadas: %d (%s), %d (%s)",
			first.status, first.body,
			second.status, second.body,
		)
	}

	var a, b models.CheckoutResponse

	if err := json.Unmarshal([]byte(first.body), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(second.body), &b); err != nil {
		t.Fatal(err)
	}

	if a.OrderID != b.OrderID ||
		a.PaymentIntentID != b.PaymentIntentID {
		t.Fatal("requisições retornaram pedidos ou pagamentos diferentes")
	}

	var orders, payments, stock int

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id=$1",
		userID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM payment_records WHERE order_id=$1",
		a.OrderID,
	).Scan(&payments); err != nil {
		t.Fatal(err)
	}

	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id=$1",
		productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if orders != 1 || payments != 1 ||
		stock != 3 || gateway.created != 1 {
		t.Fatalf(
			"pedidos=%d pagamentos=%d estoque=%d pagamentos simulados=%d",
			orders, payments, stock, gateway.created,
		)
	}
}

// Simula falha do PostgreSQL após o Stripe aceitar o PaymentIntent.
func TestCheckoutHTTPRecoversAfterPaymentRegistrationFailure(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"registration-failure@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, 50, 5) RETURNING id`,
		userID, "Produto normal",
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := usecase.NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	gateway := &concurrentCheckoutStripe{}
	pc := NewPaymentController(uc)
	pc.paymentIntents = gateway

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/checkout", func(c *gin.Context) {
		c.Set("user_id", userID)
		pc.Checkout(c)
	})

	body := `{"idempotency_key":"registration-failure-001",` +
		`"user_email":"registration-failure@example.invalid",` +
		`"items":[{"product_id":` +
		strconv.Itoa(productID) + `,"quantity":2}]}`

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(
			http.MethodPost, "/api/checkout",
			strings.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	const constraint = "test_checkout_reject_payment_registration"

	// A restrição é aplicada somente ao PostgreSQL isolado de testes.
	_, err = conn.Exec(
		`ALTER TABLE payment_records
		 ADD CONSTRAINT ` + constraint + `
		 CHECK (FALSE) NOT VALID`,
	)
	if err != nil {
		t.Fatal(err)
	}

	removeConstraint := func() error {
		_, err := conn.Exec(
			`ALTER TABLE payment_records
			 DROP CONSTRAINT IF EXISTS ` + constraint,
		)
		return err
	}

	// Garante limpeza mesmo se o teste falhar.
	t.Cleanup(func() {
		if err := removeConstraint(); err != nil {
			t.Errorf("remover restrição de teste: %v", err)
		}
	})

	first := send()
	if first.Code != http.StatusInternalServerError {
		t.Fatalf(
			"primeira tentativa: HTTP %d, esperado 500: %s",
			first.Code, first.Body.String(),
		)
	}

	var pending, payments, stock int

	err = conn.QueryRow(
		`SELECT COUNT(*) FROM orders
		 WHERE user_id = $1
		   AND status = 'pending'
		   AND payment_intent_id IS NULL`,
		userID,
	).Scan(&pending)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`SELECT COUNT(*) FROM payment_records
		 WHERE order_id IN (
		   SELECT id FROM orders WHERE user_id = $1
		 )`,
		userID,
	).Scan(&payments)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if pending != 1 || payments != 0 || stock != 3 ||
		gateway.created != 1 {
		t.Fatalf(
			"falha não preservou estado: pendentes=%d pagamentos=%d estoque=%d Stripe=%d",
			pending, payments, stock, gateway.created,
		)
	}

	// Recupera a persistência e repete a mesma tentativa.
	if err := removeConstraint(); err != nil {
		t.Fatal(err)
	}

	second := send()
	third := send()

	if second.Code != http.StatusOK ||
		third.Code != http.StatusOK {
		t.Fatalf(
			"recuperação: HTTP %d (%s), repetição: HTTP %d (%s)",
			second.Code, second.Body.String(),
			third.Code, third.Body.String(),
		)
	}

	var a, b models.CheckoutResponse

	if err := json.Unmarshal(second.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(third.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}

	if a.OrderID != b.OrderID ||
		a.PaymentIntentID != b.PaymentIntentID {
		t.Fatal("pedido ou pagamento mudou após recuperação")
	}

	var orders int

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orders)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM payment_records WHERE order_id = $1",
		a.OrderID,
	).Scan(&payments)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if orders != 1 || payments != 1 ||
		stock != 3 || gateway.created != 1 {
		t.Fatalf(
			"estado final: pedidos=%d pagamentos=%d estoque=%d Stripe=%d",
			orders, payments, stock, gateway.created,
		)
	}
}
