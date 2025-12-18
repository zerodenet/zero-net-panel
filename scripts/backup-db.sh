#!/usr/bin/env bash
set -euo pipefail

# Lightweight DB backup helper. Choose driver via ZNP_DB_DRIVER (mysql|postgres).
# For MySQL:
#   ZNP_DB_DRIVER=mysql ZNP_DB_HOST=127.0.0.1 ZNP_DB_PORT=3306 ZNP_DB_USER=root ZNP_DB_PASSWORD=secret ZNP_DB_NAME=znp ./scripts/backup-db.sh backup.sql
# For Postgres:
#   ZNP_DB_DRIVER=postgres ZNP_DB_HOST=127.0.0.1 ZNP_DB_PORT=5432 ZNP_DB_USER=postgres ZNP_DB_PASSWORD=secret ZNP_DB_NAME=znp ./scripts/backup-db.sh backup.sql

DRIVER=${ZNP_DB_DRIVER:-mysql}
OUTFILE=${1:-"backup.sql"}

HOST=${ZNP_DB_HOST:-127.0.0.1}
PORT=${ZNP_DB_PORT:-3306}
USER=${ZNP_DB_USER:-root}
PASSWORD=${ZNP_DB_PASSWORD:-}
DBNAME=${ZNP_DB_NAME:-znp}

case "${DRIVER}" in
  mysql)
    export MYSQL_PWD="${PASSWORD}"
    mysqldump -h "${HOST}" -P "${PORT}" -u "${USER}" "${DBNAME}" >"${OUTFILE}"
    ;;
  postgres|postgresql)
    export PGPASSWORD="${PASSWORD}"
    pg_dump -h "${HOST}" -p "${PORT}" -U "${USER}" -d "${DBNAME}" -F p -f "${OUTFILE}"
    ;;
  *)
    echo "Unsupported driver: ${DRIVER}. Use mysql or postgres."
    exit 1
    ;;
esac

echo "Backup completed: ${OUTFILE}"
