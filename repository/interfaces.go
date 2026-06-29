package repository

import (
	"api/models"
	"time"
)

type RefreshTokenRepositoryInterface interface {
	Create(userID int, token string, expiresAt time.Time) error
	GetByToken(token string) (*models.RefreshToken, error)
	Revoke(token string) error
	RevokeAllByUser(userID int) error
	DeleteExpired() (int64, error)
}
