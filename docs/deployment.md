# 部署

## 环境矩阵

| 环境 | Go/React 业务服务 | PostgreSQL | Redis/授权缓存 | 代理 |
|---|---|---|---|---|
| development | 本地工具链 | Docker Compose | 内存或可选 Redis | Vite/本地代理 |
| test | 指定远程物理机原生工具链 | Docker Compose | Docker Compose | 主机 Nginx |
| production | Docker Compose | Docker Compose | Docker Compose | Compose 或既定代理 |

生产环境不得使用主机进程管理器直接运行 Go/React 业务服务。数据库和其他基础设施不得安装在物理机上运行。

## 本地

```bash
make toolchain-check
make build ENV=development
make run ENV=development
```

开发环境 Compose 使用 `deploy/config.dev.yaml` 和 `deploy/environments/development.env`。数据库结构由独立迁移命令管理，API 进程不会在启动时执行 DDL。`make run ENV=development ACTION=migrate` 执行迁移；`make run ENV=development` 启动应用栈。

## 版本与工具链

根目录 `release.yaml` 是应用版本和发布 Build ID 的唯一权威来源。`releaseVersion` 使用 SemVer；`releaseBuildId` 使用 UTC `YYYYMMDDHHmmss`。开始一个新发布时运行 `make release-id` 分配并写入新的 Build ID；该命令串行分配 ID，开发构建不会改写发布值。更改 `releaseVersion` 后运行 `make release-sync`，将前台和后台 npm 包清单/锁文件中的必需版本字段同步到权威版本。`make release-check` 校验 YAML 字段、格式和同步结果。

OpenAPI 文件中的 `info.version` 表示 API 合同版本，独立于运行时应用版本。

`toolchain.versions` 是精确工具链清单：Go 1.26.8、Node.js 24.21.0、npm 11.19.1、React/react-dom 19.2.7、Vite 7.1.7。`api/go.mod` 的 `go` 指令定义最低 Go 语言/模块版本；`toolchain.versions` 单独固定实际编译器版本，若 `go.mod` 有 `toolchain` 指令，检查器也会校验它。`.node-version`、npm 包 `engines`/`packageManager`、CI 与 Docker 构建阶段由 `make toolchain-check` 检查。CI 和 Docker 安装依赖前验证并固定 npm 版本；前端使用提交的 `package-lock.json` 和 `npm ci`。

唯一构建入口是 `make build ENV=<development|test|production> [COMPONENT=all|api|web]`，运行入口是 `make run ENV=<development|test|production>`。入口解析应用版本、Build ID 和源代码修订号一次，再将相同值传给 API、前台和后台构建。development 使用 `dev-local` Build ID；test/production 必须使用 `release.yaml` 中的 Build ID、干净的已提交源码和有效版本，不能用 Git SHA 或时间戳替代应用版本/Build ID。`release.yaml` 的 Build ID 必须随发布源码提交，test 与 production 构建同一发布时使用相同 ID。

构建产物包括 API/迁移 Docker 镜像、含前台和后台静态资源的 Web 镜像，以及写入两个静态目录的 `build-info.json`。镜像标签使用 `<releaseVersion>-<releaseBuildId>-<environment>`，容器 OCI 标签记录应用版本、Build ID、源码修订号和环境；Go `/v1/version` 与后台运行信息也返回版本、Build ID 和源码修订号。构建成功后，入口在操作系统临时目录写入 artifact manifest，列出每个组件、镜像引用和共同构建元数据；CI 会上传 test/production 两份 manifest。部署使用的 `config.yaml`、密钥、日志和数据库状态仍由运行环境管理，不进入镜像构建上下文中的配置挂载。

构建和依赖安装的工作目录不写入仓库。Go 测试缓存、临时文件和 artifact manifest 放在操作系统临时目录；Docker 镜像本身由 Docker 管理。应用运行配置由 Compose 显式挂载的 `config.yaml` 路径选择，日志与运行状态遵循该配置和容器可写层约定。

数据库结构由独立迁移命令管理，API 进程不会在启动时执行 DDL。首次运行或发布新版本前执行：

```bash
make run ENV=production ACTION=migrate
```

生产容器中对应的可执行文件为 `/app/auth-template-migrate`；迁移完成后再启动 API。迁移会记录 SQL 文件 checksum，已执行迁移被修改时会失败，必须新增前向迁移文件。

当前 Session Store 使用 PostgreSQL；生产环境授权缓存使用 Redis。测试环境可根据 `REDIS_URL` 选择共享 Redis，开发环境留空时使用内存缓存。

生产环境 API 启动前必须提供 `METRICS_TOKEN`。`/metrics` 不公开访问，只接受 `Authorization: Bearer <METRICS_TOKEN>`；反向代理来源网段通过 `TRUSTED_PROXY_CIDRS` 显式声明，应用只从这些网段信任 `X-Forwarded-Host`。

## 配置

应用配置字段见 `config.example.yaml`，Docker/Compose 变量见 `.env.example`。真实配置由 `config.yaml` 或部署环境注入；Identity Client Secret、access token 和数据库凭据不进入 Git、镜像或前端构建产物。Manifest 同步复用 `IDENTITY_CLIENT_ID/IDENTITY_CLIENT_SECRET` 获取 Bearer access token，不再配置独立的 Manifest 凭据。

## 前后台发布

前台位于 `web/`，构建到 `web/dist`；后台位于 `web/admin/`，构建到 `web/admin/dist`。两个目录仍是独立 Vite 项目，各自执行依赖安装和生产构建。生产镜像由 `deploy/Dockerfile.web` 将两个构建产物分别复制到 `/usr/share/nginx/html` 和 `/usr/share/nginx/html/admin`。Nginx 必须将 `/admin/` 回退到后台 `index.html`，根路径回退到前台 `index.html`。

后台前端与前台共用 Go API 和 Session Cookie。后台入口要求 `admin:access`；订单和 Manifest API 继续由 Go 授权中间件进行最终权限校验。

## 验收

```bash
curl --fail-with-body "$API_ORIGIN/healthz"
curl --fail-with-body "$API_ORIGIN/readyz"
curl -i "$API_ORIGIN/v1/auth/session"
curl --fail-with-body -H "Authorization: Bearer $METRICS_TOKEN" "$API_ORIGIN/metrics"
```

状态变更测试必须使用隔离测试账号和测试域名。生产环境不得执行测试账号、测试凭据或破坏性 smoke test。
