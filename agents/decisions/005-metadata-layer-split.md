# 005 Devflow 作为元数据层

- 决策：
  - 将 `devflow` 定位为元数据层 / 控制面，保留资源模型、状态读写、查询 API、意图编排与事件回写。
  - 将直接操作 Tekton 与 Argo 的执行逻辑逐步拆到独立服务。
- 原因：
  - 当前 `ManifestService` 直接执行 Tekton 构建，`JobService` 直接执行 Argo 发布，控制面与执行面耦合过深。
  - `pkg/config/config.go` 启动时同时依赖 Mongo、Tekton、Argo，使元数据 API 无法作为轻量服务独立运行。
  - 构建与发布链路的伸缩、容错、权限边界与元数据查询需求完全不同，拆分后更清晰。
- 影响：
  - 本仓库未来优先承载 CRUD、状态模型、查询接口、事件写回与编排意图。
  - Tekton / Argo 客户端与 informer 迁移到执行器或状态采集服务。
  - `Manifest.Create` 与 `Job.Create` 从“直接执行”改为“落库 + 发意图 + 等外部回写”。
  - 元数据层启动不再强依赖 kubeconfig、Tekton、Argo。
