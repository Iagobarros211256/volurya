CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


/*

token TEXT sem limite de tamanho
JWTs têm tamanho previsível. Limitar evita inserções abusivas:
sqltoken VARCHAR(512) UNIQUE NOT NULL

🟡 Sem índice em user_id
Buscar tokens por usuário (para revogar todos na troca de senha, por exemplo) seria um full scan:
sqlCREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

🟡 Tokens expirados acumulam indefinidamente
Não há mecanismo de limpeza. Com o tempo a tabela cresce sem limite. Opções:
sql-- Job periódico na aplicação:
DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE;

-- Ou particionamento por data para tabelas muito grandes

🟢 Sem índice composto para o caso de uso mais comum
A query mais frequente provavelmente é WHERE token = ? AND revoked = FALSE AND expires_at > NOW(). Um índice parcial ajuda:
sqlCREATE INDEX idx_refresh_tokens_active ON refresh_tokens (token)
WHERE revoked = FALSE;


*/