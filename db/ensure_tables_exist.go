package db

import "database/sql"

func EnsureTablesExist(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS products (
	        id          SERIAL PRIMARY KEY,
	        user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	        name        TEXT NOT NULL,
	        description TEXT,
	        price       NUMERIC(10,2) NOT NULL CHECK (price >= 0),
	        stock       INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
	        created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()

		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

/*



 Duplicação do schema com as migrations
Essa função recria manualmente tabelas que já existem nas migrations. Isso significa que o schema está definido em dois lugares — qualquer alteração precisa ser feita em ambos. Já está desatualizado:

users aqui não tem updated_at
products aqui não tem image_url
Nenhuma das outras tabelas está incluída (orders, carts, cart_items, refresh_tokens, payment_records)

A solução correta é usar as próprias migrations nos testes:
gofunc SetupTestDB() *sql.DB {
    db := connectTestDB()
    if err := RunMigrations(db); err != nil {
        panic(err)
    }
    return db
}

🔴 Erros não wrappados
goreturn err
Sem contexto de qual query falhou:
goif _, err := db.Exec(q); err != nil {
    return fmt.Errorf("ensure tables failed: %w", err)
}

🟡 Schema de teste diverge do schema de produção
users aqui usa TIMESTAMP sem timezone, mas a migration de produção usa TIMESTAMP WITH TIME ZONE em products. Testes rodando com schema diferente do produção podem passar localmente e falhar em produção.

🟡 Formatação inconsistente das queries
go`CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    			        // linha em branco com tabs
    email TEXT...
Indentação mista entre as duas queries — tabs vs espaços, alinhamentos diferentes.

🟢 Função deveria não existir
Com RunMigrations já disponível no projeto, EnsureTablesExist é redundante. O único caso de uso é nos testes, e mesmo assim deveria usar as migrations reais. Candidate a remoção completa.



*/
