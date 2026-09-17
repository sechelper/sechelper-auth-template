# 配置中心接入指南

本文档是业务模块接入配置中心的操作说明。配置中心的存储、权限和生命周期约束以 [`configuration.md`](configuration.md) 为准。

## 适用场景

配置中心适合保存业务模块需要的运行参数，例如：

- `PAYMENT_PROVIDER_URL`
- `PAYMENT_TIMEOUT_SECONDS`
- `PAYMENT_CACHE_NAMESPACE`
- `ORDER_EXPORT_BUCKET`

Key 必须使用大写环境变量格式：首字符为大写字母，后续只能使用大写字母、数字和下划线，最长 128 个字符。建议使用业务前缀，避免不同模块之间发生命名冲突。

不要把用户密码、访问令牌、Session Cookie、客户端 Token 或完整数据库 DSN 拼入日志、API 响应或前端代码。敏感值在管理端默认按密文处理；业务模块读取到的是服务端内存中的明文，只能在服务端使用。

## 管理员配置步骤

1. 执行数据库迁移，确认 `configuration_entries` 已创建。
2. 为管理员授予 `configuration:read`，如需写入再授予 `configuration:write`。
3. 打开 `/admin/configuration`。
4. 输入 Key、Value 和说明，敏感凭据保持“敏感值”选中。
5. 保存后检查版本号和更新时间，并在操作审计中确认 `SECURITY_CONFIGURATION_CHANGED`。

敏感值保存后，列表和详情不会回显明文。编辑敏感值时必须重新输入完整的新值；留空不会保留旧值，也不会触发更新。

## 后端模块接入

业务模块通过 `application.ConfigurationAware` 接收宿主提供的 `application.ConfigurationProvider`。模块不得读取 `os.Getenv`、配置 YAML、Viper 或配置中心表。

```go
type Module struct {
    config application.ConfigurationProvider
}

func (m *Module) SetConfiguration(provider application.ConfigurationProvider) {
    m.config = provider
}

func (m *Module) timeoutSeconds(ctx context.Context) (int, error) {
    raw, ok := m.config.Get(ctx, "PAYMENT_TIMEOUT_SECONDS")
    if !ok {
        return 15, nil // 仅使用明确、安全的业务默认值
    }
    value, err := strconv.Atoi(raw)
    if err != nil || value < 1 || value > 300 {
        return 0, errors.New("invalid PAYMENT_TIMEOUT_SECONDS")
    }
    return value, nil
}
```

配置读取的失败策略由业务模块负责，但必须满足以下要求：

- 关键配置缺失或格式错误时，在业务操作前返回明确的依赖错误，不能静默使用危险值；
- 连接 URL、密码和令牌不得写入日志或审计 metadata；
- 配置 key 固定为常量，不接受浏览器请求直接指定任意 key；
- 需要启动时校验的配置，在模块构造或第一次受控初始化时校验，并使用 bounded timeout；
- 模块测试使用内存 fake `ConfigurationProvider`，不连接配置中心数据库。

业务模块若需要配置值来创建数据库、Redis 或外部客户端，应在宿主构造阶段显式读取并注入 typed dependency，而不是在业务 handler 中反复读取。例如：

```go
func New(clientFactory ClientFactory, config application.ConfigurationProvider) (*Service, error) {
    endpoint, ok := config.Get(context.Background(), "PAYMENT_PROVIDER_URL")
    if !ok || endpoint == "" {
        return nil, errors.New("PAYMENT_PROVIDER_URL is required")
    }
    client, err := clientFactory.New(endpoint)
    if err != nil {
        return nil, err
    }
    return &Service{client: client}, nil
}
```

当前宿主在业务模块注册后调用 `SetConfiguration`。如果模块需要在构造函数阶段读取配置，应将读取所需的 typed factory 或初始化步骤放到宿主装配流程中，并保持业务模块不直接访问存储。

## 生效与回滚

配置中心依赖启动阶段注入的 PostgreSQL bootstrap 连接和独立的配置中心加密密钥。数据库连接本身、配置中心加密密钥和安装密钥由部署环境的敏感变量提供，不能依赖同一个尚未建立的配置中心。

框架启动流程会先从 `config.yaml` 读取非敏感启动结构，并从部署环境注入 bootstrap 数据库连接和密钥，读取并解密配置中心，再覆盖支持的 Redis、Identity、Session、日志和其他框架运行参数，并重新构建依赖。数据库连接、bootstrap 密钥、安装密钥、构建元数据和配置文件路径不允许由配置中心覆盖。

普通业务配置通过 provider 读取时可在运行期间取得最新值。数据库连接池、Redis 客户端等已经创建的基础设施不会因为后台修改自动重建；修改这类配置后应执行受控重启，并先验证新值，再下线旧实例。

回滚配置时应通过管理端写回上一版本的值或删除错误的 Key，并保留审计记录。不要直接修改或删除已应用的 `008_configuration_center.sql` 迁移。
