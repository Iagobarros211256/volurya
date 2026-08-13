//go:build integration

package db

import "testing"

func TestSetupTestDB_Smoke(t *testing.T) {
	db := SetupTestDB(t)

	if err := db.Ping(); err != nil {
		t.Fatalf("database not reachable: %v", err)
	}

	// Verifica que as migrations criaram as tabelas essenciais
	tables := []string{
		"users", "products", "orders", "order_items",
		"carts", "cart_items", "refresh_tokens", "payment_records",
	}
	for _, table := range tables {
		var exists bool
		err := db.QueryRow(
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s does not exist after migrations", table)
		}
	}
}
