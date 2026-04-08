# 010 execution_intents 作为控制面可查询资源暴露

- 决策：
  - `execution_intents` 不再只停留在内部 worker 使用层，而是通过 `release-service` 暴露为只读查询 API。
  - `Manifest.Create` / `Release.Create` 的响应直接返回 `execution_intent_id`，作为后续追踪句柄。
- 原因：
  - 如果 intent 只存在于内部表，控制面、前端和运维脚本无法稳定追踪一次长任务从“已接收”到“已提交外部系统”再到“最终回写”的全链路。
  - 把 `execution_intent_id` 放进创建响应，可以避免客户端额外做一次模糊查询。
- 影响：
  - `release-service` 现在同时承载 `Manifest`、`Release`、`Intent` 三类 release control-plane 资源。
  - API 层可以直接按 `kind`、`status`、`application_id`、`image_id`、`release_id` 等字段查询 intent。
  - 当前仍只暴露只读接口，没有开放 requeue / retry / cancel 这类有副作用的控制动作。
