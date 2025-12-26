package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func SetupTestDB() *sql.DB {
	// valores padrão (caso não use env ainda)
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "test")
	password := getEnv("TEST_DB_PASSWORD", "test")
	dbname := getEnv("TEST_DB_NAME", "volurya_test")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open test db: %v", err)
	}

	// valida conexão de verdade
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}

	return db
}
