//go:build integration

package usecase

import (
	"api/db"
	"api/models"
	"api/repository"
	"errors"
	"testing"
)

func TestCheckoutRollbackWhenRepeatedProductExceedsStock(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		"INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id",
		"checkout-rollback@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"INSERT INTO products (user_id, name, price, stock) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, "Produto de teste", 19.90, 1,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	paymentUC := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	_, err = paymentUC.CreateOrderForCheckout(userID, []models.CheckoutItemInput{
		{ProductID: productID, Quantity: 1},
		{ProductID: productID, Quantity: 1},
	})
	if !errors.Is(err, repository.ErrCheckoutInsufficientStock) {
		t.Fatalf(
			"esperado ErrCheckoutInsufficientStock, recebido: %v",
			err,
		)
	}

	var stock int
	if err := conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1", productID,
	).Scan(&stock); err != nil {
		t.Fatal(err)
	}

	if stock != 1 {
		t.Errorf("estoque inconsistente após falha: esperado 1, obtido %d", stock)
	}

	// O rollback também deve remover o pedido e seus itens.
	var orderCount int
	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orderCount)
	if err != nil {
		t.Fatal(err)
	}
	if orderCount != 0 {
		t.Errorf("pedido persistido após rollback: %d", orderCount)
	}

	var itemCount int
	err = conn.QueryRow(
		"SELECT COUNT(*) FROM order_items WHERE product_id = $1",
		productID,
	).Scan(&itemCount)
	if err != nil {
		t.Fatal(err)
	}
	if itemCount != 0 {
		t.Errorf("itens persistidos após rollback: %d", itemCount)
	}
}

// Testa atomicidade quando o PostgreSQL rejeita o UPDATE do pedido.
func TestPaymentSuccessRollbackWhenOrderUpdateFails(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, orderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"payment-atomicity@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO orders (user_id, total, status)
		 VALUES ($1, 19.90, 'pending') RETURNING id`,
		userID,
	).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	payment, err := uc.CreatePaymentRecord(
		orderID, "pi_atomicity_test", 1990, "BRL", "",
	)
	if err != nil {
		t.Fatal(err)
	}

	// Provoca falha especificamente ao marcar o pedido como pago.
	_, err = conn.Exec(
		`ALTER TABLE orders ADD CONSTRAINT
		 test_payment_atomicity_prevent_paid
		 CHECK (status <> 'paid')`,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := conn.Exec(
			`ALTER TABLE orders DROP CONSTRAINT
			 IF EXISTS test_payment_atomicity_prevent_paid`,
		)
		if err != nil {
			t.Errorf("remover restrição de teste: %v", err)
		}
	})

	err = uc.HandlePaymentSuccess("pi_atomicity_test")
	if err == nil {
		t.Fatal("esperava erro ao atualizar pedido")
	}

	var paymentStatus, orderStatus string

	err = conn.QueryRow(
		"SELECT status FROM payment_records WHERE id = $1",
		payment.ID,
	).Scan(&paymentStatus)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		orderID,
	).Scan(&orderStatus)
	if err != nil {
		t.Fatal(err)
	}

	if paymentStatus != "requires_payment_method" {
		t.Errorf("pagamento alterado sem commit: %s", paymentStatus)
	}

	if orderStatus != "pending" {
		t.Errorf("pedido inconsistente: %s", orderStatus)
	}
}

func TestConfirmPaymentSucceededIdempotent(t *testing.T) {
	conn := db.SetupTestDB(t)

	const intentID = "pi_idempotency_integration_test"

	var userID, orderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"idempotency@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO orders
		 (user_id, total, status, payment_intent_id)
		 VALUES ($1, 19.90, 'pending', $2)
		 RETURNING id`,
		userID, intentID,
	).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}

	paymentRepo := repository.NewPaymentRepository(conn)

	_, err = paymentRepo.CreatePaymentRecord(models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: intentID,
		Amount:          1990,
		Currency:        "BRL",
		Status:          models.PaymentStatusRequiresPaymentMethod,
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := paymentRepo.ConfirmPaymentSucceeded(intentID)
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("primeira confirmação deveria atualizar os registros")
	}

	second, err := paymentRepo.ConfirmPaymentSucceeded(intentID)
	if err != nil {
		t.Fatal(err)
	}
	if second {
		t.Fatal("segunda confirmação deveria ser idempotente")
	}

	var paymentStatus, orderStatus, orderPaymentStatus string

	err = conn.QueryRow(
		`SELECT p.status, o.status, o.payment_status
		 FROM payment_records p
		 JOIN orders o ON o.id = p.order_id
		 WHERE p.payment_intent_id = $1`,
		intentID,
	).Scan(&paymentStatus, &orderStatus, &orderPaymentStatus)
	if err != nil {
		t.Fatal(err)
	}

	if paymentStatus != "succeeded" {
		t.Errorf("status do pagamento: %s", paymentStatus)
	}
	if orderStatus != "paid" {
		t.Errorf("status do pedido: %s", orderStatus)
	}
	if orderPaymentStatus != "succeeded" {
		t.Errorf("status financeiro do pedido: %s", orderPaymentStatus)
	}
}

