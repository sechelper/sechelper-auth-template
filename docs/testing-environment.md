# 远程测试环境

## 删除活跃会话管理功能重新部署（2026-09-16）

- 部署源：基于远端既有测试源码提交 `5cd064995034d4fab23f276d2d8d65f2e981933e` 与本次授权路径增量，快照标识 `snapshot-5cd0649-ce0bb31cc9bcc71a`；生成清单保存在备份 `artifact-manifest.json`。用 `scp` 逐文件同步授权范围内的 Go、Admin、OpenAPI 和平台说明文件；未传输 `.git/`、`.env`、`config.yaml`、密钥或构建产物。远端测试源码移除了 Admin Sessions page 与 Account session service 两个文件。
- 目标与范围：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`，服务 `order-test.service`。删除后台“活跃会话”菜单/页面、列出及撤销会话的 `/v1/admin/account/sessions` API、对应权限声明和 OpenAPI DTO；保留认证登录/刷新/登出、当前账号上下文、会话存储与自然过期，不执行数据库迁移。
- 验证与构建：远端 Go `1.26.8`、Node.js `24.21.0`（`/usr/local/bin/node`）、npm `11.19.1`（`/root/.local/bin/npm`）；`make release-check`、`make toolchain-check`、`make architecture-check`、`make test` 全部通过（Go 全量、Orders 示例、Web 8 项、Admin 17 项）。远端 Redocly CLI lint 在 npm registry 安装阶段停滞后停止；同一工作树本地 `make openapi-lint` 已通过，产生 10 条建议级既有警告。以远端原生 Go/Vite test 工具链构建 API、迁移器、Web 和 Admin，Vite 分别转换 39/54 个模块；`web/node_modules` 按锁文件安装，未修改依赖清单。Build ID 与应用版本维持 `20260915150445` / `0.1.0`。
- 部署与验收：2026-09-16 10:10 UTC，`order-test.service` active；`/admin/`、JS/CSS、`/healthz`、`/readyz` 均 HTTP 200。未认证读取旧 `/v1/admin/account/sessions` 返回 404；当前账号接口仍返回 401（预期未认证），`/v1/auth/session` 返回 200。`/v1/version` 返回环境 `test`、版本 `0.1.0`、Build ID `20260915150445` 和快照修订标识。重启后首次就绪探测短暂返回 502，随后恢复为 200；最终服务与就绪检查均正常。
- 备份与影响：`/opt/sechelper-auth-template/.deploy-backup-remove-active-sessions-20260916/` 保存部署前 API 二进制、完整 `web/dist`、被删除源码、所有同步源码及 artifact manifest；回滚可恢复这些文件后重启 `order-test.service`。未执行迁移或修改数据库/Redis/Nginx 配置，也未执行 Compose 生命周期操作；仅重启了测试业务服务。当前远端仍将 `/admin/sessions` 作为业务路由保留名单中的旧路径，以避免业务模块占用；该路径没有 Admin 菜单、页面或 API。

## 服务器资源正常状态提示文案移除部署（2026-09-16）

- 部署源：基于测试机上次同步状态，仅部署当前授权工作树的 `web/admin/src/modules/dashboard/DashboardPage.jsx`；SHA-256 `d20d74f67d42be0f33e151aef9c5a675c056e33480446f55f820d2a7a0e5bbe9`。远端原源码哈希与上一版部署记录一致；文件经 `scp` 上传至临时目录、核验后同步，未复制 `.git/`、凭据或配置。
- 目标与范围：`root@47.116.4.57:/opt/sechelper-auth-template`，`order-test.sechelper.com`。正常状态不再显示“500ms 实时刷新”；资源请求异常仍显示“正在重试”。只替换后台静态资源 `/opt/sechelper-auth-template/web/dist/admin`，沿用原 `build-info.json`。
- 验证：Node.js `24.21.0`（`/usr/local/bin/node`）、npm `11.19.1`（`/root/.local/bin/npm`）；工具链检查、`make architecture-check`、Admin 17 项测试通过。远端 Vite `7.1.7` test 构建成功，转换 55 个模块；临时构建输出已清理。构建使用远端已有锁定依赖，未改依赖或安装配置。
- 验收与回滚：2026-09-16 09:16 UTC，线上 `/admin/`、新 JS/CSS、`/healthz`、`/readyz` 均返回 HTTP 200，Admin 静态目录权限为 `755`，`order-test.service` active。备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-refresh-label-20260916/`，包含部署前 Dashboard 源码和 Admin 静态目录；可用其中 `admin-dist-live/` 恢复静态页面、用 `source/` 恢复源码。
- 影响范围：未重建或重启 API，未改数据库、Redis、Nginx、systemd、运行配置和业务数据；服务未重启。版本与 Build ID 保持不变。

## APPLICATION 运行信息分隔线移除与上移部署（2026-09-16）

- 部署源：基于测试机已同步源码提交 `5cd064995034d4fab23f276d2d8d65f2e981933e` 的当前授权工作树静态界面增量；`DashboardPage.jsx` SHA-256 `b2933a16dae4a93cb5d6838b0e6c37618ad4280f5e77116d57112c9975be29d2`，`style.css` SHA-256 `ac358eea74128b3f6e2e47df6aad5d827003cb363995747207c107edf55b1194`，`docs/dashboard.md` SHA-256 `084a0a096bf97e99bd82a62c3cac6a13091ebee24569682e9994eebb27a6e9f2`。经 `scp` 传输到远端临时目录并核对哈希后同步；未传输 `.git/`、凭据或配置。
- 目标与范围：`root@47.116.4.57:/opt/sechelper-auth-template`，`order-test.sechelper.com`。移除运行时长/最近检查上方重复分隔线并压缩留白，使信息上移；运行时长与最近检查时间仍根据应用启动时间和实际检查时间动态显示。同步两份 Admin 源码和 `docs/dashboard.md`，替换 `/opt/sechelper-auth-template/web/dist/admin`。
- 验证：远端 Node.js `24.21.0`（`/usr/local/bin/node`）、npm `11.19.1`（`/root/.local/bin/npm`）；工具链检查、架构检查通过，Admin 测试 17 项通过，Vite `7.1.7` test 构建转换 55 个模块成功。此次为未提交静态源码增量，未伪装为提交构建；在远端以 Vite test 模式构建至 `/tmp`，沿用原 `build-info.json` 中的版本及 Build ID，构建临时目录已清理。
- 验收与回滚：2026-09-16 09:05 UTC，线上 `/admin/`、新 JS/CSS、`/healthz`、`/readyz` 均返回 HTTP 200；Admin 静态目录权限为 `755`，`order-test.service` active。备份在 `/opt/sechelper-auth-template/.deploy-backup-admin-resource-layout-20260916/`，包括部署前源码和 Admin 静态目录；可用其中 `admin-dist-live/` 恢复页面，并用 `source/` 恢复源码。`/v1/version` API 源提交仍为 `2a5e49811fca873b18bee1ac98f6badf532fab92`，API 未重建；PostgreSQL、Redis、Nginx、systemd、运行配置和业务数据均未修改，也未重启服务。

## Manifest 信息排版与 APPLICATION 标题行移除部署（2026-09-16）

