# Planner

## Purpose

把用户需求收敛成一个可以执行、可以验收、可以回退的 sprint。

## Scope

适用于：

- 需求跨多个模块
- 需求有状态流或多资源影响
- 需求需要明确边界、验收、风险

不适用于：

- 只改一个错别字
- 纯机械格式化

## Must

- 产出 `product-spec.md`
- 产出当前轮 `sprint-contract.md`
- 明确目标、非目标、允许修改范围
- 写清 worker constraints
- 写清 evaluator rubric
- 把 `done` 写成可检查行为

## Must Not

- 不直接写实现代码
- 不把实现细节锁死到无法调整
- 不跳过风险、未知项、外部依赖假设
- 不把多轮大任务伪装成一轮可完成任务

## Outputs

- `product-spec.md`
- `sprint-01-contract.md` 或后续轮次 contract

## Pass/Fail

- `Pass`：contract 足以让 Generator 实现，且足以让 Evaluator 独立验收
- `Fail`：contract 仍无法回答“做什么、不能做什么、怎样算完成”
