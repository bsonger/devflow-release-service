# API Spec

## Purpose

`devflow-release-service` defines the converged public HTTP API surface for the release control-plane resources:
- `Manifest`
- `Release`
- `Intent`

## Endpoint Groups

### `Manifest`
- `GET /api/v1/manifests`
- `POST /api/v1/manifests`
- `GET /api/v1/manifests/{id}`
- `PATCH /api/v1/manifests/{id}`

### `Release`
- `GET /api/v1/releases`
- `POST /api/v1/releases`
- `GET /api/v1/releases/{id}`

### `Intent`
- `GET /api/v1/intents`
- `GET /api/v1/intents/{id}`

## Request Rules

- `POST /api/v1/manifests` accepts `application_id`, optional `configuration_revision_id`, optional `runtime_spec_revision_id`, and optional `branch`
- `repo_address` and manifest naming are resolved during manifest creation
- `POST /api/v1/releases` accepts `manifest_id`, optional `env`, and optional release `type`
- release creation validates that the referenced manifest has a valid `runtime_spec_revision_id` bound through runtime-service
- `PATCH /api/v1/manifests/{id}` only supports patch fields such as `commit_hash` and `digest`
- list endpoints use `page` and `page_size`

## Response Rules

- create endpoints return `201` with `{ "data": ... }`
- get endpoints return `200` with `{ "data": ... }`
- list endpoints return `{ "data": [...], "pagination": { "page", "page_size", "total" } }`
- `PATCH /api/v1/manifests/{id}` returns `204 No Content`
- `Intent` is a control-plane query resource and does not expose a general public update/delete CRUD surface

## Error Rules

- invalid ID or malformed request body -> `400 invalid_argument`
- resource not found -> `404 not_found`
- manifest/runtime binding precondition failure -> `409 failed_precondition`
- storage/execution uncategorized internal error -> `500 internal`

## Boundary Note

For repo scope and non-goals, see `docs/architecture.md`.

## Swagger Note

Generated Swagger artifacts must stay aligned with the current PostgreSQL-backed API contract. Regenerate them after route, request, or response changes.
