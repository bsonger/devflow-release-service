# Tekton resources

This directory is the repo-managed source of truth for DevFlow image-build Tekton resources.

## Naming rules

- all names use the `devflow-tekton-` prefix
- pipelines:
  - `devflow-tekton-image-build`
  - `devflow-tekton-two-service-build-deploy`
- task names describe the exact action, for example:
  - `devflow-tekton-git-clone`
  - `devflow-tekton-go-test`
  - `devflow-tekton-image-build-and-push`
  - `devflow-tekton-image-patch`
  - `devflow-tekton-kubectl-deploy`

## Apply

```bash
kubectl apply -k deploy/tekton/base -n tekton-pipelines
```

## Current build entrypoint

`devflow-release-service` build creation uses `devflow-tekton-image-build` when it submits Tekton `PipelineRun`s.

The image patch step defaults to the in-cluster staging release-service endpoint:

- `http://devflow-release-service.devflow-staging.svc.cluster.local:8083`

## Push-only build entrypoint for stale staging artifacts

When staging is blocked by stale mutable `:staging` tags and the goal is to refresh images before a separately controlled rollout, use the repo-managed push-only pipeline:

- pipeline: `devflow-tekton-image-build-push-only`
- example `PipelineRun`: `deploy/tekton/base/pipelineruns/devflow-tekton-image-build-push-only-staging-example.yaml`

Apply the example with:

```bash
kubectl apply -f deploy/tekton/base/pipelineruns/devflow-tekton-image-build-push-only-staging-example.yaml
```

The committed example includes two independent `PipelineRun` specs for the current M001 remediation targets:
- `devflow-app-service` from `git@github.com:bsonger/devflow-app-service.git`
- `devflow-platform-orchestrator` from `git@github.com:bsonger/devflow-platform-orchestrator.git`

The `source` workspace explicitly sets `storageClassName: local-path` because this cluster does not auto-assign a default storage class for Tekton-generated PVCs. The SSH workspace uses the existing `git-ssh-secret` secret in `tekton-pipelines`.

After the builds finish, verify the new digest is actually in use in `devflow-staging` before treating the publication as successful.

## Two-service build + direct deploy entrypoint

`devflow-tekton-two-service-build-deploy` is a generic pipeline for building two services in parallel and deploying them directly with `kubectl` after both builds succeed.

Behavior:
- clones the same git repo into two workspaces
- builds and pushes `service-a` and `service-b` in parallel
- waits for both builds to finish successfully
- optionally applies manifests from each checked-out workspace
- runs `kubectl set image` against each target deployment
- waits for both deployment rollouts to complete

Example staging `PipelineRun` file:

- `deploy/tekton/base/pipelineruns/devflow-tekton-two-service-build-deploy-staging-example.yaml`

Apply it with:

```bash
kubectl apply -f deploy/tekton/base/pipelineruns/devflow-tekton-two-service-build-deploy-staging-example.yaml
```

Notes:
- `manifest-path` is optional. If omitted, the deploy task skips `kubectl apply` and only updates the deployment image.
- The deploy task assumes the Tekton runtime service account already has permission to `apply`, `set image`, and read rollout status in the target namespace.
- This pipeline performs direct cluster deployment and intentionally bypasses the Argo/release-service runtime deployment path.
