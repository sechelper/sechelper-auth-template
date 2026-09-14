# Admin Web

独立的 React 管理端入口，生产环境部署在 `/admin`，与 `web/` 前台共用同一个 Go API 服务、HttpOnly Session Cookie 和授权体系。

管理端入口要求 `admin:access`。订单、Manifest 和会话页面分别通过业务权限控制；服务端仍是最终授权边界。

管理端统一提供 Sneat 风格掌机错误状态页：没有 `admin:access` 时显示 403，未知管理端路径显示 404，认证服务异常显示 500。三类页面共用掌机外壳、屏幕内俄罗斯方块演示、分数/等级/行数数据区、方向键、A/B/X/Y 动作键和“选择/开始”系统键；403 使用 warning 色掌机与数字，404 使用 primary 色，500 使用 danger 色。页面均提供返回概览、返回应用入口或重试操作。

后台视觉使用仓库内 `docs/ai-design-reference/sneat-bootstrap-free/` 的 Sneat Bootstrap FREE 参考：浅灰蓝画布、白色浮层卡片、固定侧边菜单、顶部导航、紫色主操作色和 Bootstrap 栅格密度。当前 `/admin` 首页已重做为 Command Center：欢迎操作区、运行概览、常用操作、系统状态和最近订单；订单摘要复用 `GET /v1/orders`，权限与会话数据仍来自服务端。实现样式集中在 `src/style.css`，不加载第三方 Bootstrap CSS。

左上角品牌锁定由 `src/platform/brand/CommunityBrand.jsx` 提供，使用 `public/assets/logo/logo.png` 中的 system-design 原始助安社区 Logo，显示“助安社区 - 模版演示”和 `SECHELPER COMMUNITY` 两行文字，并链接回应用入口。系统名来自 `globalThis.__APP_CONFIG__.systemName`，未注入时使用模板演示默认值；生产部署可通过运行时配置替换，不需要修改组件。
