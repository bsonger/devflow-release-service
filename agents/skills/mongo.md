# Mongo Skill

## Purpose

定义当前仓库中 Mongo 使用时的职责和安全约束。

## Scope

适用于：

- Mongo 集合访问
- 索引、更新、查询策略
- `pkg/service` 内的数据读写

不适用于：

- 与 Mongo 无关的业务内存处理

## Must

- 优先使用明确的集合和索引策略，避免隐式创建
- 更新操作优先使用原子更新，避免读改写竞态
- 读写分离或事务策略必须在 `pkg/service` 层落地

## Must Not

- 不在 handler 里直接访问 Mongo
- 不用整对象覆盖更新掩盖状态竞争
- 不把 Mongo 技术细节泄漏到 API 语义里

## Outputs

- 清晰的数据访问边界
- 原子且一致的持久化逻辑

## Pass/Fail

- `Pass`：Mongo 访问集中、原子、边界清晰
- `Fail`：handler 直连 Mongo 或持久化语义混乱
