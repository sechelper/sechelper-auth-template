# 订单示例模块

`api/internal/modules/orders/` 是框架的第一个业务接入示例。它只包含订单领域模型、查询用例、PostgreSQL Repository、HTTP DTO 和权限声明，不读取认证平台配置，也不直接访问 Session Store。

## API

- `GET /v1/orders`：需要 `order:read`，支持 `page[size]`，服务端限制最大 100。
- `GET /v1/orders/{orderId}`：需要 `order:read`。

## 数据

表结构由 `api/migrations/002_orders.sql` 定义。金额使用 `total_minor` 整数和三位 `currency`，时间使用 UTC。

## 接入边界

订单路由在组合根中注册统一授权中间件，但订单 Handler 只依赖订单 Application Service。后续新增库存或报表模块应遵循相同结构，不修改认证和授权内部实现。
