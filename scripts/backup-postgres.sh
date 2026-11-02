#!/usr/bin/env bash
set -euo pipefail
umask 077

BACKUP_DIR="/var/backups/volurya"
CONTAINER="volurya-postgres-1"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$BACKUP_DIR"

TEMP_FILE="$(mktemp "$BACKUP_DIR/.postgres-XXXXXXXX.tmp")"
FINAL_FILE="$BACKUP_DIR/volurya-$TIMESTAMP.dump"

trap 'rm -f "$TEMP_FILE"' EXIT

docker exec "$CONTAINER" \
  pg_dump -U volurya -d volurya_db -Fc > "$TEMP_FILE"

test -s "$TEMP_FILE"

docker exec -i "$CONTAINER" \
  pg_restore --list < "$TEMP_FILE" > /dev/null

mv "$TEMP_FILE" "$FINAL_FILE"

find "$BACKUP_DIR" -maxdepth 1 -type f \
  -name 'volurya-*.dump' -mtime +13 -delete

echo "Backup concluído: $FINAL_FILE"
