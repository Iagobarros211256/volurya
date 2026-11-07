# Volurya — Pré-venda pelo WhatsApp

**Revisão:** 07/11/2026
**Branch:** task/volurya-deploy-staycloud
**Estado:** funcional localmente, ainda não publicado.

## Objetivo

Permitir a pré-venda assistida da camiseta Volurya.
O cliente seleciona o tamanho e abre uma conversa no
WhatsApp do Heitor, com produto e preço preenchidos.

Disponibilidade, pagamento e entrega são confirmados
diretamente com Heitor. Não há pagamento online.

## Produto

| Tamanho | Preço |
| --- | ---: |
| P, M, G, GG | R$ 50,00 |
| XG, XGG | R$ 55,00 |

Imagem: views/static/imagens/volurya-camiseta-colapsu-prevenda.webp

## Implementação

- Configuração pública do modo WhatsApp.
- Produtos com sales_type standard ou preorder.
- Variantes com preços em centavos inteiros.
- Migration 000013_add_product_variants.
- API de produtos retornando variantes.
- Seleção obrigatória de tamanho e preço dinâmico.
- Mensagem do WhatsApp gerada automaticamente no link.
- Bloqueio de pré-vendas no carrinho e na criação de pedidos.
- Restrição SQL de estoque a produtos standard.
- Resposta HTTP 409 para tentativas de compra indevidas.
- Implementação de idempotência de checkout em revisão.

## Configuração local

- VOLURYA_SALES_MODE=whatsapp
- VOLURYA_SALES_WHATSAPP=558581842950
- PAYMENTS_ENABLED=false

## Validações realizadas

- Compilação Go aprovada.
- Oito testes JavaScript aprovados.
- Migration 13 executada no PostgreSQL isolado.
- API local retornando as seis variantes.
- Loja visualizada no Firefox.
- Tamanho XGG com preço R$ 55,00 confirmado.
- Abertura do WhatsApp do Heitor confirmada.
- Quebras de linha da mensagem corrigidas.

## Pendências

- Testes Go de integração com banco descartável.
- Rejeição de preorder com estoque positivo.
- Verificação de ausência de pedidos e descontos indevidos.
- Revisão do checkout idempotente e legado.
- Testes de respostas HTTP e responsividade mobile.
- Cadastro reproduzível da camiseta no banco de destino.
- Revisão de segurança e credenciais antes do push.
- Deploy e validação na VPS StayCloud.

O cadastro demonstrativo da camiseta está somente no
PostgreSQL temporário. Ainda não existe publicação desse
cadastro no ambiente de destino.

## Segurança e publicação

O fluxo do WhatsApp foi validado localmente, mas ainda
não está disponível no site público.

Pagamentos online devem permanecer desabilitados.

Este documento registra o progresso, não uma aprovação
para produção.
