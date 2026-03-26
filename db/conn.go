package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrNoConnectionString = errors.New("no database connection string provided")
	ErrConnectionFailed   = errors.New("failed to connect to database")
)

func ConnectDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not set — using fallback")
		dsn = buildFallbackDSN()
	}

	if dsn == "" {
		return nil, ErrNoConnectionString
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configurações de pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Retry mais agressivo para Render
	const maxRetries = 15
	for i := 1; i <= maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Printf("Database connected successfully")
			return db, nil
		}

		log.Printf("Ping attempt %d/%d failed: %v", i, maxRetries, err)

		sleepTime := time.Duration(i+1) * time.Second
		if sleepTime > 10*time.Second {
			sleepTime = 10 * time.Second
		}
		time.Sleep(sleepTime)
	}

	db.Close()
	return nil, fmt.Errorf("%w after %d attempts: %v", ErrConnectionFailed, maxRetries, err)
}

func buildFallbackDSN() string {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DB", "volurya_db")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

// maskDSN para não logar senha
func maskDSN(dsn string) string {
	if idx := strings.Index(dsn, "password="); idx != -1 {
		end := strings.Index(dsn[idx:], " ")
		if end == -1 {
			end = len(dsn)
		}
		return dsn[:idx] + "password=******" + dsn[idx+end:]
	}
	return dsn
}
