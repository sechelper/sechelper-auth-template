# ADR 001：只支持 Confidential Client

## 决策

业务框架只支持 Confidential Client。Client Secret、access token 和 refresh token 全部由 Go 后端管理，React 只通过 HttpOnly 会话 Cookie访问业务 API。

## 原因

- 统一认证平台已经提供服务端 Token 兑换能力；
- 业务系统主要是内部或受控业务应用；
- 避免浏览器持久化 Token 的 XSS 风险；
- 将 Token 刷新、撤销、授权失效和审计集中在服务端。

## 影响

纯前端 Public Client 不属于当前框架支持范围。任何需要浏览器直接使用 Token 的业务必须另建经过安全评审的客户端方案。
