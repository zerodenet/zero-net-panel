#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
OUTPUT_DIR=${1:-${REPO_ROOT}/docs/api-generated}

if ! command -v goctl >/dev/null 2>&1; then
    echo "goctl command is required but not found in PATH" >&2
    exit 1
fi

mkdir -p "${OUTPUT_DIR}"

pushd "${REPO_ROOT}" >/dev/null

goctl api doc --dir api --o "${OUTPUT_DIR}"

popd >/dev/null
