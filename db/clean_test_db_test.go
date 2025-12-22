package db

import "testing"

func TestCleanTestDB(t *testing.T) {
	database := SetupTestDB()
	defer database.Close()

	// Cria tabela users
	_, err := database.Exec(`CREATE TABLE IF NOT EXISTS users (
         id SERIAL PRIMARY KEY,
         email VARCHAR(255) UNIQUE NOT NULL,
         password VARCHAR(255) NOT NULL,
         role VARCHAR(50) NOT NULL,
         created_at TIMESTAMP NOT NULL DEFAULT NOW()
    )`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Cria tabela products
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS products (
           id SERIAL primary key,
           name VARCHAR(50) not null,
           description VARCHAR(200) not null,
           price NUMERIC(10, 2) not null,
           stock INTEGER DEFAULT 0 not null  
    )`)
	if err != nil {
		t.Fatalf("failed to create products table: %v", err)
	}

	// Agora chama o CleanTestDB
	if err := CleanTestDB(database); err != nil {
		t.Fatalf("clean failed: %v", err)
	}
}
