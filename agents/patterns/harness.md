# 长任务 Harness 模式

- 目标：把长任务拆成小而可验收的工件，避免上下文漂移、自评失真和跨资源回归漏检。
- 工作流：
  1. 初始化 run 目录并记录原始需求。
  2. `Planner` 先写产品/改动 spec。
  3. `Planner` 写本轮 sprint contract。
  4. `Generator` 只实现一个最小切片。
  5. `Evaluator` 独立验收，不接受“看起来差不多”。
  6. 输出 evaluator report。
  7. `Pass` 或非阻断 `Pass with risks` 后写 handoff。
  8. `Fail` 时自动生成下一轮 contract，再进入下一轮。

## Devflow 切片建议

- 优先按资源拆：
  - `Application`
  - `Manifest`
  - `Release`
  - `Configuration`
- 其次按状态流拆：
  - 创建
  - 激活 manifest
  - Tekton 构建结果回写
  - Argo 同步
  - 回滚
- 避免一次 sprint 同时覆盖多条状态流。

## Contract 最低要求

- 明确本轮目标与非目标。
- 明确允许修改的文件范围。
- 明确 worker 约束：
  - 禁止顺手扩散
  - 禁止自定义 done
  - 禁止把未验证写成已验证
- 把 `done` 写成可检查行为：
  - HTTP 状态码与响应体
  - PostgreSQL 字段变化
  - `Application` / `Manifest` / `Release` 状态值
  - 日志关键点
  - Swagger 是否需要同步
- 明确验证方法：
  - `go test ./...`
  - 路由级请求检查
  - 文档一致性检查
- 明确 evaluator rubric：
  - `Pass`
  - `Pass with risks`
  - `Fail`

## 自动化要求

- 非简单需求默认不手动跳步。
- 没有 `product-spec`，不能直接开始实现。
- 没有 `evaluator-report`，不能直接宣布完成。
- `Fail` 不能停在结论本身，必须转成下一轮最小 contract。

## Evaluator 检查项

- 功能正确：handler、service、状态流是否符合 contract。
- 资源一致：资源 ID、名称、状态来源是否前后一致。
- 可靠性：错误是否上抛，更新是否幂等，关键步骤是否留日志。
- 运维约束：不越权改外部系统，不引入未声明配置。
- 文档同步：`agents/`、Swagger、README 是否需要同步。
- 可观测性：trace / log / metrics / profile 入口是否仍满足 observability contract。

## 适合当前仓库的示例切片

- Sprint A：`Application.UpdateActiveManifest` 的校验与错误语义。
- Sprint B：`Manifest.Patch` 的字段边界与 Swagger 同步。
- Sprint C：`Release.Create -> syncArgo -> status` 链路的一致性。
- Sprint D：Tekton / Argo 事件回写的状态来源说明。
