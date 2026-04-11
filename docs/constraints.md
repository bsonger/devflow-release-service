# Constraints

## Ownership

- `Image`、`Manifest`、`Release`、`Intent` 属于 release-service 独占边界
- `Intent` 是 release/build 执行意图记录，不是通用跨仓库扩展点

## Prohibited

- 不得在本仓库新增 `Project`、`Application`、`AppConfig`、`WorkloadConfig`、`Service`、`Route`、`Verify` 对外 API
- 不得把 verify webhook 入口放回 release-service
- 不得在 metrics label 中写入高基数业务主键

## Data Rules

- 任何状态推进都必须保证幂等
- 任何写操作都必须更新 `updated_at`
- 终态资源不应被非受控流程反复覆盖
