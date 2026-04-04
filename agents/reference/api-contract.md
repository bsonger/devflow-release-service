# API Contract

本文件是 agent 参考摘要，正式规范以 `docs/api-spec.md` 为准。

当前仓库只允许：

- `Manifest`
- `Release`
- `Intent`

禁止出现：

- `Project`
- `Application`
- `Configuration`
- `Verify`

要求：

- handler、router、Swagger 三者一致
- 列表接口遵循统一分页
- 非法 ID 返回 `400`
- 缺失资源返回 `404`
- 正式接口说明更新到 `docs/api-spec.md`
