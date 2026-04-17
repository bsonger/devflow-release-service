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
- `environment_id` such as `staging` or `base`
- `image_id`

The service resolves the effective owner resources, validates route targets, chooses the image reference with `digest` first and `tag` second, and persists:
- rendered workload Deployment now includes `imagePullSecrets: [aliyun-docker-config]`
- rendered workload Deployment mounts app-config YAML files from the environment-level `*-config` ConfigMap using the frozen `app_config_snapshot.mount_path`
- rendered workload Deployment / PodTemplate now includes `devflow.application/id` and `devflow.environment/id` labels so live Pod watchers can route cluster state back to runtime-service
- source snapshots
- rendered objects
- rendered multi-document YAML
- OCI artifact publication metadata when registry publishing is enabled

Manifest creation does **not** look up or persist `runtime_spec_revision_id`.
Runtime binding is validated later during release creation from `Image.runtime_spec_revision_id`.

Environment resolution rules:
- `environment_id` is the platform environment key, not a UUID
- app-level `Service` / `Route` resources are treated as the shared base
- `AppConfig` / `WorkloadConfig` first try the target environment
- if the target environment has no matching config entry, release-service falls back to the application base entry

## OCI packaging

When manifest registry publishing is enabled, manifest creation also:
- packages deployable `manifest.yaml` into a single OCI tar+gzip layer; manifest metadata stays in the OCI config payload
- pushes the frozen bundle as an OCI artifact owned by `release-service`
- tags the OCI artifact as `service-name-YYYYMMDD-HHMMSS`
- records:
  - `artifact_repository`
  - `artifact_tag`
  - `artifact_ref`
  - `artifact_digest`
  - `artifact_media_type`
  - `artifact_pushed_at`

Registry config resolution order:
- `config.yaml` `manifest_registry.registry`
- `config.yaml` `manifest_registry.namespace`
- optional `config.yaml` `manifest_registry.repository` with default `manifests`
- optional `config.yaml` `manifest_registry.plain_http=true` for in-cluster HTTP registries
- fallback to `config.yaml` `image_registry.*`

## Returned shape

Manifest detail includes:
- source snapshot fields
- `rendered_objects`
- `rendered_yaml`
- OCI artifact metadata fields when available
- `status`

List responses return manifest metadata only.

## Ownership rules

- `release-service` owns `Manifest`
- `verify-service` does not own manifest CRUD
- `platform-orchestrator` may aggregate manifest-related views, but must not become the source of truth

## Deployment meaning

`Manifest` is the expected deployment bundle.
It is not the live cluster observation view.
