# Admin Web

独立的 React 管理端入口，源码位于 `web/admin/`，生产环境部署在 `/admin`。它与上级 `web/` 业务前端保持独立的 Vite 配置、依赖清单和构建流程，但共用同一个 Go API 服务、HttpOnly Session Cookie 和授权体系。

管理端入口要求 `admin:access`。订单、Manifest 和会话页面分别通过业务权限控制；服务端仍是最终授权边界。

管理端统一提供掌机风格错误状态页：没有 `admin:access` 时显示 403，未知管理端路径显示 404，认证服务异常显示 500。三类页面共用掌机外壳、屏幕内俄罗斯方块演示、分数/等级/行数数据区、方向键、A/B/X/Y 动作键和“选择/开始”系统键；403 使用 warning 色掌机与数字，404 使用 primary 色，500 使用 danger 色。页面均提供返回概览、返回应用入口或重试操作。

后台布局、颜色和组件样式由 `src/style.css` 维护；平台首页只展示认证、授权和运行状态，业务页面必须通过业务模块注册协议接入。管理端不加载第三方 Bootstrap CSS。

左上角品牌锁定由 `src/platform/brand/CommunityBrand.jsx` 提供，使用 `public/assets/logo/logo.png` 中的 system-design 原始助安社区 Logo，仅显示“助安社区”和 `SECHELPER COMMUNITY` 两行文字，并链接回应用入口。页面标题和系统名称来自服务端 `/v1/runtime-config` 的 `systemName`，其唯一来源是配置中心 `APP_NAME`；未加载配置前不伪造应用名称。
