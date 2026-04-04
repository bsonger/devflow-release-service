#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: create_github_repo.sh <repo-name> <description>

requires:
  GITHUB_TOKEN with repository creation permission
EOF
  exit 1
}

[ "$#" -ge 2 ] || usage
[ -n "${GITHUB_TOKEN:-}" ] || {
  echo "GITHUB_TOKEN is required" >&2
  exit 1
}

repo_name="$1"
shift
description="$*"

payload="$(cat <<EOF
{
  "name": "${repo_name}",
  "description": "${description}",
  "private": false,
  "auto_init": false
}
EOF
)"

curl -fsSL \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/user/repos \
  -d "${payload}"
