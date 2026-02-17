package usecase

import (
	"api/auth"
	"api/models"
	"api/repository"
	"errors"
	"strings"

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

func (a *AuthUsecase) Signup(email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	// Validações básicas
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("invalid email format")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Checa duplicata ANTES de gerar hash
	existing, err := a.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Email:    email,
		Password: string(hash),
		Role:     "user", // default seguro – admin via seed ou painel
	}

	return a.userRepo.Create(user)
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

	return auth.GenerateToken(user.ID, user.Role)
}