- 部署源：基于测试机已同步源码提交 `5cd064995034d4fab23f276d2d8d65f2e981933e`，仅部署当前授权工作树中的两份后台前端文件；`DashboardPage.jsx` SHA-256 `a61fc6bda0d67d497acc7794781919dd9d403d3e61cb0d5d1a0e0144a71b9cf1`，`style.css` SHA-256 `2fca9199e523b56332c19f3cebea126c4cf6a6dc0fe447e046cb99ffb2ecf513`。以 `scp` 传入远端临时目录后核验，再同步至 `/opt/sechelper-auth-template`；未复制 `.git/`、密钥或配置。
- 范围：删除 APPLICATION 内 `SYSTEM STATUS / 系统状态与依赖服务` 标题行；Manifest 版本、服务端修订和最近更新时间以对齐的语义化标签—数值行展示。只构建并替换 `/opt/sechelper-auth-template/web/dist/admin`。
- 验证：远端 `/root/.local/bin/npm` 11.19.1 与 Node.js 24.21.0；`make architecture-check` 通过，Admin 测试 17 项通过，Vite 7.1.7 test 构建转换 55 个模块成功。由于此次是已授权的未提交静态源码增量，未将其标记为干净提交构建；通过远端 Admin Vite 命令将产物生成在 `/tmp` 并直接更新静态资源，构建临时目录已清理。保留现有 `build-info.json`，应用版本及 Build ID 未变。
- 部署与验收：2026-09-16 08:53 UTC，Admin 页面、新 JS/CSS、`/healthz` 和 `/readyz` 均返回 HTTP 200；线上 bundle 含三个 Manifest 标签且不含已删除标题；`order-test.service` 为 active。切换后首次探测发现构建目录权限为 `700` 导致 Nginx 无法读取，已将 `/opt/sechelper-auth-template/web/dist/admin` 权限修正为 `755` 并复验成功。未改 Nginx 配置或 reload。
- 备份与影响：部署前源码和 Admin 静态目录备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-manifest-layout-20260916/`；回滚可用其中 `admin-dist-live/` 替换当前 Admin 静态目录，并恢复 `source/` 文件。API、PostgreSQL、Redis、systemd、运行配置、其他静态站点和业务数据未修改；服务未重启。`/v1/version` 仍报告 API 源提交 `2a5e49811fca873b18bee1ac98f6badf532fab92`，API 本次未重建。

## 平台总览 APPLICATION 与资源卡片布局部署（2026-09-16）

- 源码：本地工作站 `/Users/cookun/GolandProjects/sechelper-auth-template`，分支 `codex/full-reset-deploy-20260916`，提交 `5cd064995034d4fab23f276d2d8d65f2e981933e`。使用 `git archive` 生成不含 `.git/`、凭据、配置和依赖缓存的源码快照，经 `scp` 上传至测试主机；归档 SHA-256 为 `4ba57a500c38981273359c9b80a948fae76fc30859a4597de24bf53490152e88`。
- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`，服务 `order-test.service`。远端部署前 Dashboard、样式和页面文档源码均与已部署的 `2a5e49811fca873b18bee1ac98f6badf532fab92` 快照一致；未发现远端源码冲突。
- 构建与工具链：远端 Go `1.26.8`（`/usr/local/bin/go`）、Node.js `24.21.0`（`/usr/local/bin/node`）、npm `11.19.1`（`/root/.local/bin/npm`）。默认 PATH 中 `/usr/local/bin/npm` 为 `11.19.0`，不符合项目要求；本次显式将 `/root/.local/bin` 放在 PATH 前端。`make toolchain-check`、`make architecture-check` 通过；`npm ci --prefix web/admin --no-audit --no-fund` 后 Admin 测试 17 项通过；`WORKTREE_CLEAN=true CHECKED_OUT_REVISION=5cd064995034d4fab23f276d2d8d65f2e981933e make build ENV=test COMPONENT=web` 成功，Vite 前台/后台分别转换 39/55 个模块。npm 对锁定的 `esbuild@0.25.12` install script 给出 allowScripts 提示，未修改依赖配置。构建临时目录已清理。
- 发布范围：仅同步 `web/admin/src/modules/dashboard/DashboardPage.jsx`、`web/admin/src/style.css`、`docs/dashboard.md`，并替换 `/opt/sechelper-auth-template/web/dist/admin`。将系统状态与依赖状态并入 APPLICATION，移除依赖数量汇总；磁盘和 Goroutines 改为紧凑并列趋势卡片。备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-overview-5cd0649/`，含部署前源码、完整 Admin 静态目录及 `artifact-manifest.json`。回滚时用备份 `admin-dist-live/` 替换当前 `web/dist/admin/` 并恢复 `source/` 中三份文件；不需要重启服务或 reload Nginx。
- 验收：2026-09-16 08:41 UTC，线上 `/admin/`、本次 Admin JS/CSS、`/healthz`、`/readyz` 均返回 HTTP 200；新 bundle 包含 `SYSTEM STATUS`、`Goroutines`，不含旧“项正常”汇总文字；`order-test.service` 为 active。`/v1/version` 仍显示 API 源提交 `2a5e49811fca873b18bee1ac98f6badf532fab92`，因为本次仅更新后台静态资源，未重建或重启 API；应用版本和 Build ID 仍为 `0.1.0` 与 `20260915150445`。
- 影响范围：未重启业务服务，未修改 API、PostgreSQL、Redis、Nginx、systemd、运行配置或业务数据；未执行迁移或 Compose 生命周期命令。

## 全量测试环境重建部署（2026-09-16 02:51 Asia/Shanghai）

- 源码：本地分支 `codex/full-reset-deploy-20260916`，测试源提交 `ebceb01a9a0e68f213adbdb22fe9c4473e5ed64a`；由 `git archive` 打包并经 `scp` 上传，归档 SHA-256 `5b65f96ac07e79b484b6ddb923699b3bda2c7ef2edf0bc722cf4e7fd5d28d7b5`。快照不含 `.git/`、`.env`、`config.yaml`、依赖缓存或构建产物。远端源码已同步到该提交；清除两处旧源码目录。保留远端 `.env`、日志、bin/dist 旧版本回滚副本、IDE 状态、node_modules 缓存和 82 个历史 `.deploy-backup-*` 目录。
- 验证：远端 `make test`、`make architecture-check`、`make release-check`、`make toolchain-check` 全部通过。Go 全量测试及 Orders 示例测试通过；Web 8 项、Admin 17 项测试通过。`make build ENV=test WORKTREE_CLEAN=true CHECKED_OUT_REVISION=<完整提交 SHA>` 使用主机原生工具链成功构建，Vite test 构建分别转换 39/55 个模块；API/迁移器为静态 amd64 Go 可执行文件。manifest 验证了版本 `0.1.0`、Build ID `20260915150445`、提交 SHA `ebceb01a9a0e68f213adbdb22fe9c4473e5ed64a` 及各真实产物路径。临时构建目录已清理，manifest 副本保留在重建备份中。`npm ci` 对 `esbuild@0.25.12` 有安装脚本白名单提示，前后台构建仍成功。
- 工具链：主机 Go 1.26.8（`/usr/local/bin/go`，GOROOT `/usr/local/go`）、Node.js 24.21.0（`/usr/local/bin/node`）、npm 11.19.1（`/root/.local/bin/npm`，registry `registry.npmjs.org`）。缺失的 `rg` 14.1.0 安装于 `/root/.local/bin/rg`，来自主机 Ubuntu 24.04 配置的软件源。`/root/.profile` 已加入 `/root/.local/bin`，修改前备份为 `/root/.profile.codex-tools-backup-20260916`。
- 配置：按授权从远端 `/opt/sechelper-auth-template/config.yaml` 移除了已废弃的 `app.version` 单一键，保留 `app.name` 和其余配置。修改前版本以权限 600 保存于 `/opt/sechelper-auth-template-reset-backup-20260915T172015Z/config.yaml-before-app-version-removal`；运行时 `.env` 未变，systemd EnvironmentFile 与单元配置未变。
- 数据：仅重建项目专属容器 `sechelper-auth-template-dev-postgres-1` 中的 `auth_template`。成功应用 `001_authentication.sql`、测试业务 `002_orders.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql`、`005_audit_event_model.sql` 和 `006_platform_users.sql`；迁移账本 6 条、公共表 8 张。验收时 Session、登录事务、审计事件、用户映射及 Orders 均为 0；Manifest 启动同步状态为 1 条。浏览器验收触发的一条临时登录事务已清除。其他 PostgreSQL、Redis、Docker 卷和 Compose 项目未操作。
- 部署与验收：原子替换 `/opt/sechelper-auth-template/bin/auth-template`、迁移器及 `/opt/sechelper-auth-template/web/dist` 前后台静态产物；`order-test.service` 当前 active，重启计数 0。`/healthz`、`/readyz`、`/admin/`、Admin JS 均返回 HTTP 200；未认证访问 `/v1/admin/dashboard/overview` 返回预期 HTTP 401；Session 为未认证。`/v1/version` 返回环境 `test`、版本 `0.1.0`、Build ID `20260915150445`、源提交 `ebceb01a9a0e68f213adbdb22fe9c4473e5ed64a`。由于数据库重建清除了旧 Session，浏览器打开 Admin 后按统一 OIDC 流程跳转至测试身份中心登录页；没有输入或读取凭据，重新登录后即可查看资源卡片。
- 恢复：备份目录 `/opt/sechelper-auth-template-reset-backup-20260915T172015Z/` 保留重建前源码归档、旧 API 二进制、旧 `web/dist`、数据库自定义转储（`pg_restore --list` 验证通过，SHA-256 `e3a04a1f930d54465c9bd2493aa8de7a27f60d7cb481613d56a1b2c85d17ddd8`）、构建 manifest、配置修改前副本及新源码归档。恢复数据库需先停止 `order-test.service`，重建 `auth_template` 后从 `auth_template-before-reset.dump` 执行 `pg_restore`；应用回滚使用备份的旧二进制与静态资源后启动现有 systemd 单元。Nginx 配置未修改、未 reload。

## 平台总览服务器资源 API 远程验证（2026-09-16 00:07 Asia/Shanghai）

- 验证源：当前未提交工作树的后台 Dashboard/API 资源改动；在测试主机 `/tmp/sechelper-resource-api-check.8ehPva` 临时源码快照中验证，未替换 `/opt/sechelper-auth-template` 下的服务源码。
- 工具链与结果：Go 1.26.8 `go test ./...`、API `go build` 及后续 Dashboard 资源采集器定向测试通过；Node.js v24.21.0、npm 11.19.0 下 Admin 测试 17 项通过，Vite 7.1.7 production 构建转换 50 个模块成功。OpenAPI YAML 语法及 Dashboard resources schema 结构检查通过；Redocly CLI lint 未完成，因远端 npm 缓存缺少 `@opentelemetry/api-logs`（`ENOTCACHED`），未安装或改动项目依赖。
- 清理与运行影响：临时源码、Go 缓存、编译产物、Admin dist 和 npm cache 已清理；`order-test.service` 保持 active。没有部署或替换线上静态资源，未重启服务、未修改 API 二进制、数据库、Redis、Nginx、systemd、运行配置或业务数据。

## 平台总览夜间样式远程验证（2026-09-15 22:25 Asia/Shanghai）

- 仅将获准的 `web/admin/src/style.css` 临时同步到测试主机进行验证；Admin 测试 15 项通过，Vite 7.1.7 production 构建转换 50 个模块成功。
- 验证后恢复远端原样式，清理本次生成的 `web/admin/dist` 和临时备份；未替换线上 `/admin` 静态资源，未重启服务或修改 API、数据库、Redis、Nginx、systemd 及业务数据。

## 平台总览夜间适配与 Manifest 按钮远程部署（2026-09-15 22:33 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`。
- 部署源：当前授权工作树的 `web/admin/src/modules/dashboard/DashboardPage.jsx`（SHA-256 `6808613c7c1a3f9c177bcc9b3761e34ecad8190baa047bbb92ef09e003543167`）和 `web/admin/src/style.css`（SHA-256 `ab6a6600ecac5bb68942966a2c580fe32a7aa9be593751e015e659d1e27e0387`）；仅同步这两个文件并更新后台静态资源 `/opt/sechelper-auth-template/web/dist/admin`。
- 构建验证：复用远程 Node.js `v24.21.0`、npm `11.19.0`、Vite `7.1.7`；`npm test` 15 项通过，`npm run build` 转换 50 个模块成功。架构检查尝试执行但未完成，因主机缺少 `rg`（退出码 127）；未安装额外工具或修改检查脚本。
- 验收：线上新 JS 不再含“查看权限清单”按钮文字，CSS 包含夜间状态色板与资源/Manifest 等高布局；后台及其 JS/CSS、`/healthz`、`/readyz` 均返回 HTTP 200；未认证访问 `/v1/admin/dashboard/overview` 返回预期 HTTP 401；`order-test.service` 保持 active。
- 运行影响：仅原子替换后台静态资源；未重启服务或 reload Nginx，未修改 API、PostgreSQL、Redis、Nginx 配置、systemd、运行配置或业务数据。未部署尚待授权的服务器资源采集 API。
- 回滚：部署前两份源码及原后台静态目录位于 `/opt/sechelper-auth-template/.deploy-backup-platform-overview-night-layout-20260915T223200/`；恢复该目录下 `source/` 和 `admin-dist/` 即可回退。

