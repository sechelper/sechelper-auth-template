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
cp config.example.yaml config.yaml
cp .env.example .env
docker compose -f deploy/compose.dev.yaml up -d postgres
cd api
APP_CONFIG_FILE=../config.yaml go run ./cmd/server
cd web && npm install && npm run dev
```

数据库结构由独立迁移命令管理，API 进程不会在启动时执行 DDL。首次运行或发布新版本前执行：

```bash
cd api
APP_CONFIG_FILE=../config.yaml go run ./cmd/migrate
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
