package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func SetupDB() *sql.DB {
	databaseURL := os.Getenv("DATABASE_URL")

	// Se existir DATABASE_URL (Render)
	if databaseURL != "" {
		db, err := sql.Open("postgres", databaseURL)
		if err != nil {
			log.Fatalf("failed to open db connection: %v", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("failed to connect to db: %v", err)
		}

		return db
	}

	// Fallback para ambiente local (Docker)
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "volurya")

	dsn := "host=" + host +
		" port=" + port +
		" user=" + user +
		" password=" + password +
		" dbname=" + dbname +
		" sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	return db
}
