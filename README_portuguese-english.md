 Volurya API

Backend API in Go for managing band-related products (t-shirts, caps, socks, drumsticks, etc).  
**Live deployment:** https://volurya.onrender.com

Personal learning project / showcase built to practice real-world backend concepts.

![Badge](https://img.shields.io/badge/Go-1.25-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Ready-blue)
![Badge](https://img.shields.io/badge/JWT-Authentication-blue)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)


# Documentação Volurya API

## Visão Geral

- **[PRD](./volurya-prd.md)** — Requisitos de produto e escopo
- **[Tech Spec](./volurya-tech-spec.md)** — Arquitetura técnica e decisões
- **[ADRs](./architecture/)** — Architecture Decision Records

## ADRs

- [ADR 001 — Clean Architecture](./architecture/adr-001-clean-architecture.md)
- [ADR 002 — JWT com Refresh Token Rotation](./architecture/adr-002-jwt-refresh-rotation.md)
- [ADR 003 — Cloudflare R2 para Storage](./architecture/adr-003-cloudflare-r2.md)

## 🎯 Project Goals

- Practice **Clean Architecture** (simplified)
- Implement secure **JWT authentication** (bcrypt + refresh token rotation)
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

Clone the repo:

```bash
git clone https://github.com/Iagobarros211256/volurya.git
cd volurya
```

Create a `.env` file in the root:

```env
JWT_SECRET=your_secret_here
POSTGRES_HOST=volurya_postgres
POSTGRES_PORT=5432
POSTGRES_USER=volurya
POSTGRES_PASSWORD=volurya
POSTGRES_DB=volurya_db
PAGSEGURO_TOKEN=your_token_here
PAGSEGURO_EMAIL=your_email_here
PAGSEGURO_SANDBOX=true
PAGSEGURO_WEBHOOK_URL=http://localhost:8080/api/webhook
ACCESS_TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=7
```

Start services:

```bash
docker compose up --build -d
```

API available at: http://localhost:8080

Stop everything:

```bash
docker compose down -v
```

## 🗄️ Database Access (Local)

```bash
docker exec -it volurya_postgres psql -U volurya -d volurya_db
```

Tables are created automatically by migrations on startup. No manual SQL commands needed.

Quick checks:

```sql
\dt                    -- list tables
SELECT * FROM users;   -- see users
SELECT * FROM products;
```

## 🔐 Authentication Flow (JWT)

Signup or Login return two tokens:
- `access_token` — short-lived (15 min), used on protected routes
- `refresh_token` — long-lived (7 days), used to renew the access_token

**Signup:**

```bash
curl -X POST https://volurya.onrender.com/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'
```

**Login:**

```bash
curl -X POST https://volurya.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'
```

**Refresh (when access_token expires):**

```bash
curl -X POST https://volurya.onrender.com/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

**Logout (revokes refresh_token):**

```bash
curl -X POST https://volurya.onrender.com/api/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

**Protected route example:**

```bash
curl -X GET https://volurya.onrender.com/api/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🔌 Main Endpoints

| Method | Endpoint | Description | Auth? |
|--------|----------|-------------|-------|
| POST | /api/signup | Create user (role: user) | No |
| POST | /api/login | Generate JWT tokens | No |
| POST | /api/refresh | Generate new token pair via refresh_token | No |
| POST | /api/logout | Revoke refresh_token | No |
| GET | /api/products | List products (cursor pagination) | Yes |
| POST | /api/products | Create product (with ownership) | Yes |
| GET | /api/products/:productId | Get product by ID | Yes |
| PUT | /api/products/:productId | Update product | Yes |
| DELETE | /api/products/:productId | Delete product (ownership check) | Yes |
| GET | /ping | Healthcheck | No |

## 🧪 Tests

Start isolated test database:

```bash
docker compose -f docker-compose.test.yml up -d
```

Run all tests:

```bash
go test ./... -v
```

Coverage report:

```bash
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out
```

Includes:
- Unit tests (usecases with mocked repo)
- Integration tests (HTTP + real PostgreSQL)
- DB setup/cleanup helpers

## 🛠️ Architectural Decisions

- Simplified Clean Architecture
- JWT authentication with refresh token rotation
- Token duration configurable via environment variables
- Automatic migrations with golang-migrate
- Real PostgreSQL in integration tests
- Ownership enforcement on products (user_id)
- Docker as standard environment

## ❄️ Current Status

Backend in active development — refresh token, migrations and env-based config implemented.

## 🧠 Quick Mental Model
Signup/Login  → access_token (15min) + refresh_token (7 days)
access_token  → sent in Authorization: Bearer ...
Middleware    → validates token and injects user_id
UseCase       → business rules
Repository    → database
refresh_token → used in /api/refresh to renew access_token
logout        → revokes refresh_token in the database

## 📬 Contact

Iago Barros  
@Iagobarros2112  
Fortaleza, Brazil – 2026  
Made with ❤️, anger and lots of coffee to learn and demonstrate backend thinking.

---

# Volurya API (Português)

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
- Implementar autenticação **JWT segura** (bcrypt + refresh token com rotação)
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

Clone o repositório:

```bash
git clone https://github.com/Iagobarros211256/volurya.git
cd volurya
```

Crie um arquivo `.env` na raiz:

```env
JWT_SECRET=sua_chave_secreta_aqui
POSTGRES_HOST=volurya_postgres
POSTGRES_PORT=5432
POSTGRES_USER=volurya
POSTGRES_PASSWORD=volurya
POSTGRES_DB=volurya_db
PAGSEGURO_TOKEN=seu_token_aqui
PAGSEGURO_EMAIL=seu_email_aqui
PAGSEGURO_SANDBOX=true
PAGSEGURO_WEBHOOK_URL=http://localhost:8080/api/webhook
ACCESS_TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=7
```

Suba os containers:

```bash
docker compose up --build -d
```

API disponível em: http://localhost:8080

Parar tudo:

```bash
docker compose down -v
```

## 🗄️ Acessando o Banco (Local)

```bash
docker exec -it volurya_postgres psql -U volurya -d volurya_db
```

As tabelas são criadas automaticamente pelas migrations ao iniciar a API. Nenhum comando SQL manual é necessário.

Comandos rápidos:

```sql
\dt                    -- lista tabelas
SELECT * FROM users;   -- ver usuários
SELECT * FROM products;
```

## 🔐 Fluxo de Autenticação (JWT)

Signup ou Login retornam dois tokens:
- `access_token` — curto prazo (15 min), usado nas rotas protegidas
- `refresh_token` — longo prazo (7 dias), usado para renovar o access_token

**Cadastro:**

```bash
curl -X POST https://volurya.onrender.com/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'
```

**Login:**

```bash
curl -X POST https://volurya.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'
```

**Refresh (quando o access_token expirar):**

```bash
curl -X POST https://volurya.onrender.com/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "SEU_REFRESH_TOKEN"}'
```

**Logout (revoga o refresh_token):**

```bash
curl -X POST https://volurya.onrender.com/api/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "SEU_REFRESH_TOKEN"}'
```

**Rota protegida:**

```bash
curl -X GET https://volurya.onrender.com/api/products \
  -H "Authorization: Bearer SEU_ACCESS_TOKEN"
```

## 🔌 Principais Endpoints

| Método | Endpoint | Descrição | Auth? |
|--------|----------|-----------|-------|
| POST | /api/signup | Cria usuário (role: user) | Não |
| POST | /api/login | Gera tokens JWT | Não |
| POST | /api/refresh | Gera novo par de tokens via refresh_token | Não |
| POST | /api/logout | Revoga o refresh_token | Não |
| GET | /api/products | Lista produtos (paginação por cursor) | Sim |
| POST | /api/products | Cria produto (com ownership) | Sim |
| GET | /api/products/:productId | Busca produto por ID | Sim |
| PUT | /api/products/:productId | Atualiza produto | Sim |
| DELETE | /api/products/:productId | Deleta produto (verifica ownership) | Sim |
| GET | /ping | Healthcheck | Não |
| GET | /api/cart | Ver carrinho atual | Sim |
| POST | /api/cart/items | Adicionar produto ao carrinho | Sim |
| PUT | /api/cart/items/:itemId | Atualizar quantidade do item | Sim |
| DELETE | /api/cart/items/:itemId | Remover item do carrinho | Sim |
| POST | /api/cart/checkout | Finalizar compra e gerar links de pagamento | Sim |

## 🧪 Testes

Banco de teste isolado:

```bash
docker compose -f docker-compose.test.yml up -d
```

Rodar todos os testes:

```bash
go test ./... -v
```

Ver cobertura:

```bash
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out
```

Inclui:
- Testes unitários (usecase com repo mockado)
- Testes de integração (HTTP + banco real)
- Helpers de setup/cleanup do banco

## 🛠️ Decisões de Arquitetura

- Clean Architecture simplificada
- Refresh token com rotação e revogação no banco
- Duração dos tokens configurável via variáveis de ambiente
- Migrations automáticas com golang-migrate
- Banco real nos testes de integração
- Ownership em produtos (user_id)
- Docker como ambiente padrão
- Carrinho persistido no banco (persiste entre sessões)
- Upsert no AddItem (adicionar produto já no carrinho incrementa quantidade)
- Ownership enforcement nos itens do carrinho
- Carrinho limpo automaticamente após checkout

## ❄️ Status Atual

Backend em evolução ativa — carrinho de compras, refresh token, migrations e configuração via env implementados.

## 🧠 Modelo Mental Rápido
Signup/Login  → access_token (15min) + refresh_token (7 dias)
access_token  → enviado no header Authorization: Bearer ...
Middleware    → valida token e injeta user_id
UseCase       → regras de negócio
Repository    → banco
refresh_token → usado em /api/refresh para renovar o access_token
logout        → revoga o refresh_token no banco

## 📬 Contato

Iago Barros  
@Iagobarros2112  
Fortaleza, Brasil – 2026  
Feito com ❤️, raiva e muito café para aprender e mostrar como penso arquitetura backend.



## 🗺️ Roadmap

### ✅ Concluído
- [x] Clean Architecture (controller → usecase → repository)
- [x] JWT authentication com bcrypt
- [x] CRUD de produtos com ownership
- [x] Paginação por cursor
- [x] Refresh token com rotação e revogação no banco
- [x] Logout com invalidação do token
- [x] Duração dos tokens configurável via env
- [x] Migrations automáticas com golang-migrate
- [x] Testes unitários e de integração
- [x] Docker + deploy no Render
- [x] Carrinho de compras com persistência no banco
- [x] Upload de imagem para produtos (Cloudflare R2)
- [x] Logs estruturados com slog (JSON em produção, texto em desenvolvimento)
- [x] Request ID por requisição
- [x] user_id nos logs de rotas autenticadas

### 🔜 Features
- [ ] Verificação de estoque no CreateOrder
- [ ] Webhook do PagSeguro (atualizar status do pedido para `paid`)
- [ ] Guard de role `admin` nas rotas de criação/deleção de produtos
- [ ] Cache de produtos com Redis
- [ ] Compressão gzip nas respostas

### 🔒 Segurança
- [ ] Rate limiting nas rotas públicas (signup, login, refresh)
- [ ] CORS configurado corretamente
- [ ] Helmet headers (X-Content-Type-Options, X-Frame-Options, etc.)
- [ ] Validação de input robusta com `go-playground/validator`
- [ ] Sanitização de input
- [ ] Expiração automática de refresh tokens no banco

### 📊 Observabilidade e Performance
- [x] Logs estruturados com `slog`
- [x] Request ID por requisição
- [x] Métricas com Prometheus (`/metrics`)
- [ ] Índices no banco (`token` em refresh_tokens, `user_id` em products)
- [ ] Compressão gzip nas respostas
- [ ] Connection pool configurável via env
- [ ] Health check detalhado (banco, R2, etc.)
- [ ] Cache de produtos com Redis


### 🔧 Concorrência e Performance (Go avançado)
- [ ] Worker pool para processamento de imagens (goroutines + channels)
- [ ] Pipeline de processamento de imagem assíncrono (validar → redimensionar → comprimir → salvar no R2)
- [ ] SSE (Server-Sent Events) para notificações em tempo real ao frontend

### 💡 Por que essas features?
Essas três implementações foram escolhidas para demonstrar o que Go faz melhor:
- **Worker pool** — controle fino de concorrência com goroutines e channels
- **Pipeline** — padrão produtor/consumidor encadeado, idiomático em Go
- **SSE** — Go lida naturalmente com milhares de conexões longas e concorrentes
Juntas formam um sistema de processamento de imagem assíncrono profissional.


### 🎨 Frontend
- [x] Página do carrinho visual com listagem de itens e total
- [x] Página "About" com bio completa e galeria de fotos com lightbox
- [ ] Botão de logout no navbar quando usuário estiver logado
- [ ] Badge de quantidade no ícone do carrinho no navbar
- [ ] Redirecionamento inteligente — login redireciona de volta para a página anterior
- [ ] Usar `image_url` real do produto na store (hoje está hardcoded)
- [ ] Skeleton loading nos cards de produto
- [ ] Filtro e busca de produtos na store
- [ ] Página de confirmação de pedido após checkout
- [ ] Página 404 customizada com identidade visual da banda



## 📚 Documentação

- [PRD](./docs/volurya-prd.md) — Requisitos e visão de produto
- [Tech Spec](./docs/volurya-tech-spec.md) — Arquitetura e decisões técnicas
- [ADRs](./docs/architecture/) — Architecture Decision Records