## Frest 组件参考目录远程部署（2026-09-15 14:38 Asia/Shanghai）

- 目标：测试主机 `47.116.4.57`，项目路径 `/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`。源代码来自本地未提交工作树中的 `web/admin/src/business/component-reference/`；本次部署快照以同步文件校验和标识：`ComponentReferencePage.jsx` SHA-256 `adc05993f7eae5ba6209c17d202055ad6e06f9be92cf8c215884dac9d24c55e5`，`component-reference.css` SHA-256 `893409eab954469637eebaa0c8839b1f6499158c76aa63b3e0aa18e8df0e626d`。Git 修订号未记录，因为部署源含未提交改动。
- 范围：仅同步组件参考业务模块三个文件；远程以 `npm ci --no-audit --no-fund` 安装锁文件依赖，执行后台 `npm test` 和 `npm run build -- --mode test`，并将测试模式 Admin 静态构建替换至 `web/dist/admin`。构建包含 `/admin/component-reference` 测试路由，不含 API 或数据库改动。
- 工具链与运行：复用远程 Node.js `v24.21.0`、npm `11.19.0`（已存在的 root 用户工具链）；依赖安装作用域为项目 `web/admin/node_modules`。`npm ci` 提示 `esbuild@0.25.12` 安装脚本尚未列入 `allowScripts`，但安装、测试与构建均成功。测试通过 15 项；Vite `7.1.7` 测试构建成功，转换 55 个模块。构建/测试命令为 `npm test` 与 `npm run build -- --mode test`；本次为静态资源更新，没有执行服务启动命令。`order-test.service` 保持 active，由现有主机 Nginx 提供前后台静态资源；部署未重启服务、未 reload Nginx。
- 验收：`https://order-test.sechelper.com/admin/` 与 `/admin/component-reference` 均返回 HTTP 200；部署静态资源包含组件参考路由；服务状态为 active。`make architecture-check` 在远端未能完成，因主机缺少 `rg`，检查脚本按设计以扫描失败退出；同一工作树部署前本地 `make architecture-check` 已通过，未安装额外扫描工具或修改检查脚本。
- 数据与清理：未连接或操作 PostgreSQL、Redis、业务数据；未执行迁移。保留远端 `web/admin/node_modules` 供后续锁文件构建复用；部署后清理本次生成的 `web/admin/dist`，线上静态产物保留在 `web/dist/admin`。部署前源码与 Admin 静态产物备份位于 `/opt/sechelper-auth-template/.deploy-backup-component-reference-20260915T063624Z`，含 `web/admin/src/business/component-reference/` 和原 `web/dist/admin/`；目录内另有部署切换时保存的原线上静态目录 `web/dist/admin-live-before-swap/`。
- 回滚：恢复备份中的 `web/admin/src/business/component-reference/`，并用备份中的 `web/dist/admin/` 替换 `/opt/sechelper-auth-template/web/dist/admin/`；无需重启服务或 reload Nginx。

## 业务目录基线清库全量重部署（2026-09-14 22:16 Asia/Shanghai）

- 源码：以本地 `main` 的 `ed93095750f2b6b4745d6c23ee4047d2482d2e3a` 加当前未提交工作树为事实版本，全量同步到 `root@47.116.4.57:/opt/sechelper-auth-template`；同步包排除 `.git/`、`.env`、`config.yaml`、密钥、依赖缓存、构建产物和日志。远程旧的 `admin/`、`shared/`、`e2e/` 及旧源码目录已由当前代码结构替换。
- 构建验证：远程使用 Go 1.26.8、Node v24.21.0 和 npm 11.19.0；Go 全量测试、`example` Orders 测试、API 与迁移器编译、公共前台测试/生产构建、管理前台测试/生产构建全部通过。第一次解包因 macOS AppleDouble `._*` 旁车文件导致前台测试失败，在停服和清库前已中止；重新生成无 xattr/AppleDouble 的同步包后全量验证通过。
- 数据清理：仅停止 `order-test.service` 并删除、重建项目专用容器 `sechelper-auth-template-dev-postgres-1` 中的测试数据库 `auth_template`。统一认证 PostgreSQL 未操作；服务使用的 Redis 为统一认证共享实例，未清理、未重启、未修改配置或数据。
- 迁移：递归读取 `/opt/sechelper-auth-template/api/migrations`，成功应用 `001_authentication.sql`、业务迁移 `business/orders/002_orders.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql` 和 `005_audit_event_model.sql`。`schema_migrations` 为 5；认证 Session、登录事务、审计事件和订单均为 0。服务启动后 Manifest 自动同步生成 1 条状态记录，状态为 `applied`、版本 2、服务端修订 2。
- 部署结果：`order-test.service` 已启动且为 active；`https://order-test.sechelper.com/healthz` 与 `/readyz` 均返回 200，前台 `/`、后台 `/admin/` 和 `/v1/auth/session` 均返回 200，会话响应为未认证。当前源码结构已验证存在 `api/internal/business/orders/`、`web/src/framework/`、`web/admin/src/business/orders/` 和 `api/migrations/business/orders/002_orders.sql`，旧 `api/internal/modules/orders/` 不存在。
- 恢复点：`/opt/sechelper-auth-template/.deploy-backup-business-layout-20260914T221614`，包含权限为 600 的 `.env`、`config.yaml`、部署前源码归档和 PostgreSQL 自定义格式备份 `postgres-before-reset.dump`。本次临时上传目录、远程编译目录和临时 Go 缓存已清理；Nginx 配置未修改且无需 reload。

## 后台导航分组远程重新部署（2026-09-14 19:56 Asia/Shanghai）

