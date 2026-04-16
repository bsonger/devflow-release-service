# DevFlow Release Service

`devflow-release-service` is the backend owner for `Manifest`, `Image`, `Release`, and `Intent`.

## Role

- own build/release control-plane resources
- persist frozen manifest snapshots before release execution
- receive build/release control commands
- persist control-plane lifecycle records
- trigger execution-side resources through adapters
- submit build requests to the repo-managed Tekton pipeline `devflow-tekton-image-build`
- expose public `Image` APIs as the build-facing product surface
- expose public `Manifest` APIs for rendered deployment bundles
- accept authenticated Tekton task/status writeback from `devflow-resource-observer`
- staging now runs in direct execution mode so image builds create Tekton `PipelineRun`s immediately

## Tekton source of truth

Build-image Tekton resources now live in this repository:

- `deploy/tekton/base/pipelines/`
- `deploy/tekton/base/tasks/`

Naming is normalized with the `devflow-tekton-` prefix.

## Key Commands

- `go run ./cmd`
- `go build ./cmd/main.go`
- `go test ./...`
- Swagger UI: `/swagger/index.html`
- Staging Swagger UI: `/api/v1/release/swagger/index.html`

## Key Docs

- `docs/README.md`
- `scripts/README.md`
- `docs/architecture.md`
- `docs/constraints.md`
- `docs/observability.md`
- `docs/api-spec.md`
- `docs/resources/README.md`
- `docs/generated/swagger/swagger.yaml`
