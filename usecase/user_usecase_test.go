package usecase

import (
	"api/models"
	"errors"
	"testing"
)

type fakeUserRepo struct {
	users map[string]models.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users: make(map[string]models.User),
	}
}

func (f *fakeUserRepo) GetByEmail(email string) (*models.User, error) {
	user, ok := f.users[email]
	if !ok {
		return nil, nil
	}
	return &user, nil
}

func (f *fakeUserRepo) Create(user models.User) error {
	if _, exists := f.users[user.Email]; exists {
		return errors.New("duplicate")
	}
	f.users[user.Email] = user
	return nil
}

func TestUserUseCase_Create_Success(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	user := models.User{
		Email:    "test@test.com",
		Password: "123456",
	}

	err := uc.Create(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	savedUser, _ := repo.GetByEmail("test@test.com")

	if savedUser == nil {
		t.Fatal("user was not saved")
	}

	if savedUser.Password == "123456" {
		t.Fatal("password was not hashed")
	}

	if savedUser.Role != "user" {
		t.Fatalf("expected role 'user', got %s", savedUser.Role)
	}
}

func TestUserUseCase_Create_DuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	user := models.User{
		Email:    "test@test.com",
		Password: "123456",
	}

	_ = uc.Create(user)

	err := uc.Create(user)
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}
