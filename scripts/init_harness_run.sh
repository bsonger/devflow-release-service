#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <slug> <request text>" >&2
  exit 1
fi

slug="$1"
shift
request_text="$*"

date_prefix="$(date +%Y%m%d)"
run_dir="agents/runs/${date_prefix}-${slug}"

mkdir -p "${run_dir}"

cat > "${run_dir}/request.md" <<EOF
# Request

- Raw request:
  - ${request_text}
EOF

cp agents/templates/product-spec.md "${run_dir}/product-spec.md"
cp agents/templates/sprint-contract.md "${run_dir}/sprint-01-contract.md"
cp agents/templates/evaluator-report.md "${run_dir}/evaluator-report.md"
cp agents/templates/handoff.md "${run_dir}/handoff.md"

echo "${run_dir}"
