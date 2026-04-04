# Workflow Protocol

## Purpose

定义一条用户需求如何被自动串联到实现和验收。

## Scope

适用于所有走 Harness 的任务。

## State Machine

`Request -> Product Spec -> Sprint Contract -> Generator Output -> Evaluator Report -> Handoff`

若 `Evaluator = Fail`：

`Evaluator Report -> Next Sprint Contract -> Generator Output -> Evaluator Report`

若 `Evaluator = Pass with risks`：

- 非阻断：进入 `Handoff`
- 仍有阻断风险：进入下一轮 contract

## Must

- `Planner` 先于 `Generator`
- `Generator` 先于 `Evaluator`
- `Evaluator` 结论先于 `Handoff`
- 下一轮 contract 必须比上一轮更小、更具体

## Must Not

- 不得跳步
- 不得并行写多个互相冲突的 contract
- 不得在没有 evaluator 结论时结束流程

## Outputs

- 每一轮的 contract / evaluator report / handoff

## Pass/Fail

- `Pass`：流程从 request 串到 handoff 或明确 fail boundary
- `Fail`：任一关键工件缺失或顺序被打乱
