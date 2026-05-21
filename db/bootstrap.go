package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func BootstrapAdminUser(db *sql.DB) error {
	email := strings.TrimSpace(strings.ToLower(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")

	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD must be set together")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL has invalid format")
	}
	if len(password) < 8 {
		return errors.New("BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash bootstrap admin password: %w", err)
	}

	if _, err := db.Exec(
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, 'admin')
		 ON CONFLICT (email)
		 DO UPDATE SET password = EXCLUDED.password, role = 'admin'`,
		email,
		string(hashedPassword),
	); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	slog.Info("bootstrap admin user ready", "email", email)
	return nil
}

/*

DO UPDATE SET password atualiza senha a cada boot
sqlON CONFLICT (email)
DO UPDATE SET password = EXCLUDED.password, role = 'admin'
Se o admin alterou a própria senha via aplicação, ela é sobrescrita com a do .env a cada restart. Use DO NOTHING ou atualize apenas se necessário:
sqlON CONFLICT (email) DO NOTHING
Ou se quiser forçar reset explicitamente, adicione uma env BOOTSTRAP_FORCE_RESET=true.

🔴 Senha logada indiretamente via slog
O log atual só mostra o email, o que está correto. Mas se alguém adicionar "password", password futuramente seguindo o padrão, a senha vaza. Considere limpar a variável após o uso:
godefer func() { password = "" }()

🟡 Validação de email muito fraca
goif !strings.Contains(email, "@") || !strings.Contains(email, ".") {
"@." passaria nessa validação. Use net/mail:
goimport "net/mail"

if _, err := mail.ParseAddress(email); err != nil {
    return errors.New("BOOTSTRAP_ADMIN_EMAIL has invalid format")
}

🟡 bcrypt.DefaultCost pode ser insuficiente no futuro
DefaultCost é 10 atualmente. Para uma conta admin, considere bcrypt.DefaultCost + 2 (12) — o custo extra é irrelevante para um único hash no boot:
gobcrypt.GenerateFromPassword([]byte(password), 12)

🟢 Sem log de distinção entre criação e atualização
Com DO UPDATE, não dá pra saber pelo log se o admin foi criado ou atualizado. Útil para auditoria:
go// Verificar se já existia antes do upsert
var exists bool
db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
// ... executar upsert ...
if exists {
    slog.Info("bootstrap admin user updated", "email", email)
} else {
    slog.Info("bootstrap admin user created", "email", email)
}

*/
