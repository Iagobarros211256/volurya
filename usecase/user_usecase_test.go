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

/*

fakeUserRepo.Create tem assinatura diferente do repositório real
gofunc (f *fakeUserRepo) Create(user models.User) error  // mock
func (r *UserRepository) Create(user models.User) (int, error)  // real
O mock não implementa a interface real — o teste compila com uma interface implícita diferente. Se NewUserUseCase aceita a interface real, esse teste nem deveria compilar.

🔴 fakeUserRepo.GetByEmail retorna nil, nil para não encontrado
goif !ok {
    return nil, nil
}
Propaga o anti-padrão do repository real em vez de usar erro sentinela. Se UserUsecase for corrigido para tratar ErrUserNotFound, o mock vai precisar ser atualizado.

🟡 Senha fraca nos testes
goPassword: "123456"
Senha com 6 caracteres provavelmente falha na validação de comprimento mínimo (8 chars) do usecase — o teste TestUserUseCase_Create_Success pode estar testando um caminho de erro mascarado. Use senha válida:
goPassword: "validPassword123"

🟡 Setup duplicado
gorepo := newFakeUserRepo()
uc := NewUserUseCase(repo)
Repetido em ambos os testes. Extrai helper:
gofunc setupUserUsecase(t *testing.T) (*UserUsecase, *fakeUserRepo) {
    t.Helper()
    repo := newFakeUserRepo()
    return NewUserUseCase(repo), repo
}

🟡 Faltam testes importantes

Email inválido → erro
Senha muito curta → erro
Senha muito longa → erro
GetByEmail retornando usuário existente antes do Create


🟢 TestUserUseCase_Create_DuplicateEmail não verifica a mensagem de erro
goerr := uc.Create(user)
if err == nil {
    t.Fatal("expected error for duplicate email, got nil")
}
Qualquer erro passa — inclusive erros de infraestrutura. Verifique o tipo:
goif !errors.Is(err, ErrEmailAlreadyExists) {
    t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
}

*/
