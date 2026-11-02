package db

import (
	"database/sql"
	"fmt"
	"testing"
)

// ensureTestDatabase impede operações destrutivas fora dos bancos de teste.
func isAllowedTestDatabase(name, user string) bool {
	return user == "test" &&
		(name == "volurya_test" || name == "volurya_payment_test")
}

func ensureTestDatabase(db *sql.DB) error {
	var name, user string

	if err := db.QueryRow(
		"SELECT current_database(), current_user",
	).Scan(&name, &user); err != nil {
		return fmt.Errorf("identificar banco de testes: %w", err)
	}

	if !isAllowedTestDatabase(name, user) {
		return fmt.Errorf(
			"operação destrutiva recusada: banco=%q usuário=%q",
			name, user,
		)
	}

	return nil
}

// CleanTestDB trunca todas as tabelas do banco de teste.
// Ordem respeitando foreign keys — da mais dependente para a menos.
func CleanTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if err := ensureTestDatabase(db); err != nil {
		t.Fatal(err)
	}

	_, err := db.Exec(`
		TRUNCATE TABLE
			payment_records,
			order_items,
			cart_items,
			refresh_tokens,
			orders,
			carts,
			products,
			users
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("failed to clean test db: %v", err)
	}
}

// CleanTestDBErr é a versão que retorna erro — para uso fora de testes.
func CleanTestDBErr(db *sql.DB) error {

	if err := ensureTestDatabase(db); err != nil {
		return err
	}
	_, err := db.Exec(`
		TRUNCATE TABLE
			payment_records,
			order_items,
			cart_items,
			refresh_tokens,
			orders,
			carts,
			products,
			users
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("clean test db failed: %w", err)
	}
	return nil
}
