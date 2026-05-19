# 📱 Guia de Testes Mobile - Volurya API

Teste realista da Volurya API usando seu celular via **Tailscale** e **Termius**.

---

## 🔧 **PASSO 1: Preparar seu PC**

### 1a. Configurar a API para aceitar conexões remotas

Sua API precisa escutar em `0.0.0.0` (todas as interfaces), não só `localhost`.

**Verifique o arquivo `cmd/main.go`:**

```go
// Procure pela linha onde a API inicia:
srv := &http.Server{
    Addr:    ":8080",  // ✅ Isso está correto (0.0.0.0:8080)
    Handler: r,
}
log.Printf("Server running on %s", srv.Addr)
```

Se estiver com `localhost:8080`, mude para `:8080`.

### 1b. Verificar os IPs disponíveis

**Seu PC tem estes IPs:**

```
🔵 Tailscale:   100.85.88.73  (usa VPN Tailscale)
🟢 WiFi Local:  192.168.100.243 (mesma rede WiFi que seu celular)
🔴 Docker:      172.19.0.1    (rede Docker interna)
```

**Qual usar?**
- ✅ **100.85.88.73** (Tailscale) — Se seu celular está conectado no Tailscale
- ✅ **192.168.100.243** (WiFi) — Se seu celular está na mesma rede WiFi
- ❌ 172.19.0.1 (Docker) — Não funciona, é rede interna

### 1c. Certifique-se que a API está rodando

```bash
# No seu PC:
cd ~/Desktop/volurya-origin/volurya
docker compose up -d
```

Verifique se tá rodando:
```bash
curl http://localhost:8080/ping
# Resposta esperada: {"message":"pong"}
```

---

## 📱 **PASSO 2: Configurar seu Celular**

### 2a. Instalar Termius (se ainda não tem)

- **Android**: Play Store → Termius
- **iOS**: App Store → Termius

Termius tem cliente **HTTP integrado** (sem precisa de app separada).

### 2b. Conectar no Tailscale (ou usar WiFi local)

**Opção A - Tailscale (mais seguro):**
1. Abra Tailscale no celular
2. Verifique que está conectado
3. Você verá seu PC listado como um dispositivo

**Opção B - WiFi Local (mais rápido):**
1. Conecte seu celular na mesma rede WiFi do PC
2. Pronto!

---

## 🌐 **PASSO 3: Testar Conexão Básica**

### Via Termius (App HTTP)

1. Abra **Termius**
2. Vá em **HTTP** (ou **Snippets** → **HTTP**)
3. Crie uma nova requisição:

```
GET http://100.85.88.73:8080/ping
```

(Ou `192.168.100.243:8080/ping` se usar WiFi)

**Resposta esperada:**
```json
{"message":"pong"}
```

✅ Se receber resposta, a conexão está funcionando!

---

## 🔓 **PASSO 4: Testar Fluxo de Autenticação**

### 4a. Signup (Criar Usuário)

```http
POST http://100.85.88.73:8080/api/signup
Content-Type: application/json

{
  "email": "mobile@volurya.com",
  "password": "TestPassword123"
}
```

**Resposta esperada (201):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1
}
```

✅ Copie o `access_token` para usar nos próximos testes!

### 4b. Login

```http
POST http://100.85.88.73:8080/api/login
Content-Type: application/json

{
  "email": "mobile@volurya.com",
  "password": "TestPassword123"
}
```

Mesma resposta do signup.

### 4c. Health Check

```http
GET http://100.85.88.73:8080/api/health
```

**Resposta esperada (200):**
```json
{
  "status": "healthy",
  "timestamp": "2026-05-19T11:10:40Z",
  "database": "connected",
  "version": "1.0.0"
}
```

---

## 📝 **PASSO 5: Testar Endpoints Autenticados**

Use o `access_token` que recebeu no signup:

### 5a. Listar Produtos

```http
GET http://100.85.88.73:8080/api/products
Authorization: Bearer YOUR_ACCESS_TOKEN
```

### 5b. Carrinho - Ver

```http
GET http://100.85.88.73:8080/api/cart
Authorization: Bearer YOUR_ACCESS_TOKEN
```

### 5c. Carrinho - Adicionar Item

```http
POST http://100.85.88.73:8080/api/cart/items
Authorization: Bearer YOUR_ACCESS_TOKEN
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 2
}
```

### 5d. Carrinho - Checkout

```http
POST http://100.85.88.73:8080/api/cart/checkout
Authorization: Bearer YOUR_ACCESS_TOKEN
```

---

## 🔐 **PASSO 6: Testes de Segurança**

### ⚠️ 6a. Teste de Rate Limiting

Envie **6 requisições de login em rápida sucessão** (limite é 5/min):

```http
POST http://100.85.88.73:8080/api/login
Content-Type: application/json

