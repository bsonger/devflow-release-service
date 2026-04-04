# Runs Reference

## Purpose

定义每条真实需求的 run 目录如何组织和持续迭代。

## Scope

适用于：

- 所有走 Harness 的非简单需求
- 需要保存 request / spec / contract / evaluator / handoff 的任务

不适用于：

- 单条即时问答
- 不需要 run 工件的简单修订

## Must

- 每个 run 对应一条用户需求
- 目录命名使用 `agents/runs/YYYYMMDD-<slug>/`
- 默认至少包含：
  - `request.md`
  - `product-spec.md`
  - `sprint-01-contract.md`
  - `evaluator-report.md`
  - `handoff.md`
- 若同一需求进入多轮，继续追加：
  - `sprint-02-contract.md`
  - `evaluator-report-02.md`
  - `handoff-02.md`
- `Planner`、`Generator`、`Evaluator` 必须围绕同一个 run 目录工作
- `Evaluator` 若未给出 `Pass`，必须进入下一轮 contract，而不是直接结束

## Must Not

- 不把多个无关需求混进同一个 run
- 不跳过 request/spec/contract 直接产出 handoff
- 不在 run 目录里只保留最终结果而丢掉中间验收痕迹

## Outputs

- 可审计的 run 工件目录
- 每轮 contract / evaluator / handoff 记录

## Pass/Fail

- `Pass`：run 目录能独立复盘一条需求的完整生命周期
- `Fail`：工件缺失、命名混乱或多轮记录断裂