- 当前变更：同步 `admin/src/app/navigation.jsx`，移除“权限与资源”分组，将权限 Manifest、权限清单、资源目录和活跃会话归入“账号与安全”，将订单归入同级的“业务运营”。
- 远程目标：`root@47.116.4.57:/opt/sechelper-auth-template`；仅重建并替换后台静态资源，未修改 API、PostgreSQL、Redis、Nginx、`.env`、`config.yaml` 或业务数据。
- 验证：远程 Admin Vite 生产构建通过；`order-test.service` 为 active；`/healthz`、`/readyz` 和 `/admin/` 返回 HTTP 200；远程静态资源已包含“账号与安全”和“业务运营”。
- 部署备份：`/opt/sechelper-auth-template/.deploy-backup-navigation-20260914T1956`，包含部署前导航文件和后台静态资源。

## Frest 左侧导航远程构建验证（2026-09-14 18:13 Asia/Shanghai）

- 验证源码：本地未提交工作树中的 `admin/src/app/navigation.jsx`、`admin/src/app/components.jsx`、`admin/src/style.css` 及共享错误页依赖；本次仅构建验证，未发布到线上。
- 远程目标：`root@47.116.4.57`；使用临时目录完成 Admin 生产构建，Node `v24.21.0`、npm `11.19.0`、Vite `7.1.7`，构建通过后已清理临时目录。
- 验证范围：Frest Demo 1 风格的 `260px` 展开栏、`80px` 收起栏、分组标题、线性 SVG 图标、活动态和桌面折叠交互；未修改 `order-test.service`、Nginx、PostgreSQL、Redis、运行配置或业务数据。

## Frest 顶部导航远程部署（2026-09-14 18:19 Asia/Shanghai）

- 当前变更：覆盖管理端顶部导航为 Frest `62px` 顶栏，加入搜索入口、语言/主题/快捷入口/通知图标、用户状态点和账号菜单；与左侧导航样式一并同步。
- 远程目标：`root@47.116.4.57:/opt/sechelper-auth-template`；Go 全量测试、API 编译、前后台生产构建、`order-test.service` 重启和 Nginx reload 均成功。
- 验收：浏览器强制刷新后，`/admin/audit-events` 已显示新版顶部搜索、语言、主题、快捷入口、通知和用户状态点；旧卡片式 Header 已移除。`/healthz`、`/readyz`、前台和 `/admin/` 均返回 HTTP 200，`/v1/auth/session` 返回未认证状态。
- 部署备份：`/opt/sechelper-auth-template/.deploy-backup-topbar-20260914T181848`。Nginx 仅保留既存 TLS protocol 重复配置 warning，语法检查成功。
- 约束：保留远程 `.env`、`config.yaml`、PostgreSQL、Redis 和业务数据；不执行数据库迁移，不同步 `.git/`、密钥、依赖缓存或构建产物。

## system-design 助安社区 Logo 远程构建验证（2026-09-14 18:22 Asia/Shanghai）

- 当前变更：`CommunityBrand` 改用 system-design 规范原始资产 `admin/public/assets/logo/logo.png`，保留双行品牌文字和无障碍名称。
- 远程临时 Admin 生产构建通过，构建包包含 `dist/assets/logo/logo.png`；临时目录已清理。本次未替换线上静态文件、未重启服务，PostgreSQL、Redis、Nginx 和运行配置未修改。

## system-design 助安社区 Logo 远程发布（2026-09-14 18:23 Asia/Shanghai）

- 已同步 `admin/public/assets/logo/logo.png`、`admin/src/platform/brand/CommunityBrand.jsx` 及相关文档；Go 全量测试、API 编译、前后台生产构建均通过。
- `order-test.service` 已重启，Nginx 配置检查通过并 reload。重启瞬间首次健康探测返回 502，服务恢复后复验 `/healthz`、`/readyz`、前台、`/admin/` 和 Logo 静态资源均为 HTTP 200；服务当前 active。
- 浏览器强制刷新确认左上角显示 system-design 原始助安社区 Logo，顶部 Frest 导航与侧栏样式保持生效。部署备份：`/opt/sechelper-auth-template/.deploy-backup-logo-20260914T182305`。
- PostgreSQL、Redis、`.env`、`config.yaml` 和业务数据未修改，未执行数据库迁移。

## 概览平台卡片迁移远程构建验证（2026-09-14 18:28 Asia/Shanghai）

- 当前变更：将当前会话、访问控制诊断和系统状态内容集中迁移到 `admin/src/modules/dashboard/DashboardPage.jsx`；移除三个独立后台导航入口与直达路由，保留活跃会话、权限清单和审计页面。
- 远程临时 Admin 生产构建通过，Vite `7.1.7` 构建包生成成功；本次仅构建验证，未发布线上、未重启服务，未修改 PostgreSQL、Redis、Nginx、运行配置或业务数据。

## 当前工作树全量同步重新部署（2026-09-14 19:45 Asia/Shanghai）

- 已按当前本地工作树全量同步到 `root@47.116.4.57:/opt/sechelper-auth-template`，排除 `.git/`、`.env`、`config.yaml`、依赖缓存、构建产物和日志；远程新增同步 `api/internal/platform/security/` 与 `docs/capability-matrix.md`。
- 远程 Go 全量测试、API 编译、`web` 与 `admin` 生产构建均通过；`order-test.service` 已重启并 active，Nginx 配置检查成功并 reload。
- 重启瞬间首次健康探测返回 502，延迟复验恢复正常：`/healthz`、`/readyz`、前台、`/admin/` 和 Logo 静态资源均 HTTP 200；`/v1/auth/session` 返回未认证状态。
- 部署备份：`/opt/sechelper-auth-template/.deploy-backup-full-20260914T194526`。PostgreSQL、Redis、`.env`、`config.yaml` 和业务数据未修改，未执行数据库迁移。

## 项目身份

## Manifest 对接修订与验证（2026-09-14 17:14 Asia/Shanghai）

- 按统一认证平台业务集成契约更新 Manifest 客户端：机器令牌改用 `/api/v1/provisioning/token`，请求 `grant_type=client_credentials`、`scope=authorization.manifest` 和 audience `provisioning-api`；状态和提交路径改为 `/api/v1/provisioning/authorization-manifest/status` 与 `/api/v1/provisioning/authorization-manifest`。
- Manifest Registry 在计算 canonical hash 前规范化 method、路径和文本字段，缺省 `risk_level` 为 `normal`；新增客户端契约和 canonical 规范化测试。
- 远程目标：`root@47.116.4.57:/opt/sechelper-auth-template`。已同步选定 Go 源码、测试和集成文档，远程 Go 全量测试及构建通过；生成的新二进制已部署到 `/opt/sechelper-auth-template/bin/auth-template`。部署前二进制备份位于 `/opt/sechelper-auth-template/.deploy-backup-manifest-20260914171222`，运行配置备份位于 `/opt/sechelper-auth-template/.env.backup-manifest-20260914171307`。
- 远程验证：`order-test.service` active，`/healthz` HTTP 200，`/readyz` HTTP 200；Provisioning Token HTTP 200；Manifest 状态 HTTP 200，远端状态为 `consistent`、版本 1、server revision 1。未修改 PostgreSQL、Redis、Nginx 或业务数据，未执行 SSH 端口转发。

- 本地项目根目录：`/Users/cookun/GolandProjects/sechelper-auth-template`
- 远程 SSH 目标：`root@47.116.4.57`
- 远程主机名：`iZuf6664owh4cneengotp1Z`
- 远程操作系统：Ubuntu 24.04
- 远程架构：x86_64
- 目标环境：test
- 目标域名：`order-test.sechelper.com`
- 记录状态：前后台分离版本已重新部署到测试域名并通过健康检查；源码与构建产物位于 `/opt/sechelper-auth-template`

## 最近一次认证框架回归

- 时间：2026-09-14 Asia/Shanghai；源码已同步到 `/opt/sechelper-auth-template`，Go 服务和后台构建产物已重新部署。
- IdP：测试 Passport 的 OIDC `issuer` 为 `https://passport-test.sechelper.com`，JWKS 使用 `/oauth2/jwks`；服务配置通过 systemd drop-in 注入 `IDENTITY_JWKS_URL`，不记录任何凭据。
- 结果：Go 全量测试通过；测试浏览器完成 OIDC 授权码回调、ID Token 签名/issuer/client audience/nonce/subject 校验、Session 创建、前台会话读取、后台权限负向测试、订单页、Manifest 查看、Manifest 同步和后台登出回归。
- 发现并修复：Manifest 状态响应补充 lowerCamelCase JSON 字段；后台登出后不再被自动登录副作用带回 IdP。
- 测试数据：曾为当前测试账号增加临时本地会话权限以验证后台成功路径；测试结束后已撤销该账号的测试会话、清空临时权限并清理对应用户会话索引。
- 最终状态：`order-test.service` active，`/healthz` 返回 200，`/readyz` 返回 200；PostgreSQL、Redis 和现有业务数据未重建或清空。
- 可重复 Refresh Rotation 测试：`go test ./internal/modules/authentication/application -run 'TestRefresh'` 使用 Mock Identity Provider 固定返回两代 token，验证 Session 有效期更新、密文轮换，以及 Provider 返回 `invalid_grant` 时 Session 被撤销；本地全量 Go 测试通过。
- 多实例认证基础设施：PostgreSQL Store 通过事务和 `SELECT ... FOR UPDATE` 对单个 Session 的 Refresh Rotation 加锁，提交新密文和版本递增；`api/migrations/003_authentication_transactions.sql` 提供 `authentication_sessions.version`。真实多副本和测试数据库并发验证尚未在本轮执行。
- 认证限流：测试服务注入登录/回调/刷新限额 `10/20/30` 每分钟；连续请求 `/v1/auth/login` 的结果为前 10 次 `302`、第 11 次 `429`，测试事务随后已清理。
- 浏览器 E2E：新增 `e2e/` Playwright 工程和 CI workflow；公开测试覆盖前台壳、未认证 Session/Authorization、metrics，真实 IdP 登录/登出仅在注入 `E2E_USERNAME`/`E2E_PASSWORD` 时启用。当前 macOS 沙箱中的 Chromium 无头启动被系统 Mach port 权限阻断，API-only E2E 已通过；CI 使用 Ubuntu Chromium。
- 运行指标：测试服务 `/metrics` 已由 Nginx 代理，暴露认证请求、限流和远端 revoke 失败计数器；告警规则位于 `deploy/prometheus/auth-template-alerts.yml`。

