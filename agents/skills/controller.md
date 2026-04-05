# Controller Skill

## Purpose

定义 handler / controller 层的职责边界。

## Scope

适用于：

- `pkg/api`
- `pkg/router`
- 与 HTTP 请求/响应直接相关的入口逻辑

不适用于：

- `pkg/service` 内的业务决策
- PostgreSQL、Argo、Tekton 等依赖调用实现

## Must

- 负责参数校验、请求/响应模型转换、错误码映射
- 把业务逻辑下沉到 `pkg/service`
- 保持统一返回结构和明确错误语义

## Must Not

- 不直接访问数据库或外部系统
- 不在 controller 里拼装复杂业务状态
- 不把 service 逻辑搬回 handler

## Outputs

- 已校验的入参
- 统一的 HTTP 响应

## Pass/Fail

- `Pass`：controller 只承担入口职责，业务逻辑边界清晰
- `Fail`：controller 直接碰数据源或承载业务编排
