package db

import (
	"database/sql"
	"fmt"
)

// Limpa todas as tabelas relevantes entre testes
func CleanTestDB(db *sql.DB) error {
	_, err := db.Exec(`
		TRUNCATE TABLE
			users,
			products
		RESTART IDENTITY
		CASCADE
	`)
	if err != nil {
		return fmt.Errorf("clean test db failed: %w", err)
	}
	return nil

}
