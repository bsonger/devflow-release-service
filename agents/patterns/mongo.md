# Mongo Pattern

## Purpose

定义当前仓库里 Mongo 读写的最小安全写法。

## Scope

适用于：

- 所有 Mongo 更新语句
- `pkg/service` 内的状态写回和元数据更新

不适用于：

- HTTP 参数解析
- 与 Mongo 无关的内存态处理

## Must

- 优先使用 `$set` / `$unset` / `$inc` 做原子更新
- 需要更新时间时同步写 `updated_at`
- 优先单次写入完成状态变更，避免不必要的先读后写
- 更新语义必须与领域状态保持一致

## Must Not

- 不用日志替代 Mongo 中的真实状态
- 不把竞态窗口留在“先查再整对象覆盖更新”的路径里
- 不在 handler 层直接访问 Mongo

## Outputs

- 原子更新语句
- 一致的状态落库结果

## Pass/Fail

- `Pass`：Mongo 写入原子、可追踪、与领域状态一致
- `Fail`：存在明显竞态、整对象覆盖、或状态语义错乱
