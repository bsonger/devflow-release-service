# Escalation Protocol

## Purpose

定义何时继续自动推进，何时必须显式升级风险。

## Scope

适用于所有自动串联流程。

## Must

- 出现以下情况必须升级为显式风险：
  - 外部系统无法验证
  - 用户已有改动与当前任务冲突
  - 需要越权修改真相来源
  - 需要新增未声明配置、端口、协议
  - 需要跨仓库改动
- `Evaluator` 若发现阻断项，必须输出 `Fail` 或 `Pass with risks`

## Must Not

- 不得把阻断风险静默吞掉
- 不得在高风险假设下继续把结果宣告为完成

## Outputs

- 风险说明
- 影响范围
- 下一轮最小补救动作

## Pass/Fail

- `Pass`：风险已显式登记并进入 contract/handoff
- `Fail`：风险被隐藏或模糊处理
