# 012 verify-service 写接口使用可选共享 token

- 决策：
  - `verify-service` 的写接口在设置 `VERIFY_SERVICE_SHARED_TOKEN` 后，统一要求 `X-Devflow-Verify-Token` 请求头。
  - `healthz` 保持无鉴权。
- 原因：
  - 在把 `devflow-controller`、Tekton 观察器、Argo 观察器这类外部写回方接入后，`/api/v1/verify/*` 不应继续裸露。
  - 共享 token 是在当前拆分阶段最小可落地的保护手段，不需要先引入完整的 service-to-service 身份体系。
- 影响：
  - controller / observer 进程需要持有同一份共享 token。
  - 未设置环境变量时，仍保持向后兼容，不影响本地开发。
