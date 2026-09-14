# 配置契约

应用配置由根目录 `config.yaml`（模板为 [`config.example.yaml`](../config.example.yaml)）定义，Go 服务通过独立的 Viper 实例在启动时读取一次、严格反序列化并完成语义校验。启动参数 `--config <path>` 优先级最高，其次是 `APP_CONFIG_FILE`；仅 development/test 允许回退到工作目录的 `config.yaml` 或 `config.yml`。production 必须显式指定且可读的配置文件。

`.env` 只属于 Docker Compose：用于 `${VAR}` 插值，并且只有 Compose 文件显式列出的变量才会传入容器。它不是 Viper 配置文件，不能被复制进镜像或源代码快照；实际文件由 Git 忽略。`config.yaml` 只放非敏感应用设置，密码、token、私钥和凭据通过显式环境变量或部署密钥服务注入。

## YAML 结构

当前分组为 `app`、`server`、`database`、`identity`、`manifest`、`session`、`rateLimit`、`cors` 和 `log`，键名使用 lower-camel-case；未知键会导致启动失败。数据库连接由 `database.host/port/name/user/password/sslMode` 在内存中组装，禁止使用含凭据的 DSN。`manifest` 仅配置同步周期；Manifest API 使用 `identity.clientId/clientSecret` 获取 access token。

环境变量覆盖映射如下（环境变量优先于 YAML）：

| YAML | 环境变量 |
|---|---|
| `app.environment` / `app.version` | `APP_ENV` / `APP_VERSION` |
| `app.listenAddr` | `LISTEN_ADDR` |
| `app.publicWebOrigin` / `app.apiOrigin` | `PUBLIC_WEB_ORIGIN` / `API_ORIGIN` |
| `server.port` / `server.*Timeout` | `SERVER_PORT` / `SERVER_READ_TIMEOUT` 等 |
| `database.host/port/name/user/password/sslMode` | `DATABASE_HOST/PORT/NAME/USER/PASSWORD/SSL_MODE` |
| `identity.*` | 同名 `IDENTITY_*`；`identity.jwksUrl`、`identity.revocationEndpoint`、`identity.endSessionEndpoint` 分别对应 `IDENTITY_JWKS_URL`、`IDENTITY_REVOCATION_ENDPOINT`、`IDENTITY_END_SESSION_ENDPOINT` |
| `manifest.syncInterval` | `MANIFEST_SYNC_INTERVAL` |
| `session.*` | 同名 `SESSION_*` |
| `cors.allowedOrigins` | `ALLOWED_ORIGINS` |
| `log.*` | 同名 `LOG_*`；`log.debug` 为 `DEBUG` |

逗号分隔的 `IDENTITY_SCOPES` 与 `ALLOWED_ORIGINS` 会被拆分。`SESSION_ENCRYPTION_KEY` 必须是 RawStdEncoding 编码的 32 字节随机值；production 还要求 `session.secure: true`、非空 `app.redisUrl` 和 HTTPS 的 `identity.jwksUrl`。Origin 必须含 scheme、不得含路径、凭据或通配符。

日志字段配置位于 `log` 分组：`mode` 支持 `file` 和 `stdout`，`encoding` 固定为 `json`；`timeKey`、`levelKey`、`messageKey`、`callerKey`、`stacktraceKey` 控制 Zap 的标准字段名，时间编码为 `iso8601`，级别编码为 `lowercase`。`development`、`disableCaller`、`disableStacktrace` 和 `sampling` 只控制 Zap 的诊断行为，不能关闭脱敏、输出限制、轮转或保留策略；`debug` 是服务端诊断开关，不能由请求输入控制。

相对 `log.output` 路径以服务二进制所在目录为基准，因此默认文件是二进制旁的 `logs/app.log`，日志目录权限为 `0700`、文件权限为 `0600`。文件按日期轮转，并同时受 `retentionDays` 与 `maxAgeDays` 中较短的期限、以及 `maxBackups` 数量限制；超过 `maxSizeMB` 也会提前轮转，`compress: true` 会压缩已轮转文件。进程退出时会 flush 日志。日志不得包含密码、token、Cookie、Authorization header、私钥、完整环境变量或不受限个人数据。
