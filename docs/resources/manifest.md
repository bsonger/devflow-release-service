# Manifest

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/manifest.go`
- authoritative API doc: `docs/api-spec.md`

## Purpose

`Manifest` 表示一次打包完成后冻结下来的期望状态快照。
它绑定交付产物本身，以及当时使用的配置 revision 和 runtime revision。

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
| `execution_intent_id` | `*uuid.UUID` | optional | system | 关联的 build intent |
| `application_id` | `uuid.UUID` | required | user | 目标应用 ID |
| `configuration_revision_id` | `*uuid.UUID` | optional | user/system | 绑定的配置 revision |
| `runtime_spec_revision_id` | `*uuid.UUID` | optional | user/system | 绑定的运行期望态 revision |
| `name` | `string` | system-derived | system | manifest 名称 |
| `branch` | `string` | optional | user | Git 分支；为空默认 `main` |
| `repo_address` | `string` | system-derived | system | Git 仓库地址 |
| `commit_hash` | `string` | optional | system | 构建对应的 commit |
| `digest` | `string` | optional | system | 构建产物 digest |
| `pipeline_id` | `string` | optional | system | Tekton pipelineRun ID |
| `steps` | `[]ManifestStep` | system-managed | system | 构建步骤状态 |
| `status` | `ManifestStatus` | system-managed | system | 构建状态 |

## Create / update rules

### Create
- required:
  - `application_id`
- optional:
  - `configuration_revision_id`
  - `runtime_spec_revision_id`
  - `branch`
- server-managed:
  - `name`
  - `repo_address`
  - `status`
  - `steps`

### Update
- mutable fields:
  - `commit_hash`
  - `digest`
- immutable snapshot bindings:
  - `application_id`
  - `configuration_revision_id`
  - `runtime_spec_revision_id`

## Validation notes

- `application_id` 必须引用存在的 `Application`
- `repo_address` 从应用元数据派生
- `Manifest` 不再拥有副本数、发布策略或 service 暴露快照

## Source pointers

- router: `pkg/router/manifest.go`
- handler: `pkg/api/manifest.go`
- service: `pkg/service/manifest.go`
- model: `pkg/model/manifest.go`

## Image alias note

The product-facing build surface now uses `Image` routes, but those APIs are backed by the same underlying `Manifest` persistence model during the transition.
