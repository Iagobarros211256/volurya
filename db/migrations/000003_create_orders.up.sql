CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    total NUMERIC(10,2) NOT NULL,
    status TEXT NOT NULL,
    charge_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


/*

product_id direto na ordem — design não escala
sqlproduct_id INTEGER NOT NULL REFERENCES products(id),
quantity INTEGER NOT NULL,
Uma ordem com múltiplos produtos é impossível nessa estrutura. O padrão correto é uma tabela order_items:
sql-- orders: apenas cabeçalho
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    total NUMERIC(10,2) NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- order_items: os produtos da ordem
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10,2) NOT NULL
);

🔴 status TEXT sem constraint
Igual ao role em users — qualquer string é aceita:
sqlstatus TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'cancelled', 'refunded'))
-- ou
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

🔴 Sem ON DELETE nas foreign keys
sqluser_id INTEGER NOT NULL REFERENCES users(id),     -- o que acontece se user for deletado?
product_id INTEGER NOT NULL REFERENCES products(id) -- e se produto for deletado?
Sem política explícita, o banco bloqueia o delete do pai. Defina intencionalmente:
sqluser_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT
RESTRICT é provavelmente o correto aqui — não deve ser possível deletar usuário ou produto com ordens ativas.

🟡 quantity sem CHECK
sqlquantity INTEGER NOT NULL
Quantidade zero ou negativa é aceita. Adicione:
sqlquantity INTEGER NOT NULL CHECK (quantity > 0)

🟡 total pode ficar dessincronizado do preço real
total é armazenado mas não há constraint garantindo que seja quantity * price. Uma migration não resolve isso sozinha, mas vale registrar que a aplicação precisa calcular e validar antes de inserir.

🟡 Sem updated_at
Status muda ao longo do tempo (pending → paid → refunded). Sem updated_at não há como saber quando a última mudança ocorreu.

🟢 Sem índices
sqlCREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
Buscas por usuário e por status são as mais comuns em e-commerce.


*/