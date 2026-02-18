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
	var dsn string

	// Produção / Cloud (Render, etc.)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		dsn = url
	} else {
		log.Println("DATABASE_URL not set — using fallback")

		// Dev / Local fallback
		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("POSTGRES_USER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("POSTGRES_PASSWORD")
		if password == "" {
			password = "postgres"
		}
		dbname := os.Getenv("POSTGRES_DB")
		if dbname == "" {
			dbname = "volurya_db"
		}
		sslmode := os.Getenv("POSTGRES_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)
	}

	if dsn == "" {
		return nil, ErrNoConnectionString
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Pool config (já tinha)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Retry no ping (já tinha, mas confirme)
	const maxRetries = 5
	for i := 1; i <= maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Printf("Database connected successfully (DSN: %s)", maskDSN(dsn))
			return db, nil
		}

		log.Printf("Ping attempt %d/%d failed: %v", i, maxRetries, err)
		time.Sleep(time.Second * time.Duration(i))
	}

	db.Close()
	return nil, fmt.Errorf("%w after %d attempts: %v", ErrConnectionFailed, maxRetries, err)
}

// maskDSN pra não logar senha (opcional, mas bom)
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