## 已验证工具链

| 工具 | 版本或路径 |
|---|---|
| Go | `/usr/local/bin/go`, Go 1.26.8 linux/amd64 |
| Node | `/usr/local/bin/node`, v24.21.0 |
| npm | `/usr/local/bin/npm`, 11.19.0 |
| Corepack | `/usr/local/bin/corepack`, 0.36.0 |
| Docker | `/usr/bin/docker`, 29.8.0 |
| Docker Compose | v5.5.1 |
| Nginx | `/usr/sbin/nginx`, 1.24.0 |

## 现有远程状态

- 已存在项目目录：`/opt/order-test`
- 已存在 systemd 服务：`order-test.service`
- 已存在 Nginx 配置：`/etc/nginx/sites-enabled/order-test.sechelper.com.conf`
- 已存在业务监听：`127.0.0.1:8080`
- 已存在 PostgreSQL 容器：`unified-login-postgres-1`，绑定 `127.0.0.1:25432`
- 已存在 Redis 容器：`unified-login-redis-1`，绑定 `127.0.0.1:26379`
- 已存在业务容器：`order-test-order-test-1`，绑定 `127.0.0.1:23000`
- Nginx 正在监听 80/443

## 当前部署状态

- OIDC 配置按最新 JSON 重新校准：2026-09-14 12:42 Asia/Shanghai。将 Client、OIDC 端点、JWKS、Application Code、Audience、回调地址、Scope、CORS 和 Manifest 凭据重新写入 systemd 使用的 `/opt/sechelper-auth-template/.env`，备份位于 `/opt/sechelper-auth-template/.deploy-backup-config-json-reapply-20260914T124249`；`.env` 权限为 600。服务重启后 active，`/healthz` 和 `/readyz` 返回 200；使用伪造授权码探测 Token Endpoint 返回 `INVALID_OR_EXPIRED_CODE` 而非 `INVALID_CLIENT`，证明 Client ID/Secret 基础认证已通过。当前浏览器旧授权码不可重放，需重新发起登录。

- 独立 PostgreSQL 接管：2026-09-14 12:31 Asia/Shanghai。使用仓库 `deploy/compose.dev.yaml` 以 Compose 项目 `sechelper-auth-template-dev` 启动独立容器 `sechelper-auth-template-dev-postgres-1`，绑定 `127.0.0.1:5432`，数据库 `auth_template`、用户 `app`；已将远程 `/opt/sechelper-auth-template/config.yaml` 的数据库连接切换到该实例，并应用全部 4 个项目迁移。`order-test.service` 已重启且 active，`/healthz`、`/readyz` 返回 200，前台返回 200，新增后台接口未认证均返回 401。独立 PostgreSQL healthy；原有 `unified-login-postgres-1`、`unified-login-redis-1` 保持 running，未修改其数据和配置。

- 最新清空数据后全量部署：2026-09-14 11:05 Asia/Shanghai。停止 `order-test.service` 后删除并重新创建 `unified-login-postgres-1` 中的 `auth` 数据库，清空 `unified-login-redis-1`，同步最新代码并应用 `001_authentication.sql`、`002_orders.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql`；远程 Go 全量测试、API 构建、`web` 与 `admin` 构建均通过，已启动 `order-test.service` 并 reload Nginx。当前认证 Session、登录事务、Manifest 状态、审计记录和订单均为 0，Redis key 为 0；`/healthz` 返回 200，前台、后台返回 200，未认证访问授权及新增后台接口均返回 401。`/readyz` 当前返回 503，启动 Manifest 同步仍因远程保留的 Manifest 凭据无效而未完成；服务保持 active 但尚未 ready。部署前数据库备份位于 `/opt/sechelper-auth-template/.deploy-backup-clear-redeploy-20260914T110423/postgres-before-drop.dump`。

- systemd 配置回退：2026-09-14 10:37 Asia/Shanghai。恢复测试物理机通过 `/opt/sechelper-auth-template/.env` 加载运行时环境变量，并保留 `/opt/sechelper-auth-template/config.yaml` 作为 Viper 应用配置；移除了 drop-in 中会清空 `EnvironmentFile` 的空指令，确认 `systemctl show order-test.service -p EnvironmentFiles` 指向该 `.env`。服务当前 active，`/healthz` 返回 200，`/readyz` 仍为 503，原因是远程 Manifest 凭据仍被统一认证平台返回 `401 INVALID_SYNC_CREDENTIAL` 拒绝；未修改数据库、Redis、项目数据或 Nginx。

- 最新全量重部署：2026-09-14 10:15 Asia/Shanghai。远程重新创建了已确认删除的 `auth` 数据库并应用 `001_authentication.sql`、`002_orders.sql`、`003_authentication_transactions.sql`、`004_audit_events.sql`；同步当前代码并完成 Go 全量测试、API 构建、`web` 与 `admin` 生产构建，`order-test.service` 已启动，Nginx 已通过 `nginx -t` 并 reload。当前数据为空（认证 Session、登录事务、Manifest 状态、审计记录、订单均为 0，Redis key 为 0）。`/healthz` 返回 200，前台和后台返回 200，未认证访问授权及所有新增后台接口返回 401；`/readyz` 当前返回 503，因为清空本地 Manifest 状态后，统一认证平台拒绝远程保留的 Manifest 凭据并返回 `401 INVALID_SYNC_CREDENTIAL`。服务保持 active 但尚未 ready；需要更新有效的测试 Manifest 凭据后重启服务完成首次同步。数据库备份位于 `/opt/sechelper-auth-template/.deploy-backup-clear-target-20260914T100535/postgres.dump`。

- 本次全量代码重新部署：2026-09-14 09:42 Asia/Shanghai。已将当前本地工作树源码（排除 `.git/`、`.env`、`config.yaml`、依赖缓存、已有构建产物和日志）同步到 `/opt/sechelper-auth-template`，远程使用 Go 1.26.8、Node 24.21.0 和 npm 11.19.0 完成 Go 全量测试、API 构建、前后台生产构建；后台产物已更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-full-20260914T093740`。已重启 `order-test.service` 并 reload Nginx；`nginx -t` 成功，`/healthz`、`/readyz` 返回 200，前台、后台页面返回 200，未认证访问 `/v1/authorization/me` 和 `/v1/admin/dashboard/overview` 均返回 401。PostgreSQL、Redis、现有业务数据和远程运行配置未重建、未清空、未修改；未执行数据库迁移。

- `order-test.service` 已切换到 `/opt/sechelper-auth-template/bin/auth-template`，工作目录为 `/opt/sechelper-auth-template`。
- 运行配置位于远程 `/opt/sechelper-auth-template/.env`，由测试数据库中的客户端配置和 Manifest 凭据生成；不记录具体密钥值。
- Nginx 配置 `/etc/nginx/sites-enabled/order-test.sechelper.com.conf` 使用 `/opt/sechelper-auth-template/web/dist` 作为前台静态根目录，`/admin/` 回退到前台静态根下的后台构建目录，并将 `/v1/`、`/healthz`、`/readyz` 代理到 `127.0.0.1:8080`。
- 部署前备份：`/opt/order-test/.deploy-backup-auth-template-20260913T175801`。
- PostgreSQL、Redis 和现有业务容器未重启、未重建、未修改；新服务复用了现有测试数据库和 Redis。

远程依赖源只做了连通性检查，未写入系统或用户配置。可达候选为 `https://goproxy.cn/`、`https://registry.npmmirror.com/` 和 `https://registry.npmjs.org/`，等待部署方批准后使用。

