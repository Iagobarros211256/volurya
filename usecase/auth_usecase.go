package usecase

import (
	"api/auth"
	"api/models"
	"api/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo repository.UserRepository
}

func NewAuthUsecase(repo repository.UserRepository) AuthUsecase {
	return AuthUsecase{
		userRepo: repo,
	}
}

func (a *AuthUsecase) Signup(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user := models.User{
		Email:    email,
		Password: string(hash),
		Role:     "admin",
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
