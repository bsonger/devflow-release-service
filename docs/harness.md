# Harness

## Default Flow

默认工作流固定为：

- `Planner`
- `Generator`
- `Evaluator`

## Delegation Rule

- 运行环境支持 sub-agent 时，必须真实启动这 3 个角色
- 不支持 delegation 时，也必须保留 spec、contract、evaluator report、handoff

## Required Artifacts

- `request.md`
- `product-spec.md`
- `sprint-01-contract.md`
- `evaluator-report.md`
- `handoff.md`
