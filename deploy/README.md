# devflow-release-service deploy

Apply the RBAC resources before rolling out the deployment so in-cluster Kubernetes calls succeed.

## Apply order

```bash
# 1. Service account
kubectl apply -f deploy/serviceaccount.yaml

# 2. Tekton permissions
kubectl apply -f deploy/role-tekton.yaml
kubectl apply -f deploy/rolebinding-tekton.yaml

# 3. Argo CD permissions (applications + appprojects)
kubectl apply -f deploy/role-argocd.yaml
kubectl apply -f deploy/rolebinding-argocd.yaml

# 4. Bootstrap permissions (namespaces + secrets cluster-wide)
kubectl apply -f deploy/clusterrole-bootstrap.yaml
kubectl apply -f deploy/clusterrolebinding-bootstrap.yaml

# 5. Workload
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
```

## Required RBAC summary

| Resource | Name | Namespace | Scope | Purpose |
|----------|------|-----------|-------|---------|
| ServiceAccount | devflow-release-service | devflow-staging | — | workload identity |
| Role | devflow-release-service-tekton | tekton-pipelines | namespace | Tekton pipeline operations |
| RoleBinding | devflow-release-service-tekton | tekton-pipelines | namespace | — |
| Role | devflow-release-service-argocd | argocd | namespace | Argo CD applications + appprojects |
| RoleBinding | devflow-release-service-argocd | argocd | namespace | — |
| ClusterRole | devflow-release-service-bootstrap | — | cluster | namespace + secret bootstrap prerequisites |
| ClusterRoleBinding | devflow-release-service-bootstrap | — | cluster | — |

## Tekton resources

Repo-managed image-build Tekton resources live under `deploy/tekton/base`.

Apply them with:

```bash
kubectl apply -k deploy/tekton/base
```
