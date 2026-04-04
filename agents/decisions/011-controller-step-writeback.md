# 011 controller 风格的 step 同步通过 verify-service 落库

- 决策：
  - 吸收 `devflow-controller` 的 step 同步模型，但不让 controller 直接写当前仓库的 Mongo。
  - `Release` 增加 `steps` 字段，`verify-service` 增加 `POST /api/v1/verify/release/steps` 作为统一写回入口。
- 原因：
  - `devflow-controller` 已经证明了 `Application` / `Deployment` / `Rollout` 事件同步到 `release.steps` 的模型是有效的。
  - 但在 service split 后，controller 直接写库会绕过 control-plane 边界，破坏服务职责。
- 影响：
  - `release-service` 创建 Release 时会初始化默认 step 模板，便于 controller 后续按名称更新。
  - `verify-service` 成为 build / release 运行态回写的统一入口。
  - 未来若单独部署 `devflow-controller`，它应调用 verify API，而不是持有 metadata 库写权限。
