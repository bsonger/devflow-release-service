# Harness Skill

## Purpose

定义当前仓库处理长任务时的三角色 Harness 规则。

## Scope

适用于：

- 需求跨 `pkg/api`、`pkg/service`、`pkg/router`、`agents/` 多个位置
- 同时影响 `Application`、`Manifest`、`Job` 中两个及以上资源
- 预计需要多轮迭代或明确 handoff 的任务

不适用于：

- 单文件错别字修复
- 纯机械格式化

## Must

- 默认加载 `Planner`、`Generator`、`Evaluator` 三个角色
- 若运行时允许 delegation，应物理创建 3 个 agent
- 若运行时不允许 delegation，也必须按这 3 个角色串行执行
- 非简单需求开始前先创建 `agents/runs/YYYYMMDD-<slug>/`
- 启动入口以 `agents/manifest.yaml` 和 `agents/protocols/startup.md` 为准
- `Planner` 负责 spec 与 contract
- `Generator` 必须遵守 `reference/worker-constraints.md`
- `Evaluator` 必须遵守 `reference/evaluator-rubric.md`
- 切片优先按 `Application` / `Manifest` / `Job` 或单条状态流拆分
- 若改动 API、路由或响应模型，必须判断 Swagger 是否需要同步
- 若外部系统不可改，只能写清假设、边界和未验证项
- 每轮必须产出当前 contract、evaluator report、handoff
- 若 Evaluator 给出 `Fail`，默认自动进入下一轮 contract
- 每轮 contract 必须显式写出 `worker constraints`
- 每轮 contract 必须显式写出 `evaluator rubric` 或引用固定 rubric

## Must Not

- 不跳过 run 目录直接实现
- 不在没有 contract 时直接进入实现
- 不在没有 evaluator 结论时宣布完成
- 不把多轮大任务伪装成一轮可完成任务
- 不用“主体功能已完成”替代阻断项验收

## Outputs

- `agents/runs/YYYYMMDD-<slug>/`
- `product-spec.md`
- 当前轮 `sprint-contract`
- `evaluator-report`
- `handoff`

## Pass/Fail

- `Pass`：3 角色、工件、质量门都被执行，流程可复盘
- `Fail`：角色缺失、contract 缺失、或 evaluator 被跳过
