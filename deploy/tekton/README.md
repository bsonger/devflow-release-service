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
