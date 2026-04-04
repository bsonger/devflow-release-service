# Handoff Protocol

## Purpose

定义一轮任务结束后，如何把最小但足够的上下文交给下一轮。

## Scope

适用于所有通过 Evaluator 或进入下一轮的任务。

## Must

- 记录：
  - 本轮结论
  - 已完成内容
  - 修改文件
  - 验证记录
  - 未解决风险
  - 下一轮最小切片
  - Evaluator 重点复查项
  - Evaluator 结论
- 内容必须对应真实改动，不能复制模板占位

## Must Not

- 不得把 changelog 当 handoff
- 不得省略风险和下一轮动作
- 不得用模糊表述替代结论

## Outputs

- `handoff.md`

## Pass/Fail

- `Pass`：下一位 agent 不看聊天记录也能继续
- `Fail`：handoff 仍依赖口头上下文
