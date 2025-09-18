#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
API_ENTRY=${1:-${REPO_ROOT}/api/znp.api}
OUTPUT_DIR=${2:-${REPO_ROOT}/internal}

if ! command -v goctl >/dev/null 2>&1; then
    echo "goctl command is required but not found in PATH" >&2
    exit 1
fi

pushd "${REPO_ROOT}" >/dev/null

goctl api format -dir api

goctl api go --style goZero \
    --api "${API_ENTRY}" \
    --dir "${OUTPUT_DIR}"

popd >/dev/null
