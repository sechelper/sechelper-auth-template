# 统一认证业务框架架构

## 当前范围

本框架仅支持 Confidential Client。Client Secret、access token 和 refresh token 只允许存在服务端；浏览器通过 HttpOnly、Secure、SameSite Cookie 持有业务会话。

框架提供认证、会话、Application 授权、通用资源访问决策、Manifest、Account/Session、Audit、Operations、配置、日志、HTTP 基础设施和部署基线。用户和角色由统一认证平台管理；本项目只消费身份与权限，并由各业务模块执行自己的资源范围策略。业务模块位于 `api/internal/modules/<module>/`，不直接依赖统一认证客户端或平台存储。当前 `orders` 是用于验证接入边界的只读示例模块。

## 依赖方向

```text
transport -> application -> domain
                         -> ports
persistence -> application ports
platform identity client -> authentication application
```

`api/cmd/server` 只负责加载并校验配置、构建依赖、注册模块和启动进程。

## 当前实现边界

第一阶段已建立认证登录、回调、会话查询、登出、refresh token 刷新、AES-GCM 密文存储、PKCE 生成、CSRF 双提交校验、PostgreSQL Session Store、授权上下文、授权缓存、权限中间件、Manifest 远程同步和平台 Dashboard。共享授权缓存和远端 Token 撤销仍属于后续实现。

## 前后台入口与管理端

项目使用一个 Go API 服务和两个独立的 React 前端入口：`web/` 只负责 `/` 前台，`admin/` 只负责 `/admin` 后台。两个入口共用同一域名、HttpOnly Session Cookie、CSRF 机制和 `/v1/` API，但拥有独立的构建产物、布局、导航和页面边界。

403、404、500 为全局前端错误页：业务端 `web/` 支持 `/403`、`/404`、`/500` 及未知前台路径，后台 `admin/` 根据权限、路径和认证服务状态渲染对应页面。两端统一复用 `shared/error-pages/GlobalErrorPage.jsx` 和 `shared/error-pages/error-pages.css`，避免维护两套错误页。后端仍返回标准 HTTP 状态码和 JSON 错误响应，前端错误页仅负责展示；两端使用一致的仿真掌机视觉和状态色。

后台通过 `GET /v1/authorization/me` 获取当前 Application 权限，必须具备 `admin:access` 才能进入管理端。平台 Dashboard 通过 `GET /v1/admin/dashboard/overview` 只展示应用、认证、授权和依赖状态，不读取业务数据。订单 API 要求 `admin:access` 与 `order:read`，Manifest 查看要求 `admin:access` 与 `auth:manifest:read`，Manifest 同步额外要求 `auth:manifest:sync`。服务端同步 Manifest 时使用 Identity Confidential Client 的 access token，不使用独立 Manifest 凭据。前端权限控制只负责导航和操作展示，服务端 API 继续通过授权中间件执行最终校验。

生产静态目录为 `/usr/share/nginx/html` 和 `/usr/share/nginx/html/admin`。Nginx 将 `/admin/` 子路径回退到 `/admin/index.html`，将其他前端路径回退到根 `index.html`，保证两个 SPA 的深层路径刷新不会串台。后台专用 API 预留在 `/v1/admin/`，不与浏览器页面路径 `/admin` 混用。

后台视觉基线采用仓库内 `docs/ai-design-reference/sneat-bootstrap-free/`：浅灰蓝页面画布、白色固定菜单与顶部浮层、`#696cff` 主色、低强度阴影、紧凑卡片和表格密度。后台样式集中在 `admin/src/style.css`，不引入第二套 Bootstrap CSS 运行时依赖。

管理端左上角使用 `admin/src/platform/brand/CommunityBrand.jsx` 的社区品牌锁定：图标资源位于 `admin/public/assets/logo/logo.svg`，上方显示 `助安社区 - 模版演示`，下方显示 `SECHELPER COMMUNITY`。业务系统名通过 `globalThis.__APP_CONFIG__.systemName` 注入，缺省值仅用于本模板演示；完整锁定名称提供一个返回应用入口的可访问链接。
