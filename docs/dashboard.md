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
- `resources`：总览加载时附带的最新服务器资源快照，包含 Linux 主机 CPU、内存、API 工作目录所在文件系统用量，以及当前 Go 进程 Goroutines 数。CPU 使用 `/proc/stat` 相邻采样差计算，主机内存使用 `MemTotal - MemAvailable`，磁盘空间取工作目录所在文件系统。单项不可采集时该项字段为 `null`；采集异常不会导致总览接口失败。接口不返回文件系统路径或业务数据。

`GET /v1/admin/dashboard/resources` 使用相同的 Session 与 `admin:access` 权限，仅返回最新资源快照，不探测依赖或查询业务数据。服务端以 500ms 周期采样；页面同样每 500ms 请求该轻量接口，并在浏览器保留最近 60 个不同采样点（约 30 秒）绘制 CPU、内存、磁盘和 Goroutines 趋势。资源接口临时失败时，页面保留最近一次成功快照与曲线并自动重试；首次采样尚未就绪时显示等待状态。概览页的访问控制诊断表单调用 `POST /v1/admin/access-decisions/check`。页面不聚合业务指标。

依赖状态包括 `healthy`、`degraded`、`unavailable`、`not_configured` 和 `not_synced`。依赖探测使用短超时，单个依赖失败不会导致 Dashboard 整体失败。

## 前端边界

Dashboard 页面位于 `web/admin/src/modules/dashboard/`。管理台外壳、导航和权限守卫位于 `web/admin/src/app/`；Manifest、活跃会话和审计页面属于框架模块，订单等项目页面位于 `web/admin/src/business/<module>/`。

概览页将应用信息、系统状态、访问控制诊断和 Manifest 状态集中展示；所有展示的读取数据来自平台接口。平台 Dashboard 可以跳转到业务模块，但不能读取或聚合业务模块数据库，也不能展示订单数量、订单金额、最近订单等业务指标。
