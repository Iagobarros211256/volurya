# Volurya — Procedimento de deploy na StayCloud

Status: planejamento. Procedimento ainda não executado na VPS.

## 1. Pré-requisitos

- [ ] Confirmar o diretório da aplicação na VPS.
- [ ] Confirmar DNS e domínio definitivo.
- [ ] Confirmar portas 80 e 443 disponíveis.
- [ ] Preparar arquivo .env real, fora do Git.
- [ ] Configurar COMPOSE_PROJECT_NAME=volurya.
- [ ] Configurar JWT_SECRET seguro e não vazio.
- [ ] Configurar VOLURYA_SALES_MODE=whatsapp.
- [ ] Configurar VOLURYA_SALES_WHATSAPP com número autorizado.
- [ ] Conferir ALLOWED_ORIGINS e VOLURYA_DOMAIN.
- [ ] Manter PAYMENTS_ENABLED=false.
- [ ] Confirmar backup externo e procedimento de restauração.
- [ ] Confirmar procedimento de cadastro inicial do catálogo.

## 2. Configuração Docker

Utilizar compose.prod.yml com o mesmo arquivo de ambiente
na interpolação do Compose e no serviço api.

Convenção:

    APP_DIR=/caminho/real/volurya
    export VOLURYA_ENV_FILE="$APP_DIR/.env"

Exemplo de validação, após definir APP_DIR:

    docker compose \
      --project-directory "$APP_DIR" \
      --env-file "$VOLURYA_ENV_FILE" \
      -f "$APP_DIR/compose.prod.yml" \
      config --quiet

Não publicar saídas contendo credenciais.

## 3. Backup e migrações

A API executa db.RunMigrations durante a inicialização.

Antes de atualizar uma base existente:

1. Identificar a versão atual da aplicação e do banco.
2. Gerar backup e verificar o arquivo produzido.
3. Garantir uma cópia fora da VPS.
4. Confirmar procedimento testado de restauração.
5. Somente depois autorizar a atualização da aplicação.

Não executar docker compose down --volumes em produção.

## 4. Backup automatizado

Arquivos:

- scripts/backup-postgres.sh
- deploy/volurya-backup.service
- deploy/volurya-backup.timer
- deploy/volurya-backup.env.example

Instalação prevista:

- Script: /usr/local/sbin/volurya-backup-postgres
- Serviço e timer: /etc/systemd/system/
- Configuração: /etc/volurya/backup.env

Configurar VOLURYA_PROJECT_DIR com o caminho real.
Verificar permissões, executar backup manual e testar restauração
antes de habilitar o timer.

## 5. Publicação — requer autorização

- Conferir versão e origem do código na VPS.
- Validar configuração final do Compose.
- Confirmar backup recuperável.
- Construir e iniciar os serviços aprovados.
- Verificar migrations, healthchecks e logs.
- Conferir DNS, HTTPS, páginas e API.
- Validar catálogo, variantes e contato WhatsApp.
- Confirmar que o checkout Stripe continua bloqueado.

Nenhuma dessas operações está autorizada por este documento.

## 6. Rollback

Preservar a identificação da imagem e do código anteriores.

Rollback da aplicação não implica rollback automático do banco.
Se uma migration incompatível for aplicada, avaliar recuperação
a partir de backup validado, considerando perda de dados
posteriores ao ponto de restauração.

Documentar a estratégia específica antes do primeiro deploy.

## 7. Pendências

Consultar docs/PENDENCIAS_PRE_DEPLOY.md.

Pagamentos reais permanecem desabilitados até auditoria
e validação específicas.