## 远程执行约束

- 未执行远程 Git 操作。
- 未复制 `.git/`、`.env`、密钥、依赖缓存或构建产物。
- 已按用户授权修改 Nginx 和 systemd；未修改 Docker、PostgreSQL、Redis 配置或现有业务数据。
- 未创建 SSH 端口转发。
- 已同步源码到 `/opt/sechelper-auth-template`，未同步 `.git/`、`.env`、密钥或依赖缓存；前端构建产物由远程 Node 工具链生成。

## 最近预检

- 本次资源授权内核与后台平台能力远程重新部署：2026-09-14 13:12 Asia/Shanghai。目标 `root@47.116.4.57`，项目目录 `/opt/sechelper-auth-template`。已同步当前源码（排除 `.git/`、`.env`、`config.yaml`、依赖缓存、构建产物和日志），远程完成 Go 全量测试、API 编译、`web` 与 `admin` 生产构建；部署备份位于 `/opt/sechelper-auth-template/.deploy-backup-rac-20260914T1310`。已原子替换 API 二进制、更新后台静态产物、重启 `order-test.service` 并确认 active；`nginx -t` 成功后 reload。`/healthz`、`/readyz`、前台和 `/admin/` 返回 HTTP 200；未认证访问 `/v1/admin/dashboard/overview`、`/v1/admin/resources` 和 `/v1/admin/access-decisions/check` 均返回 HTTP 401。PostgreSQL、Redis、现有数据和运行配置未修改，未执行数据库迁移；Nginx 仅存在既有的重复 TLS protocol warning，语法测试成功。

- 部署时间：2026-09-13 21:52 Asia/Shanghai
- 本次源状态：本地工作树（当前项目未包含 Git 元数据）；远程构建产物 `/opt/sechelper-auth-template/bin/auth-template` 于 21:51 更新。
- 结果：远程 Go 全量测试通过；Go 构建通过；`web` 和 `admin` 均完成远程 `npm ci` 与 Vite 生产构建；`order-test.service` 为 active；`nginx -t` 通过并已 reload；`https://order-test.sechelper.com/`、`/admin/`、`/admin/orders` 和后台静态资源均返回 HTTP 200；`/healthz`、`/readyz` 均返回 HTTP 200；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401。
- 远程构建提示：前后台 `npm ci` 均报告 1 个 high severity audit vulnerability；本次未执行 `npm audit fix --force`，避免未授权升级依赖。
- 本次部署备份：`/opt/sechelper-auth-template/.deploy-backup-admin-split-20260913T214949`。
- Manifest 补同步：部署初次启动时身份平台返回 HTTP 502，导致启动同步失败；身份平台恢复后于 21:56 重启业务服务触发同步。远端 Manifest 当前为版本 3、服务端修订 3、状态 `consistent`；本地 `authorization_manifest_state` 已包含 `admin:access`，状态为 `applied`。
- 后台样式更新：2026-09-13 23:10 Asia/Shanghai 在远程主机完成 `admin/` 的 `npm ci` 和 Vite 生产构建，部署管理端样式到 `/admin`；`/admin/`、`/admin/orders`、新版 CSS 静态资源均返回 HTTP 200。备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-style-20260913T231020`。Go 服务、PostgreSQL、Redis 和 Nginx 配置未修改；Nginx 仅执行了配置验证，未重新加载。
- 左上角社区标识更新：2026-09-13 23:21 Asia/Shanghai 在远程主机同步 `admin/src/`、`admin/public/`、`admin/index.html`、`admin/package.json`、`admin/package-lock.json` 和 `admin/vite.config.js`，完成远程 `npm ci` 与 Vite 生产构建，并将产物同步到 `/opt/sechelper-auth-template/web/dist/admin`。`https://order-test.sechelper.com/admin/`、`/admin/assets/logo/logo.svg`、`/healthz`、`/readyz` 返回成功；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。本次 npm 审计仍报告 1 个 high severity vulnerability，未执行强制依赖升级。
- 错误状态页更新：2026-09-13 23:28 Asia/Shanghai 在远程主机同步 `admin/src/app/AdminApp.jsx`、`admin/src/style.css` 和 `admin/README.md`，由远程 Node 工具链完成 Vite 生产构建，并将后台产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。新增统一 403、404、500 页面及返回/重试操作；`/admin/` 与新版 CSS 返回 HTTP 200，`/healthz`、`/readyz` 返回 HTTP 200，`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。本次部署备份位于 `/opt/sechelper-auth-template/.deploy-backup-error-pages-20260913T232821`。
- 部署期间修复两处兼容性问题：身份平台端点校验错误限制 OAuth 路径；Manifest 状态持久化未写入现有数据库要求的 `canonical_json` 字段。对应代码已同步到本地源码和远程构建目录。
- 首次启动失败日志仍保留在远程 `/opt/sechelper-auth-template/logs/auth-template-debug.log` 和 `auth-template.jsonl`，最新启动已记录 `server.started`，服务当前正常。
- 后台主页重新部署：2026-09-14 00:00 Asia/Shanghai 在远程主机同步 `admin/src/`、`admin/public/`、`admin/index.html`、`admin/package.json`、`admin/package-lock.json`、`admin/vite.config.js` 和 `admin/README.md`，远程执行 `npm ci --no-audit --no-fund` 与 `npm run build`，将构建产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-home-20260914T000034`。`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 错误页面重新部署：2026-09-14 00:07 Asia/Shanghai 在远程主机同步 `admin/src/app/AdminApp.jsx`、`admin/src/style.css` 和 `admin/README.md`，远程执行 `npm ci --no-audit --no-fund` 与 `npm run build`，将 403/404/500 页面构建产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-errors-20260914T000703`。远程 CSS 含 `error-screen` 且不再包含旧 `incident-screen`；`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 正式后台掌机错误页远程部署：2026-09-14 01:05 Asia/Shanghai 在远程主机同步 `admin/src/app/AdminApp.jsx`、`admin/src/style.css`、`admin/public/`、`admin/index.html`、`admin/package.json`、`admin/package-lock.json`、`admin/vite.config.js` 和 `admin/README.md`，远程执行 `npm ci --no-audit --no-fund` 与 `npm run build`，将完整掌机版 403/404/500 页面构建产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-handheld-20260914T010517`。线上构建已包含 `handheld-error`、屏幕游戏数据面板以及 `handheld-forbidden`/`handheld-server` 颜色变体；`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页远程部署：2026-09-14 01:19 Asia/Shanghai 在远程主机同步 `shared/error-pages/`、`web/src/app/App.jsx`、`web/src/style.css`、`web/vite.config.js`、`admin/src/app/AdminApp.jsx`、`admin/vite.config.js` 及相关前端清单，远程分别执行 `web` 与 `admin` 的 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 及其 `admin` 子目录。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-global-errors-20260914T011621`。验收确认前台和后台构建包均包含共享 `global-error-page`、`SECHELPER COMMUNITY` 和全局 403/404/500 组件；`/`、`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页纵向布局远程部署：2026-09-14 01:22 Asia/Shanghai 在远程主机同步 `shared/error-pages/GlobalErrorPage.jsx`、`shared/error-pages/error-pages.css`、`web/src/`、`admin/src/` 及前后台清单，远程分别完成 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-global-errors-layout-20260914T012223`。验收确认前台和后台均包含“顶部大号状态数字 + 下方掌机”布局、共享错误页样式和 403/404/500 状态色；`/`、`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页交互与居中修复远程部署：2026-09-14 01:26 Asia/Shanghai 在远程主机同步 `shared/error-pages/GlobalErrorPage.jsx`、`shared/error-pages/error-pages.css`、`web/src/`、`admin/src/` 及前后台清单，远程分别完成 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-error-pages-interactive-20260914T012627`。浏览器验收确认 403 页面中的方向键、A/B/X/Y 和“开始”按钮已渲染为可交互按钮，点击“开始”并点击“下移”后游戏可响应；状态数字与掌机整体垂直居中；前台和后台构建包均包含共享错误页；`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页按键布局远程部署：2026-09-14 01:30 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-button-layout-20260914T013023`。浏览器刷新验收确认方向键靠掌机左侧、A/B/X/Y 靠掌机右侧，底部选择/开始保持居中；点击“开始”后屏幕显示运行中的方块，点击“下移”可响应；`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- A/B/X/Y 按键变形修复远程部署：2026-09-14 01:41 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，将动作键从旋转 Grid 改为固定绝对定位菱形布局，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-button-fix-20260914T014139`。浏览器刷新确认动作键位置稳定，4 个按钮均可定位和点击；点击“开始”、A、B 后页面保持正常；`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- A/B/X/Y 角度与 X-B 对齐远程部署：2026-09-14 01:47 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，将 X 调整到与 B 同一垂直中心线，并统一动作键倾斜角度，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-action-angle-20260914T014728`。线上新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页按键几何与交互远程部署：2026-09-14 01:37 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-dpad-face-buttons-20260914T013759`。浏览器刷新验收确认十字键四个三角通过 CSS 几何定位居中，A/B/Y/X 按固定菱形位置排列；点击“开始”后方块运行，点击“下移”可响应；`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 全局错误页十字键修复远程部署：2026-09-14 01:34 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-dpad-layout-20260914T013401`。浏览器验收确认方向键恢复为完整十字底板、箭头位置正确，A/B/X/Y、选择/开始和掌机布局正常；点击“开始”后游戏启动，点击“下移”后方块正常响应；`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- X 键垂直对齐修复远程部署：2026-09-14 01:55 Asia/Shanghai 在远程主机同步 `shared/error-pages/error-pages.css`，将 X 键横坐标由 `23px` 调整为与 B 键相同的 `32px`，远程分别完成前台与后台 `npm ci --no-audit --no-fund` 和 `npm run build`，更新 `/opt/sechelper-auth-template/web/dist` 与 `/opt/sechelper-auth-template/web/dist/admin`。浏览器实测 B、X 按钮中心横坐标均为 `759px`、尺寸均为 `35px × 35px`，点击“开始”和 X 键均正常；`/healthz`、`/readyz` 返回成功，`order-test.service` 保持 active，`nginx -t` 通过。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx；回滚可使用上一版按键布局备份 `/opt/sechelper-auth-template/.deploy-backup-action-angle-20260914T014728`。
# 最新全量重建部署（2026-09-14 20:51 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，业务服务 `order-test.service`。
- 数据库：仅重建项目专用 Compose 容器 `sechelper-auth-template-dev-postgres-1` 中的 `auth_template` 数据库；其他 PostgreSQL、Redis、业务容器和业务数据未操作。
- 备份：`/opt/sechelper-auth-template/.deploy-backup-full-rebuild-20260914T204703/auth_template-before-rebuild.dump`，备份文件已确认非空。
- 发布：当前工作树已排除 `.git/`、`.env`、`config.yaml`、依赖缓存、构建产物和日志后同步；远程 Go 全量测试、API/迁移编译、Web/Admin 测试和生产构建通过。
- 数据库：显式使用 `MIGRATIONS_DIR=/opt/sechelper-auth-template/api/migrations` 应用全部 5 个迁移；`schema_migrations` 记录数为 5。
- 验证：`order-test.service` active；`/healthz`、`/readyz`、前台 `/`、后台 `/admin/` 和 `/v1/auth/session` 均返回 HTTP 200；Nginx `nginx -t` 成功并已 reload。
- 回滚：API 旧二进制和旧前端静态资源位于同一备份目录；数据库可使用上述 dump 恢复。未执行 Redis 清空或数据库之外的破坏性操作。

