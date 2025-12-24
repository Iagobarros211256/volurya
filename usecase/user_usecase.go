package usecase

import (
	"api/models"
	"errors"
)

type UserRepository interface {
	GetByEmail(email string) (*models.User, error)
	Create(user models.User) error
}

type UserUseCase struct {
	repo UserRepository
}

func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Create(user models.User) error {
	// regra 1: email não pode existir
	existing, err := uc.repo.GetByEmail(user.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("email already exists")
	}

	// regra 2: role default
	if user.Role == "" {
		user.Role = "user"
	}

	// regra 3: senha obrigatória
	if user.Password == "" {
		return errors.New("password is required")
	}

	return uc.repo.Create(user)
}
