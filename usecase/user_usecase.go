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

/*

UserRepository.Create retorna error mas repository.UserRepository.Create retorna (int, error)
go// usecase — interface
Create(user models.User) error

// repository — implementação real
func (r *UserRepository) Create(user models.User) (int, error)
A implementação real não satisfaz a interface — o código não compila em produção. Uma das duas precisa mudar. O id retornado pelo banco é útil (atribuir ao usuário criado), então a interface deveria ser:
goCreate(user models.User) (int, error)

🔴 Create não retorna o usuário ou ID criado
gofunc (uc *UserUseCase) Create(user models.User) error
Após criar o usuário, o caller não tem acesso ao ID gerado pelo banco. AuthUsecase.Signup precisa do ID para gerar o JWT — por isso tem seu próprio fluxo de criação paralelo. Retorne o ID:
gofunc (uc *UserUseCase) Create(user models.User) (int, error)

🔴 Duplicação com AuthUsecase.Signup
UserUseCase.Create e AuthUsecase.Signup fazem essencialmente a mesma coisa — validam email, hashiam senha, criam usuário. Dois caminhos paralelos para a mesma operação. Um deveria usar o outro, ou deveriam ser unificados.

🟡 Validação de email ausente
AuthUsecase.Signup usa net/mail.ParseAddress para validar o email, mas UserUseCase.Create não valida o formato — aceita qualquer string como email:
goif _, err := mail.ParseAddress(user.Email); err != nil {
    return errors.New("invalid email format")
}

🟡 Validação de senha mínima ausente
goif user.Password == "" {
    return errors.New("password is required")
}
Só verifica se está vazio — não valida comprimento mínimo de 8 caracteres como em AuthUsecase. Isso explica por que o teste passa com senha "123456".

🟡 Race condition no check de email duplicado
Mesmo problema do AuthUsecase.Signup — SELECT antes do INSERT permite que dois usuários com o mesmo email sejam criados simultaneamente. Confie na constraint UNIQUE do banco.

🟡 Erros sem sentinelas
goreturn errors.New("email already exists")
return errors.New("password is required")
Deveriam usar as constantes de errors.go:
goreturn ErrEmailAlreadyExists

🟢 UserUseCase vs AuthUsecase — nomenclatura inconsistente
gotype UserUseCase struct  // PascalCase com "Case" maiúsculo
type AuthUsecase struct  // camelCase com "case" minúsculo
Padronize para UserUsecase.

*/