# 前端业务模块自动发现重新部署（2026-09-14 22:34 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，业务服务 `order-test.service`。
- 发布：同步前后台业务模块自动发现注册器、后台直达路由权限校验、禁用的订单演示声明及对应文档；未同步 `.git/`、`.env`、`config.yaml`、密钥、依赖缓存或本地构建产物。
- 验证：远程 Go 全量测试、API/迁移编译、Web/Admin 单元测试和生产构建通过；Vite 分别转换 37 和 49 个模块。`order-test.service` 保持 active；前台、`/admin/`、`/healthz`、`/readyz` 返回 HTTP 200，未认证访问 `/v1/authorization/me` 返回 HTTP 401。
- 运行影响：仅更新远端前后台静态产物；Go 二进制、systemd、Nginx 配置、PostgreSQL、Redis、迁移和业务数据均未修改，服务未重启，Nginx 未 reload。`nginx -t` 成功，仍有主机既存的重复 TLS protocol options 警告。
- 依赖：前后台按锁文件执行 `npm ci --no-audit --no-fund`，未升级依赖或执行强制漏洞修复；npm 提示 `esbuild@0.25.12` 的安装脚本尚未列入 `allowScripts`，生产构建仍成功。
- 回滚：被替换的注册器源码和部署前完整静态产物位于 `/opt/sechelper-auth-template/.deploy-backup-module-discovery-20260914T222921`；恢复该目录中的 `web/dist` 即可回退静态站点。

# Orders 示例实例启用（2026-09-14 22:48 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，业务服务 `order-test.service`。
- 构建：默认模式 Go 全量测试、Web/Admin 单元测试通过；`make example-orders-build` 完成 `example` 标签下的 Go 全量测试和 API 编译，订单 Application 测试通过；前后台 Vite 生产构建通过。
- 启用：后台 `web/admin/src/business/orders/admin-module.js` 已启用；API 使用 `-tags=example` 构建，通过受控业务运行时注册 `order:read`、`example-order` 资源和 `/v1/orders` 路由，路由同时要求 `admin:access` 与 `order:read`。
- 发布：原子替换 `/opt/sechelper-auth-template/bin/auth-template`、更新前后台静态产物并重启 `order-test.service`。服务为 active；`/healthz`、`/readyz`、`/admin/` 返回 HTTP 200，未认证访问 `/v1/orders` 由启用前的 HTTP 404 变为 HTTP 401，确认路由存在且认证保护生效。`/readyz` 为 200，启动 Manifest 同步成功。
- 数据：沿用 2026-09-14 20:51 已应用的 `api/migrations/business/orders/002_orders.sql`，本次未执行迁移，未修改 PostgreSQL、Redis、Nginx、运行配置或现有业务数据。
- 回滚：旧 API 二进制、旧订单前端声明和部署前完整静态产物位于 `/opt/sechelper-auth-template/.deploy-backup-orders-enabled-20260914T223934`。恢复该目录中的 `bin/auth-template` 与 `web/dist` 后重启 `order-test.service`；源码回滚时恢复备份的 `main.go` 和订单声明并删除新增的三个 `business_runtime*.go` 文件。

# 自动发现浏览器兼容修复（2026-09-15）

- 原因：Vite 会将 `import.meta.glob` 编译为静态模块表，但前端运行时条件判断错误地把发现表置空，导致即使当前会话拥有 `order:read`，后台也不显示订单入口。
- 修复：前后台注册器改用可被 Vite 正确替换、且不影响 Node 单元测试的环境判断；重新生成 Web/Admin 静态产物。未修改 API 二进制、数据库、Redis、Nginx 或 systemd。
- 浏览器验收：刷新已登录的管理端后，左侧显示“业务运营 → 订单查询（示例）”，当前权限包含 `order:read`；后台主页和订单路由资源返回成功。

# 二级菜单与电商订单列表部署（2026-09-15）

- 远程同步后台分组折叠组件、菜单分组模型、订单列表筛选界面和 Frest 风格统计卡片；`make architecture-check`、后台 9 项单元测试和远程 Vite 生产构建通过（50 个模块）。
- 浏览器验收确认“业务运营”二级分组可展开/收起，当前路由自动展开；“订单查询（示例）”进入后显示订单概览卡片、订单号搜索、状态筛选、刷新按钮和空状态。
- 仅更新后台静态产物，未修改 API、数据库、Redis、Nginx 或 systemd；回滚备份为 `/opt/sechelper-auth-template/.deploy-backup-admin-secondary-order-20260915T070509`。

# 分组折叠策略优化部署（2026-09-15）

- `业务运营` 改为不可折叠分组标题；业务节点支持递归两级折叠（订单 → 订单1 → 订单2）；其他分组默认继续保留折叠能力。
- 远程 `make architecture-check`、管理后台 6 项测试和 Vite 生产构建通过（50 个模块）；后台静态目录已更新，`order-test.service` active，`/admin/` 返回 HTTP 200。
- 回滚备份：`/opt/sechelper-auth-template/.deploy-backup-admin-collapse-config-20260915T071831`；未修改数据库、Redis、API 二进制、Nginx 或 systemd。

# 管理界面底色与边界阴影远程部署（2026-09-15 08:58 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`。
- 范围：仅同步获准的 `web/admin/src/style.css` 并重新生成 Admin 静态资源。原有颜色令牌及配色不变；页面根层、应用壳、主内容区保持既有页面底色；侧边栏右边界和顶部栏下边界增加低强度中性阴影。
- 验证：远程架构检查通过；Admin 测试 6 项通过；Vite 7.1.7 构建转换 51 个模块并成功。`order-test.service` 为 active，`/admin/`、`/healthz`、`/readyz` 均返回 HTTP 200。
- 运行影响：仅更新后台静态产物；API、PostgreSQL、Redis、Nginx、systemd、运行配置和业务数据均未修改；未重启服务或 reload Nginx。部署源为获准的本地工作树单文件，未在远端执行 Git 操作。
- 回滚：原样式与完整部署前静态目录位于 `/opt/sechelper-auth-template/.deploy-backup-global-layout-shadows-20260915T085632`。

# 后台嵌套菜单白屏修复部署（2026-09-15）

