# Volurya API

Backend API in Go for managing band-related products (t-shirts, caps, socks, drumsticks, etc).  
**Live deployment:** https://volurya.onrender.com

Personal learning project / showcase built to practice real-world backend concepts.

![Badge](https://img.shields.io/badge/Go-1.24-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Ready-blue)
![Badge](https://img.shields.io/badge/JWT-Authentication-blue)
![Badge](https://img.shields.io/badge/Stripe-Payments-blue?logo=stripe&logoColor=white)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)
![Badge](https://img.shields.io/badge/CI-GitHub_Actions-success?logo=githubactions&logoColor=white)

## 📚 Documentação

- **[PRD](./docs/volurya-prd.md)** — Requisitos de produto e escopo
- **[Tech Spec](./docs/volurya-tech-spec.md)** — Arquitetura técnica e decisões
- **[Docs Index](./docs/index.md)** — Índice completo da documentação
- **[ADRs](./docs/architecture/)** — Architecture Decision Records

## 🎯 Project Goals

- Practice **Clean Architecture** (simplified)
- Implement secure **JWT authentication** (bcrypt + HttpOnly cookies + refresh token rotation)
- Work with real **PostgreSQL** (even in tests)
- Build proper **unit + integration tests**
- Standardize environment with **Docker**
- Create a clean, explainable showcase project
- **Security hardening** (XSS prevention, rate limiting, CSRF, webhook validation, HttpOnly cookies)
- **Full payment flow** (Stripe Payment Element, webhook, order confirmation)
- **Enterprise-ready features** (health checks, request tracing, monitoring, CI pipeline)

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
STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
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

`BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` are optional. When set, the API creates or updates that user with `role=admin` on startup.

Start services:

```bash
docker compose -f local/docker-compose.yml --env-file .env up --build -d
```

API available at: http://localhost:8080

Stop everything:

```bash
docker compose -f local/docker-compose.yml --env-file .env down -v
```

## 🗄️ Database Access (Local)

```bash
docker exec -it volurya_postgres psql -U volurya -d volurya_db
```

Tables are created automatically by migrations on startup. No manual SQL commands needed.

```sql
\dt                      -- list tables
SELECT * FROM users;
SELECT * FROM products;
SELECT * FROM orders;
SELECT * FROM order_items;
```

## 🔐 Authentication Flow (JWT + HttpOnly Cookies)

Login and Signup set tokens as **HttpOnly cookies** — JavaScript cannot access them, protecting against XSS.

```bash
# Login
curl -c /tmp/cookies.txt -X POST https://volurya.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "fan@volurya.com", "password": "senhaSegura123"}'

# Protected route — cookie sent automatically
curl -b /tmp/cookies.txt https://volurya.onrender.com/api/products

# Check auth status
curl -b /tmp/cookies.txt https://volurya.onrender.com/api/auth/me

# Logout — revokes refresh token
curl -b /tmp/cookies.txt -X POST https://volurya.onrender.com/api/logout
```

## 💳 Payment Flow (Stripe)
User adds items to cart

↓

POST /api/checkout → creates order + Stripe PaymentIntent

↓

Frontend receives client_secret → Stripe.js Payment Element

↓

User completes payment on /checkout page

↓

Stripe sends webhook → POST /api/webhook (HMAC signature validated)

↓

Order status updated → paid

↓

User redirected to /order/success

## 🔌 Main Endpoints

| Method | Endpoint | Description | Auth? |
|--------|----------|-------------|-------|
| POST | /api/signup | Create user (role: user) | No |
| POST | /api/login | Set HttpOnly auth cookies | No |
| POST | /api/refresh | Renew tokens via cookie | No |
| POST | /api/logout | Revoke refresh token + clear cookies | No |
| GET | /api/auth/me | Check auth status | Yes |
| GET | /api/health | Healthcheck with DB status | No |
| GET | /api/config | Stripe publishable key | No |
| POST | /api/webhook | Stripe webhook (HMAC validated) | No |
| GET | /api/products | List products (cursor pagination) | Yes |
| POST | /api/products | Create product (admin only) | Yes |
| GET | /api/products/:id | Get product by ID | Yes |
| PUT | /api/products/:id | Update product (admin only) | Yes |
| DELETE | /api/products/:id | Delete product (admin only) | Yes |
| POST | /api/checkout | Create Stripe PaymentIntent | Yes |
| GET | /api/cart | View cart | Yes |
| POST | /api/cart/items | Add item to cart | Yes |
| PUT | /api/cart/items/:id | Update item quantity | Yes |
| DELETE | /api/cart/items/:id | Remove item | Yes |
| GET | /api/events | SSE real-time notifications | Yes |
| GET | /ping | Simple healthcheck | No |

## 🧪 Tests

```bash
# Unit tests (no external dependencies)
JWT_SECRET=test_secret go test ./middleware/... -v

# Integration tests (requires test DB on port 5433)
docker compose -f local/docker-compose.test.yml up -d
go test ./... -v

# Coverage
go test ./middleware/... -coverprofile=cover.out
go tool cover -html=cover.out
```

CI runs on every push and PR via GitHub Actions — see `.github/workflows/ci.yml`.

## 📁 Repository Structure
.github/workflows/   # CI/CD — GitHub Actions

cmd/                 # Application entrypoint

config/              # Configuration (Stripe, env vars)

controller/          # HTTP handlers

db/                  # Migrations (embedded via go:embed)

docs/                # PRD, Tech Spec, ADRs

local/               # Local dev resources (docker-compose, prometheus)

middleware/          # JWT auth, rate limiting, CORS, CSRF, logging

models/              # Domain entities

repository/          # Database access layer

usecase/             # Business logic

views/               # HTML templates + static assets

## 🛠️ Architectural Decisions

- Simplified Clean Architecture
- JWT with HttpOnly cookies — tokens inaccessible to JavaScript
- Refresh token rotation and revocation in database
- Stripe Payment Element for checkout (PCI compliant)
- Webhook HMAC signature validation — prevents fraudulent payment confirmations
- Atomic stock decrement — `WHERE stock >= quantity` prevents overselling
- Migrations embedded in binary via `go:embed` — no filesystem dependencies
- Cloudflare R2 for image storage — no egress fees
- Async image processing via worker pool
- Prometheus + Grafana for monitoring
- SSE for real-time notifications

## ❄️ Status

Active development — hardened and working end-to-end payment flow.

### ✅ Semana 0 — Infrastructure
- [x] Credentials removed from docker-compose.yml
- [x] main.go compilation bug fixed
- [x] Swagger disabled in production
- [x] /metrics protected by IP
- [x] log.Fatalf replaced with slog
- [x] Graceful shutdown timeout 15s
- [x] Prometheus scraping fixed

### ✅ Semana 1 — Financial security
- [x] Stripe webhook with real HMAC signature validation
- [x] PagSeguro removed completely
- [x] order_items persisted in database
- [x] Atomic stock decrement (race condition protected)
- [x] CancelOrder for orphan order compensation
- [x] HandlePaymentSuccess/Failed idempotent

### ✅ Semana 2 — General security
- [x] Tokens migrated to HttpOnly cookies
- [x] innerHTML → textContent in all JS (XSS prevention)
- [x] /api/auth/me endpoint
- [x] JWT middleware reads cookie first, Authorization header as fallback
- [x] Token removed from SSE query string
- [x] Automatic token refresh on 401
- [x] Rate limiter: sync.Map eliminates race condition and deadlock
- [x] request_id.go and logger.go unified with UUID

### ✅ Semana 3 — Quality
- [x] go:embed for migrations — no filesystem dependency
- [x] Dockerfile upgraded to golang:1.24-alpine with HEALTHCHECK
- [x] Admin panels unified (storeadmin.html deleted)
- [x] image_url from API used in store (was hardcoded)

### ✅ Semana 4 — Complete Stripe payment flow
- [x] GET /api/config returns publishable key
- [x] checkout.html with Stripe.js Payment Element (night theme)
- [x] order-success.html with payment status verification
- [x] Migration 000010: drop NOT NULL from legacy orders columns
- [x] StatementDescriptor → StatementDescriptorSuffix fix
- [x] End-to-end payment tested with real Stripe test keys

### ✅ Semana 5 — Repository reorganization
- [x] docker-compose.yml moved to local/
- [x] prometheus.yml moved to local/
- [x] MOBILE_TESTING_GUIDE.md moved to docs/
- [x] docs/index.md created
- [x] .github/workflows/ directory created
- [x] Compiled binary removed from repo and added to .gitignore
- [x] README updated with new paths

### ✅ Semana 6 — CI/CD
- [x] GitHub Actions pipeline on push and PR to main
- [x] Build validation in CI
- [x] Unit tests running in CI (middleware + usecase)
- [x] Coverage report generated

### ✅ Hardening — Production readiness
- [x] Bootstrap admin uses DO NOTHING — password never overwritten on restart
- [x] Refresh token cleanup job — expired/revoked tokens deleted every 24h
- [x] 409 Conflict for duplicate email signup (was 500)
- [x] ErrEmailAlreadyExists sentinel error propagated correctly
- [x] Security headers middleware — X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Referrer-Policy, Permissions-Policy
- [x] CSRF protection for JSON requests — Double Submit Cookie Pattern
- [x] RefreshTokenRepositoryInterface — usecase decoupled from concrete type
- [x] usecase tests with proper mocks (8 tests)

## 🔒 Security

- [x] HttpOnly cookies — tokens inaccessible to JavaScript
- [x] XSS prevention — no innerHTML with API data
- [x] Rate limiting on public routes
- [x] CORS configured per environment
- [x] Email validation (RFC 5321)
- [x] Password validation (8-128 chars)
- [x] CSRF protection middleware
- [x] Stripe webhook HMAC validation
- [x] /metrics restricted by IP
- [x] Swagger disabled in production
- [x] Helmet headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy)
- [x] CSRF protection for JSON requests — Double Submit Cookie Pattern

## 📊 Observability

- [x] Structured logs with slog (JSON in production)
- [x] UUID request ID per request
- [x] Prometheus metrics (/metrics — IP restricted)
- [x] Grafana dashboards
- [x] Health check with DB status
- [x] GitHub Actions CI

## 🔧 Concurrency & Performance

- [x] Worker pool for image processing (goroutines + channels)
- [x] Async pipeline (validate → resize → compress → upload to R2)
- [x] SSE (Server-Sent Events) for real-time notifications
- [x] sync.Map in rate limiter — no mutex contention

## 🧠 Quick Mental Model
Signup/Login  → HttpOnly cookies (access 15min + refresh 7 days)

Cookie        → sent automatically by browser on every request

Middleware    → reads cookie → validates JWT → injects user_id

UseCase       → business rules

Repository    → database

401           → tryRefreshToken() → new cookie pair or redirect to login

Stripe        → PaymentIntent created on checkout

client_secret → Stripe.js confirms payment

webhook HMAC validated → order status → paid

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

![Badge](https://img.shields.io/badge/Go-1.24-blue?logo=go&logoColor=white)
![Badge](https://img.shields.io/badge/Gin-1.10-green)
![Badge](https://img.shields.io/badge/PostgreSQL-16-blue)
![Badge](https://img.shields.io/badge/Docker-Pronto-blue)
![Badge](https://img.shields.io/badge/JWT-Autenticação-blue)
![Badge](https://img.shields.io/badge/Stripe-Pagamentos-blue?logo=stripe&logoColor=white)
![Badge](https://img.shields.io/badge/Deploy-Render-success?logo=render&logoColor=white)
![Badge](https://img.shields.io/badge/CI-GitHub_Actions-success?logo=githubactions&logoColor=white)

## 🎯 Objetivos do Projeto

- Praticar **Clean Architecture** simplificada
- Implementar autenticação **JWT segura** (bcrypt + cookies HttpOnly + rotação de refresh token)
- Trabalhar com **PostgreSQL real** (inclusive nos testes)
- Fazer **testes reais** (unitários + integração)
- Padronizar ambiente com **Docker**
- Criar um projeto fácil de explicar e apresentar
- **Hardening de segurança** (XSS, rate limiting, CSRF, validação de webhook, cookies HttpOnly)
- **Fluxo de pagamento completo** (Stripe Payment Element, webhook, confirmação de pedido)
- **Features production-ready** (healthchecks, rastreamento, monitoring, CI pipeline)

## 🌐 API Online (Render)

- URL base: https://volurya.onrender.com
- Health check: https://volurya.onrender.com/ping → `{"message": "pong"}`

**Aviso**: Render free tier tem cold start — primeira requisição após ~15 min de inatividade pode demorar 10–30 segundos.

## 🧱 Arquitetura

Camadas separadas (Clean Architecture simplificada):

- **Controllers** → camada HTTP (Gin) — só entrada/saída
- **UseCases** → regras de negócio + invariantes
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
STRIPE_PUBLISHABLE_KEY=pk_test_sua_chave
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
docker compose -f local/docker-compose.yml --env-file .env up --build -d
```

API disponível em: http://localhost:8080

Parar tudo:

```bash
docker compose -f local/docker-compose.yml --env-file .env down -v
```

## 🔐 Autenticação (JWT + Cookies HttpOnly)

Login e Signup setam tokens como **cookies HttpOnly** — JavaScript não consegue acessá-los, protegendo contra XSS.

```bash
# Login
curl -c /tmp/cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@volurya.com", "password": "sua_senha"}'

# Rota protegida — cookie enviado automaticamente
curl -b /tmp/cookies.txt http://localhost:8080/api/products

# Verificar status de autenticação
curl -b /tmp/cookies.txt http://localhost:8080/api/auth/me

# Logout — revoga refresh token
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/logout
```

## 💳 Fluxo de Pagamento (Stripe)
Usuário adiciona itens ao carrinho

↓

POST /api/checkout → cria ordem + PaymentIntent no Stripe

↓

Frontend recebe client_secret → Stripe.js Payment Element

↓

Usuário completa pagamento na página /checkout

↓

Stripe envia webhook → POST /api/webhook (assinatura HMAC validada)

↓

Status da ordem atualizado → paid

↓

Usuário redirecionado para /order/success

## 🔌 Principais Endpoints

| Método | Endpoint | Descrição | Auth? |
|--------|----------|-----------|-------|
| POST | /api/signup | Cria usuário | Não |
| POST | /api/login | Seta cookies HttpOnly | Não |
| POST | /api/refresh | Renova tokens via cookie | Não |
| POST | /api/logout | Revoga token + limpa cookies | Não |
| GET | /api/auth/me | Verifica status de autenticação | Sim |
| GET | /api/health | Healthcheck com status do BD | Não |
| GET | /api/config | Publishable key do Stripe | Não |
| POST | /api/webhook | Webhook Stripe (HMAC validado) | Não |
| GET | /api/products | Lista produtos (paginação por cursor) | Sim |
| POST | /api/products | Cria produto (admin only) | Sim |
| GET | /api/products/:id | Busca produto por ID | Sim |
| PUT | /api/products/:id | Atualiza produto (admin only) | Sim |
| DELETE | /api/products/:id | Deleta produto (admin only) | Sim |
| POST | /api/checkout | Cria PaymentIntent no Stripe | Sim |
| GET | /api/cart | Ver carrinho | Sim |
| POST | /api/cart/items | Adicionar produto ao carrinho | Sim |
| PUT | /api/cart/items/:id | Atualizar quantidade | Sim |
| DELETE | /api/cart/items/:id | Remover item | Sim |
| GET | /api/events | SSE notificações em tempo real | Sim |
| GET | /ping | Healthcheck simples | Não |

## 🧪 Testes

```bash
# Testes unitários (sem dependências externas)
JWT_SECRET=test_secret go test ./middleware/... -v

# Testes de integração (requer banco de teste na porta 5433)
docker compose -f local/docker-compose.test.yml up -d
go test ./... -v

# Coverage
go test ./middleware/... -coverprofile=cover.out
go tool cover -html=cover.out
```

CI roda em todo push e PR via GitHub Actions — veja `.github/workflows/ci.yml`.

## 🛠️ Decisões de Arquitetura

- Clean Architecture simplificada
- JWT com cookies HttpOnly — tokens inacessíveis ao JavaScript
- Stripe Payment Element para checkout (PCI compliant)
- Validação HMAC do webhook — previne confirmações de pagamento fraudulentas
- Decremento atômico de estoque — `WHERE stock >= quantity` previne overselling
- Migrations embutidas no binário via `go:embed`
- Cloudflare R2 para storage de imagens — sem taxa de egress
- Worker pool para processamento assíncrono de imagens
- Prometheus + Grafana para monitoring

## 🧠 Modelo Mental Rápido
Login/Signup  → cookies HttpOnly (access 15min + refresh 7 dias)

Cookie        → enviado automaticamente pelo browser

Middleware    → lê cookie → valida JWT → injeta user_id

UseCase       → regras de negócio

Repository    → banco

401           → tryRefreshToken() → novo cookie ou redirect para login

Stripe        → PaymentIntent criado no checkout

client_secret → Stripe.js confirma pagamento

webhook HMAC validado → ordem → paid

## 📬 Contato

Iago Barros  
@Iagobarros2112  
Fortaleza, Brasil – 2026  
Feito com ❤️, raiva e muito café para aprender e mostrar como penso arquitetura backend.

## 📚 Documentação

- [PRD](./docs/volurya-prd.md) — Requisitos e visão de produto
- [Tech Spec](./docs/volurya-tech-spec.md) — Arquitetura e decisões técnicas
- [Docs Index](./docs/index.md) — Índice completo
- [ADRs](./docs/architecture/) — Architecture Decision Records