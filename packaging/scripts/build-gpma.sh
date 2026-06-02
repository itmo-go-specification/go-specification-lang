#!/usr/bin/env bash
# Build the gpma CLI binary for linux/amd64 distribution.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST="${ROOT}/packaging/dist"

mkdir -p "${DIST}"

cd "${ROOT}/framework"
echo "Generating ANTLR parser..."
go generate ./...

echo "Building gpma binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" \
  -o "${DIST}/gpma" ./cli/main.go

echo "Built ${DIST}/gpma"
