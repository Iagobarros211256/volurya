# Volurya API

Backend API in Go for managing band-related products (t-shirts, caps, socks, drumsticks, etc).  
**Live deployment:** https://volurya.onrender.com

Personal learning project / showcase built to practice real-world backend concepts.

![Badge](https://img.shields.io/badge/Go-1.25-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Ready-blue)
![Badge](https://img.shields.io/badge/JWT-Authentication-blue)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)

## 🎯 Project Goals

- Practice **Clean Architecture** (simplified)
- Implement secure **JWT authentication** (bcrypt + validation)
- Work with real **PostgreSQL** (even in tests)
- Build proper **unit + integration tests**
- Standardize environment with **Docker**
- Create a clean, explainable showcase project

## 🌐 Live API (Render)

- Base URL: https://volurya.onrender.com
- Health check: https://volurya.onrender.com/ping → `{"message": "pong"}`

**Note**: Render free tier has cold starts — first request after ~15 min inactivity may take 10–30 seconds.

## 🧱 Architecture Overview

Simplified Clean Architecture layers:

- **Controllers** → HTTP layer (Gin) — request/response only
- **UseCases** → Business rules & invariants (e.g. password never stored plain)
- **Repositories** → Data access (PostgreSQL)
- **Models** → Domain entities

## 🚀 Running Locally (Docker)

1. Clone the repo
```bash
git clone https://github.com/Iagobarros211256/volurya.git
cd volurya


Start services:

docker compose up --build -d
API available at: http://localhost:8080

Stop everything:

Bash    docker compose down -v


🗄️ Database Access (Local)


docker exec -it volurya_postgres psql -U volurya -d volurya_db
Quick checks on database :

SQL\dt                  # list tables
SELECT * FROM user; # see users
SELECT * FROM products; # dosent exist (fix this problem now)



🔐 Authentication Flow (JWT)

Signup : (creates normal user – role: "user")

Bash  curl -X POST https://volurya.onrender.com/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'

Login : (get token)

Bash curl -X POST https://volurya.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'

Copy the access_token.

Protected route example:

Bash curl -X GET https://volurya.onrender.com/api/products \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"



🔌 Main Endpoints

Method      Endpoint                      Description                              HasAuth?
POST       /api/signup                Create user (role: user)                      No Auth 
POST       /api/login                 Generate   JWT token                          No
GET        /api/products              List products (cursor pagination)             Yes
POST       /api/products              Create product (with ownership)               Yes
GET        /api/products/:id          Get product by ID                             Yes
PUT        /api/products/:id          Update product                                Yes
DELETE     /api/products/:id          Delete product (ownership check)              Yes
GET        /ping                      Healthcheck                                   No


🧪 Tests
Bash# Start isolated test database
docker compose -f docker-compose.test.yml up -d

# Run all tests
go test ./... -v

# Coverage report
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out
Includes:

Unit tests (usecases with mocked repo)
Integration tests (HTTP + real PostgreSQL)
DB setup/cleanup helpers

🛠️ Architectural Decisions

Simplified Clean Architecture
JWT authentication (no refresh token yet)
Real PostgreSQL in integration tests
Ownership enforcement on products (user_id)
Docker as standard environment

❄️ Current Status
Backend frozen and deployed — fully achieves its learning/showcase goals.
Further features only if new context or real need arises.


🧠 Quick Mental Model
textSignup → creates user with role "user"
Login  → returns JWT
JWT    → sent in Authorization: Bearer ...
Middleware → validates token & injects user_id
UseCase → business rules
Repository → database



📬 Contact
Iago Barros
@Iagobarros2112
Fortaleza, Brazil – February 2026
Made with ❤️, anger and lots of coffee to learn and demonstrate backend thinking.



README.md em PORTUGUÊS 


# Volurya API

Backend em Go para gerenciar produtos da banda Volurya (camisetas, bonés, meias, baquetas, etc).  

**Deploy atual:** https://volurya.onrender.com

Projeto pessoal / laboratório de aprendizado para praticar conceitos reais de backend.

![Badge](https://img.shields.io/badge/Go-1.25-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Pronto-blue)
![Badge](https://img.shields.io/badge/JWT-Autenticação-blue)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)

## 🎯 Objetivos do Projeto

- Praticar **Clean Architecture** simplificada
- Implementar autenticação **JWT segura** (bcrypt + validação)
- Trabalhar com **PostgreSQL real** (inclusive nos testes)
- Fazer **testes reais** (unitários + integração)
- Padronizar ambiente com **Docker**
- Criar um projeto fácil de explicar e apresentar

## 🌐 API Online (Render)

- URL base: https://volurya.onrender.com
- Health check: https://volurya.onrender.com/ping → `{"message": "pong"}`

**Aviso**: Render free tier tem cold start — primeira requisição após ~15 min de inatividade pode demorar 10–30 segundos.

## 🧱 Arquitetura

Camadas separadas (Clean Architecture simplificada):

- **Controllers** → camada HTTP (Gin) — só entrada/saída
- **UseCases** → regras de negócio + invariantes (ex: senha nunca em plain text)
- **Repositories** → acesso ao banco (PostgreSQL)
- **Models** → entidades do domínio

## 🚀 Como Rodar Localmente (Docker)

Clone o repositório :

git clone https://github.com/Iagobarros211256/volurya.git
cd volurya

Suba os containers :

docker compose up --build -d
API disponível em: http://localhost:8080

Parar tudo :

docker compose down -v

🗄️ Acessando o Banco (Local):

docker exec -it volurya_postgres psql -U volurya -d volurya_db

Comandos rápidos:
\dt                  # lista tabelas
SELECT * FROM user; # ver usuários
SELECT * FROM products;

🔐 Fluxo de Autenticação (JWT)

Cadastro (cria usuário normal – role: "user")

curl -X POST https://volurya.onrender.com/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "fã@volurya.com", "password": "senhaSegura123"}'

Login (gera token)

curl -X POST https://volurya.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "fã@volurya.com", "password": "senhaSegura123"}'

Copie o access_token da resposta.

Rota protegida

curl -X GET https://volurya.onrender.com/api/products \
  -H "Authorization: Bearer SEU_TOKEN_AQUI"

🔌 Principais Endpoints

Metodo      Endpoint                      Descricao                              Autorizacao?
POST       /api/signup                Create user (role: user)                      nao
POST       /api/login                 Generate   JWT token                          nao
GET        /api/products              List products (cursor pagination)             sim
POST       /api/products              Create product (with ownership)               sim
GET        /api/products/:id          Get product by ID                             sim
PUT        /api/products/:id          Update product                                sim
DELETE     /api/products/:id          Delete product (ownership check)              sim
GET        /ping                      Healthcheck                                   nao

🧪 Testes
 Banco de teste isolado
docker compose -f docker-compose.test.yml up -d

# Rodar todos os testes
go test ./... -v

# Ver cobertura
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out

Inclui:

Testes unitários (usecase com repo mockado)
Testes de integração (HTTP + banco real)
Helpers de setup/cleanup do banco

🛠️ Decisões de Arquitetura

Clean Architecture simplificada
Autenticação JWT (sem refresh token por enquanto)
Banco real nos testes de integração
Ownership em produtos (user_id)
Docker como ambiente padrão

❄️ Status Atual
Backend finalizado e deployado no Render (free tier) — objetivos de aprendizado atingidos.
Novas features só se surgir necessidade real ou novo contexto.


🧠 Modelo Mental Rápido
Signup → cria usuário com role "user"
Login  → retorna JWT
JWT    → enviado no header Authorization: Bearer ...
Middleware → valida token e injeta user_id
UseCase → regras de negócio
Repository → banco


📬 Contato
Iago Barros
@Iagobarros2112
Fortaleza, Brasil – Fevereiro 2026
Feito com ❤️,raiva e muito café para aprender e mostrar como penso arquitetura backend.

