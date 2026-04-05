#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: export_service_repo.sh <app-service|config-service|release-service|verify-service> [output-root]
EOF
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
service="${1:-}"
output_root="${2:-/tmp/devflow-split}"

[ -n "${service}" ] || usage

require_cmd rsync
require_cmd perl
require_cmd git

repo_name=""
module_path=""
platform_dir=""
binary_name=""
service_title=""

case "${service}" in
  app-service)
    repo_name="devflow-app-service"
    module_path="github.com/bsonger/devflow-app-service"
    platform_dir="platform/app-service"
    binary_name="devflow-app-service"
    service_title="Devflow App Service"
    ;;
  config-service)
    repo_name="devflow-config-service"
    module_path="github.com/bsonger/devflow-config-service"
    platform_dir="platform/config-service"
    binary_name="devflow-config-service"
    service_title="Devflow Config Service"
    ;;
  release-service)
    repo_name="devflow-release-service"
    module_path="github.com/bsonger/devflow-release-service"
    platform_dir="platform/release-service"
    binary_name="devflow-release-service"
    service_title="Devflow Release Service"
    ;;
  verify-service)
    repo_name="devflow-verify-service"
    module_path="github.com/bsonger/devflow-verify-service"
    platform_dir="platform/verify-service"
    binary_name="devflow-verify-service"
    service_title="Devflow Verify Service"
    ;;
  *)
    usage
    ;;
esac

target_dir="${output_root}/${repo_name}"

if [ -e "${target_dir}" ]; then
  echo "target already exists: ${target_dir}" >&2
  echo "remove it manually or choose another output-root" >&2
  exit 1
fi

mkdir -p "${output_root}" "${target_dir}"

rsync -a \
  --exclude='.git/' \
  --exclude='.idea/' \
  --exclude='.vscode/' \
  --exclude='tmp/' \
  --exclude='node_modules/' \
  "${repo_root}/" "${target_dir}/"

rm -rf "${target_dir}/cmd"
mkdir -p "${target_dir}/cmd"
cp "${target_dir}/${platform_dir}/cmd/main.go" "${target_dir}/cmd/main.go"

rm -rf \
  "${target_dir}/platform/app-service" \
  "${target_dir}/platform/config-service" \
  "${target_dir}/platform/release-service" \
  "${target_dir}/platform/verify-service"

mkdir -p "${target_dir}/${platform_dir}"
cp "${repo_root}/${platform_dir}/README.md" "${target_dir}/${platform_dir}/README.md"
if [ -f "${repo_root}/${platform_dir}/controller-integration.md" ]; then
  cp "${repo_root}/${platform_dir}/controller-integration.md" "${target_dir}/${platform_dir}/controller-integration.md"
fi

cat > "${target_dir}/README.md" <<EOF
# ${service_title}

This repository was exported from \`github.com/bsonger/devflow\`.

GitHub target:

- \`git@github.com:bsonger/${repo_name}.git\`

Go module:

- \`${module_path}\`

Current scope:

- service entrypoint from \`${platform_dir}/cmd/main.go\`
- shared bootstrap from \`platform/shared/bootstrap\`
- current shared domain/runtime packages from \`pkg/\`

Notes:

- This is a first-stage split repo.
- Shared packages are still copied from the monorepo so the service can compile independently.
- A later cleanup phase can move stable shared pieces into \`devflow-common\` or another shared module.
EOF

cat > "${target_dir}/Dockerfile" <<EOF
FROM registry.cn-hangzhou.aliyuncs.com/devflow/golang:1.25.7 AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOROOT=\$(go env GOROOT) swag init -g cmd/main.go --parseDependency -o docs/generated/swagger
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${binary_name} ./cmd

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/${binary_name} ./${binary_name}
COPY --from=builder /app/docs ./docs

RUN adduser -D devuser
USER devuser

EXPOSE 8080

ENTRYPOINT ["./${binary_name}"]
EOF

perl -0pi -e "s|^module github.com/bsonger/devflow\$|module ${module_path}|m" "${target_dir}/go.mod"

while IFS= read -r -d '' file; do
  perl -0pi -e "s|github\\.com/bsonger/devflow(?!-)|${module_path}|g" "${file}"
done < <(find "${target_dir}" -type f \( \
  -name '*.go' -o \
  -name '*.mod' -o \
  -name '*.sum' -o \
  -name '*.md' -o \
  -name '*.yaml' -o \
  -name '*.yml' -o \
  -name '*.json' \
\) -print0)

git -C "${target_dir}" init -b main >/dev/null
git -C "${target_dir}" remote add origin "git@github.com:bsonger/${repo_name}.git"

if command -v swag >/dev/null 2>&1; then
  (
    cd "${target_dir}"
    GOROOT="$(go env GOROOT)" swag init -g cmd/main.go --parseDependency -o docs/generated/swagger >/dev/null
  )
fi

(
  cd "${target_dir}"
  go test -mod=mod ./... >/dev/null
)

echo "${target_dir}"
