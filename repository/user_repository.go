package repository

import (
	"api/models"
	"database/sql"
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

func (r *UserRepository) Create(user models.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (email, password, role) VALUES ($1, $2, $3)",
		user.Email,
		user.Password,
		user.Role,
	)
	return err
}
