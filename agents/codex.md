# Codex Repo Context

This file provides working context only.

Read authoritative documentation here:

- global architecture: `devflow-control/docs/system/architecture.md`
- boundaries: `devflow-control/docs/system/boundaries.md`
- repo-local architecture: `docs/architecture.md`
- repo-local API surface: `docs/api-spec.md`
- repo-local resources: `docs/resources/*.md`

Do not treat this file as a source of resource or boundary truth.

## 模型归属
- `Application`、`Manifest`、`Release`、`Configuration` 当前定义在 `pkg/model/`。
- `devflow-common` 保留为 client 层依赖，不再作为本仓库的领域模型权威来源。

## 典型请求流程
1. 客户端发起 HTTP 请求。
2. `pkg/router` 将请求分发到对应 Handler。
3. `pkg/api` 中的 Handler 解析请求并调用 `pkg/service`。
4. `pkg/service` 执行业务逻辑，必要时访问外部系统或数据库。
5. 返回响应给客户端。

## 接口示例
以下示例以 `http://localhost:8080` 为基础地址，按需替换。

### 获取应用列表
```bash
curl -X GET "http://localhost:8080/api/v1/applications"
```

示例响应：
```json
[
  {
    "id": "app_001",
    "name": "web-portal",
    "project_name": "demo",
    "repo_url": "https://git.example.com/demo/web-portal",
    "replica": 2,
    "internet": "external",
    "status": "Running"
  }
]
```

### 创建 Manifest
```bash
curl -X POST "http://localhost:8080/api/v1/manifests" \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "app_001",
    "application_name": "web-portal",
    "branch": "main",
    "git_repo": "https://git.example.com/demo/web-portal",
    "replica": 2,
    "internet": "external"
  }'
```

示例响应：
```json
{
  "id": "manifest_001",
  "name": "web-portal-main-20240201",
  "application_id": "app_001",
  "branch": "main",
  "git_repo": "https://git.example.com/demo/web-portal",
  "status": "Pending"
}
```

### Patch Manifest（更新 digest / commit_hash）
```bash
curl -X PATCH "http://localhost:8080/api/v1/manifests/manifest_001" \
  -H "Content-Type: application/json" \
  -d '{
    "commit_hash": "a1b2c3d4",
    "digest": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
  }'
```

示例响应：
```json
{
  "message": "ok"
}
```

## 配置与运行
- 本地运行：`go run ./cmd`（默认读取 `config/config.yaml`）。
- 构建二进制：`go build -o devflow cmd/main.go`。
- 生成 Swagger：`swag init -g cmd/main.go --parseDependency -o docs/generated/swagger`。

## 运行时依赖
- 需要有效 kubeconfig：`$HOME/.kube/config` 或集群内配置。
- 若使用容器构建：`docker build -t devflow:local .`（会生成 Swagger）。

## 维护与扩展建议
- 新增 API 时优先在 `pkg/api` 与 `pkg/service` 分层实现。
- 配置项新增时同步更新 `config/config.yaml`。
- 若路由或模型变更，重新生成 Swagger 并检查 `docs/`。
