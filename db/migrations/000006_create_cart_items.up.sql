CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);


/*


ON DELETE CASCADE em product_id é perigoso
sqlproduct_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE
Se um produto for deletado, os itens do carrinho somem silenciosamente — o usuário perde itens sem aviso. O mais seguro seria ON DELETE RESTRICT para impedir deleção de produtos com itens ativos, ou ON DELETE SET NULL com product_id nullable para preservar o histórico:
sqlproduct_id INTEGER REFERENCES products(id) ON DELETE RESTRICT

🟡 Sem updated_at
Quando a quantidade é atualizada, não há registro de quando isso ocorreu:
sqlupdated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

🟢 added_at poderia se chamar created_at
Por consistência com o resto do schema — todas as outras tabelas usam created_at:
sqlcreated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()


*/