func TestLatePaymentFailedDoesNotRegressSucceeded(t *testing.T) {
	conn := db.SetupTestDB(t)

	const intentID = "pi_late_failure_test"

	var userID, orderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"late-failure@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO orders
		 (user_id, total, status, payment_intent_id)
		 VALUES ($1, 19.90, 'pending', $2)
		 RETURNING id`,
		userID, intentID,
	).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}

	paymentRepo := repository.NewPaymentRepository(conn)

	_, err = paymentRepo.CreatePaymentRecord(models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: intentID,
		Amount:          1990,
		Currency:        "BRL",
		Status:          models.PaymentStatusRequiresPaymentMethod,
	})
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		paymentRepo,
		repository.NewProductRepository(conn),
		nil,
	)

	if err := uc.HandlePaymentSuccess(intentID); err != nil {
		t.Fatal(err)
	}

	if err := uc.HandlePaymentFailed(intentID, "old failure event"); err != nil {
		t.Fatal(err)
	}

	var paymentStatus, orderStatus, orderPaymentStatus string
	var errorMessage string

	err = conn.QueryRow(
		`SELECT p.status, o.status, o.payment_status,
		        COALESCE(p.error_message, '')
		 FROM payment_records p
		 JOIN orders o ON o.id = p.order_id
		 WHERE p.payment_intent_id = $1`,
		intentID,
	).Scan(
		&paymentStatus,
		&orderStatus,
		&orderPaymentStatus,
		&errorMessage,
	)
	if err != nil {
		t.Fatal(err)
	}

	if paymentStatus != "succeeded" {
		t.Errorf("pagamento regrediu para %q", paymentStatus)
	}
	if orderStatus != "paid" {
		t.Errorf("pedido regrediu para %q", orderStatus)
	}
	if orderPaymentStatus != "succeeded" {
		t.Errorf("status financeiro inesperado: %q", orderPaymentStatus)
	}
	if errorMessage != "" {
		t.Errorf("evento atrasado alterou erro: %q", errorMessage)
	}
}

func TestPaymentFailureRollbackWhenOrderUpdateFails(t *testing.T) {
	conn := db.SetupTestDB(t)

	const intentID = "pi_failure_atomicity_test"

	var userID, orderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"failure-atomicity@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO orders
		 (user_id, total, status, payment_intent_id)
		 VALUES ($1, 19.90, 'pending', $2) RETURNING id`,
		userID, intentID,
	).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}

	paymentRepo := repository.NewPaymentRepository(conn)

	_, err = paymentRepo.CreatePaymentRecord(models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: intentID,
		Amount:          1990,
		Currency:        "BRL",
		Status:          models.PaymentStatusRequiresPaymentMethod,
	})
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		paymentRepo,
		repository.NewProductRepository(conn),
		nil,
	)

	// Impede a atualização financeira do pedido.
	_, err = conn.Exec(
		`ALTER TABLE orders ADD CONSTRAINT
		 test_prevent_failed_payment_status
		 CHECK (payment_status <> 'failed')`,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := conn.Exec(
			`ALTER TABLE orders DROP CONSTRAINT
			 IF EXISTS test_prevent_failed_payment_status`,
		)
		if err != nil {
			t.Errorf("remover restrição de teste: %v", err)
		}
	})

	err = uc.HandlePaymentFailed(intentID, "card declined")
	if err == nil {
		t.Fatal("esperava falha ao atualizar o pedido")
	}

	var paymentStatus, orderStatus, orderPaymentStatus string
	var errorMessage string

	err = conn.QueryRow(
		`SELECT p.status, o.status, o.payment_status,
		        COALESCE(p.error_message, '')
		 FROM payment_records p
		 JOIN orders o ON o.id = p.order_id
		 WHERE p.payment_intent_id = $1`,
		intentID,
	).Scan(
		&paymentStatus,
		&orderStatus,
		&orderPaymentStatus,
		&errorMessage,
	)
	if err != nil {
		t.Fatal(err)
	}

	if paymentStatus != "requires_payment_method" {
		t.Errorf("pagamento alterado sem commit: %q", paymentStatus)
	}
	if orderStatus != "pending" {
		t.Errorf("pedido inconsistente: %q", orderStatus)
	}
	if orderPaymentStatus != "pending" {
		t.Errorf("estado financeiro inconsistente: %q", orderPaymentStatus)
	}
	if errorMessage != "" {
		t.Errorf("mensagem de erro persistida após rollback: %q", errorMessage)
	}
}

