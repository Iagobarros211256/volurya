package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
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

	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL has invalid format")
	}
	if len(password) < 8 {
		return errors.New("BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}

	// Verifica se o usuário já existe antes de criar
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check bootstrap admin user: %w", err)
	}

	if exists {
		// Usuário já existe — não sobrescreve senha nem role
		slog.Info("bootstrap admin user already exists, skipping", "email", email)
		return nil
	}

	// Cria apenas se não existe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("failed to hash bootstrap admin password: %w", err)
	}

	if _, err := db.Exec(
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, 'admin')
		 ON CONFLICT (email) DO NOTHING`,
		email,
		string(hashedPassword),
	); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	slog.Info("bootstrap admin user created", "email", email)
	return nil
}
