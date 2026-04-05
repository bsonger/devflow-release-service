# Application Reference Redirect

This file is index-only.

Authoritative sources:

- `devflow-control/docs/resources/application.md`
- `devflow-control/docs/services/app-service.md`
- `devflow-app-service/docs/resources/application.md`

This file must not define application semantics.
  - `status`
- `active_manifest_id` / `active_manifest_name` 指向当前生效的 Manifest
- 若请求提供 `project_id`，服务会校验项目存在并把 `project_name` 对齐到项目名称
- 读写由应用 API 管理

## Must Not

- 不把 `Application` 当作执行记录
- 不把运行态真相只绑定到本地数据库推断

## Outputs

- `Application` 元数据
- 当前生效 Manifest 关联

## Pass/Fail

- `Pass`：`Application` 仍是元数据资源，而不是执行器
- `Fail`：`Application` 语义被混入 Release/Manifest 执行状态细节
