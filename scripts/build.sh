#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${DIR}"

./scripts/regen-swagger.sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/devflow-release-service ./cmd
