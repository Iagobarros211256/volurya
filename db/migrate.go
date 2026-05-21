package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Tenta primeiro relativo ao executável (produção)
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	migrationsPath := filepath.Join(filepath.Dir(execPath), "db", "migrations")

	// Se não existir (go run em dev), usa o diretório atual
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		migrationsPath = filepath.Join(wd, "db", "migrations")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

/*

 Migrations embutidas no binário seria mais robusto
A abordagem atual depende de arquivos no filesystem em runtime — se o diretório db/migrations não estiver disponível no container/deploy, as migrations falham silenciosamente com path errado. O padrão moderno em Go é usar embed:
goimport "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Usar source/iofs em vez de source/file
import "github.com/golang-migrate/migrate/v4/source/iofs"

d, err := iofs.New(migrationsFS, "migrations")
m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
Isso elimina completamente a lógica de busca de path.

🔴 Path do executável pode apontar para /tmp em alguns sistemas
goexecPath, err := os.Executable()
Em alguns sistemas Linux, os.Executable() pode retornar o path em /tmp/go-build... durante go test, fazendo o fallback falhar também. Com embed esse problema desaparece.

🟡 Sem log do path de migrations usado
Se as migrations falharem, não há como saber qual path foi tentado:
goslog.Info("running migrations", "path", migrationsPath)

🟡 Sem verificação de versão atual antes de aplicar
Útil para diagnóstico em produção:
goversion, dirty, err := m.Version()
if err != nil && err != migrate.ErrNilVersion {
    slog.Warn("failed to get migration version", "error", err)
} else {
    slog.Info("migration version", "version", version, "dirty", dirty)
}

🟡 dirty state não tratado
Se uma migration falhou pela metade anteriormente, o banco fica em estado dirty e m.Up() retorna erro. Isso deveria ser tratado explicitamente com log claro em vez de um erro genérico:
goif dirty {
    return fmt.Errorf("database is in dirty state at version %d — manual intervention required", version)
}

🟢 Driver não é fechado após uso
godriver, err := postgres.WithInstance(db, &postgres.Config{})
O driver do migrate deveria ser fechado após o uso, embora na prática raramente cause problema pois as migrations rodam uma vez no startup:
godefer driver.Close()


*/
