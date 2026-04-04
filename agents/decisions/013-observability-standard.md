# ADR 013: Observability Standard For Split Services

## Status

Accepted

## Context

`devflow` 正在从单体 API 拆为多个 control-plane service，同时还有 `devflow-controller`、后续 executor、worker 等外围进程。

如果不提前统一观测约定，后果会很直接：

- trace 无法跨 service 串起来
- logs 无法和 trace 对齐
- metrics label 失控
- 一旦 worker / verify / controller 出现偶发错误，很难定位根因

## Decision

统一采用以下标准：

- traces：OpenTelemetry
- metrics：OpenTelemetry
- logs：结构化 JSON + trace correlation
- profiling：`pprof` + Pyroscope
- exporter 路径：服务 -> OTel Collector -> 实际后端

## Why

- OTel 负责统一采集 API 和上下文传播，适合作为语言无关的基线。
- logs 需要先解决“可关联”和“字段统一”，不必在第一阶段引入 OTel Logs 复杂度。
- Go 服务的火焰图和持续 profiling，`pprof + Pyroscope` 组合最直接。
- Collector 可以把 vendor lock-in 从服务代码里拿掉。

## Consequences

后续新增 service 时必须满足：

- 所有入站/出站调用可挂 trace
- 所有日志可按 `trace_id` / `span_id` 关联
- 所有关键控制面路径有基础 metrics
- release worker / verify / controller 可抓 profile

不再接受：

- 每个服务自己接不同 exporter
- 把高基数业务 ID 放进 metrics label
- 只有日志、没有 trace 的跨服务调用
- 把 profiling 当成“以后再说”

## Follow-up

- 在共享 bootstrap 里收敛统一 telemetry 初始化
- 为 HTTP client / router / worker loop 增加统一 instrumentation
- 为 verify / release worker 补充最小 profiling 接入
