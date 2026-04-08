# Tekton resources

This directory is the repo-managed source of truth for DevFlow image-build Tekton resources.

## Naming rules

- all names use the `devflow-tekton-` prefix
- pipeline name: `devflow-tekton-image-build`
- task names describe the exact action, for example:
  - `devflow-tekton-git-clone`
  - `devflow-tekton-go-test`
  - `devflow-tekton-image-build-and-push`
  - `devflow-tekton-image-patch`

## Apply

```bash
kubectl apply -k deploy/tekton/base -n tekton-pipelines
```

## Current build entrypoint

`devflow-release-service` build creation uses `devflow-tekton-image-build` when it submits Tekton `PipelineRun`s.

The image patch step defaults to the in-cluster staging release-service endpoint:

- `http://devflow-release-service.devflow-staging.svc.cluster.local:8083`
