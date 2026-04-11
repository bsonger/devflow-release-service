# Observability

## Purpose

`devflow-release-service` emits the shared backend telemetry baseline plus build/release lifecycle context.

## Logs

Required structured fields:
- `resource`
- `resource_id`
- `application_id`
- `image_id`
- `release_id`
- `intent_id`
- `dependency`
- `result`
- `error_code`

## Metrics

- use shared `devflow_http_*` ingress metrics
- use shared `devflow_dependency_*` metrics for runtime-service, Tekton, Argo CD, and Git interactions
- forbid high-cardinality labels such as commit hashes, branch names, or pipeline task IDs

## Tracing

- every business HTTP request should create a server span
- every outbound dependency call should create a client span
- preserve correlation across build/release orchestration and verify writeback flows

## Health and readiness

- expose `/healthz`, `/readyz`, and `/metrics`
- exclude `/swagger/*` and diagnostics endpoints from business telemetry rollups

## Failure modes

Watch for:
- manifest rendering and binding failures
- build trigger failures
- runtime binding/precondition failures
- external execution dependency failures

## Dashboards and runbooks

Use the shared backend dashboard/runbook set plus release-specific dashboards when they exist.