func TestConcurrentPaymentSuccessAndFailure(t *testing.T) {
	conn := db.SetupTestDB(t)

	const intentID = "pi_concurrent_events_test"

	var userID, orderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"concurrent-events@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO orders
		 (user_id, total, status, payment_intent_id)
		 VALUES ($1, 19.90, 'pending', $2)
		 RETURNING id`,
		userID, intentID,
	).Scan(&orderID)
	if err != nil {
		t.Fatal(err)
	}

	paymentRepo := repository.NewPaymentRepository(conn)

	_, err = paymentRepo.CreatePaymentRecord(models.PaymentRecord{
		OrderID:         orderID,
		PaymentIntentID: intentID,
		Amount:          1990,
		Currency:        "BRL",
		Status:          models.PaymentStatusRequiresPaymentMethod,
	})
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		paymentRepo,
		repository.NewProductRepository(conn),
		nil,
	)

	start := make(chan struct{})
	results := make(chan error, 2)

	go func() {
		<-start
		results <- uc.HandlePaymentSuccess(intentID)
	}()

	go func() {
		<-start
		results <- uc.HandlePaymentFailed(intentID, "concurrent failure")
	}()

	close(start)

	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Errorf("erro ao processar webhook concorrente: %v", err)
		}
	}

	var paymentStatus, orderStatus, orderPaymentStatus string
	var errorMessage string

	err = conn.QueryRow(
		`SELECT p.status, o.status, o.payment_status,
		        COALESCE(p.error_message, '')
		 FROM payment_records p
		 JOIN orders o ON o.id = p.order_id
		 WHERE p.payment_intent_id = $1`,
		intentID,
	).Scan(
		&paymentStatus,
		&orderStatus,
		&orderPaymentStatus,
		&errorMessage,
	)
	if err != nil {
		t.Fatal(err)
	}

	if paymentStatus != "succeeded" {
		t.Errorf("pagamento terminou em %q", paymentStatus)
	}
	if orderStatus != "paid" {
		t.Errorf("pedido terminou em %q", orderStatus)
	}
	if orderPaymentStatus != "succeeded" {
		t.Errorf("estado financeiro terminou em %q", orderPaymentStatus)
	}
	if errorMessage != "" {
		t.Errorf("erro antigo persistiu: %q", errorMessage)
	}
}

func TestCancelOrderRestoresStockExactlyOnce(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"cancel-stock@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	checkStock := func(stage string, expected int) {
		t.Helper()

		var stock int
		err := conn.QueryRow(
			"SELECT stock FROM products WHERE id = $1",
			productID,
		).Scan(&stock)
		if err != nil {
			t.Fatal(err)
		}

		if stock != expected {
			t.Errorf(
				"%s: estoque esperado=%d, obtido=%d",
				stage, expected, stock,
			)
		}
	}

	checkStock("após checkout", 3)

	if err := uc.CancelOrder(order.ID); err != nil {
		t.Fatal(err)
	}

	checkStock("após primeiro cancelamento", 5)

	// O segundo cancelamento não pode devolver estoque novamente.
	if err := uc.CancelOrder(order.ID); err != nil {
		t.Fatal(err)
	}

	checkStock("após segundo cancelamento", 5)

	var status string
	err = conn.QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		order.ID,
	).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "cancelled" {
		t.Errorf("status esperado=cancelled, obtido=%s", status)
	}
}

func TestCancelOrderRejectsPaidOrder(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"paid-order-cancel@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Simula um pedido cujo pagamento já foi confirmado.
	_, err = conn.Exec(
		"UPDATE orders SET status = 'paid' WHERE id = $1",
		order.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	// O cancelamento deve ser recusado.
	if err := uc.CancelOrder(order.ID); err == nil {
		t.Fatal("cancelamento de pedido pago deveria ser recusado")
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf(
			"estoque de pedido pago foi alterado: esperado=3, obtido=%d",
			stock,
		)
	}

	var status string
	err = conn.QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		order.ID,
	).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "paid" {
		t.Errorf("pedido pago foi alterado para %q", status)
	}
}

func TestCancelOrderRollbackWhenStatusUpdateFails(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"cancel-rollback@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Impede o UPDATE final do pedido, após a tentativa
	// de devolução de estoque dentro da transação.
	_, err = conn.Exec(`
		ALTER TABLE orders
		ADD CONSTRAINT test_prevent_cancel_status
		CHECK (status <> 'cancelled')
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := conn.Exec(`
			ALTER TABLE orders
			DROP CONSTRAINT IF EXISTS test_prevent_cancel_status
		`)
		if err != nil {
			t.Errorf("remover restrição de teste: %v", err)
		}
	})

	if err := uc.CancelOrder(order.ID); err == nil {
		t.Fatal("esperava erro ao atualizar status do pedido")
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf(
			"rollback do estoque falhou: esperado=3, obtido=%d",
			stock,
		)
	}

	var status string
	err = conn.QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		order.ID,
	).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "pending" {
		t.Errorf(
			"pedido alterado apesar do rollback: %q",
			status,
		)
	}
}

