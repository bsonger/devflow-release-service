# 004 长任务 Harness

- 决策：复杂任务采用 `Planner -> Generator -> Evaluator` 三段式，并通过文件工件传递上下文；简单任务仍优先单回合直接完成。
- 原因：
  - Devflow 变更常跨 `pkg/api`、`pkg/service`、`pkg/router` 与 `agents/` 文档，长对话容易丢失资源关系与状态流上下文。
  - `Application`、`Manifest`、`Release` 存在串联状态，自评容易漏掉跨资源回归。
  - 项目依赖 Argo CD、Tekton、PostgreSQL、OTel，很多正确性必须靠明确验收，而不是主观判断。
- 影响：
  - 项目启动时默认先进入 `Planner -> Generator -> Evaluator` 角色模型。
  - 若运行时允许 sub-agent delegation，则优先实际创建 3 个 agent；否则串行执行同一协议。
  - 先产出高层 spec，再按 sprint contract 切成单资源或单状态流任务。
  - 每轮必须留下 handoff，记录改动、验证、风险和下一轮切片。
  - Evaluator 以编译、测试、HTTP 行为、状态一致性和文档同步为准；任一阻断项失败即 sprint 失败。
  - 若任务只涉及单文件文档或简单修订，不启用完整 Harness，保持最小复杂度。
