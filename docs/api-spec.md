# API Spec

## Resources

- `Manifest`
  - `GET /api/v1/manifests`
  - `POST /api/v1/manifests`
  - `GET /api/v1/manifests/:id`
  - `PATCH /api/v1/manifests/:id`
- `Release`
  - `GET /api/v1/releases`
  - `POST /api/v1/releases`
    - platform-relevant request fields: `manifest_id`, `configuration_id` (optional), `env` (optional, default `prod`), `type`
  - `GET /api/v1/releases/:id`
- `Intent`
  - `GET /api/v1/intents`
  - `GET /api/v1/intents/:id`

## Response Rules

- 创建接口返回统一创建响应
- 列表接口遵循统一分页参数和分页响应头
- `Intent` 为控制面查询资源，不在本仓库对外提供通用更新 CRUD

## Error Rules

- 非法 ObjectID 返回 `400`
- 资源不存在返回 `404`
- 存储层或执行层未分类错误返回 `500`

## Swagger

Swagger 必须只包含 `Manifest`、`Release`、`Intent` 接口。
