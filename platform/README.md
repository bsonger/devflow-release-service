# Platform Services

本目录用于承载拆分后的平台服务骨架。

当前规划：

- `app-service/`：`Project`、`Application` 元数据 CRUD 与 active manifest 绑定。
- `config-service/`：`Configuration` 元数据 CRUD。
- `release-service/`：`Manifest`、`Job`、`Intent` 查询、发布意图与后续 release 编排。
- `verify-service/`：Tekton / Argo / controller 事件接入、校验与状态回写。

共享基础：

- 领域模型统一来自 `pkg/model/`
- 当前业务逻辑仍复用 `pkg/api`、`pkg/service`、`pkg/router`
- `devflow-common` 继续仅作为外部 client 来源使用
- `pkg/runtime` 控制当前进程是 `direct` 还是 `intent` 执行模式
- `execution_intents` 用于承载 build / release 执行意图
- 所有 service / worker 统一遵守 `agents/reference/observability.md`

现阶段目标：

- 先把服务边界和独立入口分出来
- 保持单体入口兼容
- 再逐步把 `release-service` 与 `verify-service` 从直接执行逻辑中抽离

运行时约定：

- 每个 service 会以自己的名字覆盖 OTel `service.name`
- `*_METRICS_PORT` 用于暴露 `/metrics`
- `*_PPROF_PORT` 用于暴露 `/debug/pprof/*`

仓库拆分：

- 目标拆分仓库定义见 `platform/repositories.md`
- 导出单个服务仓库：`bash scripts/export_service_repo.sh <service>`
- 批量导出 4 个服务仓库：`bash scripts/export_service_repos.sh`
- 若本机具备 `GITHUB_TOKEN`，可继续执行 `bash scripts/create_service_repos.sh`
