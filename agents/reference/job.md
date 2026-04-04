# Job Reference

## Purpose

说明 `Job`、`JobStatus`、`steps` 以及当前确认的发布链路。

## Scope

适用于：

- `Job` 资源语义
- release 相关状态流
- step 回写链路

## Must

- `Job` 表示一次发布、回滚或同步任务记录
- 典型字段：
  - `id`
  - `execution_intent_id`
  - `application_id`
  - `manifest_id`
  - `status`
  - `type`
  - `env`
  - `steps`
- 状态枚举：
  - `Pending`
  - `Running`
  - `Succeeded`
  - `Failed`
  - `RollingBack`
  - `RolledBack`
  - `Syncing`
  - `SyncFailed`
- 状态变化可由内部流程和外部事件共同驱动，但运行态和终态优先由外部系统回写
- 当前确认的创建链路：
  - `pkg/api/job.go` 通过 `POST /api/v1/jobs` 调用 `JobService.Create`
  - `pkg/service/job.go` 创建记录时先写 `Pending`
  - 若请求未提供 `steps`，会按应用发布策略自动初始化默认 steps
  - `intent` 模式下会创建 release intent 并把 `execution_intent_id` 回填到 Job
  - `direct` 模式下会继续调用 Argo，并在提交前把 Job 更新为 `Syncing`
  - 若 direct 提交 Argo 失败，会把 Job 更新为 `SyncFailed`
- 当前确认的 step 回写链路：
  - `POST /api/v1/verify/release/steps` 更新 `job.steps`
  - 目标 step 不存在时，service 会自动补创建
  - 每次 step 更新后，service 会从全部 steps 自动收敛 `job.status`

## Must Not

- 不把“已创建 Job 记录”写成“发布已成功”
- 不把“已发起 Argo 请求”写成“外部系统已确认”
- 不绕开外部事件直接伪造运行终态

## Outputs

- Job 元数据
- release 过程的 `steps`
- 与 release intent 的关联

## Pass/Fail

- `Pass`：Evaluator 能区分本地同步链路状态和外部执行终态
- `Fail`：`Job` 状态在本地日志、Mongo、外部系统之间语义不一致
