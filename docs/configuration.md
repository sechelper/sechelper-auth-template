# 配置契约

数据库使用一个项目专用 PostgreSQL 数据库，并按职责隔离 schema：框架表位于 `framework`，配置中心表位于 `configuration`，业务模块表位于 `business_<module>`。配置中心的 bootstrap 数据库连接只负责建立初始连接，不能从尚未加载的配置中心读取。

应用配置由根目录 `config.yaml`（模板为 [`config.example.yaml`](../config.example.yaml)）定义，Go 服务通过独立的 Viper 实例在启动时读取一次、严格反序列化并完成语义校验。启动参数 `--config <path>` 优先级最高，其次是 `APP_CONFIG_FILE` 选择配置文件路径；配置文件同目录存在 `.env` 时会作为本地/测试敏感配置输入读取，显式进程环境变量优先。服务运行配置只来自 YAML、bootstrap 环境输入和配置中心。

配置优先级为：代码默认值 < `config.yaml` < `.env` 中的 bootstrap/数据库/Redis 连接配置 < 配置中心运行配置。配置中心在服务启动阶段加载并覆盖运行时框架配置；`.env` 不覆盖配置中心中的 OIDC、Session、日志、限流和业务配置。

应用发布版本不属于运行配置：`app.version` 和 `APP_VERSION` 不再受支持。升级现存部署配置时应删除 `app.version`；严格解码会拒绝未知字段。版本由根目录 `project.yaml` 在规范构建过程中注入。

`.env` 是部署环境的敏感配置文件：由 systemd `EnvironmentFile=` 或 Compose 外部注入，保存 `DATABASE_URL`、`REDIS_URL`、`CONFIG_CENTER_ENCRYPTION_KEY` 和 `CONFIG_CENTER_INSTALL_KEY`，真实文件不得进入 Git。`config.yaml` 只保存非敏感启动结构和默认值；Session、OIDC Client 和 Metrics 等运行敏感配置保存到配置中心。`bootstrap.installLockFile` 指向服务端安装锁文件，默认是 `logs/install.lock`。

## YAML 结构

当前分组为 `bootstrap`、`app`、`server`、`identity`、`manifest`、`session`、`rateLimit`、`cors`、`security` 和 `log`，键名使用 lower-camel-case；未知键会导致启动失败。服务监听地址由 `app.listenAddr` 配置，安装引导不再要求填写独立的 `SERVER_HOST` 或 `SERVER_PORT`；`server.readTimeout`、`server.writeTimeout`、`server.idleTimeout` 和 `server.shutdownTimeout` 由配置中心保存，配置文件仅保留非敏感启动结构。数据库连接由部署 `.env` 中的 `DATABASE_URL` 注入，Redis 连接由 `REDIS_URL` 注入，不再由配置中心保存或覆盖；`manifest` 仅配置同步周期；Manifest API 使用 `identity.clientId/clientSecret` 获取 access token。

`security.trustedProxyCidrs` 是唯一允许应用信任 `X-Forwarded-Host` 的来源网络；未命中该列表时，应用忽略转发主机并使用 TCP 请求地址中的 Host。`security.metricsToken` 在 production 必须在配置文件或配置中心中提供，`/metrics` 只接受对应的 Bearer Token。Host 必须匹配 `app.apiOrigin` 或 `app.publicWebOrigin`，Origin 必须匹配 CORS 白名单。

配置中心保存的键仍使用大写环境变量格式，便于部署引导和业务模块读取；这些名称只是配置中心的键名，不再表示进程环境变量。`DATABASE_URL` 和 `REDIS_URL` 是部署环境专属连接配置，不允许通过配置中心写入。`IDENTITY_SCOPES`、`ALLOWED_ORIGINS` 和 `TRUSTED_PROXY_CIDRS` 支持填写多个值，统一使用英文逗号分隔，例如 `openid,profile,email`、`https://example.com,https://admin.example.com` 和 `127.0.0.1/32,10.0.0.0/8`。Session 加密密钥必须是 RawStdEncoding 编码的 32 字节随机值；production 还要求 `session.secure: true`、非空 `REDIS_URL` 和 HTTPS 的 `identity.jwksUrl`。Origin 必须含 scheme、不得含路径、凭据或通配符。

日志字段配置位于 `log` 分组：`mode` 支持 `file` 和 `stdout`，`encoding` 固定为 `json`；`timeKey`、`levelKey`、`messageKey`、`callerKey`、`stacktraceKey` 控制 Zap 的标准字段名，时间编码为 `iso8601`，级别编码为 `lowercase`。`development`、`disableCaller`、`disableStacktrace` 和 `sampling` 只控制 Zap 的诊断行为，不能关闭脱敏、输出限制、轮转或保留策略；`debug` 是服务端诊断开关，不能由请求输入控制。

## 配置中心

配置中心由 `configuration.configuration_entries` 表和管理端 `/admin/configuration` 页面提供。键名必须是大写环境变量格式，例如 `IDENTITY_CLIENT_ID` 或业务模块自定义的 `PAYMENT_TIMEOUT_SECONDS`；禁止保存 `DATABASE_URL`、`REDIS_URL` 及其密码。值使用会话加密密钥进行 AES-GCM 加密后保存；敏感值默认开启，列表、详情和审计事件均不返回明文。

管理端读取配置需要 `configuration:read`，新增、替换和删除需要 `configuration:write`。业务配置列表只返回元数据；非敏感详情可返回明文供管理端回显，敏感详情始终返回空值，管理端只显示固定掩码。保存时未修改的敏感项不会将掩码提交回配置中心。修改会记录 `SECURITY_CONFIGURATION_CHANGED` 审计事件。配置中心使用启动阶段由环境注入的 bootstrap 数据库连接和加密密钥作为依赖，因此这两项不能依赖配置中心本身；安装密钥也只从部署环境注入。

服务启动时先使用配置文件中的非敏感启动结构和环境注入的 bootstrap 数据库连接读取并解密配置中心，再覆盖可运行时调整的框架配置并重新构建 Redis、身份客户端、Session、日志和 HTTP 依赖。数据库连接、配置中心加密密钥、安装密钥、应用版本、构建信息和配置文件路径不允许由配置中心改变。安装模式不使用独立的环境变量开关，而由环境注入的 bootstrap 安装密钥和配置中心安装标记共同决定。

配置中心值在进程启动时生效；涉及宿主基础设施（例如数据库连接池、Redis 客户端）的变更必须重启服务。若配置中心值无法解密或校验失败，服务拒绝启动，不使用不安全的部分配置继续运行。

业务模块通过宿主注入的 `application.ConfigurationProvider` 调用 `Get(ctx, key)` 读取值，不得读取环境变量、配置文件或配置中心数据库，也不得把配置中心接口暴露给浏览器。不存在的键和解密失败都按配置缺失处理，由业务决定是否使用安全默认值或拒绝启动。

相对 `log.output` 路径以服务二进制所在目录为基准，因此默认文件是二进制旁的 `logs/app.log`，日志目录权限为 `0700`、文件权限为 `0600`。文件按日期轮转，并同时受 `retentionDays` 与 `maxAgeDays` 中较短的期限、以及 `maxBackups` 数量限制；超过 `maxSizeMB` 也会提前轮转，`compress: true` 会压缩已轮转文件。进程退出时会 flush 日志。日志不得包含密码、token、Cookie、Authorization header、私钥、完整环境变量或不受限个人数据。
