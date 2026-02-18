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
	userRepo repository.UserRepositoryInterface
}

func NewAuthUsecase(userRepo repository.UserRepositoryInterface) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
	}
}

func (a *AuthUsecase) Signup(email, password string) (string, error) { // ← muda retorno pra (string, error)
	email = strings.TrimSpace(strings.ToLower(email))

	// Validações (já deve ter)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return "", errors.New("invalid email format")
	}
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}

	// Checa duplicata
	existing, err := a.userRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := models.User{
		Email:    email,
		Password: string(hash),
		Role:     "user",
	}

	// Cria e pega o ID gerado
	id, err := a.userRepo.Create(user)
	if err != nil {
		return "", err
	}

	// Gera o token imediatamente com o ID novo
	token, err := auth.GenerateToken(id, user.Role, 24*time.Hour) // ajuste duração
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUsecase) Login(email, password string) (string, error) {
	user, err := a.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	return auth.GenerateToken(user.ID, user.Role, 24*time.Hour) // ou 1*time.Hour pra teste
}
