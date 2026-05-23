#!/bin/bash

# 📱 Script de Teste da API Volurya - para usar via Termius SSH
# Uso: ./mobile-test.sh

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_URL="${1:-http://localhost:8080}"
EMAIL="mobile_test_$(date +%s)@volurya.com"
PASSWORD="TestPassword123"

echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}📱 TESTE DE API VOLURYA - Mobile Testing${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "🎯 Alvo: ${YELLOW}${API_URL}${NC}"
echo -e "📧 Email: ${YELLOW}${EMAIL}${NC}"
echo -e "🔑 Senha: ${YELLOW}${PASSWORD}${NC}"
echo ""

# TESTE 1: Health Check
echo -e "${BLUE}[1/10] Health Check${NC}"
HEALTH=$(curl -s -X GET "${API_URL}/api/health")
echo -e "  Resposta: ${GREEN}${HEALTH}${NC}\n"

# TESTE 2: Ping
echo -e "${BLUE}[2/10] Ping${NC}"
PING=$(curl -s -X GET "${API_URL}/ping")
echo -e "  Resposta: ${GREEN}${PING}${NC}\n"

# TESTE 3: Signup
echo -e "${BLUE}[3/10] Signup (Criar Usuário)${NC}"
AUTH=$(curl -s -X POST "${API_URL}/api/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
echo -e "  Resposta: ${GREEN}${AUTH}${NC}"

# Extrair tokens
ACCESS_TOKEN=$(echo $AUTH | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $AUTH | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo -e "${RED}❌ Erro: Não conseguiu extrair access_token${NC}\n"
  exit 1
fi

echo -e "  ${GREEN}✅ Access Token: ${ACCESS_TOKEN:0:20}...${NC}"
echo -e "  ${GREEN}✅ Refresh Token: ${REFRESH_TOKEN:0:20}...${NC}\n"

# TESTE 4: Login
echo -e "${BLUE}[4/10] Login${NC}"
LOGIN=$(curl -s -X POST "${API_URL}/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
echo -e "  Resposta: ${GREEN}${LOGIN:0:100}...${NC}\n"

# TESTE 5: Listar Produtos (protegido)
echo -e "${BLUE}[5/10] Listar Produtos (com auth)${NC}"
PRODUCTS=$(curl -s -X GET "${API_URL}/api/products" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")
echo -e "  Resposta: ${GREEN}${PRODUCTS:0:150}...${NC}\n"

# TESTE 6: Ver Carrinho
echo -e "${BLUE}[6/10] Ver Carrinho${NC}"
CART=$(curl -s -X GET "${API_URL}/api/cart" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")
echo -e "  Resposta: ${GREEN}${CART}${NC}\n"

# TESTE 7: Rate Limiting (6 requisições rápidas)
echo -e "${BLUE}[7/10] Teste de Rate Limiting (6 logins rápidos)${NC}"
for i in {1..6}; do
  RATE_TEST=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
  HTTP_CODE=$(echo "$RATE_TEST" | tail -n1)
  
  if [ "$HTTP_CODE" = "429" ]; then
    echo -e "  ${GREEN}✅ Requisição #${i}: ${HTTP_CODE} (Rate limit ativo!)${NC}"
    break
  else
    echo -e "  ⏳ Requisição #${i}: ${HTTP_CODE}"
  fi
done
echo ""

# TESTE 8: Email Validation (inválido)
echo -e "${BLUE}[8/10] Teste de Email Validation (inválido)${NC}"
INVALID_EMAIL=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/signup" \
  -H "Content-Type: application/json" \
  -d '{"email":"@invalido.com","password":"TestPassword123"}')
HTTP_CODE=$(echo "$INVALID_EMAIL" | tail -n1)
BODY=$(echo "$INVALID_EMAIL" | head -n-1)

if [ "$HTTP_CODE" = "400" ]; then
  echo -e "  ${GREEN}✅ Email inválido rejeitado (400)${NC}"
  echo -e "  Erro: ${GREEN}${BODY}${NC}\n"
else
  echo -e "  ${RED}❌ Email inválido não foi rejeitado (${HTTP_CODE})${NC}\n"
fi

# TESTE 9: Password Validation (muito curta)
echo -e "${BLUE}[9/10] Teste de Password Validation (muito curta)${NC}"
INVALID_PASS=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"test_$(date +%s)@volurya.com\",\"password\":\"123\"}")
HTTP_CODE=$(echo "$INVALID_PASS" | tail -n1)
BODY=$(echo "$INVALID_PASS" | head -n-1)

if [ "$HTTP_CODE" = "400" ]; then
  echo -e "  ${GREEN}✅ Senha muito curta rejeitada (400)${NC}"
  echo -e "  Erro: ${GREEN}${BODY}${NC}\n"
else
  echo -e "  ${RED}❌ Senha muito curta não foi rejeitada (${HTTP_CODE})${NC}\n"
fi

# TESTE 10: Refresh Token
echo -e "${BLUE}[10/10] Teste de Refresh Token${NC}"
REFRESH=$(curl -s -X POST "${API_URL}/api/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"${REFRESH_TOKEN}\"}")
echo -e "  Resposta: ${GREEN}${REFRESH:0:100}...${NC}\n"

# Resumo
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ TODOS OS TESTES CONCLUÍDOS COM SUCESSO!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}📊 Resumo de Testes:${NC}"
echo -e "  ✅ Health check"
echo -e "  ✅ Ping"
echo -e "  ✅ Signup"
echo -e "  ✅ Login"
echo -e "  ✅ Listar produtos (auth)"
echo -e "  ✅ Ver carrinho"
echo -e "  ✅ Rate limiting"
echo -e "  ✅ Email validation"
echo -e "  ✅ Password validation"
echo -e "  ✅ Refresh token"
echo ""
echo -e "${YELLOW}📝 Tokens para teste manual:${NC}"
echo -e "  Access:  ${ACCESS_TOKEN}"
echo -e "  Refresh: ${REFRESH_TOKEN}"
echo ""

#:<<Senha exposta no output do terminal
#bashecho -e "🔑 Senha: ${YELLOW}${PASSWORD}${NC}"
#E no final:
#bashecho -e "  Refresh: ${REFRESH_TOKEN}"
#Tokens e senha aparecem em claro no terminal — ficam no histórico do shell e em logs de CI. Mascare:
#bashecho -e "🔑 Senha: ${YELLOW}****${NC}"
#echo -e "  Access:  ${ACCESS_TOKEN:0:20}..."  # já faz isso em alguns lugares
#echo -e "  Refresh: ${REFRESH_TOKEN:0:20}..."

#🔴 set -e + exit 1 manual é inconsistente
#bashset -e  # para no primeiro erro
# ...
#if [ -z "$ACCESS_TOKEN" ]; then
#  exit 1  # manual
#fi
#Com set -e, qualquer comando que falhe já para o script. Mas grep retorna código 1 quando não encontra nada — com set -e isso mata o script silenciosamente antes do if. Troque por:
#bashset -euo pipefail  # -u: variável não definida é erro, -o pipefail: erro em pipe
#ACCESS_TOKEN=$(echo "$AUTH" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
#Ou use jq se disponível:
#bashACCESS_TOKEN=$(echo "$AUTH" | jq -r '.access_token // empty')

#🔴 Parse de JSON com grep e cut é frágil
#bashACCESS_TOKEN=$(echo $AUTH | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
#Quebra se o JSON vier formatado, com espaços, ou em ordem diferente. Use jq:
#bashACCESS_TOKEN=$(echo "$AUTH" | jq -r '.access_token')
#REFRESH_TOKEN=$(echo "$AUTH" | jq -r '.refresh_token')

#🟡 echo $AUTH sem aspas — word splitting
#bashecho $AUTH | grep ...
#ACCESS_TOKEN=$(echo $AUTH | ...)
#Sem aspas, espaços no JSON quebram o echo. Use sempre "$AUTH":
#bashecho "$AUTH" | jq -r '.access_token'

#🟡 Resumo final sempre imprime "TODOS OS TESTES CONCLUÍDOS COM SUCESSO"
#bashecho -e "${GREEN}✅ TODOS OS TESTES CONCLUÍDOS COM SUCESSO!${NC}"
#Essa mensagem aparece mesmo se vários testes falharam — set -e para o script em erros de comando, mas falhas de validação (HTTP_CODE != 400) apenas imprimem ❌ e continuam. O resumo deveria ser condicional baseado em um contador de falhas:
#bashFAILURES=0

#if [ "$HTTP_CODE" != "400" ]; then
#    echo -e "  ${RED}❌ Falhou${NC}"
#    FAILURES=$((FAILURES + 1))
#fi

# No final:
#if [ $FAILURES -eq 0 ]; then
#    echo -e "${GREEN}✅ TODOS OS TESTES PASSARAM${NC}"
#else
#    echo -e "${RED}❌ ${FAILURES} TESTE(S) FALHARAM${NC}"
#    exit 1
#fi

#🟡 Testes de admin não incluídos
#Não há testes para:

#Criar produto (admin)
#Upload de imagem (admin)
#Deletar produto (admin)
#Checkout/pagamento

#O script testa apenas o fluxo de usuário comum.

#🟢 Dependência de curl não verificada
#bash#!/bin/bash
# usa curl extensivamente mas não verifica se está instalado
#bashcommand -v curl >/dev/null 2>&1 || { echo "curl não encontrado"; exit 1; }
#command -v jq >/dev/null 2>&1 || { echo "jq não encontrado"; exit 1; }