- 原因：订单模块切换为递归 `children` 导航后，远端仍使用旧的平面导航校验器，启动时抛出 `Invalid admin navigation in orders`，导致后台无法渲染。
- 修复：同步递归导航校验器及菜单渲染代码，远程架构检查、后台 6 项测试和 Vite 生产构建通过（50 个模块）；后台静态产物已更新。
- 浏览器验收：刷新后台后无白屏，显示“业务运营 → 订单 → 订单1 → 订单2”；服务保持 active，未修改 API、数据库、Redis、Nginx 或 systemd。
- 回滚备份：`/opt/sechelper-auth-template/.deploy-backup-admin-bugfix-20260915T073000`。

# 同级分组不可折叠部署（2026-09-15）

- 同步分组模型：`账号与安全`、`监控与审计`、`业务运营`仅作为同级标题展示；订单内部的 `订单`、`订单1`继续支持折叠，`订单2`为页面入口。
- 远程架构检查、管理后台 6 项测试和 Vite 生产构建通过（50 个模块）；`order-test.service` active，`/admin/` 与 `/readyz` 返回 HTTP 200。
- 回滚备份：`/opt/sechelper-auth-template/.deploy-backup-admin-sibling-static-20260915T072345`；未修改 API、数据库、Redis、Nginx 或 systemd。

# 管理界面底色统一与交界阴影部署（2026-09-15 08:57 Asia/Shanghai）

- 范围：获准的框架文件 `web/admin/src/style.css`。管理端根节点、应用壳、主内容区和 body 统一沿用既有 现有页面背景变量，没有修改主题颜色令牌；固定侧边栏右边缘增加低强度中性投影，sticky 页眉下缘增加低强度分隔阴影。
- 远程目标：`root@47.116.4.57:/opt/sechelper-auth-template`，域名 `order-test.sechelper.com`。仅同步该 CSS 文件并重建 Admin 静态资源；本次使用当前获准工作树样式文件，未在本地运行构建，也未在远端使用 Git。
- 验证：远程 `make architecture-check`、Admin 测试通过（6 项）、Vite 生产构建通过（51 modules）；`order-test.service` 保持 active；`/admin/`、`/healthz`、`/readyz` 返回 HTTP 200。
- 未修改：API 二进制、PostgreSQL、Redis、Nginx 配置、systemd、运行配置和业务数据；无服务重启或 Nginx reload。
- 回滚：部署前后台静态产物备份位于 `/opt/sechelper-auth-template/.deploy-backup-global-layout-shadows-20260915T085632/web/dist`；原 CSS 位于同目录 `web/admin/src/style.css`。

# 页眉主内容交界阴影复修部署（2026-09-15 09:20 Asia/Shanghai）

- 原因：浏览器计算样式和远端源码确认仍引用旧页眉样式（`--frest-menu-bg` 与低强度 box-shadow），此前版本未正确进入当前线上静态 bundle。
- 修复：同步获准的 `web/admin/src/style.css`；页眉和主内容统一使用既有 现有页面背景变量，页眉交界 `box-shadow: none`。颜色令牌未变。
- 验证：远程架构检查、Admin 测试 6 项、Vite 生产构建（51 modules）通过；服务 active，`/admin/` 与 `/readyz` 返回 HTTP 200。浏览器刷新后页眉和主内容计算背景均为 `rgb(245, 245, 249)`，页眉阴影为 `none`。
- 未修改：API、数据库、Redis、Nginx、systemd、运行配置和业务数据；无服务重启或 Nginx reload。
- 回滚：部署前样式与静态目录备份位于 `/opt/sechelper-auth-template/.deploy-backup-header-seam-fix-20260915T092011`。

# 业务运营分组临时置于监控审计之前（2026-09-15 09:04 Asia/Shanghai）

- 范围：获准框架文件 `web/admin/src/app/AdminApp.jsx`。组合导航数组时将业务模块导航直接插入 `监控与审计` 分组起始位置，即 `账号与安全 → 业务运营 → 监控与审计`；未增加通用自定义排序字段或排序机制。
- 远程目标：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`。仅同步获准的 AdminApp 源码并重建后台静态资源；以当前获准工作树文件部署，本次未执行本地构建或 Git 操作。
- 验证：远程架构检查通过；Admin 测试 6 项通过；Vite 7.1.7 构建转换 51 个模块成功。`order-test.service` active，`/admin/` 和 `/readyz` 返回 HTTP 200。
- 未修改：API、数据库、Redis、Nginx、systemd、运行配置及业务数据；未重启服务或 reload Nginx。
- 回滚：部署前 AdminApp 文件和完整后台静态目录备份位于 `/opt/sechelper-auth-template/.deploy-backup-business-before-monitor-20260915T090423/web/`。

# 侧边栏底部折叠与悬停展开部署（2026-09-15）

- 折叠按钮移至侧边栏底部；收起状态宽度为 80px，鼠标移入自动展开至 260px，鼠标移出自动恢复收起，主内容区边距同步调整。
- 远程架构检查、管理后台 6 项测试和 Vite 生产构建通过（50 个模块）；浏览器实测底部按钮、悬停展开和移出隐藏均正常。
- 未修改 API、数据库、Redis、Nginx 或 systemd。回滚备份：`/opt/sechelper-auth-template/.deploy-backup-sidebar-hover-20260915T072852`。

# 框架注册与生产产物隔离验证（2026-09-15）

- 验证方式：将当前工作树中排除 `.git/`、根目录 `.env*`、`config.yaml`、`node_modules`、缓存、构建产物及 macOS `._*` 元数据的源码快照传至测试主机临时目录；Go 缓存、编译二进制和前端 `dist` 均在临时目录中。验证进程退出时由钩子清理临时目录；未更新 `/opt/sechelper-auth-template`，未部署或重启 `order-test.service`，未连接或操作数据库、Redis、Nginx及业务数据。
- 后端：Go 默认模式和 `-tags=example` 全量测试均通过；默认服务/迁移器和 `example` 服务/迁移器均成功编译。Go 测试与 `example` 案例测试代码不进入默认服务构建。
- 前端：Web 测试 5 项、Admin 测试 15 项通过。Admin Vite production 构建为 50 个模块，产物未含 `/admin/component-reference` 或 `/admin/orders`；test 构建为 55 个模块，含两个测试/案例路由；development 构建为 52 个模块，不含组件参考路由。Web production 构建为 37 个模块。
- 结果：生产与非生产共用构建入口、配置载入代码和 API/迁移程序入口；由构建模式和 Go build constraints 排除测试/案例模块。未构建容器镜像，未执行 Compose 生命周期操作。
- 提示：`npm ci --no-audit --no-fund` 提示锁定的 `esbuild@0.25.12` 安装脚本尚未列入 `allowScripts`；构建成功。本次未修改依赖或锁文件。
- 架构检查：修正 `deploy/scripts/check-boundaries.sh`，改为扫描存在的 `web/admin/src/app`，排除测试夹具，并在扫描路径缺失或 ripgrep 发生扫描错误时明确失败。修正后 `make architecture-check` 与 `git diff --check` 均通过；框架反向依赖业务、业务绕过基础设施及默认宿主引用订单案例均无命中。
- OpenAPI：测试机 npm registry 探测返回 HTTP 200，但临时快照中的 `npx --package=@redocly/cli@1.34.0` 在 180 秒限定时间内未完成，合同 lint 未验证。临时源文件与 npm 缓存目录已清理；未修改依赖、检查脚本或正式环境。

# 平台总览页面远程部署（2026-09-15 22:08 Asia/Shanghai）

- 目标：`root@47.116.4.57:/opt/sechelper-auth-template`，测试域名 `order-test.sechelper.com`。
- 范围：仅同步 `web/admin/src/modules/dashboard/DashboardPage.jsx` 与 `web/admin/src/style.css`，远程执行 Admin 测试及 Vite production 构建，并更新 `/opt/sechelper-auth-template/web/dist/admin`。
- 验证：Admin 测试 15 项通过；Vite 7.1.7 production 构建转换 50 个模块成功；后台 JS/CSS、`/admin/`、`/healthz`、`/readyz` 返回 HTTP 200，`order-test.service` 保持 active；未认证访问 `/v1/admin/dashboard/overview` 返回预期 HTTP 401。远程 `make architecture-check` 未完成：主机缺少 `rg`，退出码 127；这是扫描工具缺失，不能据此判定边界检查通过或发现违规。
- 未修改：API 二进制、PostgreSQL、Redis、Nginx、systemd、运行配置及业务数据；未重启服务或 reload Nginx。
- 回滚：远端原始两份源码及完整后台静态产物备份位于 `/opt/sechelper-auth-template/.deploy-backup-dashboard-overview-20260915T221000/`；替换前原静态目录另保留于 `/opt/sechelper-auth-template/web/dist/.admin-previous-dashboard-overview-20260915T221000`。
