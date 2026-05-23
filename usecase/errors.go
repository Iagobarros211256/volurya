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

/*

Erros declarados mas nem todos usados
Cruzando com os usecases analisados:

ErrInvalidEmail — declarado mas auth_usecase.go retorna errors.New("invalid email format") diretamente em vez de usar essa constante
ErrPasswordTooShort — mesmo problema, auth_usecase.go retorna errors.New("password must be between 8 and 128 characters") com mensagem diferente
ErrEmailAlreadyExists — declarado mas auth_usecase.go retorna errors.New("email already registered") diretamente

Os erros existem mas o código não os usa — perdendo o benefício de errors.Is() nos controllers e testes.

🟡 Faltam erros que aparecem como strings nos usecases
go// Em cart_usecase.go — deveriam estar aqui:
errors.New("item not found")
errors.New("cart is empty")
errors.New("forbidden: item does not belong to your cart")
errors.New("product not found")

// Em auth_usecase.go:
errors.New("refresh token not found")
errors.New("refresh token revoked")
errors.New("refresh token expired")

🟢 ErrPasswordTooShort não reflete a validação real
goErrPasswordTooShort = errors.New("password must be at least 8 characters")
O auth_usecase.go valida len(password) < 8 || len(password) > 128 — há um limite máximo também. O nome e mensagem do erro deveriam refletir isso:
goErrInvalidPasswordLength = errors.New("password must be between 8 and 128 characters")

*/
