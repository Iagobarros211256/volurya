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
	Create(user models.User) error
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
