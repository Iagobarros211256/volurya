CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

/*

user_id em produtos é um design questionável
sqluser_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
Produtos pertencem à loja, não a um usuário específico. Se o admin que criou o produto for deletado, todos os produtos somem com CASCADE. Produtos deveriam ser entidades independentes — remova user_id ou torne-o nullable como created_by:
sqlcreated_by INTEGER REFERENCES users(id) ON DELETE SET NULL

🟡 Inconsistência com users — timezone
users.created_at usa TIMESTAMP (sem timezone), products usa TIMESTAMP WITH TIME ZONE. Padronize todas as tabelas com TIMESTAMP WITH TIME ZONE.

🟡 updated_at não é atualizado automaticamente
sqlupdated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
O DEFAULT só define o valor na inserção. Updates não atualizam automaticamente. Adicione um trigger:
sqlCREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

🟡 Sem índice em name para busca
Se houver busca por nome de produto, um índice acelera bastante:
sqlCREATE INDEX idx_products_name ON products (name);
-- ou para busca parcial:
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);

🟢 description permite NULL implicitamente
Sem NOT NULL e sem DEFAULT ''. Isso é uma escolha válida mas deve ser intencional — garanta que a aplicação trata NULL e string vazia de forma consistente.

*/