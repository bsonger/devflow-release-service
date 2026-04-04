#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${repo_root}/scripts/create_github_repo.sh" "devflow-app-service" "Devflow app metadata service"
"${repo_root}/scripts/create_github_repo.sh" "devflow-config-service" "Devflow configuration metadata service"
"${repo_root}/scripts/create_github_repo.sh" "devflow-release-service" "Devflow release intent and orchestration service"
"${repo_root}/scripts/create_github_repo.sh" "devflow-verify-service" "Devflow verify and state writeback service"
