# Release

## Ownership

- owner repo: `devflow-release-service`
- authoritative model file: `pkg/model/release.go`
- authoritative API doc: `docs/api-spec.md`

## Purpose

`Release` 是一次执行记录。
它指向一个已经冻结好的 `Manifest`，并记录发布动作、目标环境和执行状态。
Manifest 可以携带已经推送到 OCI 仓库的部署制品引用；Release 消费这个 frozen manifest，而不是重新读取 service/route/config 输入。
当 Manifest 已带有 OCI artifact 时，Release 创建 Argo CD Application 会优先直接指向这个 OCI 制品；只有在没有 artifact 的旧路径下，才回退到 repo plugin 流程。
因为 OCI repository 路径本身按 `application + environment` 分桶，Release 在不同环境发布时会消费对应环境生成的 manifest OCI，而不是共享同一个部署包。

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
| `execution_intent_id` | `*uuid.UUID` | optional | system | 关联的 release intent |
| `application_id` | `uuid.UUID` | system-derived | system | 应用 ID |
| `manifest_id` | `uuid.UUID` | required | user | 发布基于哪个 frozen manifest |
| `image_id` | `uuid.UUID` | system-derived | system | 从 manifest 派生出的 image |
| `env` | `string` | optional | user/system | 目标环境；默认使用 image 绑定的 runtime environment |
| `type` | `string` | optional | user/system | 发布动作；为空默认 `Upgrade` |
| `steps` | `[]ReleaseStep` | system-managed | system | 发布执行步骤 |
| `status` | `ReleaseStatus` | system-managed | system | 发布状态 |
| `external_ref` | `string` | optional | verify/system | 外部系统引用 |

## Create / update rules

### Create
- required:
  - `manifest_id`
- optional:
  - `env`
  - `type`
- server-managed:
  - `application_id`
  - `image_id`
  - `status`
  - `steps`

### Update
- mutable fields:
  - `type`
  - `steps`
  - `status`
  - `external_ref`
- immutable identity fields:
  - `application_id`
  - `manifest_id`
  - `image_id`

## Validation notes

- `manifest_id` 必须引用存在且 `ready` 的 `Manifest`
- `image_id` 由 manifest 派生，不能由调用方直接指定
- `Image.runtime_spec_revision_id` 必须存在，release 才能创建
- 若显式传入 `env`，必须和 image 绑定的 runtime environment 一致

## Source pointers

- router: `pkg/router/release.go`
- handler: `pkg/api/release.go`
- service: `pkg/service/release.go`
- model: `pkg/model/release.go`
