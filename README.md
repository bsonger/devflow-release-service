# Devflow Release Service

`devflow-release-service` 只负责 `Manifest`、`Release`、`Intent`。

边界：

- 对外 HTTP 资源只有 `manifests`、`releases`、`intents`
- 不提供 `Project`、`Application`、`Configuration`、`Verify` 的 API、router 或 Swagger
- 启动骨架、通用中间件、分页/响应、观测基础设施优先复用 `../devflow-service-common`

仓库文档：

- [架构](docs/architecture.md)
- [接口规范](docs/api-spec.md)
- [约束](docs/constraints.md)
- [观测规范](docs/observability.md)
- [Harness](docs/harness.md)
- [资源说明](docs/resources/README.md)

运行约定：

- 任何调用其他服务或外部系统的代码都必须同时产出 `metrics + trace + structured log`
- 默认 harness 为 `Planner -> Generator -> Evaluator`
- 运行时支持 delegation 时，必须真实启动对应 sub-agent，不允许只在单 agent 内口头模拟

常用命令：

- `go run ./cmd`
- `go build ./cmd/main.go`
- `go test ./...`
- `swag init -g cmd/main.go --parseDependency`
