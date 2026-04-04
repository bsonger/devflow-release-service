# pkg/model

本目录承载 Devflow 自有的领域模型。

当前包含：

- `Project`
- `Application`
- `Manifest`
- `Job`
- `Intent`
- `Configuration`
- 公共状态枚举与基础方法
- 配置结构：`ServerConfig`、`MongoConfig`、`LogConfig`、`OtelConfig`、`Repo`

其中 `Intent` 用于承载 control-plane 的 build / release 执行意图，并通过 `execution_intents` 集合作为独立元数据资源持久化。

`Job` 当前还承载 `steps`，用于保存来自 controller / verify-service 的部署阶段细粒度进度。

设计约束：

- 领域模型由本仓库维护和版本化
- `pkg/api`、`pkg/service`、`platform/*` 统一依赖这里
- 外部依赖 client 可以继续复用 `devflow-common`，但不再把其 `model` 当作权威领域定义
