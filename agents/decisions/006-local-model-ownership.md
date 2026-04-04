# 006 领域模型归仓库所有

- 决策：
  - `Application`、`Manifest`、`Release`、`Configuration` 及其状态枚举、公共方法迁入本仓库的 `pkg/model/`。
  - `devflow-common` 继续提供 logging、mongo、argo、tekton、otel 等 client，不再作为本仓库领域模型的权威来源。
- 原因：
  - 当前仓库正在向 metadata/control plane 演进，领域模型应与 API、状态机、拆分边界一起版本化。
  - 若模型继续放在外部 common 包，metadata API 的演进会被外部仓库节奏反向约束。
  - `platform/*` 服务拆分后，统一依赖本仓库 `pkg/model/` 更清晰。
- 影响：
  - `pkg/api`、`pkg/service`、`pkg/config` 统一依赖 `pkg/model/`。
  - `pkg/config` 对 `devflow-common/client/*` 使用配置适配，而不是继续透传外部 model 类型。
  - 后续若需要进一步去掉 `devflow-common/model` 依赖，可继续把 client 所需配置模型或 client 实现迁入本仓库。
