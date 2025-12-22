package db

import "database/sql"

// Limpa todas as tabelas relevantes entre testes
func CleanTestDB(db *sql.DB) error {
	_, err := db.Exec(`
		TRUNCATE TABLE
			users,
			products
		RESTART IDENTITY
		CASCADE
	`)
	return err
}
