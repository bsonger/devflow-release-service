# Rollout Reference

## Purpose

说明 Argo Rollout 在当前仓库里的角色边界。

## Scope

适用于：

- 蓝绿发布
- 金丝雀发布
- Rollout 观察器回写链路

## Must

- Rollout 是发布策略状态载体
- 当前只读：只查询状态与事件，不直接变更资源
- 重点字段：
  - 当前阶段
  - 可用副本数
  - 回滚信息
- Rollout 观察器应通过 `verify-service` 的 `release/steps` 接口回写 `release.steps`

## Must Not

- 不直接写 PostgreSQL
- 不在观察器侧决定 metadata 真相来源

## Outputs

- release steps 回写信号
- 发布策略状态观察结果

## Pass/Fail

- `Pass`：Rollout 继续作为外部状态来源
- `Fail`：Rollout 观察器直接越权写 metadata 真相
