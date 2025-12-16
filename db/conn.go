package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "volurya_postgres"
	port     = 5432
	user     = "volurya"
	password = "volurya"
	dbname   = "volurya_db"
)

func ConnectDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= 10; attempt++ {
		db, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			err = db.Ping()
			if err == nil {
				fmt.Println("Connected to " + dbname)
				return db, nil
			}
		}

		fmt.Printf("Database not ready (attempt %d), retrying...\n", attempt)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to database after retries: %w", err)
}
