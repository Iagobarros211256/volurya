package repository

import (
	"api/models"
	"database/sql"
	"fmt"
	"time"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.QueryRow(
		"SELECT id, user_id, token, expires_at, revoked, created_at FROM refresh_tokens WHERE token = $1",
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(token string) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1",
		token,
	)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(userID int) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1",
		userID,
	)
	return err
}
