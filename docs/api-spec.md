# API Spec

## Purpose

`devflow-release-service` defines the public HTTP API surface for release control-plane resources:

- `Manifest`
- `Image`
- `Release`
- `Intent`

## Swagger

- local UI: `/swagger/index.html`
- generated source: `docs/generated/swagger/swagger.yaml`

## Endpoint Groups

### `Manifest`
- `GET /api/v1/manifests`
- `POST /api/v1/manifests`
- `GET /api/v1/manifests/{id}`

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

- `POST /api/v1/manifests` accepts `application_id`, `environment_id`, and `image_id`
- `environment_id` is the environment key from platform-orchestrator bindings, for example `staging` or `base`
- manifest creation validates that:
  - the image belongs to the application
  - the environment is attached to the application
  - application routes target existing services and ports
  - the app config has renderable file data
  - the workload config exists for the application + environment
- manifest rendering builds a frozen deployment bundle from services, routes, app config, workload config, and image
- rendered workloads include registry pull secrets and mount `/etc/devflow/config/config.yaml` from the environment config ConfigMap
- when an environment-specific app/workload config is missing, manifest creation falls back to the application base config entry
- rendered workload image refs prefer `name@sha256:...`; when digest is missing they fall back to `name:tag`
- manifest creation can publish the frozen bundle as an OCI artifact
- manifest OCI publishing uses `config.yaml` `manifest_registry.*`
- `manifest_registry.plain_http=true` supports in-cluster registries such as `zot`
- when manifest-specific registry fields are unset, OCI publishing falls back to `image_registry.*`
- `POST /api/v1/images` accepts `application_id`, optional `configuration_revision_id`, optional `runtime_spec_revision_id`, and optional `branch`
- image creation submits a Tekton `PipelineRun` against `devflow-tekton-image-build` when build dispatch is enabled
- observer writeback endpoints require `X-Devflow-Observer-Token` and are intended for `devflow-resource-observer` only
- `repo_address` and image naming are resolved during image creation
- image registry target is read from mounted `config.yaml` `image_registry.registry` and `image_registry.namespace`
- image names follow branch rules: `main` keeps the base name, non-`main` appends a normalized branch suffix
- image tags prefer an exact Git tag on `HEAD`; when absent they fall back to `YYYYMMDD-HHmmss`
- `POST /api/v1/releases` accepts `manifest_id`, optional `env`, and optional release `type`
- release creation validates that the referenced manifest is `ready`
- release creation derives `application_id` and `image_id` from the frozen manifest rather than re-reading mutable config/network inputs
- release creation validates that the referenced image still has a valid `runtime_spec_revision_id` bound through runtime-service
- release sync prefers the manifest OCI artifact as the Argo CD application source; when no artifact exists it falls back to the repo plugin flow
- `PATCH /api/v1/images/{id}` supports build/writeback fields such as `commit_hash`, `digest`, `tag`, and status payloads
- list endpoints use `page` and `page_size`

## Response Rules

- create endpoints return `201` with `{ "data": ... }`
- get endpoints return `200` with `{ "data": ... }`
- list endpoints return `{ "data": [...], "pagination": { "page", "page_size", "total" } }`
- manifest detail responses include:
  - `environment_id` as the environment key string
  - `rendered_objects`
  - `rendered_yaml`
  - `artifact_repository`
  - `artifact_tag` using `service-name-YYYYMMDD-HHMMSS`
  - `artifact_ref`
  - `artifact_digest`
  - `artifact_media_type`
  - `artifact_pushed_at`
- `PATCH /api/v1/images/{id}` returns `204 No Content`
- `Intent` is a control-plane query resource and does not expose a general public update/delete CRUD surface

## Error Rules

- invalid ID or malformed request body -> `400 invalid_argument`
- resource not found -> `404 not_found`
- manifest dependency/binding/rendering precondition failure -> `409 failed_precondition`
- image/runtime binding precondition failure -> `409 failed_precondition`
- storage or execution internal error -> `500 internal`

## Boundary Note

For repo scope and non-goals, see `docs/architecture.md`.
