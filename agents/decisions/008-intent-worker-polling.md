# 008 execution_intents 先采用单 worker 轮询

- 决策：
  - `execution_intents` 当前先由 `release-service` worker 通过轮询方式消费。
  - 当前实现按单 worker、顺序处理假设运行。
- 原因：
  - 先验证 intent 架构的闭环，比一开始就引入消息队列或复杂认领协议更稳。
  - 现有直接执行逻辑可以更快复用到 worker 中，减少迁移面。
- 影响：
  - 当前不适合多 worker 并发消费同一批 pending intents。
  - 后续若要横向扩展，需要引入更明确的 claim/lease/outbox 或消息队列机制。
