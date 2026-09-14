# 远程测试环境

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
- 可重复 Refresh Rotation 测试：`go test ./internal/modules/authentication/application -run TestRefreshRotatesTokenAndExtendsSessionWithoutWaitingForExpiry` 使用 Mock Identity Provider 固定返回两代 token，验证 Session 有效期更新、密文轮换和旧 token 拒绝；本地与远程全量 Go 测试均通过。
- 多实例认证基础设施：新增 `api/migrations/003_authentication_transactions.sql`，测试数据库已应用 `authentication_login_transactions` 和 `authentication_sessions.version`；远程 Go 全量测试通过，服务重启后 readiness 正常。
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
- 后台样式更新：2026-09-13 23:10 Asia/Shanghai 在远程主机完成 `admin/` 的 `npm ci` 和 Vite 生产构建，部署 Sneat Bootstrap FREE 风格样式到 `/admin`；`/admin/`、`/admin/orders`、新版 CSS 静态资源均返回 HTTP 200。备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-style-20260913T231020`。Go 服务、PostgreSQL、Redis 和 Nginx 配置未修改；Nginx 仅执行了配置验证，未重新加载。
- 左上角社区标识更新：2026-09-13 23:21 Asia/Shanghai 在远程主机同步 `admin/src/`、`admin/public/`、`admin/index.html`、`admin/package.json`、`admin/package-lock.json` 和 `admin/vite.config.js`，完成远程 `npm ci` 与 Vite 生产构建，并将产物同步到 `/opt/sechelper-auth-template/web/dist/admin`。`https://order-test.sechelper.com/admin/`、`/admin/assets/logo/logo.svg`、`/healthz`、`/readyz` 返回成功；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。本次 npm 审计仍报告 1 个 high severity vulnerability，未执行强制依赖升级。
- 错误状态页更新：2026-09-13 23:28 Asia/Shanghai 在远程主机同步 `admin/src/app/AdminApp.jsx`、`admin/src/style.css` 和 `admin/README.md`，由远程 Node 工具链完成 Vite 生产构建，并将后台产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。新增统一 403、404、500 页面及返回/重试操作；`/admin/` 与新版 CSS 返回 HTTP 200，`/healthz`、`/readyz` 返回 HTTP 200，`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。本次部署备份位于 `/opt/sechelper-auth-template/.deploy-backup-error-pages-20260913T232821`。
- 部署期间修复两处兼容性问题：身份平台端点校验错误限制 OAuth 路径；Manifest 状态持久化未写入现有数据库要求的 `canonical_json` 字段。对应代码已同步到本地源码和远程构建目录。
- 首次启动失败日志仍保留在远程 `/opt/sechelper-auth-template/logs/auth-template-debug.log` 和 `auth-template.jsonl`，最新启动已记录 `server.started`，服务当前正常。
- 后台主页重新部署：2026-09-14 00:00 Asia/Shanghai 在远程主机同步 `admin/src/`、`admin/public/`、`admin/index.html`、`admin/package.json`、`admin/package-lock.json`、`admin/vite.config.js` 和 `admin/README.md`，远程执行 `npm ci --no-audit --no-fund` 与 `npm run build`，将构建产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-home-20260914T000034`。`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
- 错误页面重新部署：2026-09-14 00:07 Asia/Shanghai 在远程主机同步 `admin/src/app/AdminApp.jsx`、`admin/src/style.css` 和 `admin/README.md`，远程执行 `npm ci --no-audit --no-fund` 与 `npm run build`，将 Sneat 风格 403/404/500 页面构建产物更新到 `/opt/sechelper-auth-template/web/dist/admin`。部署前备份位于 `/opt/sechelper-auth-template/.deploy-backup-admin-errors-20260914T000703`。远程 CSS 含 `error-screen` 且不再包含旧 `incident-screen`；`/admin/`、新版 JS/CSS、`/healthz`、`/readyz` 返回成功；`/v1/auth/session` 返回 `{"authenticated":false}`；未认证访问 `/v1/authorization/me` 返回 HTTP 401；`order-test.service` 保持 active。Go 服务、PostgreSQL、Redis、业务容器和 Nginx 配置未修改，未重启服务或重载 Nginx。
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
