package usecase

import (
	"api/auth"
	"api/models"
	"api/repository"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo         repository.UserRepositoryInterface
	refreshTokenRepo *repository.RefreshTokenRepository
}

func NewAuthUsecase(userRepo repository.UserRepositoryInterface, refreshTokenRepo *repository.RefreshTokenRepository) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (a *AuthUsecase) Signup(email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return "", "", errors.New("invalid email format")
	}
	if len(password) < 8 {
		return "", "", errors.New("password must be at least 8 characters")
	}

	existing, err := a.userRepo.GetByEmail(email)
	if err != nil {
		return "", "", err
	}
	if existing != nil {
		return "", "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	user := models.User{
		Email:    email,
		Password: string(hash),
		Role:     "user",
	}

	id, err := a.userRepo.Create(user)
	if err != nil {
		return "", "", err
	}

	accessToken, err := auth.GenerateToken(id, user.Role, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, expiresAt, err := auth.GenerateRefreshToken(id, user.Role)
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(id, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *AuthUsecase) Login(email, password string) (string, string, error) {
	user, err := a.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return "", "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, expiresAt, err := auth.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(user.ID, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *AuthUsecase) RefreshToken(refreshToken string) (string, string, error) {
	// Valida o token JWT
	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Busca no banco
	rt, err := a.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil || rt == nil {
		return "", "", errors.New("refresh token not found")
	}

	// Verifica se foi revogado ou expirou
	if rt.Revoked {
		return "", "", errors.New("refresh token revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	// Revoga o token atual (rotação)
	if err := a.refreshTokenRepo.Revoke(refreshToken); err != nil {
		return "", "", err
	}

	// Gera novos tokens
	newAccessToken, err := auth.GenerateToken(claims.UserID, claims.Role, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, expiresAt, err := auth.GenerateRefreshToken(claims.UserID, claims.Role)
	if err != nil {
		return "", "", err
	}

	if err := a.refreshTokenRepo.Create(claims.UserID, newRefreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (a *AuthUsecase) Logout(refreshToken string) error {
	return a.refreshTokenRepo.Revoke(refreshToken)
}
