//go:build integration

package usecase

import (
	"api/db"
	"api/models"
	"api/repository"
	"errors"
	"testing"
)

func TestPreorderWithPositiveStockCannotEnterCartOrCheckout(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, productID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
         VALUES ($1, $2) RETURNING id`,
		"preorder-checkout@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`INSERT INTO products
         (user_id, name, price, stock, sales_type)
         VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, "Camiseta Volurya", 50.00, 10,
		models.SalesTypePreorder,
	).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	productRepo := repository.NewProductRepository(conn)

	cartUC := NewCartUsecase(
		repository.NewCartRepository(conn), productRepo,
	)

	_, err = cartUC.AddItem(userID, productID, 1)
	if !errors.Is(err, ErrPreorderCheckout) {
		t.Fatalf("carrinho deveria rejeitar pré-venda: %v", err)
	}

	paymentUC := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		productRepo,
		nil,
	)

	_, err = paymentUC.CreateOrderForCheckout(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: productID, Quantity: 1},
		},
	)
	if !errors.Is(err, ErrPreorderCheckout) {
		t.Fatalf("checkout deveria rejeitar pré-venda: %v", err)
	}

	var orders, cartItems, stock int

	err = conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orders)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(
		`SELECT COUNT(*) FROM cart_items ci
         JOIN carts c ON c.id = ci.cart_id
         WHERE c.user_id = $1`,
		userID,
	).Scan(&cartItems)
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

	if orders != 0 || cartItems != 0 || stock != 10 {
		t.Fatalf(
			"estado alterado: pedidos=%d, itens=%d, estoque=%d",
			orders, cartItems, stock,
		)
	}
}

func TestIdempotentCheckoutRollsBackWhenCartContainsPreorder(t *testing.T) {
	conn := db.SetupTestDB(t)

	var userID, standardID, preorderID int

	err := conn.QueryRow(
		`INSERT INTO users (email, password)
         VALUES ($1, $2) RETURNING id`,
		"preorder-rollback@example.invalid", "test-password",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	for _, product := range []struct {
		id        *int
		name      string
		salesType string
	}{
		{&standardID, "Produto normal", models.SalesTypeStandard},
		{&preorderID, "Camiseta pré-venda", models.SalesTypePreorder},
	} {
		err = conn.QueryRow(
			`INSERT INTO products
             (user_id, name, price, stock, sales_type)
             VALUES ($1, $2, 50, 10, $3) RETURNING id`,
			userID, product.name, product.salesType,
		).Scan(product.id)
		if err != nil {
			t.Fatal(err)
		}
	}

	uc := NewPaymentUsecase(
		repository.NewOrderRepository(conn),
		repository.NewPaymentRepository(conn),
		repository.NewProductRepository(conn),
		nil,
	)

	_, _, err = uc.CreateOrderForCheckoutIdempotent(
		userID,
		[]models.CheckoutItemInput{
			{ProductID: standardID, Quantity: 2},
			{ProductID: preorderID, Quantity: 1},
		},
		"preorder-rollback-test-001",
		"preorder-rollback@example.invalid",
	)

	if !errors.Is(err, ErrPreorderCheckout) {
		t.Fatalf("checkout deveria rejeitar pré-venda: %v", err)
	}

	var orderCount int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&orderCount); err != nil {
		t.Fatal(err)
	}
	if orderCount != 0 {
		t.Fatalf("checkout deixou %d pedidos persistidos", orderCount)
	}

	for _, productID := range []int{standardID, preorderID} {
		var stock int
		if err := conn.QueryRow(
			"SELECT stock FROM products WHERE id = $1",
			productID,
		).Scan(&stock); err != nil {
			t.Fatal(err)
		}
		if stock != 10 {
			t.Errorf("produto %d: estoque alterado para %d", productID, stock)
		}
	}
}
