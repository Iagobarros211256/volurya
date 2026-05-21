package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func SetupTestDB() *sql.DB {
	// valores padrão (caso não use env ainda)
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "test")
	password := getEnv("TEST_DB_PASSWORD", "test")
	dbname := getEnv("TEST_DB_NAME", "volurya_test")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open test db: %v", err)
	}

	// valida conexão de verdade
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}

	return db
}

/*


log.Fatalf em função de setup de teste
golog.Fatalf("failed to open test db: %v", err)
log.Fatalf chama os.Exit(1) diretamente — bypassa o defer db.Close() de todos os testes em andamento e não integra com o framework de testes do Go. O correto é receber *testing.T:
gofunc SetupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    if err := db.Ping(); err != nil {
        t.Fatalf("failed to connect to test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
Com t.Cleanup o Close() fica garantido sem precisar de defer em cada teste.

🔴 Sem RunMigrations no setup
godb, err := sql.Open("postgres", dsn)
// ...
return db
O banco de teste é retornado sem schema. Os testes dependem de EnsureTablesExist separadamente, que como vimos está desatualizado. O setup deveria aplicar as migrations reais:
gofunc SetupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db := connectTestDB(t)
    if err := RunMigrations(db); err != nil {
        t.Fatalf("failed to run migrations: %v", err)
    }
    t.Cleanup(func() {
        CleanTestDB(t, db)
        db.Close()
    })
    return db
}

🔴 Sem pool de conexões configurado
conn.go configura pool explicitamente, mas SetupTestDB usa os defaults do database/sql — conexões ilimitadas. Em testes paralelos isso pode exaurir o banco:
godb.SetMaxOpenConns(5)
db.SetMaxIdleConns(2)

🟡 sslmode=disable hardcoded sem possibilidade de override
Diferente de conn.go que lê POSTGRES_SSLMODE via env, aqui está fixo. Consistência:
gosslmode := getEnv("TEST_DB_SSLMODE", "disable")

🟡 Importa log em vez de slog
O projeto todo usa slog estruturado mas esse arquivo usa o log antigo — mesma inconsistência vista no main.go.


*/
