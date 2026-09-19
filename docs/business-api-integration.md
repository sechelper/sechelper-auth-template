# 业务 API 接入规范

本文档是业务模块接入本框架的执行规范。它描述业务模块如何使用统一 API、认证会话、用户资料、权限、配置中心和框架能力；不替代主 OpenAPI 合同，也不授权业务模块修改框架核心。

业务接入必须作为一个纵向切片交付：后端模块、数据库迁移、业务 API 合同、前台/后台页面、权限声明、审计事件、测试和模块文档使用同一个 `<module>` 名称，并在同一变更中保持一致。

## 1. 接入边界

业务模块的目录必须与业务能力同名：

```text
api/internal/business/<module>/
api/migrations/business/<module>/
web/src/business/<module>/                 # 有公共前台时创建
web/admin/src/business/<module>/            # 有管理后台时创建
docs/contracts/business/<module>/
docs/business/<module>/                    # 业务说明需要时创建
```

业务模块只能通过宿主公开的注册协议接入。后端模块实现 `application.BusinessModule`，提供：

- `Name()`：稳定且唯一的模块名；
- `RegisterPermissions()`：注册权限码、风险级别和 API 绑定；
- `RegisterResources()`：注册资源类型、动作和资源级策略；
- `RegisterAuditEvents()`：注册模块使用的审计事件类型；
- `RegisterRoutes()`：在宿主传入的 `/v1` 路由组中注册路由，并为每条受保护路由绑定授权中间件。

需要配置时，实现 `application.ConfigurationAware`，通过 `SetConfiguration(application.ConfigurationProvider)` 接收配置提供者。配置提供者只提供 `Get(ctx, key)`；业务模块不得读取环境变量、YAML、Viper、配置中心表或框架内部包。

公共前台和管理后台通过固定的 `public-module.js`、`admin-module.js` 自描述文件接入。不得修改应用入口、全局路由、Provider、框架注册器或把公共前台代码导入管理后台。

## 2. API 合同与 URL 规则

### 2.1 路径和命名

- 所有业务 API 使用 `/v1/` 前缀，例如 `/v1/resources`、`/v1/resources/{resourceId}`。
- 管理端业务 API 使用 `/v1/admin/<resource>`；浏览器页面路径使用 `/admin/...`，两者不能混用。
- URL 使用资源名和 camelCase 参数名；路径参数应表达资源标识，例如 `{resourceId}`，不要使用数据库列名或内部表名。
- JSON 字段使用 camelCase；Go 的数据库模型、内部 DTO 不得直接作为 API 响应。
- 时间统一使用 UTC RFC3339；金额使用最小货币单位整数和明确的货币代码，不使用浮点数。
- API 合同先写入 `docs/contracts/business/<module>/openapi.yaml`，再实现 Handler。路由、参数、响应、权限、错误码和分页行为必须都能在合同中找到。

### 2.2 响应封装

成功响应使用统一封装：

```json
{
  "data": { "id": "..." },
  "meta": {}
}
```

集合响应也使用 `data` 和 `meta`：

```json
{
  "data": [{ "id": "..." }],
  "meta": {
    "hasMore": false,
    "nextCursor": null
  }
}
```

`meta` 中只放分页、版本、重试或其他协议元数据，不把业务对象重复放入 `meta`。没有元数据时仍保持响应结构稳定，不要让调用方在对象和封装对象之间猜测。

### 2.3 错误响应

