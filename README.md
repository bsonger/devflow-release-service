# Devflow Release Service

`devflow-release-service` is the backend owner for `Manifest`, `Release`, and `Intent`.

## Backend Role

- own build/release control-plane resources
- receive build/release control commands
- persist control-plane lifecycle records
- trigger execution-side resources through adapters

## Backend Architecture

This repo uses a **layered control-plane backend** with separable command/query paths:

```text
cmd
 -> config
 -> router
 -> api
 -> service
    -> store
    -> runtime / external adapters
 -> model
```

### Package responsibilities

- `cmd/`: service startup
- `pkg/config`: config loading and runtime init
- `pkg/router`: Gin router and middleware wiring
- `pkg/api`: manifest/release/intent handlers
- `pkg/service`: lifecycle rules, command/query behavior, adapter coordination
- `pkg/runtime`: execution-mode and runtime wiring
- `pkg/model`: control-plane models

## Non-Goals

- no `Project` ownership
- no `Application` ownership
- no `Configuration` ownership
- no verify ingress ownership
- no platform dashboard ownership

## Key Docs

- `docs/architecture.md`
- `docs/api-spec.md`
- `docs/constraints.md`
- `docs/resources/README.md`

## Local Run

- `go run ./cmd`
- `go build ./cmd/main.go`
- `go test ./...`
