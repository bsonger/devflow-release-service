# 007 release-service 使用 intent 模式

- 决策：
  - `release-service` 默认以 `intent` 模式运行。
  - `Manifest.Create` 与 `Release.Create` 在该模式下只负责落库并创建 `execution_intents`，不再直接触发 Tekton / Argo。
  - `verify-service` 负责把外部执行结果回写到 `Manifest`、`Release` 与 `Intent`。
- 原因：
  - 这能把 metadata API 与执行器主流程解耦，符合控制面 / 执行面拆分方向。
  - 先落意图再消费执行，便于后续引入独立 executor、队列或 outbox，而不继续把 Tekton / Argo client 固定在 API 主链路上。
- 影响：
  - `release-service` 的成功语义变为“元数据已接受并已创建 intent”，不等于执行完成。
  - 执行状态需要通过 `verify-service` 回写更新。
  - 单体入口和其他服务暂时仍可继续保持 `direct` 模式，降低迁移冲击。
