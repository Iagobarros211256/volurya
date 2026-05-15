package usecase

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
)
