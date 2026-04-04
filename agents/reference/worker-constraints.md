# Worker Constraints

## Purpose

约束 `Generator` / `Worker` 类角色，避免“顺手多改”“自评通过”“超边界扩散”。

## Scope

适用于：

- 所有 sprint contract 驱动的实现任务
- 所有代码、配置、文档、测试交付

不适用于：

- `Planner` 写 spec 和 contract
- `Evaluator` 做独立验收

## Must

- 一次只交付一个 sprint contract 定义的最小切片
- 只实现当前 sprint 的目标
- 不越过 `允许修改文件` 的边界
- 明确区分：
  - 已实现
  - 已验证
  - 仅推断
  - 未完成
- 若 contract 要求文档、Swagger、测试同步，必须显式处理
- 若发现用户已有未提交改动，必须在不回滚对方改动的前提下继续
- contract 外但明显必要的补充改动，只允许做支撑当前切片成立的最小补丁
- 优先复用现有模型、状态枚举、路由和 service 边界
- `release-service` / `verify-service` / `controller` 相关改动，必须检查 observability 是否被破坏
- 只要新增或变更 API 行为，就必须判断 Swagger 是否需要同步
- 只要涉及 `Application` / `Manifest` / `Job` 状态，就必须对齐 ADR 001 的状态来源约束

## Must Not

- 不把“推测正确”写成“已经验证”
- 不把未完成项包装成已完成
- 不顺手重构无关模块
- 不在同一轮同时覆盖多条状态流
- 不用更大范围改动绕过局部设计问题
- 不用“后面再修”代替当前 sprint 的阻断项
- 不自定义新的 done 标准
- 不引入高基数 metrics label
- 不把日志当成状态真相来源
- 不把 debug 字段写进正式 API 响应
- 不让 controller / worker 直接越权修改 metadata 真相来源

## Outputs

- 改动摘要
- 验证结果
- 未验证项
- 剩余风险
- 外部依赖假设及影响范围

## Pass/Fail

- `Pass`：交付范围受控、验证状态清楚、无越权扩散
- `Fail`：超边界、伪装完成、或破坏仓库既有状态/观测约束
