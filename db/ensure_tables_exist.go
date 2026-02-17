package db

import "database/sql"

func EnsureTablesExist(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS products (
	        id          SERIAL PRIMARY KEY,
	        user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	        name        TEXT NOT NULL,
	        description TEXT,
	        price       NUMERIC(10,2) NOT NULL CHECK (price >= 0),
	        stock       INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
	        created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()

		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
