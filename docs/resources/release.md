# Release

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/release.go`
- authoritative API doc: `docs/api-spec.md`
- generated swagger: `docs/swagger.yaml` (transitional; still reflects legacy handler layer until API migration)

## Purpose

`Release` 是部署侧控制面资源。用户点击 release 后创建，并触发 Kubernetes `Application` / Argo CD 侧执行。

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
| `execution_intent_id` | `*uuid.UUID` | optional | system | 关联 release intent |
| `configuration_id` | `*uuid.UUID` | optional | user/system | 关联的 Configuration ID |
| `configuration_revision_id` | `*uuid.UUID` | optional | user/system | 绑定的配置 revision |
| `application_id` | `uuid.UUID` | system-derived | system | 应用 ID |
| `manifest_id` | `uuid.UUID` | required | user | 发布基于哪个 manifest |
| `type` | `string` | optional | user/system | 发布动作；为空默认 `Upgrade` |
| `env` | `string` | optional | user/system | 目标环境；为空默认 `prod` |
| `status` | `ReleaseStatus` | system-defaulted | verify/system | 发布状态 |
| `steps` | `[]ReleaseStep` | optional | verify/system | 发布步骤集合 |
| `external_ref` | `string` | optional | verify/system | 外部系统引用 |

## Nested types

### `ReleaseStatus`
- `Pending`
- `Running`
- `Succeeded`
- `Failed`
- `RollingBack`
- `RolledBack`
- `Syncing`
- `SyncFailed`

### Release action values
- `Install`
- `Upgrade`
- `Rollback`

### `ReleaseStep`
- `name: string`
- `progress: int32`
- `status: StepStatus`
- `message: string`
- `start_time: *time.Time`
- `end_time: *time.Time`

## Lifecycle / status fields

- public owner: `devflow-release-service`
- verify writeback path: verify result ownership is in `devflow-verify-service`
- defaults:
  - create 时 `status = Pending`
  - 若未提供 `steps`，会根据应用策略和发布动作自动生成默认步骤
- dispatch path:
  - release 创建后会同步到 Argo / Kubernetes `Application`
  - dispatch 前可能把状态切到 `Syncing`

## Create / update rules

### Create
- entrypoint:
  - `POST /api/v1/releases`
- practical required fields:
  - `manifest_id`
- defaults / derived values:
  - `type` 为空时默认 `Upgrade`
  - `application_id` 来自 `Manifest`
  - `env` 为空时默认 `prod`
  - `status` 初始化为 `Pending`
  - `steps` 可自动生成
- side effects:
  - 创建 release record
  - 可能创建 release intent
  - 非 intent mode 下会创建 Argo/Kubernetes `Application`

### Update/writeback
- verify may update:
  - `status`
  - `steps`
- public CRUD update endpoint:
  - 当前不对外暴露通用 update 接口

## Validation notes

- `manifest_id` 必须引用存在的 `Manifest`
- 若同时传 `configuration_id` 与 `configuration_revision_id`，两者必须属于同一配置
- 当前 `env` 不是完全自由输入字段，而是服务约束下的目标环境

## Source pointers

- router: `pkg/router/release.go`
- handler: `pkg/api/release.go`
- service: `pkg/service/release.go`
- model: `pkg/model/release.go`