{
  "email": "mobile@volurya.com",
  "password": "TestPassword123"
}
```

**6ª requisição esperada (429):**
```json
{
  "error": "Too many requests"
}
```

✅ Se receber 429, o rate limiting **está funcionando**!

### ⚠️ 6b. Teste de CORS

Envie uma requisição com header `Origin`:

```http
GET http://100.85.88.73:8080/api/products
Authorization: Bearer YOUR_ACCESS_TOKEN
Origin: http://app.maliciosa.com
```

**Resposta esperada:**
```
Access-Control-Allow-Origin: (nada / seu domínio autorizado)
```

✅ Se negar acesso a domínio não autorizado, CORS **está seguro**!

### ⚠️ 6c. Teste de Email Validation

Tente criar usuário com email inválido:

```http
POST http://100.85.88.73:8080/api/signup
Content-Type: application/json

{
  "email": "@invalido.com",
  "password": "TestPassword123"
}
```

**Resposta esperada (400):**
```json
{
  "error": "invalid email format"
}
```

✅ Se rejeitar, validação **está funcionando**!

### ⚠️ 6d. Teste de Password Validation

Tente senha muito curta:

```http
POST http://100.85.88.73:8080/api/signup
Content-Type: application/json

{
  "email": "test@volurya.com",
  "password": "123"
}
```

**Resposta esperada (400):**
```json
{
  "error": "password must be between 8 and 128 characters"
}
```

✅ Se rejeitar, validação **está funcionando**!

### ⚠️ 6e. Teste de Token Expirado

Aguarde 16 minutos ou use um token fake:

```http
GET http://100.85.88.73:8080/api/products
Authorization: Bearer fake_invalid_token
```

**Resposta esperada (401):**
```json
{
  "error": "unauthorized"
}
```

✅ Se rejeitar token inválido, auth **está segura**!

### ⚠️ 6f. Teste de Refresh Token

```http
POST http://100.85.88.73:8080/api/refresh
Content-Type: application/json

{
  "refresh_token": "YOUR_REFRESH_TOKEN"
}
```

**Resposta esperada (200):**
```json
{
  "access_token": "novo_token...",
  "refresh_token": "novo_refresh_token..."
}
```

### ⚠️ 6g. Teste de Ownership (Proteção de Dados)

1. Crie um produto como usuário A
2. Tente deletar com token de usuário B
3. Deve receber **403 Forbidden**

---

## 🎯 **PASSO 7: Testes de Funcionalidade Completa**

### Fluxo de Compra Realista:

```
1. ✅ Signup
2. ✅ Login (pega access_token)
3. ✅ Ver carrinho (vazio)
4. ✅ Listar produtos
5. ✅ Adicionar item ao carrinho
6. ✅ Ver carrinho (com items)
7. ✅ Atualizar quantidade
8. ✅ Checkout
9. ✅ Logout (revoga refresh_token)
10. ✅ Tentar usar access_token antigo (deve falhar)
```

---

## 📊 **PASSO 8: Monitorar com Swagger UI**

Acesse pelo browser do PC (para ver logs em tempo real):

```
http://localhost:8080/swagger/index.html
```

Nele você pode:
- Ver documentação de todos os endpoints
- Testar diretamente (com tokens reais)
- Ver padrão de request/response

---

## 🛠️ **DICAS PRÁTICAS**

### Copiar/Colar Tokens Facilmente

No Termius, você pode:
1. Guardar tokens em **Notes** (Termius tem seção de notas)
2. Copiar direto do histórico de requisições anteriores
3. Usar **variáveis** se seu cliente HTTP suportar

### Ver Logs em Tempo Real

```bash
# Terminal do PC:
docker compose logs -f

# Vê:
# - Qual IP acessou
# - Qual endpoint chamou
# - Tempo de resposta
# - Status HTTP
# - Request ID para tracing
```

### Testar com cURL (Terminal Termius)

Se quiser, você pode usar **Termius SSH** para conectar no PC e rodar:

```bash
curl -X GET http://localhost:8080/api/health
```

---

## ⚙️ **CONFIGURAÇÃO AVANÇADA (Opcional)**

### Se quiser HTTPS (mais realista)

1. Gere certificado auto-assinado:
```bash
openssl req -x509 -newkey rsa:4096 -nodes -out cert.pem -keyout key.pem -days 365
```

2. Configure no `cmd/main.go`:
```go
srv.ListenAndServeTLS("cert.pem", "key.pem")
```

3. No celular, ignore warnings de certificado inválido

---

## 📋 **CHECKLIST DE TESTES**

- [ ] Conexão básica (ping)
- [ ] Signup com email válido
- [ ] Login com credenciais corretas
- [ ] Listar produtos (requer auth)
- [ ] Ver carrinho
- [ ] Adicionar item ao carrinho
- [ ] Fazer checkout
- [ ] Logout
- [ ] Rate limiting (429 após 5 requisições)
- [ ] Email validation (rejeita inválidos)
- [ ] Password validation (rejeita <8 chars)
- [ ] CORS protection
- [ ] Auth com token inválido (401)
- [ ] Refresh token
- [ ] Health check
- [ ] Ownership enforcement

---

## 🚀 **PRÓXIMOS PASSOS**

Se encontrar bugs:
1. Anote o erro exato
2. Veja os logs: `docker compose logs -f`
3. Procure o `request_id` nos logs para rastrear
4. Corrija e teste de novo

Bom teste! 🎉