错误使用稳定的机器可读错误码，不能把 Go 错误文本、SQL 错误或堆栈直接返回浏览器：

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "资源不存在",
    "details": {}
  }
}
```

最低状态码约定如下：

| 状态码 | 使用场景 |
| --- | --- |
| `400` | 请求结构、参数或业务输入无效 |
| `401` | 没有有效业务会话 |
| `403` | 已登录但缺少权限，或资源范围策略拒绝 |
| `404` | 资源不存在；不得泄露不应确认的资源存在性 |
| `409` | 版本冲突、幂等冲突或状态不允许 |
| `422` | 结构合法但业务规则无法接受 |
| `429` | 触发限流，必要时返回 `Retry-After` |
| `500` | 未预期的服务端错误 |
| `502/503` | 外部依赖或服务暂不可用 |

业务错误码必须稳定、带模块命名空间或明确业务语义，例如 `RESOURCE_NOT_FOUND`、`RESOURCE_STATE_CONFLICT`。前端依据 `error.code` 决定展示、重试或跳转，不匹配后端自然语言。

### 2.4 查询、分页和并发

- 列表必须限制默认和最大 page size；排序字段、方向、筛选字段和筛选值范围必须白名单化。
- 优先使用游标分页；若使用 `page[size]`，必须明确快照边界和 `hasMore` 语义，不能无限制返回全表。
- 分页、排序和筛选参数必须在业务合同中声明，并对非法值返回稳定的 `400` 错误。
- 写操作如可能被重复提交，应定义幂等键、唯一约束或版本条件，并明确 `409` 行为。
- 需要乐观并发控制时，在合同中声明版本字段或 `If-Match` 语义；不要在 Handler 中静默覆盖其他写入。

## 3. 用户登录、会话和用户资料

### 3.1 业务页面如何使用登录

浏览器只使用框架提供的 `AuthProvider`/`useAuth()` 能力，不直接拼接 OIDC URL，不读取 Cookie、Client Secret、access token 或 refresh token。

框架暴露的主要能力：

| 能力 | 业务使用方式 |
| --- | --- |
| 当前会话 | 读取 `state.status`、`state.authenticated`、`state.expiresAt` 等状态 |
| 当前用户 | 公共前台读取 `GET /v1/account/me`；管理后台读取 `GET /v1/admin/account` |
| 登录 | 调用 `login()`；需要静默检查时调用 `silentLogin()`（公共前台） |
| 刷新页面状态 | 调用 `refresh()`，重新读取会话和用户资料 |
| 刷新服务端会话 | 调用 `refreshSession()`，由框架完成 CSRF 和 Refresh Rotation |
| 退出 | 调用 `logout()`，不得自行删除 Cookie 或调用身份平台 |
| 用户设置 | 使用 `userCenterURL()` 生成统一身份中心链接 |
| 后台权限展示 | 使用 `hasPermission(code)` 控制导航和操作展示，但不能代替服务端鉴权 |

业务页面应处理 `loading`、`unauthenticated`、`authenticated`、`refreshing`、`error` 和 `reauthentication_required` 状态。未登录时显示业务自己的登录入口；业务页面不得把未登录状态当作服务端授权结论。

### 3.2 框架认证 API

| API | 业务接入规则 |
| --- | --- |
| `GET /v1/auth/public/session` / `GET /v1/auth/admin/session` | 只用于判断对应 surface 的本地会话是否有效；不返回业务用户资料 |
| `GET /v1/account/me` / `GET /v1/admin/account` | 登录后读取对应 surface 的当前用户展示资料；业务不得直接调用身份平台 UserInfo |
| `GET /v1/admin/authorization/me` | 管理端读取 admin surface 的 Application 权限；仅用于导航和交互展示 |
| `POST /v1/auth/public/refresh` / `POST /v1/auth/admin/refresh` | 由对应框架认证客户端调用；需要对应的 `X-CSRF-Token`，业务不实现 Refresh Token 轮换 |
| `POST /v1/auth/public/logout` / `POST /v1/auth/admin/logout` | 只撤销对应 surface 的本地会话；远端撤销失败不阻断本地退出 |
| `GET /v1/auth/public/login` / `GET /v1/auth/admin/login` | 由对应 Provider 通过 `login()` 调用；只允许 `prompt=none` 或 `prompt=login` |

登录回调、state、nonce、PKCE、ID Token 校验、Session Cookie、Refresh Token 密文保存和退出状态均属于框架职责。业务 API 只依赖“当前请求是否已经通过宿主授权中间件”这一结果。

### 3.3 设置、个人资料和退出

本应用不复制保存统一身份平台的密码、MFA、登录设备或个人设置。页面中的“设置”链接必须指向 `userCenterURL()` 返回的统一身份中心用户中心。

退出时必须调用 Provider 的 `logout()`：

1. 框架服务端撤销本地 Session 和 Refresh Token；
2. 框架按配置尝试远端 Token 撤销；
3. 浏览器清理本地认证状态，并回到应用当前约定页面；
4. 业务页面通过 Provider 状态变化清空业务内存状态并停止受保护请求。

业务模块不得自行实现退出按钮的 Cookie 清理、远端 logout URL、LocalStorage Token、Refresh Token 或“退出后自动登录”。

## 4. 认证授权与业务资源

每条需要登录的业务路由都必须通过宿主传入的 `RouteAuthorizer` 声明权限。例如：

```go
group.GET("", authorizer.RequirePermissions("admin:access", "resource:read"), handler.List)
group.GET("/:resourceId", authorizer.RequirePermissions("admin:access", "resource:read"), handler.Get)
```

业务模块必须同时做到：

- 权限码使用业务命名空间，如 `resource:read`、`resource:write`；
- `RegisterPermissions` 中声明权限和实际 API 绑定；
- 高风险写操作拆分独立权限，不用一个宽泛权限覆盖所有动作；
- 服务端始终是最终权限裁决者，前端权限只控制导航、按钮和交互；
- 资源所有权、租户、项目、状态和数据范围在 use case 中再次判断，不能只依赖权限中间件；
- 未登录返回 `401`，无权限或资源范围拒绝返回 `403`，授权失败默认拒绝。

跨模块访问必须通过公开用例、端口或事件合同完成，不得读取其他模块的内部包或数据库表。

## 5. 配置中心接入

### 5.1 业务配置的范围

配置中心适合保存业务运行参数，例如：

```text
PAYMENT_PROVIDER_URL
PAYMENT_TIMEOUT_SECONDS
RESOURCE_EXPORT_BUCKET
```

Key 必须符合 `[A-Z][A-Z0-9_]{0,127}`，建议使用模块前缀。业务模块不得声明或覆盖以下框架启动依赖：

- `DATABASE_URL`、`REDIS_URL`；
- 配置中心加密密钥和安装密钥；
- 应用版本、Build ID、配置文件路径；
- Client Secret、Session 加密密钥、Metrics Token 等框架敏感配置，除非由框架维护流程明确批准。

### 5.2 后端读取方式

实现 `ConfigurationAware`，保存宿主注入的 Provider，并使用固定 key 读取：

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
        return 15, nil
    }
    value, err := strconv.Atoi(raw)
    if err != nil || value < 1 || value > 300 {
        return 0, errors.New("invalid PAYMENT_TIMEOUT_SECONDS")
    }
    return value, nil
}
```

