# Evaluator Rubric

## Purpose

约束 `Evaluator` 只依据地面事实做验收，而不是“看起来差不多就过”。

## Scope

适用于：

- 所有走 Harness 的非简单需求
- 所有有代码、配置、状态流、文档影响的改动

不适用于：

- 代替 `Generator` 补实现
- 修改 contract 逃避问题

## Must

- 结论必须且只能是：
  - `Pass`
  - `Pass with risks`
  - `Fail`
- 必查维度包括：
  - contract 对齐
  - 功能正确
  - 一致性
  - 可验证性
  - 可运维性
  - 边界控制
- 检查 handler / service / router 行为是否满足 contract
- 检查 `Application` / `Manifest` / `Job` / `Intent` 的 ID 与状态关系是否一致
- 区分真正跑过的测试和仅推断可过的部分
- 检查 observability 是否退化：
  - trace 是否还能串起关键链路
  - logs 是否还能按 `trace_id` / `span_id` 关联
  - metrics 是否保持低基数
- 对当前仓库专项检查：
  - `Intent` 模式下，创建成功不得等价于执行成功
  - `verify-service` 回写后，状态字段与 step 字段必须保持收敛关系
  - worker / controller 改动后，至少检查 trace/log/metrics 没被破坏
  - 新增 observability 端点后，不应把 `/metrics`、`/healthz`、`/readyz`、`/debug/pprof/*` 计入业务指标

## Must Not

- 不接受“基本可以”“问题不大”这类模糊结论
- 不忽略关键 contract 项未完成
- 不忽略状态来源被写乱
- 不忽略明显回归
- 不在观测链路被破坏但未声明时给通过

## Outputs

- `Pass` / `Pass with risks` / `Fail`
- findings
- 风险范围
- 下一轮最小补救动作

## Pass/Fail

- `Pass`：阻断项通过，功能、状态、观测、边界都成立
- `Fail`：阻断测试失败、关键项未完成、越权改动、状态来源错乱、明显回归或观测链路被破坏

## Pass With Risks

- 仅当主体行为完成，但依赖外部系统无法完整验证、存在已声明技术债、或有非阻断已知风险时使用
