#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${DIR}"

export GOROOT="$(go env GOROOT)"
swag init -g cmd/main.go --parseDependency -o docs/generated/swagger
