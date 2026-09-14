# 平台 Dashboard

## 定位

后台首页 `/admin` 是认证业务框架的平台总览，不是具体业务数据看板。它只展示应用运行信息、认证状态、授权 Manifest 状态和基础设施健康状态，不依赖 `orders` 或其他具体业务模块。

订单、财务、用户运营等业务指标应由各自业务模块提供独立页面和接口，例如 `/admin/orders/overview` 和 `/v1/orders/dashboard`。

## API

`GET /v1/admin/dashboard/overview` 需要当前 Session 和 `admin:access` 权限，返回以下数据：

- `application`：应用名称、版本、运行环境和启动时间；
- `currentSession`：当前管理员的认证上下文；概览页同时从当前前端会话展示 Email、有效期和全部权限；
- `dependencies`：API、PostgreSQL、Redis 和 Manifest 的健康状态；
- `manifest`：当前 Manifest 版本、状态、Content Hash 和服务端修订号。

概览页还集中承载访问控制诊断表单（调用 `POST /v1/admin/access-decisions/check`）和系统状态接口（调用 `GET /v1/admin/operations/overview`）的依赖卡、应用版本和运行指标。当前会话、访问控制诊断和系统状态不再作为独立后台导航入口。

依赖状态包括 `healthy`、`degraded`、`unavailable`、`not_configured` 和 `not_synced`。依赖探测使用短超时，单个依赖失败不会导致 Dashboard 整体失败。

## 前端边界

Dashboard 页面位于 `admin/src/modules/dashboard/`。管理台外壳、导航和权限守卫位于 `admin/src/app/`；订单、Manifest、活跃会话和审计页面分别由自己的业务模块负责。

概览页将当前会话、系统状态、访问控制诊断和 Manifest 状态集中展示；系统状态由概览接口与运行状态接口共同提供。平台 Dashboard 可以跳转到业务模块，但不能读取或聚合业务模块数据库，也不能展示订单数量、订单金额、最近订单等业务指标。
