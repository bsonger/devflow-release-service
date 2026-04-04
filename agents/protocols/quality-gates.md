# Quality Gates

## Purpose

定义什么情况下可以宣布一轮任务完成。

## Scope

适用于所有非简单需求。

## Must

- 必须存在：
  - `product-spec`
  - 当前轮 `sprint-contract`
  - `evaluator-report`
  - `handoff`
- Evaluator 结论必须是三选一：
  - `Pass`
  - `Pass with risks`
  - `Fail`
- 若测试该跑未跑，必须在 evaluator report 明示

## Must Not

- 不得用“实现了大部分”代替通过
- 不得用“后面再补 evaluator”代替当前验收
- 不得把未验证风险埋进 handoff 里装作已通过

## Outputs

- 明确的 go / no-go 结论

## Pass/Fail

- `Pass`：阻断项全部满足
- `Fail`：任一阻断项缺失或未验证
