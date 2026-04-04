# Agents 文档入口

`agents/` 是本仓库的 agent control plane。

目录职责：

1. `manifest.yaml`：机器可读入口
2. `roles/`：角色定义
3. `protocols/`：启动、工作流、质量门、handoff、升级协议
4. `reference/`：历史事实和规则补充
5. `types/`：类型速查
6. `patterns/`：固定写法
7. `templates/`：标准工件模板
8. `runs/`：每条真实需求的运行工件
9. `examples/`：完整示例
10. `decisions/`：架构决策

推荐阅读顺序：

1. `AGENTS.md`
2. `agents/manifest.yaml`
3. `agents/protocols/startup.md`
4. `agents/roles/`
5. `agents/reference/`
6. `agents/patterns/`
7. `agents/templates/`
8. `agents/runs/README.md`
9. `agents/examples/`

如果 `AGENTS.md` 与其他文档冲突，以 `AGENTS.md` 为准。

当前仓库默认启动模式：

- 默认 harness 为 `Planner -> Generator -> Evaluator`
- 运行时支持 delegation 时，必须真实启动 3 个 sub-agent
- 不支持 delegation 时，也不能省略这 3 个角色的 contract、handoff 和 evaluator 结论
- 每个非简单需求先创建 `agents/runs/YYYYMMDD-<slug>/` 工件目录

快速校验：

- 结构与标准文档检查：`bash scripts/check_agents.sh`
