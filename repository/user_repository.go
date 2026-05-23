package repository

import (
	"api/models"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

type UserRepositoryInterface interface {
	GetByEmail(email string) (*models.User, error)
	Create(user models.User) (int, error)
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password, role FROM users WHERE email = $1",
		email,
	)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(user models.User) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id",
		user.Email,
		user.Password,
		user.Role,
	).Scan(&id)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return 0, errors.New("email already registered")
		}
		return 0, err
	}

	return id, nil
}

/*

Interface no lugar errado
gotype UserRepositoryInterface interface {
    GetByEmail(email string) (*models.User, error)
    Create(user models.User) (int, error)
}
A interface está no pacote repository em vez de no pacote que a consome (usecase). Em Go o padrão é definir interfaces onde são usadas — o usecase deveria definir o que precisa, não o repository definir o que oferece. Além disso, nenhum outro repository tem interface — inconsistência total.

🔴 GetByEmail retorna nil, nil para não encontrado
Mesmo problema do RefreshTokenRepository — padrão inconsistente com ProductRepository que usa erro sentinela:
govar ErrUserNotFound = errors.New("user not found")

if err == sql.ErrNoRows {
    return nil, ErrUserNotFound
}

🔴 Erros sem wrap
goreturn nil, err  // GetByEmail
return 0, err    // Create (caso não-pq)
Inconsistente — o erro de email duplicado tem mensagem customizada mas outros erros não têm contexto nenhum:
goreturn nil, fmt.Errorf("failed to get user by email: %w", err)
return 0, fmt.Errorf("failed to create user: %w", err)

🔴 "23505" hardcoded como string
goif pgErr.Code == "23505"
Use a constante do pacote pq:
goif pgErr.Code == pq.ErrorCode("23505")
// ou melhor:
const uniqueViolation = pq.ErrorCode("23505")
if pgErr.Code == uniqueViolation

🟡 GetByEmail não retorna created_at
goSELECT id, email, password, role FROM users
created_at existe no banco mas não é retornado — o model User também não o tem, problema já apontado.

🟡 Create recebe models.User com senha já hasheada
A convenção não está documentada — quem chama Create precisa saber que deve passar o hash, não a senha pura. Um tipo dedicado tornaria isso explícito:
gotype CreateUserInput struct {
    Email          string
    HashedPassword string
    Role           models.UserRole
}

🟡 errors.New("email already registered") perde o erro original
goreturn 0, errors.New("email already registered")
O erro do postgres é descartado. Use wrap para manter a rastreabilidade:
govar ErrEmailAlreadyRegistered = errors.New("email already registered")
return 0, fmt.Errorf("%w: %w", ErrEmailAlreadyRegistered, err)

🟢 Sem GetByID
Busca por ID é uma operação básica que provavelmente vai ser necessária para carregar dados do usuário autenticado:
gofunc (r *UserRepository) GetByID(id int) (*models.User, error)


*/
