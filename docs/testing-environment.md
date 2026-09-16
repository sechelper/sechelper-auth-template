# 测试环境记录

本文档记录当前仓库可由代码、配置和构建入口确认的测试环境事实。一次性部署日志、备份目录、远程命令输出和历史浏览器验收记录不在此维护；如需远程部署，应在执行前根据当前部署系统重新确认目标主机、服务和凭据。

## 项目与工具链

- 项目根目录：`/Users/cookun/GolandProjects/sechelper-auth-template`（当前开发工作站）。
- 后端：Go `1.26.8`，声明来源为 `toolchain.versions`；模块语言版本以 `api/go.mod` 为准。
- 前端：Node.js `24.21.0`、npm `11.19.1`、React `19.2.7`、Vite `7.1.7`，声明来源为 `toolchain.versions`。
- 前端依赖安装：在 `web/` 或 `web/admin/` 下执行 `npm ci`，必须使用对应锁文件。
- 本地编译、测试缓存、二进制和临时构建产物必须位于操作系统临时目录，不得写入仓库持久目录。

## 运行组件与边界

- 统一构建入口：`make build ENV=<development|test|production>`。
- 统一运行入口：`make run ENV=<development|test|production>`；迁移使用 `ACTION=migrate`。
- 本地开发 Compose 项目：`sechelper-auth-template-dev`，配置为 `deploy/compose.dev.yaml`。
- 测试 Compose 项目：`sechelper-auth-template-test`，配置为 `deploy/compose.prod.yaml` 与 `deploy/environments/test.env`。
- 生产 Compose 项目：`sechelper-auth-template`，配置为 `deploy/compose.prod.yaml` 与 `deploy/environments/production.env`。
- PostgreSQL、Redis、Mock Identity Provider 仅通过 Compose 运行；测试资源使用项目专属容器、网络和命名卷。
- 本地开发 Compose 暴露 API `127.0.0.1:8080`、Web `127.0.0.1:8088`、PostgreSQL `127.0.0.1:5432` 和 Redis `127.0.0.1:6379`。
- 健康检查：`GET /healthz`；依赖就绪检查：`GET /readyz`；版本信息：`GET /v1/version`。

## 配置与数据

- 服务配置通过 `--config` 指向配置文件，并由 `api/internal/platform/config` 统一加载和校验。
- 配置样例：`config.example.yaml`、`deploy/config.dev.yaml`、`.env.example`。
- 迁移入口：`make db-migrate`；迁移程序使用 `api/migrations/`，通过 `schema_migrations` 校验版本和 checksum，并使用数据库 advisory lock 串行执行。
- 当前框架迁移：`001_authentication.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql`、`005_audit_event_model.sql`、`006_platform_users.sql`、`007_authentication_oidc_logout.sql`。
- 当前业务迁移：`api/migrations/business/orders/002_orders.sql`。该迁移由示例业务显式启用，生产迁移默认拒绝示例迁移。
- 测试数据和示例订单数据不得写入生产数据库；本仓库没有可自动导入生产环境的种子数据。

## 可复用验证命令

在仓库根目录执行：

```text
make toolchain-check
make architecture-check
make test-go
make test-web
make test-admin
make example-orders-test
make openapi-lint
```

`make test` 会组合执行 Go、Web、Admin 和订单示例测试。`make openapi-lint` 校验主合同 `docs/contracts/openapi.yaml` 与订单业务合同 `docs/contracts/business/orders/openapi.yaml`。

测试构建使用 `ENV=test`，输出目录由 `BUILD_OUTPUT_DIR` 指向操作系统临时目录；测试 API 构建显式使用 `example` build tag，生产构建不使用该 tag。生产构建还会检查示例业务路由和示例迁移没有进入正式产物。

## 业务示例隔离

- 后端订单示例位于 `api/internal/business/orders/`，通过 `api/cmd/server/business_runtime_example.go` 和 `example` build tag 装配。
- 管理端订单示例位于 `web/admin/src/business/orders/`；组件参考页位于 `web/admin/src/business/component-reference/`。
- 前台资料演示位于 `web/src/business/profile-example/`。
- `example-module.js` 和 `test-module.js` 只允许进入开发或测试模块图；正式前端构建必须排除它们。
- 以上资产是测试和业务接入验证的一部分，不应作为无效代码删除；清理时应只移除经代码和构建检查确认的重复或失效引用。

## 远程环境说明

当前仓库文件没有声明固定远程主机、SSH 用户、systemd 服务名或测试域名；这些属于部署系统的外部运行状态，不在本记录中复制历史值。执行远程测试或部署前，应从当前授权的部署配置和目标主机重新确认：项目路径、服务单元、Compose 项目、数据库/Redis 资源、配置指纹和回滚位置。不得把凭据、令牌、完整 `.env` 内容或备份路径写入本文件。

## 当前验证记录

- 验证基准：当前工作树（尚未提交的工作区改动不伪装为 Git 提交）。
- 2026-09-16：`make toolchain-check`、`make architecture-check`、`make openapi-lint` 和 `git diff --check` 通过；OpenAPI 合同无解析错误，Redocly 保留 6 条语义性建议（健康/会话/版本端点缺少 4xx，以及登录/回调重定向端点缺少 2xx）。
- 最近一次记录应在完成代码或合同改动后更新，包含实际执行的命令、结果和仍待处理的问题。
- 仅保留最近一次可复现验证结果；更早的部署流水账、备份目录和临时故障说明应放在外部发布系统或任务交接中，而不是继续追加到本文件。
