# Observability

## Shared Baseline

This repo follows the shared telemetry contract implemented in `devflow-service-common`.

- structured logs with shared runtime fields
- `devflow_http_*` ingress metrics
- `devflow_dependency_*` outbound metrics
- standard server/client spans with service-defined business attributes
- optional diagnostics only for `pprof` and Pyroscope

## Repo-Local Focus

`devflow-release-service` should add resource context for:

- `manifest`
- `release`
- `intent`

Recommended structured fields:

- `resource`
- `resource_id`
- `application_id`
- `manifest_id`
- `release_id`
- `intent_id`
- `dependency`
- `result`
- `error_code`

## Outbound Dependencies

Calls to these systems should always go through shared dependency telemetry helpers where possible:

- runtime-service
- Tekton
- Argo CD
- Git or manifest repositories

## Profile

- `pprof` is disabled by default
- Pyroscope is disabled by default
- both are enabled only through explicit runtime configuration
