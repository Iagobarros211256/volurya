package db

import "testing"

func TestCleanTestDB_ShouldRemoveAllData(t *testing.T) {
	database := SetupTestDB()
	defer database.Close()

	if err := EnsureTablesExist(database); err != nil {
		t.Fatalf("failed to ensure tables: %v", err)
	}

	// Insere dados fake
	_, err := database.Exec(
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3)`,
		"test@test.com",
		"hashed",
		"user",
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Sanity check
	var before int
	_ = database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&before)
	if before == 0 {
		t.Fatal("expected data before clean")
	}

	// Act
	if err := CleanTestDB(database); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	// Assert
	var after int
	err = database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&after)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if after != 0 {
		t.Fatalf("expected 0 users after clean, got %d", after)
	}
}

/*


Testa apenas users — outras tabelas não são verificadas
CleanTestDB provavelmente limpa múltiplas tabelas. O teste só verifica users:
gotables := []string{"users", "products", "orders", "carts", "cart_items", "refresh_tokens", "payment_records"}
for _, table := range tables {
    var count int
    db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)).Scan(&count)
    if count != 0 {
        t.Errorf("expected 0 rows in %s after clean, got %d", table, count)
    }
}

🟡 Senha em texto puro inserida diretamente
go"hashed",  // não é um hash real
Para um teste de limpeza isso é aceitável, mas se users.password tiver uma constraint de tamanho mínimo futuramente (como sugerido na migration), o teste vai quebrar. Use um hash bcrypt real ou uma constante:
goconst testPasswordHash = "$2a$10$..." // hash bcrypt pré-computado

🟡 _ = database.QueryRow(...).Scan(&before) ignora erro
go_ = database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&before)
No sanity check o erro é ignorado. Se a query falhar, before fica 0 e o teste falha com mensagem enganosa:
goif err := database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&before); err != nil {
    t.Fatalf("failed to count users before clean: %v", err)
}

🟢 Teste não verifica ordem de limpeza
Se CleanTestDB não respeitar a ordem das foreign keys, vai falhar com erro de constraint. O teste passando implicitamente valida isso, mas seria mais claro inserir dados em múltiplas tabelas relacionadas para garantir que a limpeza respeita as dependências.



*/
