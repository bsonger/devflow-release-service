# Intent

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/intent.go`
- authoritative API doc: `docs/api-spec.md`

## Purpose

`Intent` 是异步执行协调资源。
它不再复制业务上下文，只记录某个目标资源的执行声明、抢占状态和失败信息。

## Common base fields

| Field | Type | Required | Writable | Description |
|---|---|---|---|---|
| `id` | `uuid.UUID` | server-generated | no | 主键 |
| `created_at` | `time.Time` | server-generated | no | 创建时间 |
| `updated_at` | `time.Time` | server-generated | no | 更新时间 |
| `deleted_at` | `*time.Time` | optional | system-managed | 软删除时间 |

## Field table

| Field | Type | Required | Writable | Description |
|---|---|---|---|---|
| `kind` | `IntentKind` | required | system | 执行类型，如 `build` / `release` |
| `status` | `IntentStatus` | required | system | 当前执行状态 |
| `resource_type` | `string` | required | system | 目标资源类型，如 `image` / `release` |
| `resource_id` | `uuid.UUID` | required | system | 目标资源 ID |
| `trace_id` | `string` | optional | system | trace 关联 ID |
| `message` | `string` | optional | system | 最近状态说明 |
| `last_error` | `string` | optional | system | 最近失败原因 |
| `claimed_by` | `string` | optional | worker | 当前 worker |
| `claimed_at` | `*time.Time` | optional | worker | 抢占时间 |
| `lease_expires_at` | `*time.Time` | optional | worker | 租约到期时间 |
| `attempt_count` | `int` | system-managed | system | 处理次数 |

## Query notes

- 列表支持：
  - `kind`
  - `status`
  - `resource_type`
  - `resource_id`
  - `claimed_by`

## Validation notes

- `Intent` 只拥有执行协调信息
- 应用、image、release、repo、branch 等业务上下文应从目标资源追溯

## Source pointers

- router: `pkg/router/intent.go`
- handler: `pkg/api/intent.go`
- service: `pkg/service/intent.go`
- model: `pkg/model/intent.go`
