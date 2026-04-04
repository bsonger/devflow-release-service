# Release

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/release.go`
- authoritative API doc: `docs/api-spec.md`
- swagger source: `docs/swagger.yaml`

## Purpose

`Release` 是部署侧控制面资源。用户点击 release 后创建，并触发 Kubernetes `Application` / Argo CD 侧执行。

## Common base fields

| Field | Type | Required | Writable | Description |
|---|---|---|---|---|
| `id` | `ObjectID` | server-generated | no | 主键 |
| `created_at` | `time.Time` | server-generated | no | 创建时间 |
| `updated_at` | `time.Time` | server-generated | no | 更新时间 |
| `deleted_at` | `*time.Time` | optional | system-managed | 软删除时间 |

## Field table

| Field | Type | Required | Writable | Description |
|---|---|---|---|---|
| `execution_intent_id` | `*ObjectID` | optional | system | 关联 release intent |
| `application_id` | `ObjectID` | system-derived | system | 应用 ID |
| `application_name` | `string` | system-derived | system | 应用名 |
| `project_name` | `string` | system-derived | system | 项目名/命名空间上下文 |
| `manifest_id` | `ObjectID` | required in practice | user | 发布基于哪个 manifest |
| `manifest_name` | `string` | system-derived | system | manifest 名 |
| `type` | `string` | optional | user/system | 发布动作；为空默认 `Upgrade` |
| `env` | `string` | system-derived | system | 当前实现固定写入 `prod` |
| `status` | `ReleaseStatus` | system-defaulted | verify/system | 发布状态 |
| `steps` | `[]ReleaseStep` | optional | verify/system | 发布步骤集合 |

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
- verify writeback path: `devflow-verify-service`
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
  - `manifest_name`、`application_id` 来自 `Manifest`
  - `application_name`、`project_name` 来自 `Application`
  - `env` 当前默认 `prod`
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

- `manifest_id` 必须是合法且存在的 `Manifest`
- 列表过滤里的 `application_id`、`manifest_id` 必须是合法 ObjectID
- 当前 `env` 不是用户自由输入字段，而是服务内派生值

## Source pointers

- router: `pkg/router/release.go`
- handler: `pkg/api/release.go`
- service: `pkg/service/release.go`
- model: `pkg/model/release.go`
