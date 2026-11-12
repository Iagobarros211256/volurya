#!/usr/bin/env bash
set -euo pipefail
umask 077

BACKUP_DIR="${VOLURYA_BACKUP_DIR:-/var/backups/volurya}"
PROJECT_DIR="${VOLURYA_PROJECT_DIR:?Defina VOLURYA_PROJECT_DIR}"
ENV_FILE="${VOLURYA_ENV_FILE:-$PROJECT_DIR/.env}"
COMPOSE_FILE="$PROJECT_DIR/compose.prod.yml"

test -d "$PROJECT_DIR"
test -f "$ENV_FILE"
test -f "$COMPOSE_FILE"

compose() {
  VOLURYA_ENV_FILE="$ENV_FILE" docker compose \
    --project-directory "$PROJECT_DIR" \
    --env-file "$ENV_FILE" \
    -f "$COMPOSE_FILE" "$@"
}
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$BACKUP_DIR"

TEMP_FILE="$(mktemp "$BACKUP_DIR/.postgres-XXXXXXXX.tmp")"
FINAL_FILE="$BACKUP_DIR/volurya-$TIMESTAMP.dump"

trap 'rm -f "$TEMP_FILE"' EXIT

compose exec -T postgres sh -ec \
  'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$TEMP_FILE"

test -s "$TEMP_FILE"

compose exec -T postgres \
  pg_restore --list < "$TEMP_FILE" > /dev/null

mv "$TEMP_FILE" "$FINAL_FILE"

find "$BACKUP_DIR" -maxdepth 1 -type f \
  -name 'volurya-*.dump' -mtime +13 -delete

echo "Backup concluído: $FINAL_FILE"
