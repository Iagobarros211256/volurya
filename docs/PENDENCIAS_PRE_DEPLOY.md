# Volurya — Pendências técnicas e de segurança

Data da revisão: 31/10/2026

## Objetivo

Publicar o site na VPS StayCloud sem bloquear a entrega
por problemas que não afetam a navegação pública.

Não habilitar pagamentos reais antes de corrigir
e validar as pendências críticas.

## Críticas — Antes de habilitar pagamentos

- [ ] Garantir que pedidos, itens e estoque sejam
      persistidos em uma única transação PostgreSQL.
- [ ] Reverter ou liberar estoque quando houver falha
      na criação do PaymentIntent na Stripe.
- [ ] Impedir resposta de sucesso ao checkout quando
      o registro local do pagamento não for persistido.
- [ ] Retornar erro apropriado no webhook quando
      a persistência falhar, permitindo retry da Stripe.
- [ ] Garantir atomicidade entre status do pagamento
      e status do pedido.
- [ ] Tratar webhooks duplicados e fora de ordem sem
      regredir o estado de pagamentos confirmados.
- [ ] Validar os fluxos críticos com testes de integração.

## Importantes — Revisão posterior

- [ ] Padronizar os erros de estoque entre repository,
      usecase e controller (HTTP 422).
- [ ] Avaliar substituição de float64 por valores em
      centavos inteiros para cálculos financeiros.
- [ ] Revisar política de cancelamento e reposição
      de estoque, evitando devoluções duplicadas.
- [ ] Revisar a configuração de versão da API Stripe.
- [ ] Auditar autenticação, autorização, rate limiting
      e tratamento de dados sensíveis antes de
      considerar o e-commerce integralmente validado.

## Situação atual

- PostgreSQL de integração: testado.
- Migrations no banco financeiro isolado: testadas.
- Transação de checkout: implementação local iniciada.
- Teste de rollback: ainda não validado.
- Pagamentos reais: não validados para produção.
- VPS StayCloud: deploy da aplicação pendente.

## Estratégia

1. Publicar o site e o catálogo.
2. Manter o checkout real bloqueado no backend.
3. Validar acesso, HTTPS, saúde da API e logs.
4. Corrigir as pendências financeiras posteriormente.
5. Habilitar pagamentos somente após testes e revisão.

Este documento registra riscos conhecidos, não uma
auditoria completa de segurança.
