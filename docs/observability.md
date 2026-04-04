# Observability

## Mandatory Rule

只要涉及调用其他服务或外部系统，就必须同时产出：

- metrics
- trace
- structured log

## Inbound HTTP

- 所有业务请求必须有 server span
- 记录请求次数、耗时、错误数
- `/metrics`、`/healthz`、`/readyz`、`/debug/pprof/*` 不计入业务指标

## Outbound

- 对 Tekton、Argo CD、worker dispatch 或其他外部系统的调用必须有 client span
- 必须记录调用次数、延迟、错误数和结构化日志

## Log Fields

- 基础字段：`service`、`trace_id`、`span_id`、`request_id`
- 资源字段：`application_id`、`manifest_id`、`release_id`、`intent_id`、`external_ref`

## Profile

- 保留 `/debug/pprof/*`
- worker 和 API 进程都应允许 profile 定位热点
