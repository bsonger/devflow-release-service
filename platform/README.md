# Platform Notes

This repository only owns the `devflow-release-service` boundary.

Runtime shape:

- `cmd/main.go` uses shared bootstrap from `../devflow-service-common`
- `pkg/router/` exposes only manifest/release/intent routes
- `pkg/api/` only contains release-side HTTP handler surfaces
- `pkg/service/` contains release metadata logic and execution-side integration points

Shared infra:

- pagination and response helpers come from `devflow-service-common/httpx`
- middleware and telemetry helpers come from `devflow-service-common/routercore` and `devflow-service-common/observability`

Operational rules:

- outbound service or external calls must emit `metrics + trace + structured log`
- `Planner -> Generator -> Evaluator` is the default harness
- when delegation is supported, sub-agents must be spawned
