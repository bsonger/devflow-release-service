# Reconcile Pattern

## Purpose

定义 controller / verify / worker 处理事件时的最小幂等路径。

## Scope

适用于：

- Argo / Tekton / release step 回写
- 所有把外部事件映射成领域状态的流程

不适用于：

- 单纯的本地 CRUD
- 不涉及状态收敛的纯查询逻辑

## Must

- 按 `事件输入 -> 解析对象 -> 映射为领域状态 -> 幂等写回` 的顺序执行
- 状态变化必须可重复执行，不能因为重复事件造成回退或脏写
- 关键步骤必须保留结构化日志和必要的 trace 上下文
- 终态处理必须遵守状态来源约束

## Must Not

- 不依赖单次事件成功作为唯一真相
- 不跳过幂等保护直接覆盖终态
- 不把外部事件字段原样暴露成内部 API 语义

## Outputs

- 可重复执行的 reconcile 逻辑
- 与领域模型一致的状态写回

## Pass/Fail

- `Pass`：重复事件不会破坏状态，链路可观测
- `Fail`：存在回退终态、重复写乱、或不可追踪的 reconcile
