# Project Reference Redirect

This file is index-only.

Authoritative sources:

- `devflow-control/docs/resources/project.md`
- `devflow-control/docs/services/app-service.md`
- `devflow-app-service/docs/resources/project.md`

This file must not define project semantics.
  - `name`
  - `key`
  - `description`
  - `namespace`
  - `owner`
  - `labels`
  - `status`
- `Application` 应优先通过 `project_id` 关联 `Project`
- `project_name` 可以继续保留作为兼容字段和展示字段
- 若请求提供 `project_id`，服务必须校验项目存在，并把 `project_name` 对齐为项目名称
- `GET /api/v1/projects/:id/applications` 必须能列出该项目下的应用

## Must Not

- 不让 `Project` 变成只是前端展示标签
- 不继续只依赖 `project_name` 做唯一归属语义
- 不在项目归属不一致时静默接受请求

## Outputs

- `Project` 资源
- `Application -> Project` 的明确归属关系
- 项目维度的查询入口

## Pass/Fail

- `Pass`：项目资源可独立管理，应用归属可校验、可查询
- `Fail`：项目仍然只是字符串字段，或归属关系无法验证
