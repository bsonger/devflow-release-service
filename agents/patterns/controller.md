# Controller Pattern

## Purpose

定义 handler / controller 层的最小职责。

## Scope

适用于：

- `pkg/api/*`
- 路由进入后的请求处理入口

不适用于：

- 复杂业务规则实现
- 跨资源状态编排

## Must

- 使用固定结构：
  1. 参数校验
  2. 调用 service
  3. 返回统一响应
- 错误语义至少保持：
  - 参数问题：`400`
  - 未找到：`404`
  - 业务失败：`500` 或更明确状态码

## Must Not

- 不在 controller 中写业务逻辑
- 不在 controller 中直接访问底层外部依赖
- 不在 controller 中重复实现状态机

## Outputs

- 一致的 HTTP 响应
- 明确的错误语义

## Pass/Fail

- `Pass`：controller 只做边界适配
- `Fail`：controller 开始承载业务规则或直接写依赖
