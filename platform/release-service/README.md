# Release Service

职责：

- 提供 `Manifest` 和 `Job` 的 API
- 提供 `Intent` 查询 API
- 负责 release intent、构建/发布编排的服务边界
- 后续从这里继续抽离 Tekton / Argo 的直接执行逻辑

当前运行模式：

- 默认以 `intent` 模式启动
- `Manifest.Create` 与 `Job.Create` 会落库并创建 execution intent
- 创建响应会返回 `execution_intent_id`
- `Job.Create` 会按应用发布策略初始化默认 `steps`
- 不再在服务主流程内直接调用 Tekton / Argo

控制面查询接口：

- `GET /api/v1/intents`
- `GET /api/v1/intents/:id`
- 支持按 `kind`、`status`、`application_id`、`manifest_id`、`job_id`、`claimed_by` 等字段过滤

配套 worker：

- `go run ./platform/release-service/cmd/worker`
- worker 会轮询 `execution_intents` 集合中的 `Pending` 记录
- worker 通过 `claimed_by` + `lease_expires_at` 认领待执行 intent
- 当前仍按“顺序执行一个 batch”实现，但已经避免多个 worker 同时消费同一条 pending intent

可选环境变量：

- `RELEASE_WORKER_ID`
- `RELEASE_WORKER_INTERVAL_SECONDS`
- `RELEASE_WORKER_BATCH_SIZE`
- `RELEASE_WORKER_LEASE_SECONDS`
- `RELEASE_WORKER_METRICS_PORT`
- `RELEASE_WORKER_PPROF_PORT`

当前复用的现有实现：

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

可观测性：

- traces / metrics 统一走 OpenTelemetry
- logs 必须带 `trace_id` / `span_id`
- worker 链路必须覆盖 claim / lease / dispatch / verify writeback 这些关键 span
- API 进程上报的 OTel `service.name` 为 `release-service`
- worker 进程上报的 OTel `service.name` 为 `release-worker`
- 规范见 `agents/reference/observability.md`
