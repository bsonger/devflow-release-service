# Architecture

## Purpose

`devflow-release-service` is the control-plane owner for `Image`/`Manifest`, `Release`, and `Intent`.
It receives build/release commands, stores lifecycle records, and coordinates execution-side adapters without giving up ownership semantics.
It freezes build-time repository/service metadata into `Image` records backed by the current manifest store, binds release execution to configuration references, accepts watcher-driven Tekton task/status writeback from `devflow-resource-observer`, and submits build requests into the normalized Tekton pipeline `devflow-tekton-image-build`.

## Architecture Style

This repo uses a **layered control-plane backend**:

```text
router -> api -> service -> store
                    \-> runtime / external adapters
                    \-> model
```

The service layer is the domain center:
- lifecycle defaults
- build/release command rules
- public control-plane state semantics
- coordination with Tekton / Argo / Kubernetes adapters
- resolution of app/config metadata into release-owned snapshots and references

The target relational resource model is:

- `Manifest` = build artifact + frozen repository/service snapshot
- `Release` = deploy command + config reference + rollout state
- `Intent` = long-running execution tracking record

## Request Flow

### Command path

```text
Client
  -> manifest/release handler
  -> release service logic
  -> control-plane persistence
  -> runtime / external adapter dispatch
  -> HTTP response
```

### Query path

```text
Client
  -> manifest/release/intent handler
  -> release query logic
  -> persistence store
  -> HTTP response
```

## Internal Package Layout

- `cmd/main.go`
  - process entrypoint only
- `pkg/config`
  - config loading
  - runtime initialization
- `pkg/router`
  - route registration
  - module wiring
- `pkg/api`
  - manifest/release/intent handlers
- `pkg/service`
  - lifecycle rules
  - command/query logic
  - adapter coordination
  - app/config reference resolution
- `pkg/runtime`
  - execution-mode switching
- `pkg/model`
  - `Manifest`, `Release`, `Intent`, related nested types
- `pkg/store`
  - persistence primitives

## External Dependencies

- `Gin`
- PostgreSQL persistence
- `devflow-service-common`
- Tekton-related clients / resources
- repo-managed Tekton pipeline/task YAML under `deploy/tekton/base`
- Argo CD / Kubernetes adapters

## Non-Goals

- `Project` / `Application` metadata ownership
- `Configuration` ownership
- verify-service domain ownership and verify ingress semantics
- platform UI aggregation ownership
- environment-variable ownership outside configuration references

## Swagger generation

- Run `scripts/regen-swagger.sh` (or the Dockerfile) to generate `docs/generated/swagger`.
- Use `scripts/build.sh` to regenerate swagger and rebuild the binary locally.
- Export/release tooling expects the bundle in `docs/generated/swagger`, so keep it checked in.

## Swagger generation

- Build image runs `swag init -g cmd/main.go --parseDependency -o docs/generated/swagger` before compiling the binary.
- Generated swagger files live under `docs/generated/swagger`; preserve them if you export the repo.
- `scripts/export_service_repo.sh` expects the generated bundle in `docs/generated/swagger` when copying docs.