规则：

- key 必须是代码中的常量，不能让浏览器请求传入任意 key；
- 密码、Token、连接串和密钥按敏感值保存，绝不写入日志、审计 metadata 或 API 响应；
- 缺失或格式错误的关键配置必须返回明确依赖错误或阻止受控初始化，不能静默使用危险值；
- 外部客户端、连接池和高成本依赖应在宿主装配阶段读取并构造 typed dependency，不要在每个 Handler 中重复创建；
- 业务测试使用 fake `ConfigurationProvider`，不得连接真实配置中心数据库；
- 配置中心值不向浏览器暴露。前端需要的非敏感公开配置只能使用 `GET /v1/runtime-config` 提供的框架白名单字段。

### 5.3 管理端配置

管理员通过 `/admin/configuration/business` 管理业务配置，需要 `admin:access` 和 `configuration:read`；新增、替换、删除需要 `configuration:write`。敏感值保存后不可回显，修改敏感值必须重新输入完整新值。

普通业务配置可在运行期间由 Provider 读取最新值；数据库连接池、Redis 客户端等已经创建的基础设施不会自动重建。修改这类配置必须按发布流程验证并受控重启。所有配置变更都应保留 `SECURITY_CONFIGURATION_CHANGED` 审计记录。

## 6. 框架向业务提供的能力

