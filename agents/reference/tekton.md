# Tekton Reference

## Purpose

说明 Tekton `PipelineRun` / `TaskRun` 在当前仓库中的事件语义。

## Scope

适用于：

- build 事件来源
- Tekton 状态解析

## Must

- Tekton 是 CI/CD 流水线执行记录与事件来源
- 当前只读：仅解析状态、日志、事件
- 重点字段：
  - `PipelineRun` 状态
  - `TaskRun` 状态
  - 开始时间
  - 结束时间
  - 错误信息

## Must Not

- 不在本仓库中把 Tekton 事件语义和本地默认状态混为一谈
- 不把 TaskRun/PipelineRun 当作可随意回写的本地资源

## Outputs

- build 状态回写信号
- task 级步骤信息

## Pass/Fail

- `Pass`：Tekton 仍是 build 过程的外部事实来源
- `Fail`：本地状态覆盖或替代了 Tekton 事实
