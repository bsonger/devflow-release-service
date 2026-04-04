# Repository Guidelines

## Boundary

- This repository is `devflow-release-service`.
- Public surface is `Manifest`, `Release`, and `Intent` only.
- Do not reintroduce `Project`, `Application`, `Configuration`, or `Verify` public routes, router modules, Swagger tags, or handlers.

## Structure

- `cmd/main.go` uses shared bootstrap from `../devflow-service-common`.
- `pkg/api/` contains only manifest/release/intent handlers.
- `pkg/service/` contains release-side orchestration and metadata logic.
- `pkg/router/` contains release-service route registration.
- `pkg/config/` initializes config, observability, Mongo, and runtime dependencies.
- `docs/` contains the repository-level architecture, API, constraints, observability, and harness docs.

## Required Rules

- Any outbound service or external call must emit `metrics + trace + structured log`.
- Do not add high-cardinality business IDs to metrics labels.
- Default harness is `Planner -> Generator -> Evaluator`.
- When the runtime supports delegation, the harness must spawn those roles as separate sub-agents.
- Non-trivial work should use a run directory under `agents/runs/`.

## Doc And API Hygiene

- Regenerate Swagger after route or handler changes.
- Keep `README.md`, `AGENTS.md`, `agents/protocols/startup.md`, `agents/reference/api-contract.md`, `agents/reference/observability.md`, and `docs/*.md` aligned with the actual boundary.
- Do not reintroduce dead service/model/router/bootstrap files.
