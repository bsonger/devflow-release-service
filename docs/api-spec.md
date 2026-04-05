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
- `GET /api/v1/manifests/:id`
- `PATCH /api/v1/manifests/:id`

### `Release`
- `GET /api/v1/releases`
- `POST /api/v1/releases`
- `GET /api/v1/releases/:id`

### `Intent`
- `GET /api/v1/intents`
- `GET /api/v1/intents/:id`

## Request Rules

- `POST /api/v1/manifests` accepts build command inputs such as `application_id` and `branch`
- `repo_address` and `services` are resolved from app-service and frozen into the created `Manifest`
- `POST /api/v1/releases` may accept platform-relevant fields: `manifest_id`, optional `configuration_id`, optional `configuration_revision_id`, optional `env` (default `prod`), and release `type`
- if both `configuration_id` and `configuration_revision_id` are present, the service must validate that they belong together
- `PATCH /api/v1/manifests/:id` only supports repo-defined patch fields such as `commit_hash` and `digest`
- list endpoints use the common pagination parameters and filtering conventions in this repo

## Response Rules

- create endpoints return the common create-response shape
- list endpoints return common pagination headers
- `Intent` is a control-plane query resource and does not expose a general public update/delete CRUD surface

## Error Rules

- invalid ID -> `400`
- resource not found -> `404`
- storage/execution uncategorized internal error -> `500`

## Boundary Note

For repo scope and non-goals, see `docs/architecture.md`.
