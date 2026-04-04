# API Contract Reference

## Purpose

定义当前 metadata/control plane 对外 HTTP API 的统一约定。

## Scope

适用于：

- `pkg/api`
- `pkg/router`
- Swagger 暴露的所有 `/api/v1/*` 接口

不适用于：

- 内部 Go 方法签名
- 外部系统 webhook 的原始协议细节

## Must

- 路径使用资源化命名，优先使用复数名词：
  - `/projects`
  - `/applications`
  - `/manifests`
  - `/jobs`
  - `/intents`
- 创建接口返回统一 `CreateResponse`
- 列表接口支持标准分页头，并保持过滤语义稳定
- 与软删除相关的资源，默认过滤 `deleted_at`
- 若存在资源归属关系，应优先使用明确的 ID 关联，而不是仅靠名称字符串
- 若接口行为发生变化，必须同步判断 Swagger 是否需要重生成
- metadata/control plane 的接口必须明确区分：
  - 本地记录已创建
  - 外部执行已提交
  - 外部系统已确认终态

## Must Not

- 不混用资源路径和动作路径作为主要建模方式
- 不把调试字段直接暴露成正式 API 契约
- 不把“请求已受理”写成“执行已成功”
- 不让同一字段在响应和文档里表达不同语义

## Outputs

- 稳定的 REST 风格接口
- 可同步到 Swagger 的对外契约
- 与领域模型一致的响应语义

## Pass/Fail

- `Pass`：路径、过滤、响应、状态语义一致，可被外部系统稳定消费
- `Fail`：接口语义漂移、资源归属模糊、或文档与实现不一致
