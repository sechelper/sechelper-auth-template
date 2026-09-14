# 统一认证平台集成

## 调用方向

业务服务作为 Confidential Client 调用统一认证平台。浏览器只访问业务服务，不直接调用 Token Endpoint。

## 当前客户端接口

实现位于 `api/internal/platform/identity/client.go`：

- Authorization Code + PKCE Token Exchange；
- UserInfo；
- 当前 Application Authorization。

Manifest 状态查询、提交和幂等同步已实现，调用路径为统一认证平台的 `/api/v1/provisioning/authorization-manifest/status` 和 `/api/v1/provisioning/authorization-manifest`。客户端使用 Identity 的 `client_id` 与 `client_secret` 调用 `/api/v1/provisioning/token`，以 `client_credentials`、固定 audience `provisioning-api` 和 scope `authorization.manifest` 获取 access token，再以 `Authorization: Bearer` 调用 Manifest API，并在提交时携带 `Idempotency-Key`。系统不再支持独立的 Manifest Credential。

## 安全

Client Secret 只通过服务端配置注入。Token 只在认证平台客户端和认证用例内部传递，不写入日志、数据库或浏览器响应。

## 超时和失败

当前 HTTP Client 使用 15 秒总超时，并将外部失败映射为本地认证失败。后续需要根据平台限流协议增加有限重试、429 处理、熔断和依赖健康状态。
