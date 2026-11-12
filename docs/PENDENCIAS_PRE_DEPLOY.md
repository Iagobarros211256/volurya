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

## Validação mobile — antes da publicação

- [ ] Testar larguras de tela de 320, 360, 390 e 430 px.
- [ ] Testar também orientação horizontal e tablets.
- [ ] Verificar menus, cabeçalho, rodapé e navegação por toque.
- [ ] Conferir responsividade dos cards, imagens e páginas de produtos.
- [ ] Garantir ausência de rolagem horizontal indesejada.
- [ ] Verificar legibilidade, contraste e tamanho dos elementos clicáveis.
- [ ] Testar formulários, teclado virtual e mensagens de erro.
- [ ] Verificar carrinho e apresentação do checkout desabilitado.
- [ ] Testar carregamento e comportamento em conexão móvel lenta.
- [ ] Validar em dispositivo Android/Chrome e iPhone/Safari,
      quando disponíveis.
- [ ] Registrar problemas com captura de tela, dispositivo,
      navegador, dimensão da tela e passos para reprodução.
- [ ] Corrigir problemas que impeçam a navegação antes de
      considerar a experiência mobile validada.

A simulação responsiva no navegador ajuda a encontrar problemas,
mas não substitui completamente testes em dispositivos físicos.

## Situação atual

- PostgreSQL de integração: testado.
- Migrations no banco financeiro isolado: testadas.
- Transação de checkout: testes no PostgreSQL isolado aprovados.
- Rollback financeiro: cenários de integração local aprovados.
- Validação mobile: pendente.
- Pagamentos reais: não validados para produção.
- VPS StayCloud: deploy da aplicação pendente.

## Infraestrutura e recuperação — antes da publicação

- [ ] Confirmar o nome do projeto Docker Compose e do contêiner PostgreSQL.
- [ ] Fazer o script de backup usar o banco e usuário efetivos de produção.
- [ ] Instalar o script no caminho esperado pelo serviço systemd.
- [ ] Validar o timer de backup na VPS e a retenção dos arquivos.
- [ ] Testar restauração completa em um PostgreSQL isolado.
- [ ] Definir uma cópia externa à VPS para recuperação de desastres.
- [ ] Validar o carregamento das variáveis pelo Docker Compose.
- [ ] Confirmar o WhatsApp comercial autorizado.
- [ ] Preparar o cadastro reproduzível da camiseta e suas variantes.

## Evidência de backup e restauração — 09/11/2026

Teste local realizado em PostgreSQL 16, em ambiente Docker
Compose descartável e separado dos bancos dos demais projetos.

- [x] Executar o script real `scripts/backup-postgres.sh`.
- [x] Gerar dump PostgreSQL no formato customizado.
- [x] Validar a leitura do índice com `pg_restore --list`.
- [x] Restaurar o dump em outro banco do laboratório.
- [x] Comparar os três registros originais e restaurados.

Resultado: 3/3 registros idênticos, incluindo identificadores,
descrições e valores em centavos.

Limitação: teste realizado com uma tabela de referência.
Não comprova recuperação de uma base de produção completa.

Continuam pendentes: instalação e validação do systemd na VPS,
backup externo, retenção operacional e simulação de recuperação
representativa da produção.

## Revalidação do backup — 12/11/2026

Segunda validação funcional realizada em PostgreSQL 16,
utilizando a versão atualizada de scripts/backup-postgres.sh.

- [x] Identificar o laboratório por COMPOSE_PROJECT_NAME.
- [x] Carregar o ambiente utilizando VOLURYA_ENV_FILE.
- [x] Gerar e validar um novo dump PostgreSQL.
- [x] Restaurar o dump em outro banco isolado.
- [x] Comparar os três registros: 3/3 idênticos.
- [x] Remover contêiner, volume, rede e arquivos temporários.

Resultado: segunda validação local aprovada, sem divergências.

Esta evidência não substitui a validação de recuperação
completa da base de produção, a instalação do systemd,
o teste de retenção ou a cópia externa dos backups.

## Estratégia

1. Publicar o site e o catálogo.
2. Manter o checkout real bloqueado no backend.
3. Validar acesso, HTTPS, saúde da API e logs.
4. Corrigir as pendências financeiras posteriormente.
5. Habilitar pagamentos somente após testes e revisão.

Este documento registra riscos conhecidos, não uma
auditoria completa de segurança.
