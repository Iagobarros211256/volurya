package db

import (
	"database/sql"
	"fmt"
	"testing"
)

// CleanTestDB trunca todas as tabelas do banco de teste.
// Ordem respeitando foreign keys — da mais dependente para a menos.
func CleanTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

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
