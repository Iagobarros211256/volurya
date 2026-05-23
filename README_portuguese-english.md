# Volurya API

Backend API in Go for managing band-related products (t-shirts, caps, socks, drumsticks, etc).  
**Live deployment:** https://volurya.onrender.com

Personal learning project / showcase built to practice real-world backend concepts.

![Badge](https://img.shields.io/badge/Go-1.23-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Ready-blue)
![Badge](https://img.shields.io/badge/JWT-Authentication-blue)
![Badge](https://img.shields.io/badge/Stripe-Payments-blue?logo=stripe&logoColor=white)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)

# Documentação Volurya API

## Visão Geral

- **[PRD](./docs/volurya-prd.md)** — Requisitos de produto e escopo
- **[Tech Spec](./docs/volurya-tech-spec.md)** — Arquitetura técnica e decisões
- **[ADRs](./docs/architecture/)** — Architecture Decision Records

## ADRs

- [ADR 001 — Clean Architecture](./docs/architecture/adr-001-clean-architecture.md)
- [ADR 002 — JWT com Refresh Token Rotation](./docs/architecture/adr-002-jwt-refresh-rotation.md)
- [ADR 003 — Cloudflare R2 para Storage](./docs/architecture/adr-003-cloudflare-r2.md)

## 🎯 Project Goals

- Practice **Clean Architecture** (simplified)
- Implement secure **JWT authentication** (bcrypt + refresh token rotation)
- Work with real **PostgreSQL** (even in tests)
- Build proper **unit + integration tests**
- Standardize environment with **Docker**
- Create a clean, explainable showcase project
- **Security hardening** (email validation, rate limiting, CSRF protection, webhook validation)
- **Full API documentation** (Swagger/OpenAPI)
- **Enterprise-ready features** (health checks, request tracing, monitoring)

## 🌐 Live API (Render)

- Base URL: https://volurya.onrender.com
- Health check: https://volurya.onrender.com/ping → `{"message": "pong"}`
- **Swagger UI**: https://volurya.onrender.com/swagger/index.html (development only)

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

Create a `.env` file in the root (see `.env.example`):

```env
APP_ENV=development

JWT_SECRET=your_secret_here_min_32_chars
ACCESS_TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=7

POSTGRES_HOST=volurya_postgres
POSTGRES_PORT=5432
POSTGRES_USER=volurya
POSTGRES_PASSWORD=your_postgres_password
POSTGRES_DB=volurya_db

STRIPE_SECRET_KEY=sk_test_your_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_secret_here

R2_ACCESS_KEY_ID=your_r2_key
R2_SECRET_ACCESS_KEY=your_r2_secret
R2_ACCOUNT_ID=your_account_id
R2_BUCKET_NAME=your_bucket_name
R2_PUBLIC_URL=https://your-bucket.r2.dev

BOOTSTRAP_ADMIN_EMAIL=admin@volurya.com
BOOTSTRAP_ADMIN_PASSWORD=troque-esta-senha

GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your_grafana_password
```

`BOOTSTRAP_ADMIN_EMAIL` e `BOOTSTRAP_ADMIN_PASSWORD` são opcionais. Quando definidos, a API cria ou atualiza esse usuário com `role=admin` no startup.

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
\dt                      -- list tables
SELECT * FROM users;     -- see users
SELECT * FROM products;
SELECT * FROM order_items;
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
| GET | /api/health | Healthcheck with DB status (200/503) | No |
| POST | /api/webhook | Stripe webhook (signature validated) | No |
| GET | /api/products | List products (cursor pagination) | Yes |
| POST | /api/products | Create product (admin only) | Yes |
| GET | /api/products/:productId | Get product by ID | Yes |
| PUT | /api/products/:productId | Update product (admin only) | Yes |
| DELETE | /api/products/:productId | Delete product (admin only) | Yes |
| POST | /api/checkout | Create Stripe payment intent | Yes |
| GET | /api/cart | View cart | Yes |
| POST | /api/cart/items | Add item to cart | Yes |
| PUT | /api/cart/items/:itemId | Update item quantity | Yes |
| DELETE | /api/cart/items/:itemId | Remove item | Yes |
| POST | /api/cart/checkout | Checkout cart | Yes |
| GET | /api/events | SSE real-time notifications | Yes |
| GET | /ping | Simple healthcheck | No |
| GET | /swagger/\*any | Swagger UI (dev only) | No |

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

## 🛠️ Architectural Decisions

- Simplified Clean Architecture
- JWT authentication with refresh token rotation
- Token duration configurable via environment variables
- Automatic migrations with golang-migrate
- Stripe for payments (webhook with signature validation)
- Cloudflare R2 for image storage
- Docker as standard environment
- Prometheus + Grafana for monitoring
- SSE for real-time notifications

## ❄️ Status Atual

Backend em desenvolvimento ativo.

### ✅ Semana 0 — Infraestrutura e correções críticas
- [x] Credenciais removidas do docker-compose.yml
- [x] Bug de compilação do main.go corrigido
- [x] Swagger desabilitado em produção
- [x] /metrics protegido por IP
- [x] log.Fatalf substituído por slog
- [x] Shutdown timeout aumentado para 15s
- [x] Prometheus coletando métricas corretamente

### ✅ Semana 1 — Segurança financeira
- [x] Webhook Stripe com validação real de assinatura
- [x] PagSeguro removido completamente
- [x] order_items persistidos no banco
- [x] Estoque decrementado atomicamente (race condition protegida)
- [x] CancelOrder para compensação de ordens órfãs
- [x] HandlePaymentSuccess/Failed idempotentes
- [x] JWT revalidado manualmente removido do checkout
- [x] math.Round para conversão de centavos

### 🔜 Semana 2 — Segurança geral
- [ ] Tokens em cookies HttpOnly (XSS protection)
- [ ] innerHTML → textContent no frontend (XSS prevention)
- [ ] Interfaces nos usecases e repositories
- [ ] Race condition e deadlock no rate limiter
- [ ] request_id.go e logger.go unificados

### 🔜 Semana 3 — Qualidade
- [ ] go:embed para migrations
- [ ] Testes reais nos controllers e usecases
- [ ] Dois painéis admin unificados
- [ ] image_url real na store

## 🔒 Segurança

- [x] Rate limiting nas rotas públicas (signup, login, refresh)
- [x] CORS configurado por ambiente
- [x] Email validation (RFC 5321)
- [x] Password validation (8-128 chars)
- [x] CSRF protection middleware
- [x] HTTPS redirect em produção
- [x] Stripe webhook signature validation
- [x] /metrics protegido por IP
- [x] Swagger desabilitado em produção
- [ ] Tokens em cookies HttpOnly
- [ ] Helmet headers

## 📊 Observabilidade

- [x] Logs estruturados com slog (JSON em produção)
- [x] Request ID por requisição (UUID)
- [x] Métricas com Prometheus (/metrics — acesso restrito)
- [x] Grafana para visualização
- [x] Health check com status do banco (200/503)

### 🔧 Concorrência e Performance (Go avançado)

- [x] Worker pool para processamento de imagens (goroutines + channels)
- [x] Pipeline assíncrono (validar → redimensionar → comprimir → salvar no R2)
- [x] SSE (Server-Sent Events) para notificações em tempo real

### 💡 Por que essas features?

Essas três implementações foram escolhidas para demonstrar o que Go faz melhor:
- **Worker pool** — controle fino de concorrência com goroutines e channels
- **Pipeline** — padrão produtor/consumidor encadeado, idiomático em Go
- **SSE** — Go lida naturalmente com milhares de conexões longas e concorrentes

## 🧠 Quick Mental Model
Signup/Login  → access_token (15min) + refresh_token (7 days)
access_token  → sent in Authorization: Bearer ...
Middleware    → validates token and injects user_id
UseCase       → business rules
Repository    → database
refresh_token → used in /api/refresh to renew access_token
logout        → revokes refresh_token in the database
Stripe        → payment intent created on checkout
webhook validates signature → updates order status

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

![Badge](https://img.shields.io/badge/Go-1.23-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Pronto-blue)
![Badge](https://img.shields.io/badge/JWT-Autenticação-blue)
![Badge](https://img.shields.io/badge/Stripe-Pagamentos-blue?logo=stripe&logoColor=white)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)

## 🎯 Objetivos do Projeto

- Praticar **Clean Architecture** simplificada
- Implementar autenticação **JWT segura** (bcrypt + refresh token com rotação)
- Trabalhar com **PostgreSQL real** (inclusive nos testes)
- Fazer **testes reais** (unitários + integração)
- Padronizar ambiente com **Docker**
- Criar um projeto fácil de explicar e apresentar
- **Hardening de segurança** (validação de email, rate limiting, validação de webhook)
- **Documentação completa** (Swagger/OpenAPI)
- **Features production-ready** (healthchecks, rastreamento, monitoring)

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

Crie um arquivo `.env` na raiz (veja `.env.example`):

```env
APP_ENV=development

JWT_SECRET=sua_chave_secreta_minimo_32_chars
ACCESS_TOKEN_DURATION_MINUTES=15
REFRESH_TOKEN_DURATION_DAYS=7

POSTGRES_HOST=volurya_postgres
POSTGRES_PORT=5432
POSTGRES_USER=volurya
POSTGRES_PASSWORD=sua_senha_postgres
POSTGRES_DB=volurya_db

STRIPE_SECRET_KEY=sk_test_sua_chave
STRIPE_WEBHOOK_SECRET=whsec_seu_secret

R2_ACCESS_KEY_ID=sua_chave_r2
R2_SECRET_ACCESS_KEY=seu_secret_r2
R2_ACCOUNT_ID=seu_account_id
R2_BUCKET_NAME=nome_do_bucket
R2_PUBLIC_URL=https://seu-bucket.r2.dev

BOOTSTRAP_ADMIN_EMAIL=admin@volurya.com
BOOTSTRAP_ADMIN_PASSWORD=troque-esta-senha

GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=sua_senha_grafana
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

As tabelas são criadas automaticamente pelas migrations ao iniciar a API.

```sql
\dt                        -- lista tabelas
SELECT * FROM users;       -- ver usuários
SELECT * FROM products;
SELECT * FROM order_items; -- itens de cada pedido
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

**Refresh:**

```bash
curl -X POST https://volurya.onrender.com/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "SEU_REFRESH_TOKEN"}'
```

**Logout:**

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
| POST | /api/refresh | Renova tokens via refresh_token | Não |
| POST | /api/logout | Revoga o refresh_token | Não |
| GET | /api/health | Healthcheck com status do BD | Não |
| POST | /api/webhook | Webhook Stripe (assinatura validada) | Não |
| GET | /api/products | Lista produtos (paginação por cursor) | Sim |
| POST | /api/products | Cria produto (admin only) | Sim |
| GET | /api/products/:productId | Busca produto por ID | Sim |
| PUT | /api/products/:productId | Atualiza produto (admin only) | Sim |
| DELETE | /api/products/:productId | Deleta produto (admin only) | Sim |
| POST | /api/checkout | Cria payment intent no Stripe | Sim |
| GET | /api/cart | Ver carrinho atual | Sim |
| POST | /api/cart/items | Adicionar produto ao carrinho | Sim |
| PUT | /api/cart/items/:itemId | Atualizar quantidade | Sim |
| DELETE | /api/cart/items/:itemId | Remover item | Sim |
| POST | /api/cart/checkout | Finalizar compra | Sim |
| GET | /api/events | Stream SSE de notificações | Sim |
| GET | /ping | Healthcheck simples | Não |

## 🧪 Testes

```bash
docker compose -f docker-compose.test.yml up -d
go test ./... -v
go test ./... -coverprofile=cover.out && go tool cover -html=cover.out
```

## 🛠️ Decisões de Arquitetura

- Clean Architecture simplificada
- Stripe para pagamentos com validação de webhook
- Refresh token com rotação e revogação no banco
- Duração dos tokens configurável via env
- Migrations automáticas com golang-migrate
- Cloudflare R2 para storage de imagens
- Docker como ambiente padrão
- Prometheus + Grafana para monitoring

## 🧠 Modelo Mental Rápido
Signup/Login  → access_token (15min) + refresh_token (7 dias)
access_token  → enviado no header Authorization: Bearer ...
Middleware    → valida token e injeta user_id
UseCase       → regras de negócio
Repository    → banco
refresh_token → usado em /api/refresh para renovar o access_token
logout        → revoga o refresh_token no banco
Stripe        → payment intent criado no checkout
webhook valida assinatura → atualiza status da ordem

## 📬 Contato

Iago Barros  
@Iagobarros2112  
Fortaleza, Brasil – 2026  
Feito com ❤️, raiva e muito café para aprender e mostrar como penso arquitetura backend.

## 📚 Documentação

- [PRD](./docs/volurya-prd.md) — Requisitos e visão de produto
- [Tech Spec](./docs/volurya-tech-spec.md) — Arquitetura e decisões técnicas
- [ADRs](./docs/architecture/) — Architecture Decision Records