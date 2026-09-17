# 测试环境记录

本文档记录当前测试环境的可复用事实和验收入口。密码、Token、私钥、完整配置值和数据库备份位置不写入本文档；一次性部署日志和临时故障过程不在此维护。

## 项目与工具链

- 项目根目录：`/Users/cookun/GolandProjects/sechelper-auth-template`（当前开发工作站）。
- 后端：Go `1.26.8`，声明来源为 `project.yaml`；模块语言版本以 `api/go.mod` 为准。
- 前端：Node.js `24.21.0`、npm `11.19.1`、React `19.2.7`、Vite `7.1.7`，声明来源为 `project.yaml`。
- 前端依赖安装：在 `web/` 或 `web/admin/` 下执行 `npm ci`，必须使用对应锁文件。
- 远程测试机工具链：Go `1.26.8`（`/usr/local/go/bin/go`）、Node.js `24.21.0`、npm `11.19.1`；实际版本必须与 `project.yaml` 一致。
- 本地编译、测试缓存、二进制和临时构建产物必须位于操作系统临时目录，不得写入仓库持久目录。

## 运行组件与边界

- 统一构建入口：`make build ENV=<development|test|production>`。
- 统一运行入口：`make run ENV=<development|test|production>`；迁移使用 `ACTION=migrate`。
- 本地开发 Compose 项目：`sechelper-auth-template-dev`，配置为 `deploy/compose.dev.yaml`。
- 测试远程服务：由 Linux 主机上的 `order-test.service` 管理，工作目录为 `/opt/sechelper-auth-template`；服务用户为 `order-test`。
- 测试基础设施 Compose：`deploy/compose.test.yaml`，只运行项目专用 PostgreSQL/Redis；测试 API、迁移器、前台和管理端由远程物理机原生工具链编译并由 systemd 管理。生产业务 Compose：`deploy/compose.prod.yaml`；环境名通过统一运行入口显式传入，不维护项目环境文件。
- 远程迁移注意：当前远程主机仍有历史 `sechelper-auth-template-dev-postgres-1` 容器；本次仅更新仓库编排规范，未重建或迁移远程数据库/Redis。切换到 `compose.test.yaml` 属于单独的测试基础设施迁移，需要明确授权后执行。
- PostgreSQL、Redis、Mock Identity Provider 仅通过 Compose 运行；本次远程测试服务使用专用 PostgreSQL 容器 `sechelper-auth-template-dev-postgres-1`，容器端口映射为主机 `127.0.0.1:5432`。远程应用未配置 Redis 地址，授权缓存使用进程内实现。
- 本地开发 Compose 暴露 API `127.0.0.1:8080`、Web `127.0.0.1:8088`、PostgreSQL `127.0.0.1:5432` 和 Redis `127.0.0.1:6379`。
- 测试域名：`https://order-test.sechelper.com`；前台和 API 共用该 Origin。
- 健康检查：`GET https://order-test.sechelper.com/healthz`；依赖就绪检查：`GET https://order-test.sechelper.com/readyz`；版本信息：`GET https://order-test.sechelper.com/v1/version`。安装页由 Nginx 通过内部服务端守卫控制，不提供公开安装状态接口。

## 配置与数据

- 服务配置通过 `--config` 指向配置文件，并由 `api/internal/platform/config` 统一加载和校验。
- 配置样例：`config.example.yaml`、`.env.example`。
- 远程配置文件：`/opt/sechelper-auth-template/config.yaml`；远程 Secret 文件：`/opt/sechelper-auth-template/.env`，权限为 `0600`，由 systemd `EnvironmentFile=` 加载。`DATABASE_URL`、`REDIS_URL`、`CONFIG_CENTER_ENCRYPTION_KEY` 和 `CONFIG_CENTER_INSTALL_KEY` 只从该 Secret 文件注入，配置中心不保存这些敏感配置。
- 远程配置中心数据库：本项目专用数据库 `auth_template_test`；bootstrap 数据库连接、加密密钥和安装密钥从远程 `.env` 注入，配置中心保存其余框架与业务 Key。
- 迁移入口：`make db-migrate`；迁移程序使用 `api/migrations/`，通过 `schema_migrations` 校验版本和 checksum，并使用数据库 advisory lock 串行执行。
- 当前框架迁移：`001_authentication.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql`、`005_audit_event_model.sql`、`006_platform_users.sql`、`007_authentication_oidc_logout.sql`、`008_configuration_center.sql`。
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

## 远程环境

- SSH 目标：`root@47.116.4.57`；环境：`test`；操作系统/架构：Linux `x86_64`。
- 远程项目路径：`/opt/sechelper-auth-template`。
- systemd 单元：`order-test.service`；工作目录 `/opt/sechelper-auth-template`；运行用户 `order-test`；重启命令：`systemctl restart order-test.service`。
- API 二进制：`/opt/sechelper-auth-template/bin/auth-template`；迁移器：`/opt/sechelper-auth-template/bin/auth-template-migrate`；迁移目录：`/opt/sechelper-auth-template/migrations`。
- 前台静态目录：`/opt/sechelper-auth-template/web/dist`；管理端静态目录：`/opt/sechelper-auth-template/web/admin/dist`；日志目录：`/opt/sechelper-auth-template/logs`。
- 远程 Nginx vhost：`/etc/nginx/sites-enabled/order-test.sechelper.com.conf`；管理端 `/admin/`、`/admin/assets/` 和 SPA 路由使用 `/opt/sechelper-auth-template/web/admin/dist` 的独立映射。
- 远程数据库容器：`sechelper-auth-template-dev-postgres-1`；应用数据库：`auth_template_test`；健康状态以 `docker ps` 和 PostgreSQL healthcheck 为准。
- 共享统一登录数据库所在的 `25432` 端口不属于本项目测试数据库，项目重置和迁移不得操作该资源。
- 回滚方式：先停止服务，再从部署备份恢复对应的二进制、静态目录、`config.yaml` 或数据库备份，最后重新执行迁移兼容性检查并启动服务；备份位置由每次部署任务单独记录，不写入本文档。

## 当前验证记录

- 最近验证时间：`2026-09-17 12:41:46 +08:00`。
- 远程服务状态：`order-test.service` active；`/healthz` 返回 `{"status":"ok"}`；`/readyz` 返回 `{"status":"ready"}`。
- 当前发布标识：应用版本 `0.1.2`，Build ID `20260917001357`，部署环境 `test`；源码标记为当前工作区构建，未伪装为 Git 提交。最新构建在远程物理机完成。
- 安装与配置状态：项目数据库已应用 7 条迁移，配置中心记录数为 35；敏感连接配置已迁移到远程 `.env`，首次安装锁和配置中心安装状态按当前运行环境维护。
- 浏览器验收：Nginx `/admin/` 静态入口和管理端资源均返回 200；完成首次安装后，进入后台仍由服务端校验 `admin:access`，未授予该权限的账号会返回 403，不能通过前端绕过。配置中心写入后需重启 `order-test.service` 才会让数据库、Identity、Session、日志等进程级配置生效。
- 本地回归：Go 全量测试、Web 测试、Admin 测试和 `make architecture-check` 已通过；远程构建使用主机 Go `1.26.8`、Node.js `24.21.0`、npm `11.19.1`。
- 已知限制：`v1/version` 的源码标记为 `WORKTREE_UNCOMMITTED`；下一次正式发布应使用干净且已提交的源码快照。
