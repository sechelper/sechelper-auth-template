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
- `web/src/app/**`
- `web/src/platform/**`
- `web/admin/src/app/**`
- `web/admin/src/platform/**`

如确需修改，必须先获得用户明确授权，新增 ADR、更新接口合同、补框架级测试，并由 Framework Owner 审查。不得为了适配业务而复制、绕过或放宽框架规则。

## 允许的业务扩展路径

- `api/internal/business/<module>/**`
- `api/migrations/business/**`
- `web/src/business/<module>/**`（业务前台）
- `web/admin/src/business/<module>/**`（管理后台）
- `docs/contracts/business/**`
- `examples/**`

业务模块必须通过唯一后端注册表，以及各自唯一的前台/后台注册表接入：前台只能登记到 `web/src/app/public-business-modules.js`，后台只能登记到 `web/admin/src/app/admin-business-modules.js`。前后台业务代码不得互相导入；后台路由必须位于 `/admin/` 并声明权限，前台路由不得进入 `/admin`。业务模块使用注入的 HTTP、事务、审计、授权、日志和配置接口，不得读取环境变量、创建全局连接、直接操作 Session Store 或自定义错误协议。

## AI 停止条件

当任务要求触及受保护路径、改变认证/授权语义、改变公共响应格式或绕过契约测试时，AI 必须停止实现，说明触及的规则和所需授权；不得通过修改框架代码“先让业务跑起来”。

## 合并门禁

CI 必须检查：框架到业务无反向依赖、业务无直接环境读取、路由均登记 OpenAPI、权限均登记 Manifest、默认构建不引用 `examples`，并且业务模块拥有契约测试。
