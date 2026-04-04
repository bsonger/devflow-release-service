# Release Service

职责：

- 提供 `Manifest` 的查询、创建、patch
- 提供 `Release` 的查询、创建
- 提供 `Intent` 的查询
- 负责 release/build 控制面资源和后续执行编排衔接

当前实现：

- `cmd/main.go` 通过 `devflow-service-common/bootstrap` 启动
- `pkg/api/manifest.go`
- `pkg/api/job.go`
- `pkg/api/intent.go`
- `pkg/service/manifest.go`
- `pkg/service/job.go`
- `pkg/service/intent.go`
- `pkg/router/manifest.go`
- `pkg/router/job.go`
- `pkg/router/intent.go`

建议端口：

- `RELEASE_SERVICE_PORT`
- `RELEASE_SERVICE_METRICS_PORT`
- `RELEASE_SERVICE_PPROF_PORT`

运行时：

- 上报的 OTel `service.name` 为 `release-service`
- 任何 outbound service / external call 都必须带 `metrics + trace + structured log`
- 默认 harness 为 `Planner -> Generator -> Evaluator`，并且支持 delegation 时必须真实启动 sub-agents

接口：

- `GET /api/v1/manifests`
- `POST /api/v1/manifests`
- `GET /api/v1/manifests/:id`
- `PATCH /api/v1/manifests/:id`
- `GET /api/v1/releases`
- `POST /api/v1/releases`
- `GET /api/v1/releases/:id`
- `GET /api/v1/intents`
- `GET /api/v1/intents/:id`