func TestConcurrentCancelOrderRestoresStockOnce(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"concurrent-cancel@example.invalid",
		"test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start
			results <- uc.CancelOrder(order.ID)
		}()
	}

	close(start)

	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Errorf("cancelamento concorrente: %v", err)
		}
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 5 {
		t.Errorf(
			"estoque esperado=5, obtido=%d",
			stock,
		)
	}

	var status string
	err = conn.QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		order.ID,
	).Scan(&status)
	if err != nil {
		t.Fatal(err)
	}

	if status != "cancelled" {
		t.Errorf(
			"status esperado=cancelled, obtido=%q",
			status,
		)
	}
}

func TestRegisterPaymentIntentRollbackWhenRecordInsertFails(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"atomic-payment@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Simula uma falha na inserção do registro financeiro.
	_, err = conn.Exec(`
		ALTER TABLE payment_records
		ADD CONSTRAINT test_reject_atomic_payment
		CHECK (payment_intent_id <> 'pi_atomic_rollback_test')
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := conn.Exec(`
			ALTER TABLE payment_records
			DROP CONSTRAINT IF EXISTS test_reject_atomic_payment
		`)
		if err != nil {
			t.Errorf("remover restrição de teste: %v", err)
		}
	})

	// Este método será implementado na próxima etapa.
	err = uc.RegisterPaymentIntentAtomically(
		order.ID,
		"pi_atomic_rollback_test",
		15980,
		"BRL",
		"atomic-payment@example.invalid",
	)
	if err == nil {
		t.Fatal("esperava erro ao inserir registro financeiro")
	}

	var intentID, paymentStatus string

	err = conn.QueryRow(
		`SELECT COALESCE(payment_intent_id, ''),
		        COALESCE(payment_status, '')
		 FROM orders WHERE id = $1`,
		order.ID,
	).Scan(&intentID, &paymentStatus)
	if err != nil {
		t.Fatal(err)
	}

	if intentID != "" {
		t.Errorf(
			"PaymentIntent persistido após rollback: %q",
			intentID,
		)
	}

	if paymentStatus != "pending" {
		t.Errorf(
			"status financeiro alterado após rollback: %q",
			paymentStatus,
		)
	}

	var recordCount int
	err = conn.QueryRow(
		"SELECT COUNT(*) FROM payment_records WHERE order_id = $1",
		order.ID,
	).Scan(&recordCount)
	if err != nil {
		t.Fatal(err)
	}

	if recordCount != 0 {
		t.Errorf(
			"registros financeiros persistidos: %d",
			recordCount,
		)
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf(
			"estoque alterado indevidamente: esperado=3, obtido=%d",
			stock,
		)
	}
}

func TestRegisterPaymentIntentSuccessAndIdempotency(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"payment-success@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	register := func(intentID string) error {
		return uc.RegisterPaymentIntentAtomically(
			order.ID,
			intentID,
			15980,
			"BRL",
			"payment-success@example.invalid",
		)
	}

	if err := register("pi_registration_success_test"); err != nil {
		t.Fatalf("primeiro registro: %v", err)
	}

	// Repetir a mesma operação não deve duplicar registros.
	if err := register("pi_registration_success_test"); err != nil {
		t.Fatalf("registro idempotente: %v", err)
	}

	// Não permitir substituir o PaymentIntent de um pedido.
	if err := register("pi_different_intent_test"); err == nil {
		t.Fatal("esperava rejeição de PaymentIntent diferente")
	}

	var intentID, paymentStatus string

	err = conn.QueryRow(
		`SELECT payment_intent_id, payment_status
		 FROM orders WHERE id = $1`,
		order.ID,
	).Scan(&intentID, &paymentStatus)
	if err != nil {
		t.Fatal(err)
	}

	if intentID != "pi_registration_success_test" {
		t.Errorf("PaymentIntent incorreto: %q", intentID)
	}
	if paymentStatus != "pending" {
		t.Errorf("status financeiro incorreto: %q", paymentStatus)
	}

	var count, amount int
	var currency, status string

	err = conn.QueryRow(
		`SELECT COUNT(*),
		        COALESCE(MAX(amount), 0),
		        COALESCE(MAX(currency), ''),
		        COALESCE(MAX(status), '')
		 FROM payment_records
		 WHERE order_id = $1`,
		order.ID,
	).Scan(&count, &amount, &currency, &status)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("esperado 1 registro financeiro, obtido %d", count)
	}
	if amount != 15980 {
		t.Errorf("valor esperado=15980 centavos, obtido=%d", amount)
	}
	if currency != "BRL" {
		t.Errorf("moeda incorreta: %q", currency)
	}
	if status != "requires_payment_method" {
		t.Errorf("status do pagamento incorreto: %q", status)
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf("estoque esperado=3, obtido=%d", stock)
	}
}

func TestConcurrentRegisterPaymentIntentSingleRecord(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"concurrent-registration@example.invalid",
		"test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	const intentID = "pi_concurrent_registration_test"

	start := make(chan struct{})
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start
			results <- uc.RegisterPaymentIntentAtomically(
				order.ID,
				intentID,
				15980,
				"BRL",
				"concurrent-registration@example.invalid",
			)
		}()
	}

	close(start)

	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Errorf("registro concorrente: %v", err)
		}
	}

	var storedIntent string
	err = conn.QueryRow(
		`SELECT payment_intent_id FROM orders WHERE id = $1`,
		order.ID,
	).Scan(&storedIntent)
	if err != nil {
		t.Fatal(err)
	}

	if storedIntent != intentID {
		t.Errorf(
			"PaymentIntent esperado=%q, obtido=%q",
			intentID, storedIntent,
		)
	}

	var count int
	err = conn.QueryRow(
		`SELECT COUNT(*) FROM payment_records
		 WHERE order_id = $1 AND payment_intent_id = $2`,
		order.ID, intentID,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf(
			"esperado 1 registro financeiro, obtido %d",
			count,
		)
	}

	var stock int
	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf(
			"estoque esperado=3, obtido=%d",
			stock,
		)
	}
}

func TestConcurrentRegisterDifferentPaymentIntentsRejectsOne(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"conflicting-intents@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	order, err := uc.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	type result struct {
		intentID string
		err      error
	}

	start := make(chan struct{})
	results := make(chan result, 2)

	intents := []string{
		"pi_conflict_first_test",
		"pi_conflict_second_test",
	}

	for _, intentID := range intents {
		go func(id string) {
			<-start
			err := uc.RegisterPaymentIntentAtomically(
				order.ID,
				id,
				15980,
				"BRL",
				"conflicting-intents@example.invalid",
			)
			results <- result{intentID: id, err: err}
		}(intentID)
	}

	close(start)

	successCount := 0
	failureCount := 0
	var winner string

	for i := 0; i < 2; i++ {
		res := <-results
		if res.err == nil {
			successCount++
			winner = res.intentID
		} else {
			failureCount++
			t.Logf(
				"PaymentIntent rejeitado: %s: %v",
				res.intentID,
				res.err,
			)
		}
	}

	if successCount != 1 || failureCount != 1 {
		t.Fatalf(
			"esperado 1 sucesso e 1 rejeição; sucessos=%d, rejeições=%d",
			successCount,
			failureCount,
		)
	}

	var storedIntent, paymentStatus string

	err = conn.QueryRow(
		`SELECT payment_intent_id, payment_status
		 FROM orders WHERE id = $1`,
		order.ID,
	).Scan(&storedIntent, &paymentStatus)
	if err != nil {
		t.Fatal(err)
	}

	if storedIntent != winner {
		t.Errorf(
			"pedido vinculado a %q, vencedor foi %q",
			storedIntent,
			winner,
		)
	}

	if paymentStatus != "pending" {
		t.Errorf("status financeiro inesperado: %q", paymentStatus)
	}

	var count int
	var recordedIntent string

	err = conn.QueryRow(
		`SELECT COUNT(*), COALESCE(MAX(payment_intent_id), '')
		 FROM payment_records WHERE order_id = $1`,
		order.ID,
	).Scan(&count, &recordedIntent)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("esperado 1 registro financeiro, obtido %d", count)
	}

	if recordedIntent != winner {
		t.Errorf(
			"registro financeiro=%q, vencedor=%q",
			recordedIntent,
			winner,
		)
	}

	var stock int

	err = conn.QueryRow(
		"SELECT stock FROM products WHERE id = $1",
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 3 {
		t.Errorf("estoque esperado=3, obtido=%d", stock)
	}
}

func TestCheckoutIdempotencyReusesOrderWithoutSecondStockReservation(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"checkout-replay@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 2,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	items := []models.CheckoutItemInput{
		{ProductID: productID, Quantity: 2},
	}

	const key = "checkout-replay-key-001"
	const email = "checkout-replay@example.invalid"

	first, firstReused, err := uc.CreateOrderForCheckoutIdempotent(
		userID, items, key, email,
	)
	if err != nil {
		t.Fatalf("primeiro checkout: %v", err)
	}
	if firstReused {
		t.Fatal("primeiro checkout marcado como reutilizado")
	}

	second, secondReused, err := uc.CreateOrderForCheckoutIdempotent(
		userID, items, key, email,
	)
	if err != nil {
		t.Fatalf("repetição do checkout: %v", err)
	}
	if !secondReused {
		t.Fatal("segunda chamada deveria reutilizar o pedido")
	}

	if first.ID != second.ID {
		t.Errorf(
			"pedido duplicado: primeiro=%d, segundo=%d",
			first.ID, second.ID,
		)
	}

	var orderCount, itemCount, stock int

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orderCount)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM order_items WHERE order_id = $1",
		first.ID,
	).Scan(&itemCount)
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

	if orderCount != 1 {
		t.Errorf("esperado 1 pedido, obtido %d", orderCount)
	}
	if itemCount != 1 {
		t.Errorf("esperado 1 item, obtido %d", itemCount)
	}
	if stock != 0 {
		t.Errorf("estoque esperado=0, obtido=%d", stock)
	}
}

func TestCheckoutIdempotencyRejectsDifferentCart(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"checkout-conflict@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	const key = "checkout-conflict-key-001"
	const email = "checkout-conflict@example.invalid"

	first, reused, err := uc.CreateOrderForCheckoutIdempotent(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 2},
		},
		key,
		email,
	)
	if err != nil {
		t.Fatalf("primeiro checkout: %v", err)
	}
	if reused {
		t.Fatal("primeiro pedido foi marcado como reutilizado")
	}

	_, _, err = uc.CreateOrderForCheckoutIdempotent(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 3},
		},
		key,
		email,
	)
	if !errors.Is(err, repository.ErrCheckoutIdempotencyConflict) {
		t.Fatalf(
			"esperado conflito de idempotência, recebido: %v",
			err,
		)
	}

	var count, stock, quantity int

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&count)
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

	err = conn.QueryRow(
		`SELECT quantity FROM order_items
		 WHERE order_id = $1 AND product_id = $2`,
		first.ID, productID,
	).Scan(&quantity)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("esperado 1 pedido, obtido %d", count)
	}
	if stock != 3 {
		t.Errorf("estoque esperado=3, obtido=%d", stock)
	}
	if quantity != 2 {
		t.Errorf("quantidade original esperada=2, obtida=%d", quantity)
	}
}

func TestConcurrentIdempotentCheckoutReservesStockOnce(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
		 VALUES ($1, $2) RETURNING id`,
		"concurrent-idempotency@example.invalid",
		"test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products (user_id, name, price, stock)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Camiseta Volurya", 79.90, 5,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	type result struct {
		orderID int
		reused  bool
		err     error
	}

	const key = "checkout-concurrent-key-001"
	const email = "concurrent-idempotency@example.invalid"

	start := make(chan struct{})
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func() {
			<-start

			order, reused, err := uc.CreateOrderForCheckoutIdempotent(
				userID,
				[]models.CheckoutItemInput{
					{ProductID: productID, Quantity: 2},
				},
				key,
				email,
			)

			res := result{reused: reused, err: err}
			if order != nil {
				res.orderID = order.ID
			}
			results <- res
		}()
	}

	close(start)

	first := <-results
	second := <-results

	if first.err != nil {
		t.Fatalf("primeira resposta: %v", first.err)
	}
	if second.err != nil {
		t.Fatalf("segunda resposta: %v", second.err)
	}

	if first.orderID <= 0 || first.orderID != second.orderID {
		t.Fatalf(
			"pedidos diferentes: %d e %d",
			first.orderID,
			second.orderID,
		)
	}

	if first.reused == second.reused {
		t.Errorf(
			"esperado um pedido novo e um reutilizado: %v, %v",
			first.reused,
			second.reused,
		)
	}

	var orderCount, itemCount, stock int

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orderCount)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM order_items WHERE order_id = $1",
		first.orderID,
	).Scan(&itemCount)
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

	if orderCount != 1 {
		t.Errorf("esperado 1 pedido, obtido %d", orderCount)
	}
	if itemCount != 1 {
		t.Errorf("esperado 1 item, obtido %d", itemCount)
	}
	if stock != 3 {
		t.Errorf("estoque esperado=3, obtido=%d", stock)
	}
}
