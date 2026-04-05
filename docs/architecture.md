# Architecture

## Purpose

`devflow-release-service` is the control-plane owner for `Manifest`, `Release`, and `Intent`.
It receives build/release commands, stores lifecycle records, and coordinates execution-side adapters without giving up ownership semantics.
It freezes build-time repository/service metadata into `Manifest` and binds release execution to configuration references.

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
- Argo CD / Kubernetes adapters

## Non-Goals

- `Project` / `Application` metadata ownership
- `Configuration` ownership
- verify webhook / writeback ingress
- platform UI aggregation ownership
- environment-variable ownership outside configuration references
