CREATE TABLE IF NOT EXISTS users (
    --Emails são case-insensitive por padrão mas UNIQUE em TEXT 
    --é case-sensitive no PostgreSQL. "User@example.com" e 
    --"user@example.com" seriam considerados diferentes. 
    --Normalize na aplicação (.ToLower()) ou use índice funcional
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password VARCHAR(60) NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);