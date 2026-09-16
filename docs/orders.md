# 订单示例模块

订单是业务接入示例，不属于 Framework Core。默认 Go 构建不加载它；测试环境可使用 `example` 构建标签显式启用后端订单权限、资源与路由，并同时启用后台的 `admin-module.js` 声明。

## API

- `GET /v1/orders`：需要 `order:read`，支持 `page[size]`，服务端限制最大 100。
- `GET /v1/orders/{orderId}`：需要 `order:read`。

## 数据

表结构由 `api/migrations/business/orders/002_orders.sql` 定义。金额使用 `total_minor` 整数和三位 `currency`，时间使用 UTC。

## 接入边界

订单后端位于 `api/internal/business/orders/`，订单管理页面位于 `web/admin/src/business/orders/`。带 `example` 标签的 `api/cmd/server/business_runtime_example.go` 负责构造示例模块并注册统一授权中间件，默认构建使用空业务运行时；订单 Handler 只依赖订单 Application Service。测试机启用前执行 `make example-orders-test`，普通构建继续执行 `make build`。后续真实业务模块应采用独立、明确的装配方案，不把示例订单硬编码进默认框架。
