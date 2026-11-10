# Backlog — Uso futuro do Strix | Volurya — site e loja

**ID:** VOL-STRIX-001  
**Prioridade:** P1 — alta  
**Responsável/condição:** Projeto próprio

## Para que usar

Verificar segurança da API Go, autenticação, carrinho, checkout, pedidos e áreas administrativas antes de publicar ou ampliar o e-commerce.

## Cenários de segurança propostos (não são falhas confirmadas)

- Alteração indevida de preço, desconto, frete, quantidades ou estado de pedidos.
- Acesso a pedidos/carrinhos de outra conta (IDOR/BOLA).
- Autenticação, sessões/JWT, brute force e ausência de rate limit.
- Validação de entradas, injeções e exposição de variáveis/configurações.

## Como usar (plano futuro)

1. Subir cópia local com Go, PostgreSQL e Docker, sem acesso ao provedor de pagamentos real.
2. Criar clientes, produtos, pedidos e pagamentos simulados com valores fictícios.
3. Delimitar rotas e contas autorizadas e realizar execução controlada do Strix.
4. Reproduzir achados e criar testes Go e de integração para evitar regressões.
5. Revisar resultados antes do deployment na StayCloud.

## Restrições específicas

Não criar compras, cobranças, pagamentos ou pedidos reais; não atacar a VPS pública.

## Critérios de aceite para um piloto futuro

- [ ] Proprietários/responsáveis aprovaram o escopo, os endpoints, contas e limites.
- [ ] Ambiente de teste isolado e com restauração documentada.
- [ ] Dados e credenciais sintéticos; revisão de privacidade do provedor LLM.
- [ ] Configuração, custo e tempo máximos definidos.
- [ ] Execução piloto registrada com versão do Strix, alvos, logs e evidências.
- [ ] Achados classificados em **confirmado**, **não reproduzido** ou **fora do escopo**.
- [ ] Correções e testes de regressão revisados antes de eventual uso recorrente.
- [ ] Decisão de prosseguir, limitar ou descartar documentada.

## Situação e governança

- **Status:** PENDENTE — avaliação futura; não instalado, não integrado e não executado neste projeto.
- **Data de registro:** 2026-11-09.
- **Ferramenta:** [Strix (usestrix/strix)](https://github.com/usestrix/strix), pentesting automatizado com agentes de IA.
- **Natureza do registro:** proposta técnica, não requisito implementado ou teste concluído.
- **Autorização:** testar somente código, serviços e ambientes próprios ou com autorização explícita dos responsáveis, delimitando por escrito o escopo.
- **Ambiente:** preferir laboratório isolado e reversível, nunca produção, com usuários, documentos e dados sintéticos.
- **Privacidade e custos:** revisar o código, dependências, permissões, política de envio de conteúdo ao provedor LLM, retenção, custos, limites de execução e logs antes de configurar qualquer chave.
- **Impactos:** o Strix pode enviar requisições com efeitos colaterais e tentar explorações; aplicar rate limit, backups/snapshots e rollback.
- **Validação:** separar hipótese/sinal de vulnerabilidade de evidência confirmada; reproduzir e corrigir manualmente antes de declarar resultado.
- **Operações:** este registro **não** autoriza scans, instalação, alteração de repositórios, push/pull, CI, deploy, acesso a dados reais ou alteração de bancos.
