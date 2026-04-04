# Service Pattern

## Purpose

定义 `pkg/service` 层的标准职责和最小执行顺序。

## Scope

适用于：

- 所有 `pkg/service/*` 业务方法
- 需要访问 Mongo、Tekton、Argo、verify-service 等依赖的场景

不适用于：

- HTTP 参数解析
- 路由绑定
- 直接构造 HTTP 响应

## Must

- 输入必须是已校验的业务参数
- 输出必须是领域模型、状态变化或明确错误
- 标准顺序：
  1. 读取依赖数据
  2. 执行业务规则
  3. 写入状态或结果
  4. 返回领域结果
- 所有外部依赖调用都应保留可观测性上下文

## Must Not

- 不在 service 层直接做 HTTP 参数校验
- 不在 service 层直接拼 HTTP 响应结构
- 不把日志当作状态真相来源

## Outputs

- 领域对象
- 状态更新
- 明确错误

## Pass/Fail

- `Pass`：service 层只承载业务逻辑与依赖协作
- `Fail`：service 层混入 handler 逻辑、响应拼装或无边界副作用
