# Project Reference

## Purpose

定义 `Project` 资源在 metadata/control plane 中的职责、字段和归属语义。

## Scope

适用于：

- `Project` 元数据模型
- `Application` 对 `Project` 的归属关系
- app-service 下的项目查询与聚合接口

不适用于：

- 发布执行流程
- Tekton / Argo / verify-service 回写语义

## Must

- `Project` 必须是一等元数据资源，不能只靠 `project_name` 字符串存在
- 推荐字段至少包括：
  - `id`
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
