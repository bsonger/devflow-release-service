# API Spec

## Purpose

`devflow-release-service` exposes public HTTP APIs for the release control-plane resources:
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

- `POST /api/v1/releases` may accept platform-relevant fields: `manifest_id`, `configuration_id` (optional), `env` (optional, default `prod`), `type`
- `PATCH /api/v1/manifests/:id` only supports repo-defined patch fields such as `commit_hash` and `digest`
- list endpoints use the common pagination parameters and filtering conventions in this repo

## Response Rules

- create endpoints return the common create-response shape
- list endpoints return common pagination headers
- `Intent` is a control-plane query resource and does not expose a general public update CRUD surface

## Error Rules

- invalid ObjectID -> `400`
- resource not found -> `404`
- storage/execution uncategorized internal error -> `500`

## Non-Goals

This repo does not expose public CRUD for:
- `Project`
- `Application`
- `Configuration`
- verify ingress
