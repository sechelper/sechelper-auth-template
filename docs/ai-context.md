# AI 交接说明

## 入口

- Go 组合根：`api/cmd/server/main.go`
- 配置：`api/internal/platform/config/`
- 统一认证客户端：`api/internal/platform/identity/`
- 会话：`api/internal/platform/session/`
- 认证用例：`api/internal/modules/authentication/application/`
- 认证 HTTP：`api/internal/modules/authentication/transport/http/`
- Manifest：`api/internal/modules/manifest/`
- React 认证壳：`web/src/modules/auth/`
- OpenAPI：`docs/contracts/openapi.yaml`

## 重要限制

本项目只支持 Confidential Client。不要在前端加入 Client Secret、Token Exchange 或 Token 持久化。新增业务必须作为独立模块实现，并通过显式权限注册接入。

## 默认视觉规范

除非当前需求明确指定其他样式，`web/` 与 `admin/` 的新页面、新组件和受影响区域默认采用 [`docs/development-standards.md`](development-standards.md) 规定的 Sneat Bootstrap FREE 设计参考。实现前阅读 [`docs/ai-design-reference/sneat-bootstrap-free/`](ai-design-reference/sneat-bootstrap-free/)，优先复用其中的设计令牌、布局模式和组件状态；不得因此引入完整 Bootstrap/Sneat 运行时或把业务逻辑写入参考目录。
