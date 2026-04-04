# Architecture

## Purpose

`devflow-release-service` 负责 release 控制面，只管理 `Manifest`、`Job`、`Intent`。

## Inbound Surface

- `GET/POST /api/v1/manifests`
- `GET /api/v1/manifests/:id`
- `PATCH /api/v1/manifests/:id`
- `GET/POST /api/v1/jobs`
- `GET /api/v1/jobs/:id`
- `GET /api/v1/intents`
- `GET /api/v1/intents/:id`

## Data And Dependencies

- 主存储：MongoDB
- 主要集合：`manifests`、`job`、`execution_intents`
- 运行时可能接入 Tekton、Argo CD、意图执行链路
- 启动、路由、HTTP 公共件、观测基础设施优先来自 `devflow-service-common`

## Outbound Rules

- 任何调用 Tekton、Argo CD 或其他外部系统的代码都必须同时产生 `metrics + trace + structured log`

## Non-Goals

- 不负责 `Project`、`Application` 元数据 CRUD
- 不负责 `Configuration`
- 不负责 verify webhook / writeback 入站
