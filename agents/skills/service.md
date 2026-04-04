# Service Skill

## Purpose

定义 `pkg/service` 的职责边界和实现要求。

## Scope

适用于：

- 所有业务逻辑
- 所有外部依赖协作

不适用于：

- HTTP 参数解析
- 路由绑定
- Swagger 生成产物

## Must

- 业务逻辑集中在 `pkg/service`
- 外部依赖通过配置或客户端注入
- 对外部系统调用保留清晰的错误传播和日志
- 保持可测试、可复用、可观测

## Must Not

- 不引入与业务无关的副作用
- 不在 service 层做 HTTP 响应拼装
- 不在 service 层混入无边界的文件系统写入

## Outputs

- 可复用的业务行为
- 明确的依赖协作和错误语义

## Pass/Fail

- `Pass`：service 只承载业务和依赖编排，边界稳定
- `Fail`：service 夹带入口逻辑或无关副作用
