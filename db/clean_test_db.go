package db

import (
	"database/sql"
	"fmt"
)

// Limpa todas as tabelas relevantes entre testes
func CleanTestDB(db *sql.DB) error {
	_, err := db.Exec(`
		TRUNCATE TABLE
			users,
			products
		RESTART IDENTITY
		CASCADE
	`)
	if err != nil {
		return fmt.Errorf("clean test db failed: %w", err)
	}
	return nil

}

/*



Tabelas faltando no TRUNCATE
sqlTRUNCATE TABLE users, products RESTART IDENTITY CASCADE
CASCADE propaga o truncate para tabelas dependentes, então orders, carts, cart_items, refresh_tokens e payment_records são limpas indiretamente. Mas isso é implícito e frágil — se uma nova tabela for adicionada sem FK para users ou products, não será limpa. Seja explícito:
sqlTRUNCATE TABLE
    payment_records,
    cart_items,
    refresh_tokens,
    orders,
    carts,
    products,
    users
RESTART IDENTITY CASCADE

🔴 Inconsistência com o teste
O teste em clean_test_db_test.go verifica apenas users após o clean. Com CASCADE implícito, outras tabelas são limpas mas não verificadas. E se CASCADE falhar silenciosamente numa tabela específica, o teste não detecta.

🟡 Sem verificação de ambiente
CleanTestDB pode ser chamada acidentalmente em produção. Adicione uma proteção:
gofunc CleanTestDB(db *sql.DB) error {
    if os.Getenv("ENV") == "production" {
        return errors.New("CleanTestDB cannot be called in production")
    }
    // ...
}

🟢 Nome da função não segue convenção de teste
Funções utilitárias de teste em Go costumam receber *testing.T para integrar com o framework:
gofunc CleanTestDB(t *testing.T, db *sql.DB) {
    t.Helper()
    if _, err := db.Exec(`TRUNCATE TABLE ...`); err != nil {
        t.Fatalf("failed to clean test db: %v", err)
    }
}
Isso elimina o tratamento de erro no caller e integra melhor com o output do go test.


*/
