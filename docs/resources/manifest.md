# Manifest Resource

## Purpose

`Manifest` is the frozen deployment snapshot owned by `devflow-release-service`.
It is rendered from:
- `Image`
- `Service`
- `Route`
- `AppConfig`
- `WorkloadConfig`

The rendered result is the deployment bundle that release execution should consume, not a mutable source model.

## Public API

- `POST /api/v1/manifests`
- `GET /api/v1/manifests`
- `GET /api/v1/manifests/{id}`

## Create contract

`POST /api/v1/manifests` accepts:
- `application_id`
- `environment_id`
- `image_id`

The service resolves the effective owner resources, validates route targets, chooses the image reference with `digest` first and `tag` second, and persists:
- source snapshots
- rendered objects
- rendered multi-document YAML

## Returned shape

Manifest detail includes:
- source snapshot fields
- `rendered_objects`
- `rendered_yaml`
- `status`

List responses return manifest metadata only.

## Ownership rules

- `release-service` owns `Manifest`
- `verify-service` does not own manifest CRUD
- `platform-orchestrator` may aggregate manifest-related views, but must not become the source of truth

## Deployment meaning

`Manifest` is the expected deployment bundle.
It is not the live cluster observation view.
