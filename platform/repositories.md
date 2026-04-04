# Service Repository Split

当前 `bsonger` 账号下已存在的 Devflow 相关仓库包括：

- `devflow`
- `devflow-common`
- `devflow-console`
- `devflow-controller`

基于当前 `platform/` 的服务边界，建议继续拆出 4 个独立仓库：

## Target Repositories

- `devflow-app-service`
  - 责任：`Project`、`Application` 元数据 CRUD 与 `active_manifest` 绑定
  - 入口：`platform/app-service/cmd/main.go`
  - 目标模块：`github.com/bsonger/devflow-app-service`

- `devflow-config-service`
  - 责任：`Configuration` 元数据 CRUD
  - 入口：`platform/config-service/cmd/main.go`
  - 目标模块：`github.com/bsonger/devflow-config-service`

- `devflow-release-service`
  - 责任：`Manifest`、`Job`、`Intent` 查询、release intent 编排、worker 消费
  - 入口：`platform/release-service/cmd/main.go`
  - 目标模块：`github.com/bsonger/devflow-release-service`

- `devflow-verify-service`
  - 责任：Tekton / Argo / controller 事件接入、校验与状态回写
  - 入口：`platform/verify-service/cmd/main.go`
  - 目标模块：`github.com/bsonger/devflow-verify-service`

## Current Export Strategy

- 第一阶段不是做“彻底去共享代码”的大重构。
- 第一阶段先把每个 service 导出成可独立初始化、可独立编译、可独立推送的单仓库。
- 导出后的仓库会保留当前需要的共享包副本，例如 `pkg/model`、`pkg/service`、`pkg/router`、`pkg/config`、`pkg/telemetry`、`platform/shared/bootstrap`。
- 真正的二阶段再收敛共享代码边界，逐步下沉到 `devflow-common` 或新的 shared module。

## Tooling

- 导出单个服务仓库：`bash scripts/export_service_repo.sh <service>`
- 批量导出 4 个服务仓库：`bash scripts/export_service_repos.sh`
- 创建单个 GitHub 仓库：`bash scripts/create_github_repo.sh <repo-name> <description>`
- 批量创建 4 个 GitHub 仓库：`bash scripts/create_service_repos.sh`

默认导出目录：

- `/tmp/devflow-split`

## Boundary Notes

- SSH 私钥已经可以认证到 `git@github.com:bsonger/*`
- 但“创建 GitHub 仓库”仍需要 GitHub API token 或 `gh` 登录态
- 因此当前自动化拆分分成两步：
  - 本地导出并初始化 git 仓库
  - GitHub 仓库创建成功后再 push
