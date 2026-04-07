# devflow-release-service deploy

Apply the `tekton-pipelines` and `argocd` role bindings before rolling out the deployment so in-cluster Kubernetes calls succeed.

## Tekton resources

Repo-managed image-build Tekton resources live under `deploy/tekton/base`.

Apply them with:

```bash
kubectl apply -k deploy/tekton/base
```
