# 统一认证业务框架架构

授权框架的维护边界和变更流程以[授权框架维护](authorization-framework-maintenance.md)为准；本文档只保留系统级架构和组件边界。

## 当前范围

本框架仅支持 Confidential Client。Client Secret、access token 和 refresh token 只允许存在服务端；浏览器通过 HttpOnly、Secure、SameSite Cookie 持有业务会话。

框架提供认证、会话、Application 授权、通用资源访问决策、Manifest、Account/Session、Audit、Operations、配置、日志、HTTP 基础设施和部署基线。用户和角色由统一认证平台管理；本项目只消费身份与权限，并由各业务模块执行自己的资源范围策略。后端业务模块位于 `api/internal/business/<module>/`，业务前台位于 `web/src/business/<module>/`，业务管理后台位于 `web/admin/src/business/<module>/`；业务代码不直接依赖统一认证客户端或平台存储。当前 `orders` 是用于验证接入边界的只读示例模块。

## 依赖方向

```text
transport -> application -> domain
                         -> ports
persistence -> application ports
platform identity client -> authentication application
```

后端业务通过 `api/internal/application.Registry` 接入宿主，不由框架反向导入业务实现。模块向宿主注册权限、资源定义、审计事件类型和受保护路由；每个业务路由必须在模块边界声明服务端授权要求。框架提供统一的审计 Recorder 与 HTTP 错误封装，业务不自建公共错误协议或绕过审计端口。

开发、测试、生产共享 `make build`、`make run`、配置载入包和 API/迁移程序入口；`ENV` 只选择构建配置，不创建另一套启动路径。生产 Go 构建不带 `example` build tag，订单示例代码因 build constraints 不进入编译单元；Admin 生产构建只发现 `admin-module.js`，不会把 `test-module.js` 或 `example-module.js` 纳入模块图。生产镜像还检查组件参考、订单案例路由及示例迁移是否泄漏。Go 测试文件 `_test.go` 只由测试命令编译，不属于服务二进制。

`api/cmd/server` 只负责加载并校验配置、构建依赖、注册模块和启动进程。

## 当前实现边界

第一阶段已建立认证登录、回调、会话查询、登出、refresh token 刷新、AES-GCM 密文存储、PKCE 生成、CSRF 双提交校验、PostgreSQL Session Store、授权上下文、授权缓存、权限中间件、Manifest 远程同步和平台 Dashboard。共享授权缓存和远端 Token 撤销仍属于后续实现。

## PostgreSQL schema 隔离

每个环境使用一个项目专用 PostgreSQL 数据库，但数据库对象按职责放入独立 schema：框架对象使用 `framework`，配置中心使用 `configuration`，业务模块使用 `business_<module>`（例如 `business_orders`）。运行时 SQL 必须显式使用 schema-qualified 名称，不依赖连接默认的 `search_path` 来实现边界。

迁移器在数据库中创建这些 schema，并在每个迁移事务内只将当前迁移的目标 schema 放入临时 `search_path`。迁移账本统一存放在 `framework.schema_migrations`，通过 `(scope, version)` 唯一标识，避免不同 schema 的迁移版本相互冲突。配置中心仍依赖 bootstrap 数据库连接启动，但其数据只写入 `configuration.configuration_entries`。

业务数据库重置是单向受限操作：唯一脚本只能根据模块名生成 `business_<module>` schema，并在执行前验证目标数据库存在 `framework` 基线。配置中心 schema、框架 schema、`public` schema 和数据库级销毁操作均被脚本和架构检查拒绝。

## 前后台入口与管理端

项目使用一个 Go API 服务和两个独立的 React/Vite 前端项目：`web/` 负责 `/` 前台，`web/admin/` 负责 `/admin` 后台。两个项目共用同一域名、HttpOnly Session Cookie、CSRF 机制和 `/v1/` API，但拥有独立的入口、依赖、构建配置、构建产物、布局和导航边界。公共前台明确分为 `web/src/framework/` 与 `web/src/business/`：框架目录负责应用壳、认证、运行时配置、基础样式、全局错误和业务路由调度。正式首页必须由真实业务模块提供，框架不提供正式首页回退。当前 `profile-example` 是开发/测试演示模块，不是真实业务，也不参与正式路由发现；当前没有真实业务首页，因此正式构建会失败关闭，直到真实业务模块声明 `/`。项目业务页面只能位于 `web/src/business/<module>/`。

