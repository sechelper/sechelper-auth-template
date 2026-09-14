# 统一认证业务框架

本项目是只支持 Confidential Client 的 Go/Gin + React 业务接入框架。系统由 `/` 前台、`/admin` 管理端和一个统一 Go API 服务组成；业务系统通过后端会话接入统一认证平台，浏览器只持有 HttpOnly 会话 Cookie。

详细架构、认证边界、部署和 API 契约见 [`docs/`](docs/README.md)。
