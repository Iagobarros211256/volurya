package db

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB conecta ao banco de teste e aplica as migrations reais.
// Requer banco PostgreSQL rodando na porta 5433 (docker-compose.test.yml).
// Uso: db := SetupTestDB(t) — Close() e CleanTestDB são chamados automaticamente via t.Cleanup.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

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
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("test database not available (start with docker compose -f local/docker-compose.test.yml up -d): %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Aplica migrations reais — mesmo schema de produção
	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations on test db: %v", err)
	}

	t.Cleanup(func() {
		CleanTestDB(t, db)
		db.Close()
	})

	return db
}
