# 认证设计

## 客户端类型

只支持 Confidential Client。浏览器不得获取 Client Secret、access token 或 refresh token。

## 流程

1. 浏览器访问 `GET /v1/auth/login`。
2. 服务端生成 state、nonce、PKCE verifier，并保存短期登录事务。
3. 服务端跳转统一认证平台。
4. `GET /v1/auth/callback` 接收授权码并校验 state。
5. 服务端向 Token Endpoint 兑换 Token。
6. 服务端读取 UserInfo 和当前 Application 授权。
7. 服务端创建会话并返回 HttpOnly Cookie。

前台业务进入时只自动请求本地会话；未登录时显示业务自己的登录按钮，点击后使用 OIDC `prompt=login` 发起交互式登录。已登录时由业务页面正常加载数据，不因首页未登录而自动跳转认证。

`GET /v1/auth/login` 只接受 `prompt=none` 或 `prompt=login`。认证模式绑定在一次性随机 `state` 的服务端事务中，不能由回调请求改写。业务登出后不会自动重新登录；用户再次进入业务时才会触发新的登录检查。

退出登录先撤销本地 Session 和 Refresh Token。OIDC Profile 展示页完成退出后保持在当前页面，前端重新读取本地会话并直接显示未登录状态，不跳转到身份中心退出页面。服务端仍兼容生成 OIDC RP-Initiated Logout 地址和校验退出回调，供其他 API 客户端按需使用；浏览器 Profile 页不使用该地址，因此不会离开当前页面。

服务端向统一认证平台调用业务 Manifest API 时，复用同一组 Confidential Client 的 `client_id` 与 `client_secret`，通过 Client Credentials Grant 获取短期 access token，并使用 Bearer Token 调用接口；浏览器和本地数据库都不接触该 access token。系统不支持 `manifest_id`、`manifest_secret` 或独立 Manifest 凭据。

回调完成前，服务端必须校验授权码交换返回的 ID Token：签名算法、JWKS `kid`、issuer、以 Client ID 为目标的 audience、有效期、nonce，以及 ID Token、UserInfo 和授权响应中的 subject 一致性。`identity.audience` 是业务 API audience，不用于 ID Token 的 `aud` 校验。JWKS 地址由 `IDENTITY_JWKS_URL` 配置，不能从请求参数提供。

## 会话

当前代码已提供 PostgreSQL Session Store。登录事务的 state、nonce 和 PKCE verifier 已保存到 `authentication_login_transactions`，回调通过一次性 DELETE 原子消费，支持多实例回调并防止 state 重放。Session 通过 `version` 字段进行乐观并发控制，刷新写回使用版本条件；跨实例冲突会拒绝旧版本更新。数据库结构必须先通过独立迁移命令完成；API 启动时只检查数据库连通性，不再隐式修改 schema。Redis、多实例缓存广播和远端 Token 撤销仍需在统一认证平台契约确认后实现。

`/v1/auth/login`、`/v1/auth/callback` 和 `/v1/auth/refresh` 由 `rateLimit` 配置限流。生产使用 Redis 计数器，开发环境在未配置 Redis 时使用进程内计数器；超过限制返回 `429 RATE_LIMITED` 和 `Retry-After`。

服务端已提供 `POST /v1/auth/refresh`。Refresh Token 只以 AES-GCM 密文保存在 PostgreSQL，密钥由 `SESSION_ENCRYPTION_KEY` 注入，浏览器不接触 Refresh Token。生产 PostgreSQL Store 会在事务中锁定当前 Session 行，完成 Provider 刷新、密文替换和版本递增，跨实例不依赖进程内互斥锁。统一认证平台返回新的 Refresh Token 时，服务端替换旧密文；未返回时保留原有密文。Provider 明确返回 OAuth `invalid_grant` 时视为 Rotation Reuse，服务端在同一事务内撤销该 Session，客户端必须重新登录；网络超时和其他 Provider 错误不会静默授权。

Refresh Rotation 的确定性回归测试位于 `api/internal/modules/authentication/application/service_test.go`。测试使用 Mock Identity Provider 固定返回两代 token，不依赖浏览器 Cookie 或等待 Session 自然过期，验证 Session 有效期更新、Refresh Token 密文轮换，以及旧 token 被拒绝。

## 当前仍需补齐的安全项

- 更严格的 Origin/Host Policy 校验；
- 统一登出和远端 Token 撤销；
- 会话审计；
- 登录事务的多实例共享。当前登录事务仍由认证服务内存保存，生产多副本部署前需要迁移到 Redis 或 PostgreSQL。
- Refresh Token 的跨实例乐观锁/行锁。当前已增加单进程串行保护，不能替代多实例并发控制。

退出登录要求 `X-CSRF-Token` 与非 HttpOnly 的 CSRF Cookie 一致；浏览器认证 API 客户端负责自动携带该 Header。业务写接口必须复用同一 CSRF 策略。服务端先撤销本地 Session，再按可选的 `identity.revocationEndpoint` 撤销远端 Refresh Token；配置了可选的 `identity.endSessionEndpoint` 时才生成身份平台终止会话地址，两个远端 Endpoint 都不是首次安装必填项。未配置时仍完成本地登出；远端失败不阻断本地登出，但会记录指标和告警。
