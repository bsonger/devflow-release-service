# Architecture

## Purpose

`devflow-release-service` 负责 release 控制面，只管理 `Manifest`、`Release`、`Intent`。

## Inbound Surface

- `GET/POST /api/v1/manifests`
- `GET /api/v1/manifests/:id`
- `PATCH /api/v1/manifests/:id`
- `GET/POST /api/v1/releases`
- `GET /api/v1/releases/:id`
- `GET /api/v1/intents`
- `GET /api/v1/intents/:id`

## Data And Dependencies

- 主存储：MongoDB
- 主要集合：`manifests`、`job`、`execution_intents`
- `job` 集合作为 `Release` 的历史存储名暂时保留，避免本次顺手引入数据迁移
- 运行时可能接入 Tekton、Argo CD、意图执行链路
- 启动、路由、HTTP 公共件、观测基础设施优先来自 `devflow-service-common`

## Outbound Rules

- 任何调用 Tekton、Argo CD 或其他外部系统的代码都必须同时产生 `metrics + trace + structured log`

## Non-Goals

- 不负责 `Project`、`Application` 元数据 CRUD
- 不负责 `Configuration`
- 不负责 verify webhook / writeback 入站
