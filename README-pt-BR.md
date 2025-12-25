Volurya API é um laboratório pessoal para simular práticas usadas em ambientes de produção — como JWT authentication, Clean Architecture, Docker, testes automatizados e CI/CD — aplicadas a um e-commerce simples de produtos relacionados à banda fictícia Volurya.

O projeto foi desenvolvido com foco em aprendizado prático, clareza arquitetural e testes reais, não como produto comercial ativo.

🎯 Objetivo do Projeto

Praticar arquitetura em camadas (Clean Architecture simplificada)

Implementar autenticação real com JWT

Trabalhar com banco PostgreSQL real (inclusive em testes)

Diferenciar testes unitários de integração

Usar Docker para padronizar ambiente

Criar um projeto explicável e apresentável como showcase

🧱 Arquitetura (Resumo)

Controller → camada HTTP (Gin)

UseCase → regras de negócio

Repository → acesso a dados

Models → entidades do domínio

Regras críticas (como hash de senha e validações) vivem no UseCase, não no controller nem no banco.

🐳 Subindo o Projeto com Docker
sudo docker compose up --build
sudo docker start volurya
sudo docker start volurya_postgres


Verificar containers ativos:

sudo docker ps

🗄️ Acessando o Banco de Dados (PostgreSQL)

Entrar no container do banco:

sudo docker exec -it volurya_postgres psql -U volurya -d volurya_db


Consultar dados:

SELECT * FROM product;
SELECT * FROM users;

🧾 Estrutura SQL (Exemplo)
Tabela de produtos
CREATE TABLE IF NOT EXISTS product (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  description VARCHAR(200) NOT NULL,
  price NUMERIC(10,2) NOT NULL,
  stock INTEGER DEFAULT 0 NOT NULL
);

Inserção de dados de exemplo
INSERT INTO product (name, description, price, stock)
VALUES
('Camisa volurya', 'camiseta baby look m', 100.00, 25),
('Camisa volurya', 'camiseta machao m', 100.00, 40),
('Bone', 'aba aberta', 300.00, 10),
('Meias volurya', 'meias alien m', 200.00, 25);


Resultado esperado:

 id |      name      |     description      | price  | stock
----+----------------+----------------------+--------+-------
  1 | Camisa volurya | camiseta baby look m | 100.00 |    25
  2 | Camisa volurya | camiseta machao m    | 100.00 |    40
  3 | Bone           | aba aberta           | 300.00 |    10
  4 | Meias volurya  | meias alien m        | 200.00 |    25

🔌 Testando a API (Postman / cURL)
Verificar se a API está ativa
GET http://localhost:8000/ping


Resposta esperada:

{ "message": "pong" }

🔐 Fluxo Completo de Autenticação (JWT)
1️⃣ Criar usuário (Signup)
POST http://localhost:8000/signup

{
  "email": "admin@volurya.com",
  "password": "123456"
}


Resposta esperada:

{ "message": "user created" }


Verificar no banco:

SELECT * FROM users;

2️⃣ Login (Gerar JWT)
POST http://localhost:8000/login

{
  "email": "admin@volurya.com",
  "password": "123456"
}


Resposta esperada:

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}


Copie o token sem aspas.

3️⃣ Testar rota protegida (SEM token)
GET http://localhost:8000/products


Resposta esperada:

{ "error": "unauthorized" }


Se não falhar, o middleware JWT não está ativo.

4️⃣ Testar rota protegida (COM token)
GET http://localhost:8000/products


Headers:

Authorization: Bearer SEU_TOKEN_AQUI


Resposta esperada:

Status 200

Lista de produtos

5️⃣ Criar produto (rota protegida)
POST http://localhost:8000/products


Headers:

Authorization: Bearer SEU_TOKEN
Content-Type: application/json


Body:

{
  "name": "Camiseta Volurya",
  "price": 99.90,
  "stock": 10
}

🧪 Testes Automatizados

Subir banco de testes:

docker compose -f docker-compose.test.yml up -d


Rodar testes:

go test ./...


Os testes cobrem:

regras de negócio (unitários)

fluxo HTTP + banco real (integração)

setup e limpeza de banco

📌 Decisões Conscientes

Clean Architecture simplificada

JWT para simular ambiente real

Banco real em testes (sem mocks excessivos)

Sem DTOs (evitar overengineering)

Monolito por simplicidade

Projeto congelado conscientemente

❄️ Status do Projeto

Projeto congelado.
O escopo atual cumpre totalmente seu objetivo educacional e de showcase.

🧠 Mapa Mental do JWT
POST /signup   → cria usuário
POST /login    → gera JWT
JWT            → vai no header Authorization
Middleware     → valida token
UseCase        → regra de negócio
Repository     → banco