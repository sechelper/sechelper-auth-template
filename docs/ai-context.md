# AI 交接说明

## 入口

- Go 组合根：`api/cmd/server/main.go`
- 配置：`api/internal/platform/config/`
- 配置中心：`api/internal/modules/configuration/`；业务模块使用宿主注入的 `application.ConfigurationProvider`
- 统一认证客户端：`api/internal/platform/identity/`
- 会话：`api/internal/platform/session/`
- 认证用例：`api/internal/modules/authentication/application/`
- 认证 HTTP：`api/internal/modules/authentication/transport/http/`
- Manifest：`api/internal/modules/manifest/`
- React 前台框架：`web/src/framework/`，其中认证壳位于 `web/src/framework/auth/`
- OpenAPI：`docs/contracts/openapi.yaml`
- 授权框架维护：[`docs/authorization-framework-maintenance.md`](authorization-framework-maintenance.md)

## 重要限制

本项目只支持 Confidential Client。不要在前端加入 Client Secret、Token Exchange 或 Token 持久化。新增业务必须作为独立模块实现，并通过显式权限注册接入。

## 默认视觉规范

`web/` 与 `web/admin/` 的新页面和改造区域遵守 [`docs/development-standards.md`](development-standards.md)，优先复用项目现有样式令牌、组件与布局约定。具体产品设计或当前需求另有明确要求时，以该要求为准；视觉实现不得改变认证、授权和 API 边界。
