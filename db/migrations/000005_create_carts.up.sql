CREATE TABLE IF NOT EXISTS carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

/*


Sem updated_at
Útil para saber quando o carrinho foi modificado pela última vez — especialmente para limpeza de carrinhos abandonados:
sqlupdated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

🟡 Sem limpeza de carrinhos abandonados
Não há expires_at nem mecanismo de expiração. Carrinhos antigos acumulam indefinidamente. Considere:
sqlexpires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 days')
E um job periódico:
sqlDELETE FROM carts WHERE expires_at < NOW();

🟢 Tabela muito enxuta — considere status
Para suportar carrinhos abandonados vs ativos vs convertidos em ordem:
sqlstatus TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'converted', 'abandoned



*/