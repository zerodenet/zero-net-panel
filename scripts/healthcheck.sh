#!/usr/bin/env bash
set -euo pipefail

# Lightweight health and error check.
HEALTH_URL=${ZNP_HEALTH_URL:-http://127.0.0.1:8888/api/v1/ping}
LOG_FILE=${ZNP_LOG_FILE:-/var/log/znp/znp.log}
ERROR_PATTERNS=${ZNP_ERROR_PATTERNS:-"ERROR|panic|migration failed"}

curl -fsS "${HEALTH_URL}" >/dev/null || {
  echo "health check failed: ${HEALTH_URL}"
  exit 1
}

if [[ -f "${LOG_FILE}" ]]; then
  if tail -n 200 "${LOG_FILE}" | grep -E "${ERROR_PATTERNS}" >/dev/null 2>&1; then
    echo "recent errors detected in ${LOG_FILE}"
    exit 2
  fi
fi

echo "ok"
