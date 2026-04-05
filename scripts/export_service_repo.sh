#!/usr/bin/env bash
set -euo pipefail
workspace="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
exec "${workspace}/devflow-control/scripts/export_service_repo.sh" "$@"
