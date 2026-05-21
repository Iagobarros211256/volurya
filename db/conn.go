package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrNoConnectionString = errors.New("no database connection string provided")
	ErrConnectionFailed   = errors.New("failed to connect to database")
)

func ConnectDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Warn("DATABASE_URL not set — using fallback")
		dsn = buildFallbackDSN()
	}

	if dsn == "" {
		return nil, ErrNoConnectionString
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configurações de pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Retry mais agressivo para Render
	const maxRetries = 15
	for i := 1; i <= maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			slog.Info("Database connected successfully")
			return db, nil
		}

		slog.Warn("Ping attempt failed", "attempt", i, "max", maxRetries, "error", err)

		sleepTime := time.Duration(i+1) * time.Second
		if sleepTime > 10*time.Second {
			sleepTime = 10 * time.Second
		}
		time.Sleep(sleepTime)
	}

	db.Close()
	return nil, fmt.Errorf("%w after %d attempts: %v", ErrConnectionFailed, maxRetries, err)
}

func buildFallbackDSN() string {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DB", "volurya_db")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

// maskDSN para não logar senha
func maskDSN(dsn string) string {
	if idx := strings.Index(dsn, "password="); idx != -1 {
		end := strings.Index(dsn[idx:], " ")
		if end == -1 {
			end = len(dsn)
		}
		return dsn[:idx] + "password=******" + dsn[idx+end:]
	}
	return dsn
}

/*



 maskDSN definido mas nunca usado
gofunc maskDSN(dsn string) string {
A função existe mas não é chamada em nenhum log. Se DATABASE_URL fosse logada, a senha apareceria em claro. Use ou remova:
goslog.Warn("DATABASE_URL not set — using fallback", "dsn", maskDSN(dsn))

🔴 sslmode=disable no fallback
gosslmode := getEnv("POSTGRES_SSLMODE", "disable")
SSL desabilitado por padrão é perigoso se o fallback for usado acidentalmente em produção. O default deveria ser require:
gosslmode := getEnv("POSTGRES_SSLMODE", "require")

🟡 time.Sleep bloqueia a goroutine sem context
gotime.Sleep(sleepTime)
Se o processo receber SIGTERM durante o retry, o shutdown graceful fica bloqueado pelo sleep. Use context com cancelamento:
gofunc ConnectDB(ctx context.Context) (*sql.DB, error) {
    // ...
    select {
    case <-time.After(sleepTime):
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

🟡 15 retries com até 10s cada = até 2.5 minutos bloqueando o startup
goconst maxRetries = 15
// sleepTime máximo de 10s
Em produção no Render isso pode ser necessário, mas em testes locais torna o feedback muito lento. Considere tornar configurável:
gomaxRetries := 15
if os.Getenv("ENV") == "test" {
    maxRetries = 3
}

🟡 getEnv usado mas não definido neste arquivo
gohost := getEnv("POSTGRES_HOST", "localhost")
Provavelmente está em env.go, o que é ok, mas vale confirmar que não há duplicação com funções similares no pacote config.

🟢 Pool hardcoded — considere tornar configurável
godb.SetMaxOpenConns(20)
db.SetMaxIdleConns(10)
Para ambientes com recursos limitados (Render free tier), 20 conexões pode ser muito. Via env:
gomaxConns := getEnvInt("DB_MAX_OPEN_CONNS", 20)
db.SetMaxOpenConns(maxConns)

*/
