//go:build integration

package db

import (
	"fmt"
	"testing"
)

func TestCleanTestDB_ShouldRemoveAllData(t *testing.T) {
	db := SetupTestDB(t)

	// Hash bcrypt real de "testpassword" com cost 10 — exatamente 60 chars
	const bcryptHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	var userID int
	err := db.QueryRow(
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
		"test@test.com", bcryptHash, "user",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	var productID int
	err = db.QueryRow(
		`INSERT INTO products (user_id, name, price, stock) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Test Product", 99.90, 10,
	).Scan(&productID)
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}

	// Verifica que os dados existem antes do cleanup
	tables := []string{"users", "products"}
	for _, table := range tables {
		var count int
		if err := db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)).Scan(&count); err != nil {
			t.Fatalf("failed to count %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("expected data in %s before clean", table)
		}
	}
	// CleanTestDB é chamado automaticamente pelo t.Cleanup do SetupTestDB
}
