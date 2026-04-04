#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_root="${1:-/tmp/devflow-split}"

services=(
  app-service
  config-service
  release-service
  verify-service
)

for service in "${services[@]}"; do
  "${repo_root}/scripts/export_service_repo.sh" "${service}" "${output_root}"
done

echo "exported service repositories to ${output_root}"
