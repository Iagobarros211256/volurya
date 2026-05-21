package db

import "testing"

func TestSetupTestDB_Smoke(t *testing.T) {
	db := SetupTestDB()
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("database not reachable: %v", err)
	}
}

/*


Não verifica se as tabelas foram criadas
SetupTestDB provavelmente chama EnsureTablesExist ou RunMigrations. O smoke test deveria verificar que o setup completo funcionou:
gofunc TestSetupTestDB_Smoke(t *testing.T) {
    db := SetupTestDB()
    defer db.Close()

    if err := db.Ping(); err != nil {
        t.Fatalf("database not reachable: %v", err)
    }

    // Verifica que as tabelas essenciais existem
    tables := []string{"users", "products", "orders", "carts", "cart_items", "refresh_tokens"}
    for _, table := range tables {
        var exists bool
        err := db.QueryRow(
            `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
            table,
        ).Scan(&exists)
        if err != nil || !exists {
            t.Errorf("table %s does not exist after setup", table)
        }
    }
}

🟡 Sem verificação de t.Helper() ou t.Skip() se banco não disponível
Em CI sem banco configurado, o teste vai falhar com mensagem genérica. Seria mais claro skippar:
godb, err := TrySetupTestDB()
if err != nil {
    t.Skip("database not available, skipping:", err)
}

🟢 Nome Smoke adequado mas poderia ter tag de build
Testes que dependem de banco externo deveriam ser separados dos unitários:
go//go:build integration
Assim go test ./... não tenta conectar ao banco por padrão, e go test -tags integration ./... roda os testes de integração explicitamente.



*/
