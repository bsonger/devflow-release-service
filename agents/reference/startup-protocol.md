# Startup Protocol

兼容入口：当前文件保留为旧路径别名。

规范入口改为：

- `agents/protocols/startup.md`

本仓库默认采用 3 角色 Harness 作为项目启动协议。

## 启动时默认加载

进入项目后，默认先加载以下 3 个角色：

- `Planner`
- `Generator`
- `Evaluator`

这不是“可选建议”，而是当前仓库的默认工作模式。

## 自动串联流程

收到一条非简单需求后，默认执行以下自动流程：

1. 生成 run 目录：`agents/runs/YYYYMMDD-<slug>/`
2. 写入 `request.md`
3. `Planner` 产出 `product-spec.md`
4. `Planner` 产出 `sprint-01-contract.md`
5. `Generator` 按 contract 实现
6. `Evaluator` 产出 `evaluator-report.md`
7. 若结论为：
   - `Pass`：写 `handoff.md`，流程结束
   - `Pass with risks`：写 `handoff.md`，如仍存在阻断风险则自动进入下一轮 sprint
   - `Fail`：自动进入下一轮 contract，再次交给 Generator

目标不是“讨论完就停”，而是默认推进到可交付或明确失败边界。

## 启动顺序

1. 读取根目录 `AGENTS.md`
2. 读取 `agents/README.md`
3. 读取 `agents/skills/harness.md`
4. 读取：
   - `agents/reference/worker-constraints.md`
   - `agents/reference/evaluator-rubric.md`
   - `agents/reference/observability.md`
5. 根据任务生成或更新本轮：
   - `product spec`
   - `sprint contract`
   - `handoff`

## 角色职责

### Planner

- 将需求收敛成单轮 sprint
- 明确目标、非目标、边界、验收、风险
- 不提前把实现细节锁死到无法调整

### Generator

- 只实现当前 sprint
- 严格遵守 `worker constraints`
- 不顺手扩散到无关模块

### Evaluator

- 独立按 `evaluator rubric` 给出结论
- 结论只能是：
  - `Pass`
  - `Pass with risks`
  - `Fail`

## “自动加载” 的真实含义

仓库级规则要求项目启动时默认进入这 3 个角色。

但需要区分两层：

- 角色层：必须始终存在
- 进程层：是否真的物理创建 3 个 sub-agent，取决于当前运行时是否允许 delegation

也就是说：

- 若运行时允许并且当前会话允许 delegation，应实际创建 3 个 agent
- 若运行时不允许，则由当前主 agent 串行扮演这 3 个角色，但 contract 和验收标准不能省略
- 若运行时支持本地脚本，可先调用 `scripts/init_harness_run.sh` 初始化 run 目录；若不支持，也必须按相同目录结构手动产出工件

## 适用范围

- 默认适用于当前仓库的大多数工程任务
- 简单单文件文档修订可以降级为串行单 agent，但仍保留 Planner / Generator / Evaluator 的最小检查语义

## 当前仓库的特殊要求

- 只要任务涉及 `Application` / `Manifest` / `Release` / `Intent` 之一，就优先进入完整 3 角色模式
- 只要任务涉及 `release-service` / `verify-service` / `controller` / `worker`，Evaluator 必查 observability
- 只要任务涉及 API 行为变化，必须判断 Swagger 是否需要同步
