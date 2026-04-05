# Manifest

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/manifest.go`
- authoritative API doc: `docs/api-spec.md`
- generated swagger: `docs/swagger.yaml` (transitional; still reflects legacy handler layer until API migration)

## Purpose

`Manifest` 是 build 侧控制面资源。用户点击 build 后通过 `release-service` 创建，并触发 Tekton 执行。
它冻结构建时需要的仓库地址与服务暴露快照。

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
| `execution_intent_id` | `*uuid.UUID` | optional | system | 关联 build intent |
| `application_id` | `uuid.UUID` | required | user | 目标应用 ID |
| `name` | `string` | system-generated | system | manifest 名称 |
| `branch` | `string` | optional | user | Git 分支；为空默认 `main` |
| `repo_address` | `string` | system-derived | system | Git 仓库地址 |
| `commit_hash` | `string` | optional | patch/user | 构建提交 hash |
| `replica` | `*int32` | system-derived | system | 副本数快照 |
| `digest` | `string` | optional | patch/user | 镜像 digest |
| `type` | `ReleaseType` | system-derived | system | 发布策略类型 |
| `services` | `[]ManifestService` | system-derived | system | 服务暴露快照 |
| `pipeline_id` | `string` | system-generated | verify/system | Tekton pipeline/pipelinerun 标识 |
| `steps` | `[]ManifestStep` | system-generated | verify/system | 构建步骤状态 |
| `status` | `ManifestStatus` | system-defaulted | verify/system | Manifest 状态 |

## Nested types

### `ManifestStatus`
- `Pending`
- `Running`
- `Succeeded`
- `Failed`

### `StepStatus`
- `Pending`
- `Running`
- `Succeeded`
- `Failed`

### `ManifestStep`
- `task_name: string`
- `task_run: string`
- `status: StepStatus`
- `start_time: *time.Time`
- `end_time: *time.Time`
- `message: string`

### `ManifestService`
- `name: string`
- `internet: Internet`
- `ports: []Port`

### Shared nested types
- `ReleaseType`: `normal` / `canary` / `blue-green`
- `Internet`: `internal` / `external`
- `Port`: `name` / `port` / `target_port`

## Lifecycle / status fields

- public owner: `devflow-release-service`
- writeback path: verify result ownership is in `devflow-verify-service`
- default status on create: `Pending`
- step source:
  - create 时从 Tekton Pipeline 模板初始化 steps
  - 执行中由 verify 更新

## Create / update rules

### Create
- entrypoint:
  - `POST /api/v1/manifests`
- current API behavior:
  - handler 绑定整个 `model.Manifest`
- practical required fields:
  - `application_id`
- defaults / derived values:
  - `name` 自动生成
  - `branch` 为空时默认 `main`
  - `status` 初始化为 `Pending`
  - `repo_address`, `replica`, `type`, `services` 从 app-service 元数据衍生
- side effects:
  - 创建 Tekton 相关资源
  - 可能创建 build intent

### Patch
- patch endpoint only supports:
  - `commit_hash`
  - `digest`

### Update/writeback
- verify may update:
  - `pipeline_id`
  - `steps`
  - `status`

## Validation notes

- `application_id` 必须引用存在的 `Application`
- `pipeline_id` 可在后续 verify 阶段补写
- `repo_address` 使用统一字段名，不再使用 `git_repo`

## Source pointers

- router: `pkg/router/manifest.go`
- handler: `pkg/api/manifest.go`
- service: `pkg/service/manifest.go`
- model: `pkg/model/manifest.go`
