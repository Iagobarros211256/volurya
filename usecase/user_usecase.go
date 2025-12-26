package usecase

import (
	"api/models"
	"errors"

	"golang.org/x/crypto/bcrypt"
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

	// regra 2: senha obrigatória
	if user.Password == "" {
		return errors.New("password is required")
	}

	// regra 3: hash da senha (INVARIANTE DE DOMÍNIO)
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	// regra 4: role default
	if user.Role == "" {
		user.Role = "user"
	}

	return uc.repo.Create(user)
}