同一业务能力在三端使用相同的 `<module>` 名称：后端为 `api/internal/business/<module>/`，业务前台为 `web/src/business/<module>/`，管理前台为 `web/admin/src/business/<module>/`。某端没有该业务的实现时可以不创建对应目录，但不得使用不同模块名或把业务源码放入框架目录。当前 `orders` 有后端和管理前台，没有公共前台页面。

前端业务模块采用自描述和构建期自动发现。真实业务前台模块通过 `web/src/business/<module>/public-module.js` 自描述路由，其中最多一个启用模块可以声明 `/`；正式构建只发现这类真实业务声明，并要求恰有一个业务首页。开发/测试演示模块使用 `example-module.js`，只在 Vite development 或显式 `test` 模式下于根路由无真实业务页面时动态载入，不参与正式路由发现或首页检查。管理后台模块通过 `web/admin/src/business/<module>/admin-module.js` 声明路由、导航、排序和权限。框架注册器使用固定的单层目录模式发现声明；新增、删除或调整真实业务模块不修改应用入口、全局路由或框架注册器。管理端在显示导航和渲染直达路由时都会检查声明的权限，服务端仍执行最终授权。

管理后台页面壳在主内容左上方统一显示当前菜单路径。面包屑由框架根据当前路由递归匹配基础菜单与业务模块导航生成，不包含同级分组标题；嵌套业务菜单显示从业务父项到当前路由的完整路径。业务页面不得各自重复实现该导航路径展示。

管理后台导航由固定框架菜单和业务模块声明合并。`概览`、`账号与安全`、`监控与审计`及其框架页面由框架持有；业务模块只能增加或移除自己的导航声明，不得占用固定分组或框架路由。业务分组通过导航项的 `sectionOrder` 配置位置，数值越小越靠前；未设置时排在固定分组之后。固定框架分组使用框架定义的顺序，不因业务分组的增删或排序改变。框架从合并后的菜单按当前路由递归生成页面路径，并自动展开包含当前页面的分组和父节点。

管理后台模块导航声明中的 `section` 标识分组，`sectionOrder` 控制分组间顺序，`sectionCollapsible` 控制该分组是否可折叠，`order` 控制模块导航项的稳定顺序。同一分组在多个模块声明时必须使用相同的顺序和折叠设置。业务分组可通过调整 `sectionOrder` 放到框架分组之前、之间或之后，不需要修改后台布局或菜单渲染逻辑。

403、404、500 为全局前端错误页：业务端 `web/` 支持 `/403`、`/404`、`/500` 及未知前台路径，后台 `web/admin/` 根据权限、路径和认证服务状态渲染对应页面。共享错误页位于 `web/shared/error-pages/`，两个 Vite 项目各自编译使用；后端仍返回标准 HTTP 状态码和 JSON 错误响应，前端错误页仅负责展示。

后台通过 `GET /v1/authorization/me` 获取当前 Application 权限，必须具备 `admin:access` 才能进入管理端。平台 Dashboard 通过 `GET /v1/admin/dashboard/overview` 只展示应用、认证、授权和依赖状态，不读取业务数据。订单 API 要求 `admin:access` 与 `order:read`，Manifest 查看要求 `admin:access` 与 `auth:manifest:read`，Manifest 同步额外要求 `auth:manifest:sync`。服务端同步 Manifest 时使用 Identity Confidential Client 的 access token，不使用独立 Manifest 凭据。前端权限控制只负责导航和操作展示，服务端 API 继续通过授权中间件执行最终校验。

生产静态目录为 `/usr/share/nginx/html` 和 `/usr/share/nginx/html/admin`。Nginx 将 `/admin/` 子路径回退到 `/admin/index.html`，将其他前端路径回退到根 `index.html`，保证两个 SPA 的深层路径刷新不会串台。后台专用 API 预留在 `/v1/admin/`，不与浏览器页面路径 `/admin` 混用。

后台布局与视觉样式由 `web/admin/src/style.css` 维护，组件遵循项目现有的布局、样式令牌和无障碍规范；不引入第二套 Bootstrap CSS 运行时依赖。

管理端左上角使用 `web/admin/src/platform/brand/CommunityBrand.jsx` 的社区品牌锁定：图标资源位于 `web/admin/public/assets/logo/logo.svg`，下方显示 `SECHELPER COMMUNITY`。业务系统名由服务端配置中心解析后通过 `/v1/runtime-config` 注入；完整锁定名称提供一个返回应用入口的可访问链接。
