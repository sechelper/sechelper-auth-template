# Orders 业务对接规范

本文是 `orders` 示例业务的唯一对接说明。该模块用于验证业务模块边界，当前迁移、后端、管理端路由和合同均为 `example` 能力，不得作为生产业务功能发布。

## 能力边界

后端提供只读订单查询：

- `GET /v1/orders?page%5Bsize%5D=<1..100>` 返回最多一页订单快照；不支持创建、修改、退款、导出、游标或下一页。
- `GET /v1/orders/{orderId}` 返回单个订单。
- 两个接口都要求服务端 Session、`admin:access` 和 `order:read`。前端菜单权限只控制展示，不能替代服务端授权。

管理端提供 `/admin/orders` 查询页，负责加载、空状态、失败重试、订单号本地筛选和状态筛选。它通过框架公开的认证 HTTP 适配器发送同源 Session 请求，不读取 Cookie、Token 或环境变量。

## 响应与字段

列表响应为 `{ data: Order[], meta: { count, pageSize } }`；详情响应为 `{ data: Order }`。错误响应统一为 `{ error: { code, message, requestId } }`，当前业务错误码为 `ORDER_NOT_FOUND` 和 `ORDER_QUERY_FAILED`。

`Order` 字段含义如下：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 订单唯一标识，调用详情接口时原样作为路径参数，并按 URL 规则编码。 |
| `status` | string | 服务端订单状态原值；展示端不得自行推断支付结果。 |
| `totalMinor` | int64 | 最小货币单位的非负整数，不能使用浮点数参与金额计算。 |
| `currency` | string | 三位货币代码，当前由数据库约束长度；展示时使用服务端值。 |
| `createdAt` / `updatedAt` | RFC 3339 date-time | 服务端输出 UTC 时间；客户端只负责格式化。 |

## 调用约定

业务调用方必须使用 `/v1/` 版本路径、携带浏览器 Session，并复用框架 HTTP 适配器处理 CSRF、统一错误和认证刷新。不得拼接或保存认证凭据，不得把 `admin:access` 或 `order:read` 当作前端可信授权。

`page[size]` 缺省为 20，最大为 100。响应中的 `meta.count` 是本次返回数量，`meta.pageSize` 是服务端实际采用的上限；由于当前合同没有分页游标，调用方不得依据 `count == pageSize` 推断还有下一页。

订单数据来自业务自有 `orders` 表，金额使用 `total_minor`，时间使用 `TIMESTAMPTZ`。业务代码不得读取其他模块表，也不得直接访问 Session、授权缓存、审计存储或框架配置。

## 可复用的框架能力

业务模块可以通过宿主注入的公开边界使用：

- Session 身份与当前账号信息：公共前台由 `/v1/auth/public/session`、`/v1/account/me` 提供；管理后台由对应 admin surface 接口提供；业务不接触 Cookie 或 Token。
- 服务端授权：模块注册权限、资源和受保护路由；最终决策由服务端执行。
- 标准 HTTP：统一请求 ID、错误封装和 JSON `data`/`meta` 响应形状。
- 审计与日志：通过宿主注册器/Recorder 写入，不自建审计表或日志协议。
- 事务、配置和基础设施：通过宿主注入接口获取，不读取进程环境、不创建全局连接。

## 不属于本模块的能力

认证登录、OIDC、Session 存储、权限 Manifest 同步、配置中心、审计查询、运维操作、健康检查和平台 Dashboard 都属于框架能力；订单模块只声明依赖，不重新实现这些能力。

## 验证

提交前至少运行 `make example-orders-test`、`make openapi-lint` 和 `make architecture-check`。示例迁移必须保持 `MODULE_KIND=example`，正式构建和生产迁移不得包含本模块。
