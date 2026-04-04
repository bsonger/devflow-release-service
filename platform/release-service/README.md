# Release Service Platform Notes

## Purpose

This file is the repo-local runtime note for `devflow-release-service`.
For public API shape, ownership, and resource details, prefer:
- `../README.md`
- `../docs/`
- `../docs/resources/`

## Runtime entrypoints

- process entry: `cmd/main.go`
- shared bootstrap: `../devflow-service-common/bootstrap`
- router root: `pkg/router/router.go`

## Main local code paths

- manifest routes / handlers / logic:
  - `pkg/router/manifest.go`
  - `pkg/api/manifest.go`
  - `pkg/service/manifest.go`
- release routes / handlers / logic:
  - `pkg/router/release.go`
  - `pkg/api/release.go`
  - `pkg/service/release.go`
- intent routes / handlers / logic:
  - `pkg/router/intent.go`
  - `pkg/api/intent.go`
  - `pkg/service/intent.go`

## Execution-side integration points

- Tekton build dispatch is triggered from `pkg/service/manifest.go`
- Argo / Kubernetes Application sync is triggered from `pkg/service/release.go`

## Platform dependencies

- shared response / pagination: `devflow-service-common/httpx`
- shared middleware: `devflow-service-common/routercore`
- shared observability: `devflow-service-common/observability`

## Service identity

- OTel `service.name`: `release-service`
- typical ports:
  - `RELEASE_SERVICE_PORT`
  - `RELEASE_SERVICE_METRICS_PORT`
  - `RELEASE_SERVICE_PPROF_PORT`
