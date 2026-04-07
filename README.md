# DevFlow Release Service

`devflow-release-service` is the backend owner for `Manifest`, `Release`, and `Intent`.

## Backend Role

- own build/release control-plane resources
- receive build/release control commands
- persist control-plane lifecycle records
- trigger execution-side resources through adapters
- submit build requests to the repo-managed Tekton pipeline `devflow-tekton-image-build`
- expose public `Image` APIs as the build-facing product surface
- accept authenticated Tekton task/status writeback from `devflow-resource-observer`
- staging now runs in direct execution mode so image builds create Tekton `PipelineRun`s immediately

## Tekton source of truth

Build-image Tekton resources now live in this repository:

- `deploy/tekton/base/pipelines/`
- `deploy/tekton/base/tasks/`

Naming is normalized with the `devflow-tekton-` prefix.

## Local Run

- `go run ./cmd`
- `go build ./cmd/main.go`
- `go test ./...`

## Key Docs

- `docs/architecture.md`
- `docs/api-spec.md`
- `docs/constraints.md`
- `docs/resources/README.md`
