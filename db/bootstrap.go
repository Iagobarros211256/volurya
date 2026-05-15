package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func BootstrapAdminUser(db *sql.DB) error {
	email := strings.TrimSpace(strings.ToLower(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")

	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD must be set together")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL has invalid format")
	}
	if len(password) < 8 {
		return errors.New("BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash bootstrap admin password: %w", err)
	}

	if _, err := db.Exec(
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, 'admin')
		 ON CONFLICT (email)
		 DO UPDATE SET password = EXCLUDED.password, role = 'admin'`,
		email,
		string(hashedPassword),
	); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	slog.Info("bootstrap admin user ready", "email", email)
	return nil
}
