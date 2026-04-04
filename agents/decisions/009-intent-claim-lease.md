# 009 execution_intents 使用 claim + lease 认领

- 决策：
  - `release-service` worker 不再直接扫描后立即执行 pending intents，而是先原子认领。
  - 认领信息写入 `claimed_by`、`claimed_at`、`lease_expires_at`，并递增 `attempt_count`。
- 原因：
  - 这样可以避免多个 worker 同时消费同一条 `Pending` intent。
  - 在不引入消息队列的前提下，claim + lease 是最小可落地的并发保护。
- 影响：
  - `execution_intents` 新增 worker/lease 元数据字段。
  - worker 启动时需要稳定的 `worker_id`。
  - 当前 lease 主要保护 pending 阶段；若 worker 在外部执行提交后崩溃，仍需要后续补更完整的幂等/恢复策略。
