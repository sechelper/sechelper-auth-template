# AI 与二开边界规则

本文件是业务二开和 AI 修改本仓库时的强制规则。规则优先级高于业务便利性和生成代码习惯。

## 受保护的 Framework Core

以下路径属于框架核心，业务需求和 AI 默认不得修改：

- `api/internal/platform/**`
- `api/internal/modules/authentication/**`
- `api/internal/modules/authorization/**`
- `api/internal/modules/manifest/**`
- `api/internal/modules/audit/**`
- `api/internal/modules/account/**`
- `api/internal/modules/operations/**`
- `api/cmd/server/main.go`
- `api/migrations/001_authentication.sql`
- `api/migrations/003_authentication_transactions.sql`
- `api/migrations/004_audit_events.sql`
- `api/migrations/005_audit_event_model.sql`
- `web/src/framework/**`
- `web/src/main.jsx`
- `web/admin/src/app/**`
- `web/admin/src/platform/**`

如确需修改，必须先获得用户明确授权，新增 ADR、更新接口合同、补框架级测试，并由 Framework Owner 审查。不得为了适配业务而复制、绕过或放宽框架规则。

## 允许的业务扩展路径

- `api/internal/business/<module>/**`
- `api/migrations/business/**`
- `web/src/business/<module>/**`（业务前台）
- `web/admin/src/business/<module>/**`（管理后台）
- `docs/contracts/business/**`

业务 API、认证会话、用户资料、统一身份中心设置、配置中心和框架能力的接入细则以 [`docs/business-api-integration.md`](business-api-integration.md) 为准。AI 或业务开发不得通过直接读取 Cookie、Token、Session Store、环境变量、配置中心表或身份平台 UserInfo 来绕过该规范。

业务模块必须通过唯一后端注册边界和前端自描述文件接入。业务前台在 `web/src/business/<module>/public-module.js` 声明路由，管理后台在 `web/admin/src/business/<module>/admin-module.js` 声明路由、导航和权限；框架注册器自动发现这些固定文件，普通业务任务不得修改注册器、应用入口或全局路由。只有改变声明协议或发现机制时才提交框架扩展申请。前后台业务代码不得互相导入；后台路由必须位于 `/admin/` 并声明权限，前台路由不得进入 `/admin`。业务模块使用注入的 HTTP、事务、审计、授权、日志和配置接口，不得读取环境变量、创建全局连接、直接操作 Session Store 或自定义错误协议。

用户身份及个人资料由统一认证平台作为唯一权威来源保存。业务模块不得将用户信息复制或持久化到业务数据库；如界面需要，可在浏览器端缓存展示所需信息，并在适当时机从统一认证平台刷新。浏览器缓存仅用于展示，不得作为认证、授权或用户信息权威性的依据。

## AI 停止条件

当任务要求触及受保护路径、改变认证/授权语义、改变公共响应格式或绕过契约测试时，AI 必须停止实现，说明触及的规则和所需授权；不得通过修改框架代码“先让业务跑起来”。

## 正式环境隔离

- 前后端模板、示例、测试样例、测试专用接口，以及测试数据库、测试数据库对象和测试数据只允许用于开发或测试环境，绝不允许编译、打包、迁移、初始化或部署到正式环境。
- 正式构建与发布配置必须显式排除上述模板和测试资源；正式数据库迁移与初始化不得包含测试用 schema、表、数据或接口所需的测试数据。不得依赖“生产环境不会调用”作为保留测试接口或测试资源的理由。
- 若无法确认某项资源是否会进入正式产物或正式数据库，AI 必须停止发布相关变更并报告该资源及其影响，确认隔离方式后再继续。

## 合并门禁

CI 必须检查：框架到业务无反向依赖、业务无直接环境读取、路由均登记 OpenAPI、权限均登记 Manifest、默认构建不引用 `examples`，并且业务模块拥有契约测试。
