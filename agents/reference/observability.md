# Observability

本文件是 agent 参考摘要，正式规范以 `docs/observability.md` 为准。

硬规则：

- 任何出站服务/外部调用都必须有 `metrics + trace + structured log`
- 入站 HTTP 必须有 server span、请求计数、耗时、错误数
- 高基数字段不能进入 metrics label
- 日志至少带 `service`、`trace_id`、`span_id`、`request_id`

当前仓库重点字段：

- `application_id`
- `image_id`
- `release_id`
- `intent_id`
- `external_ref`
