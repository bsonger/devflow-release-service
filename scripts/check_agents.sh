#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

fail() {
  echo "agents check failed: $*" >&2
  exit 1
}

require_path() {
  local path="$1"
  [ -e "${path}" ] || fail "missing ${path}"
}

require_text() {
  local file="$1"
  local pattern="$2"
  grep -Fq -- "${pattern}" "${file}" || fail "${file} missing pattern: ${pattern}"
}

required_paths=(
  "AGENTS.md"
  "agents/README.md"
  "agents/manifest.yaml"
  "agents/templates/standard-doc.md"
  "agents/roles/planner.md"
  "agents/roles/generator.md"
  "agents/roles/evaluator.md"
  "agents/protocols/startup.md"
  "agents/protocols/workflow.md"
  "agents/protocols/quality-gates.md"
  "agents/protocols/handoff.md"
  "agents/protocols/escalation.md"
  "agents/reference/worker-constraints.md"
  "agents/reference/evaluator-rubric.md"
  "agents/reference/observability.md"
  "scripts/init_harness_run.sh"
)

for path in "${required_paths[@]}"; do
  require_path "${path}"
done

[ -x "scripts/init_harness_run.sh" ] || fail "scripts/init_harness_run.sh is not executable"

require_text "agents/manifest.yaml" "canonical_root: agents"
require_text "agents/manifest.yaml" "run_initializer: scripts/init_harness_run.sh"
require_text "agents/manifest.yaml" "structure_checker: scripts/check_agents.sh"
require_text "agents/manifest.yaml" "- planner"
require_text "agents/manifest.yaml" "- generator"
require_text "agents/manifest.yaml" "- evaluator"

standard_files=(
  "agents/roles/planner.md"
  "agents/roles/generator.md"
  "agents/roles/evaluator.md"
  "agents/protocols/startup.md"
  "agents/protocols/workflow.md"
  "agents/protocols/quality-gates.md"
  "agents/protocols/handoff.md"
  "agents/protocols/escalation.md"
  "agents/reference/application.md"
  "agents/reference/manifest.md"
  "agents/reference/job.md"
  "agents/reference/intent.md"
  "agents/reference/project.md"
  "agents/reference/api-contract.md"
  "agents/reference/rollout.md"
  "agents/reference/tekton.md"
  "agents/reference/worker-constraints.md"
  "agents/reference/evaluator-rubric.md"
  "agents/reference/observability.md"
  "agents/patterns/service.md"
  "agents/patterns/controller.md"
  "agents/patterns/status-flow.md"
  "agents/patterns/mongo.md"
  "agents/patterns/reconcile.md"
  "agents/skills/base.md"
  "agents/skills/controller.md"
  "agents/skills/harness.md"
  "agents/skills/infra.md"
  "agents/skills/mongo.md"
  "agents/skills/service.md"
  "agents/runs/README.md"
  "agents/examples/README.md"
)

required_sections=(
  "## Purpose"
  "## Scope"
  "## Must"
  "## Must Not"
  "## Outputs"
  "## Pass/Fail"
)

for file in "${standard_files[@]}"; do
  require_path "${file}"
  for section in "${required_sections[@]}"; do
    require_text "${file}" "${section}"
  done
done

echo "agents control-plane check passed"