| 能力 | 接入方式 | 业务边界 |
| --- | --- | --- |
| 统一认证会话 | 前端 `AuthProvider`；服务端授权中间件 | 不读 Cookie、Token、Session Store |
| 当前用户资料 | `GET /v1/account/me`、`GET /v1/admin/account` / `useAuth().user` | 不复制身份平台用户主数据到业务表 |
| Application 权限 | `RouteAuthorizer`、`GET /v1/admin/authorization/me` | 服务端权威；前端只做展示控制 |
| 资源级授权 | `RegisterResources` + 业务 use case 策略 | 业务负责租户、所有权、状态和数据范围 |
| 配置中心 | `ConfigurationProvider.Get(ctx, key)` | 不读环境变量、配置文件或配置中心表 |
| API 错误协议 | 宿主 HTTP 错误封装 | 不自建另一套公共错误格式 |
| 审计注册 | `RegisterAuditEvents` 和宿主审计 Recorder | 只记录关键安全/业务动作，不记录普通读取 |
| 数据迁移 | `api/migrations/business/<module>/` | 只新增前向迁移；迁移器自动写入 `business_<module>` schema，不得写入 `public` 或其他业务 schema |
| 日志与指标 | 使用宿主注入的日志/指标边界（如模块装配已提供） | 脱敏；不记录密码、Token、Cookie 和完整个人数据 |
| 数据库与事务 | 由宿主装配依赖并在 use case 中使用 | 事务覆盖同一业务写操作，不能创建全局连接 |
| 前端运行时配置 | `GET /v1/runtime-config` 的非敏感白名单 | 不把秘密或业务私密配置发给浏览器 |
| 管理后台框架 | `admin-module.js` 路由、导航、权限声明 | 路由必须 `/admin/`，并与服务端权限一致 |

如果某项能力缺少稳定的公开注入接口，业务模块不得直接导入框架内部实现“先接起来”。应先提交框架维护变更，明确接口、兼容性、测试和回滚方案。

## 7. 前端接入约束

业务前台位于 `web/src/business/<module>/`，管理后台位于 `web/admin/src/business/<module>/`。页面必须覆盖：加载、空数据、错误、重试、未授权、登录失效和写操作进行中状态。

业务 API 请求必须复用框架提供的请求客户端，以便自动携带 HttpOnly Session Cookie、保存响应中的 CSRF Token、统一解析错误和处理同源运行时配置。不得自行使用裸 `fetch` 实现认证、CSRF 或刷新逻辑；若业务确需独立客户端，必须保留这些框架行为并经过框架维护评审。

前端不得：

- 保存或解析 access token、refresh token、Client Secret 或 Session Cookie；
- 把前端权限判断当作数据安全边界；
- 把用户设置、密码、MFA 和登录设备复制成业务表单；
- 修改主入口、全局 Provider、全局路由或框架样式来适配单个业务模块；
- 在公共前台和管理后台之间互相导入源码。

## 8. 交付与验收

提交前必须确认：

1. 业务合同、Handler、权限绑定和前端调用的路径完全一致；
2. 所有路由均以 `/v1/` 开始，并有 OpenAPI 合同；
3. 所有受保护路由均有服务端权限；资源范围策略有越权测试；
4. 登录、退出、刷新、用户资料和设置均复用框架能力；
5. 配置读取只通过 `ConfigurationProvider`，敏感值不会泄露；
6. 写操作有事务、幂等/并发策略和关键审计事件；
7. 迁移只位于业务目录且只做前向追加；
8. Go 单元测试、前端测试、OpenAPI 校验、关键路径 E2E 和架构检查通过；
9. 正式构建不包含示例模块、测试接口、测试迁移或测试数据。

建议验证命令：

```bash
make fmt-check
make test
make openapi-lint
make architecture-check
```

环境、构建、合同和测试的最终要求以 [`docs/testing-environment.md`](testing-environment.md)、[`docs/development-workflow.md`](development-workflow.md) 和主 OpenAPI 合同为准。
