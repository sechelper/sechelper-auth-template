# 平台 Dashboard

## 定位

后台首页 `/admin` 是认证业务框架的平台总览，不是具体业务数据看板。它只展示应用运行信息、认证状态、授权 Manifest 状态和基础设施健康状态，不依赖任何具体业务模块。

业务指标应由各自业务模块提供独立页面和接口，不由平台 Dashboard 聚合。

## API

`GET /v1/admin/dashboard/overview` 需要当前 Session 和 `admin:access` 权限，返回以下数据：

- `application`：应用名称、版本、运行环境和启动时间；
- `currentSession`：当前管理员的认证上下文；概览页同时从当前前端会话展示 Email、有效期和全部权限；
- `dependencies`：API、PostgreSQL、Redis 和 Manifest 的健康状态；
- `manifest`：当前 Manifest 版本、状态、Content Hash 和服务端修订号。
- `resources`：总览加载时附带的最新服务器资源快照，包含 Linux 主机 CPU、内存、API 工作目录所在文件系统用量，以及当前 Go 进程 Goroutines 数。CPU 使用 `/proc/stat` 相邻采样差计算，主机内存使用 `MemTotal - MemAvailable`，磁盘空间取工作目录所在文件系统。单项不可采集时该项字段为 `null`；采集异常不会导致总览接口失败。接口不返回文件系统路径或业务数据。

`GET /v1/admin/dashboard/resources/ws` 是使用相同 Session 与 `admin:access` 权限的 WebSocket 升级入口，生产环境通过 `wss://` 连接。连接建立后服务端按 500ms 采样节奏推送 `{ "data": <ServerResourceSnapshot> }`，不探测依赖或查询业务数据；页面不再轮询，也不保留原 HTTP 资源接口。页面在浏览器保留最近 60 个不同采样点（约 30 秒）绘制 CPU 与内存趋势。磁盘使用率和 Goroutines 展示最新数值，不绘制趋势线。连接临时断开时，页面保留最近一次成功快照与 CPU/内存曲线，并在 1 秒后自动重连；首次采样尚未就绪时不推送消息，页面显示等待状态。概览页的访问控制诊断表单调用 `POST /v1/admin/access-decisions/check`。页面不聚合业务指标。

依赖状态包括 `healthy`、`degraded`、`unavailable`、`not_configured` 和 `not_synced`。依赖探测使用短超时，单个依赖失败不会导致 Dashboard 整体失败。具备 `admin:access` 权限的管理员还会在概览页看到部署版本信息、Bootstrap 配置、配置中心和框架检查状态；这组卡片读取 `GET /v1/admin/deployment/guide`，加载失败时不影响概览的其他内容。

## 前端边界

Dashboard 页面位于 `web/admin/src/modules/dashboard/`。管理台外壳、导航和权限守卫位于 `web/admin/src/app/`；Manifest、活跃会话和审计页面属于框架模块，项目页面位于 `web/admin/src/business/<module>/`。

概览页将应用信息与系统状态、依赖服务合并在 APPLICATION 卡片中，直接展示运行时长、最近检查时间及各依赖的独立健康状态，不显示依赖健康数量汇总；服务器资源中的磁盘和 Goroutines 使用并列紧凑卡片展示当前数值与必要容量说明，不显示趋势线，并在窄屏下改为单列；CPU 与内存保留趋势图；访问控制诊断和 Manifest 状态继续集中展示。所有展示的读取数据来自平台接口。平台 Dashboard 可以跳转到业务模块，但不能读取或聚合业务模块数据库，也不能展示业务模块的业务指标。
