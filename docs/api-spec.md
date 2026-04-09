# API Spec

## Purpose

`devflow-release-service` defines the public HTTP API surface for release control-plane resources:

- `Image`
- `Release`
- `Intent`

## Swagger

- local UI: `/swagger/index.html`
- generated source: `docs/generated/swagger/swagger.yaml`

## Endpoint Groups

### `Image`
- `GET /api/v1/images`
- `POST /api/v1/images`
- `GET /api/v1/images/{id}`
- `PATCH /api/v1/images/{id}`

### Observer writeback
- `POST /api/v1/images/tekton/tasks`
- `POST /api/v1/images/tekton/status`

### `Release`
- `GET /api/v1/releases`
- `POST /api/v1/releases`
- `GET /api/v1/releases/{id}`

### `Intent`
- `GET /api/v1/intents`
- `GET /api/v1/intents/{id}`

## Request Rules

- `POST /api/v1/images` accepts `application_id`, optional `configuration_revision_id`, optional `runtime_spec_revision_id`, and optional `branch`
- image creation submits a Tekton `PipelineRun` against `devflow-tekton-image-build` when build dispatch is enabled
- observer writeback endpoints require `X-Devflow-Observer-Token` and are intended for `devflow-resource-observer` only
- `repo_address` and image naming are resolved during image creation
- image registry target is read from backend config: `IMAGE_REGISTRY` and `IMAGE_REGISTRY_NAMESPACE`
- image names follow branch rules: `main` keeps the base name, non-`main` appends a normalized branch suffix
- image tags prefer an exact Git tag on `HEAD`; when absent they fall back to `YYYYMMDD-HHmmss`
- `POST /api/v1/releases` accepts `image_id`, optional `env`, and optional release `type`
- release creation validates that the referenced image has a valid `runtime_spec_revision_id` bound through runtime-service
- `PATCH /api/v1/images/{id}` supports build/writeback fields such as `commit_hash`, `digest`, `tag`, and status payloads
- list endpoints use `page` and `page_size`

## Response Rules

- create endpoints return `201` with `{ "data": ... }`
- get endpoints return `200` with `{ "data": ... }`
- list endpoints return `{ "data": [...], "pagination": { "page", "page_size", "total" } }`
- `PATCH /api/v1/images/{id}` returns `204 No Content`
- `Intent` is a control-plane query resource and does not expose a general public update/delete CRUD surface

## Error Rules

- invalid ID or malformed request body -> `400 invalid_argument`
- resource not found -> `404 not_found`
- image/runtime binding precondition failure -> `409 failed_precondition`
- storage or execution internal error -> `500 internal`

## Boundary Note

For repo scope and non-goals, see `docs/architecture.md`.
