# 认证运行告警

API 的 `/metrics` 暴露认证计数器。生产 Prometheus 可以加载 [`deploy/prometheus/auth-template-alerts.yml`](../deploy/prometheus/auth-template-alerts.yml)。

限流 Redis 故障时，登录、回调和刷新默认 fail closed，返回 `503 RATE_LIMIT_DEPENDENCY_FAILED`；这会优先保护认证接口，必须触发 `AuthRateLimitDependencyErrors` 告警并恢复 Redis。测试和开发环境可以使用内存限流器，但不适合作为多实例生产实现。

登出流程先撤销本地 Session，再尝试调用 `identity.revocationEndpoint`。远端撤销失败不会让本地登出失败，但会增加 `auth_remote_revoke_errors_total` 并触发 `AuthRemoteRevokeErrors`。如果身份平台没有提供 Revocation Endpoint，应在 Provider 契约中明确这一点，并使用其支持的终止会话流程或后台补偿任务。
