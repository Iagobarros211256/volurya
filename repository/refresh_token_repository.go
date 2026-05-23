package repository

import (
	"api/models"
	"database/sql"
	"fmt"
	"time"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.QueryRow(
		"SELECT id, user_id, token, expires_at, revoked, created_at FROM refresh_tokens WHERE token = $1",
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(token string) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1",
		token,
	)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(userID int) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1",
		userID,
	)
	return err
}

/*


 GetByToken retorna nil, nil para token não encontrado
goif err == sql.ErrNoRows {
    return nil, nil
}
Padrão inconsistente com ProductRepository que usa erro sentinela. O caller precisa checar if token == nil em vez de if err != nil, o que é fácil de esquecer:
govar ErrRefreshTokenNotFound = errors.New("refresh token not found")

if err == sql.ErrNoRows {
    return nil, ErrRefreshTokenNotFound
}

🔴 Revoke e RevokeAllByUser sem wrap de erro
gofunc (r *RefreshTokenRepository) Revoke(token string) error {
    _, err := r.db.Exec(...)
    return err  // sem wrap
}
Inconsistente com Create e GetByToken:
goif err != nil {
    return fmt.Errorf("failed to revoke token: %w", err)
}
return nil

🔴 Revoke não verifica se o token existia
goUPDATE refresh_tokens SET revoked = TRUE WHERE token = $1
Se o token não existir, a query executa sem erro e sem efeito. O caller não sabe se a revogação foi bem-sucedida. Use RowsAffected:
goresult, err := r.db.Exec("UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1", token)
if err != nil {
    return fmt.Errorf("failed to revoke token: %w", err)
}
rows, _ := result.RowsAffected()
if rows == 0 {
    return ErrRefreshTokenNotFound
}

🟡 GetByToken não filtra tokens expirados ou revogados
goSELECT ... FROM refresh_tokens WHERE token = $1
Retorna tokens expirados e revogados — a validação fica inteiramente no usecase. É mais seguro filtrar no banco:
goSELECT ... FROM refresh_tokens
WHERE token = $1
  AND revoked = FALSE
  AND expires_at > NOW()
Se o token estiver expirado ou revogado, retorna ErrRefreshTokenNotFound — sem vazar informação sobre o motivo.

🟡 Sem limpeza de tokens expirados
Já apontado na migration — tokens expirados acumulam indefinidamente. O repository deveria ter um método de cleanup:
gofunc (r *RefreshTokenRepository) DeleteExpired() (int64, error) {
    result, err := r.db.Exec(
        "DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE",
    )
    if err != nil {
        return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
    }
    return result.RowsAffected()
}

🟡 Sem interface
Padrão recorrente do projeto.


*/
