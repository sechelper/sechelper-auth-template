# 后台平台模块

后台平台模块由 Account/Session、Permission/Manifest、Audit 和 Operations 组成。它们服务于管理台本身，不拥有订单等业务数据。

## Account / Session

- `GET /v1/admin/account`：读取当前管理员的 Application 上下文。

`authentication` 模块和 PostgreSQL Store 继续管理浏览器认证会话及其登录、刷新、登出和到期生命周期；管理后台不提供列出或撤销会话的管理入口。Account 模块仅返回当前管理员的授权上下文。

## Permission / Manifest

`GET /v1/admin/permissions` 展示当前代码注册的权限和 API 绑定，Manifest 状态和同步接口仍由 `manifest` 模块负责。权限分配的权威来源仍是统一认证平台，本项目不重复实现本地角色管理。

## Audit

`audit_events` 只记录关键认证、会话、授权、资源、Manifest 和高风险运维操作，不记录普通 HTTP 请求、页面访问、Dashboard 加载或健康检查。事件包含事件类型、分类、严重级别、结果、原因、Subject、Application Code、资源引用和 Request ID。审计记录不得包含 Client Secret、Access Token、Refresh Token、Cookie 或数据库密码；查询接口支持按事件类型、分类、严重级别、结果、操作者、资源和时间范围筛选。

## Operations

`GET /v1/admin/operations/overview` 展示应用版本、API/数据库状态和运行指标，当前由平台概览页集中呈现。基础设施探针仍使用 `/healthz` 和 `/readyz`；该接口不替代容器健康检查。

## Resource Access

- `GET /v1/admin/resources`：查看业务模块注册的资源类型、动作和所需 Application 权限；
- `POST /v1/admin/access-decisions/check`：解释当前管理员对一个已注册资源动作的授权决策。

资源注册由 `authorization` 模块提供，具体资源策略由业务模块实现。访问控制诊断表单已迁移到平台概览，仍只解释当前管理员的授权决策，不返回资源内容，也不绕过业务 API。

## 前端共享能力

`web/admin/src/shared/ui.jsx` 提供 `DataTable`、`EmptyState` 和 `RefreshButton`。页面需要自行处理加载、空数据、失败、重试和 mutation 状态，但统一使用这些基础组件保持一致体验。

管理端左侧导航采用 Frest Demo 1 的展开式菜单规格：固定 `260px` 宽度、`75px` 品牌区、`42px` 菜单行、分组标题、线性图标、淡蓝活动态和独立滚动区；桌面端支持收起为 `80px` 图标栏。导航内容、可见性和权限仍由 `web/admin/src/app/navigation.jsx` 与服务端会话权限控制，左上角继续使用 `CommunityBrand` 提供的 SECHELPER COMMUNITY 品牌锁定。

顶部导航覆盖为参考页的 `62px` Frest 顶栏规格：与侧栏同色背景、左侧搜索入口、右侧语言/主题/快捷入口/通知图标和用户头像状态点；用户头像菜单承载刷新会话与退出操作，避免为复刻视觉而移除现有会话能力。

左上角品牌锁定由 `web/admin/src/platform/brand/CommunityBrand.jsx` 提供，使用 `web/admin/public/assets/logo/logo.png` 中的 system-design 原始助安社区 Logo，显示“助安社区 - 模版演示”和 `SECHELPER COMMUNITY` 两行文字，并链接回应用入口。系统名来自 `globalThis.__APP_CONFIG__.systemName`，未注入时使用模板演示默认值。
