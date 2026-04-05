# Intent

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/intent.go`
- authoritative API doc: `docs/api-spec.md`
- generated swagger: `docs/swagger.yaml` (transitional; still reflects legacy handler layer until API migration)

## Purpose

`Intent` 是控制面执行意图记录，用于 build/release 长任务的异步编排、worker 消费与状态跟踪。

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
| `kind` | `IntentKind` | system-set | system | `build` 或 `release` |
| `status` | `IntentStatus` | system-set | system/worker/verify | 当前意图状态 |
| `resource_type` | `string` | system-set | system | 目标资源类型，如 `manifest` / `release` |
| `resource_id` | `uuid.UUID` | system-set | system | 目标资源 ID |
| `application_id` | `uuid.UUID` | system-set | system | 应用 ID |
| `manifest_id` | `*uuid.UUID` | optional | system | Manifest ID |
| `release_id` | `*uuid.UUID` | optional | system | Release ID |
| `release_type` | `string` | optional | system | 发布类型 |
| `env` | `string` | optional | system | 目标环境 |
| `repo_address` | `string` | optional | system | 仓库地址 |
| `branch` | `string` | optional | system | 分支 |
| `external_ref` | `string` | optional | worker/verify | 外部系统引用 |
| `trace_id` | `string` | optional | system/worker | 调用链 trace 标识 |
| `message` | `string` | optional | worker/verify | 状态消息 |
| `last_error` | `string` | optional | worker/system | 最后错误 |
| `claimed_by` | `string` | optional | worker/system | 被哪个 worker 认领 |
| `claimed_at` | `*time.Time` | optional | worker/system | 认领时间 |
| `lease_expires_at` | `*time.Time` | optional | worker/system | lease 过期时间 |
| `attempt_count` | `int` | optional | worker/system | 尝试次数 |

## Enums

### `IntentKind`
- `build`
- `release`

### `IntentStatus`
- `Pending`
- `Running`
- `Succeeded`
- `Failed`

## Create / update rules

### Create
- current behavior:
  - 不通过 public create API 直接创建
  - 由服务在创建 `Manifest` / `Release` 时派生创建
- build intent source:
  - `service.IntentService.CreateBuildIntent`
- release intent source:
  - `service.IntentService.CreateReleaseIntent`

### Update
- common update paths:
  - worker claim / lease 更新
  - verify 根据 build/release 结果回写状态
  - external ref / message 更新

## Query rules

- public surface:
  - `GET /api/v1/intents`
  - `GET /api/v1/intents/{id}`
- common query filters:
  - `kind`
  - `status`
  - `resource_type`
  - `resource_id`
  - `application_id`
  - `manifest_id`
  - `release_id`
  - `release_type`
  - `env`
  - `branch`
  - `claimed_by`
  - `external_ref`

## Source pointers

- router: `pkg/router/intent.go`
- handler: `pkg/api/intent.go`
- service: `pkg/service/intent.go`
- model: `pkg/model/intent.go`
