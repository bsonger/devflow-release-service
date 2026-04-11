#!/usr/bin/env bash
set -euo pipefail

go test ./...
bash scripts/build.sh
