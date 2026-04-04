# Intent Reference

## Purpose

说明 `execution_intents` 的资源语义、创建链路、消费链路与回写边界。

## Scope

适用于：

- build intent
- release intent
- worker 认领与回写语义

## Must

- `Intent` 表示控制面记录的待执行或正在执行的 build / release 编排请求
- 典型字段：
  - `id`
  - `kind`
  - `status`
  - `resource_type`
  - `resource_id`
  - `external_ref`
  - `claimed_by`
  - `lease_expires_at`
- 种类枚举：
  - `build`
  - `release`
- 状态枚举：
  - `Pending`
  - `Running`
  - `Succeeded`
  - `Failed`
- 当前确认的创建链路：
  - `POST /api/v1/manifests` 在 `intent` 模式下先落库 Manifest，再创建 build intent
  - `POST /api/v1/jobs` 在 `intent` 模式下先落库 Job，再创建 release intent
  - 创建响应会直接返回 `execution_intent_id`
- 当前确认的消费链路：
  - `platform/release-service/cmd/worker` 轮询并认领 `Pending` intents
  - `pkg/service/intent.go` 使用 `claimed_by` + `lease_expires_at` 做 claim + lease
  - worker 提交外部执行后会把 intent 更新为 `Running` 并写入 `external_ref`
- 当前确认的回写链路：
  - `verify-service` 接收 Tekton / Argo 事件后，会同步回写 `Manifest`、`Job` 与 `Intent`

## Must Not

- 不把 `Intent` 直接当作外部系统事实状态来源
- 不把控制面 accepted 写成执行完成

## Outputs

- build/release 编排记录
- worker 认领与执行进度

## Pass/Fail

- `Pass`：Intent 仍是控制面资源，而不是执行真相来源
- `Fail`：Intent 状态替代了外部事件事实
