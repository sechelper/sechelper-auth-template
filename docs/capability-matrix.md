# 能力核验矩阵

本矩阵依据当前工作树中的代码、迁移、测试和部署文件整理。`部分实现` 表示代码路径存在但仍缺少完成定义要求的生产证据；`未验证` 表示当前没有可复现的证据，不代表功能不存在。

| 能力 | 当前状态 | 代码位置 | 测试证据 | 生产风险 | 后续动作 |
|---|---|---|---|---|---|
| Confidential Client OAuth/OIDC、PKCE、State/Nonce、ID Token | 部分实现 | `api/internal/platform/identity/`、`api/internal/modules/authentication/` | identity/authentication 单元测试通过 | 缺少真实 IdP 回调和线上 E2E 证据 | 增加隔离 IdP 合约与回调失败矩阵 |
| PostgreSQL Session Store | 部分实现 | `api/internal/platform/session/session.go`、`api/migrations/001_authentication.sql` | Memory Store 版本冲突测试通过 | 未在本次运行中连接数据库验证 | 运行专用 Compose 数据库集成测试 |
| Refresh Token Rotation/并发控制 | 已实现原子轮换，待真实多副本验证 | `api/internal/modules/authentication/application/service.go`、`api/internal/platform/session/session.go` | Rotation、invalid_grant 撤销、全量 Go 测试通过 | 真实 PostgreSQL 多副本和 IdP 行为尚未在本轮运行 | 使用专用 PostgreSQL/IdP 执行并发与恢复测试 |
| Host/Origin/代理头策略 | 已实现，待远程验证 | `api/internal/platform/security/policy.go`、`api/cmd/server/main.go` | 新增未知 Host、可信转发 Host 测试通过 | 真实 Nginx 转发链和 IPv6 网段未验证 | 在测试域名执行 Nginx 回归并补 Forwarded-Proto 策略 |
| CSRF 与 Cookie 基线 | 部分实现 | `api/internal/modules/authentication/transport/http/handler.go` | 认证应用测试通过 | Cookie Domain/Path 与服务端生命周期尚无浏览器矩阵 | 增加 SameSite、Session Fixation、过期和重登录 E2E |
| Migration checksum 与并发锁 | 已实现，待数据库验证 | `api/cmd/migrate/main.go`、`api/migrations/` | 无数据库运行证据 | 尚无空库/增量/失败恢复演练 | 在专用 PostgreSQL 中执行全量与并发迁移测试 |
| Redis 授权缓存 | 部分实现 | `api/internal/modules/authorization/persistence/redis.go` | 授权应用单元测试通过 | Redis 故障 fail-closed 未完成多实例证据 | 增加 Redis 故障、失效广播和缓存击穿测试 |
| 资源级授权与租户隔离 | 部分实现 | `api/internal/modules/authorization/`、业务模块 Use Case | 授权资源单元测试通过 | 业务模块仍需补充租户/所有权过滤证据 | 将授权下沉到业务 Use Case，补越权测试 |
| Manifest 同步与审计 | 部分实现 | `api/internal/modules/manifest/`、`api/internal/modules/audit/` | manifest/audit 单元测试通过 | 缺少高风险差异预览、回滚和线上告警验证 | 补版本差异、同步互斥、失败补偿和告警 |
| 统一 API 错误/分页/并发协议 | 部分实现 | 各模块 `transport/http`、`docs/contracts/openapi.yaml` | 无完整 Contract Test | 业务分页仍需统一边界，OpenAPI 一致性未自动检查 | 建立统一 DTO、游标/排序/筛选和契约校验 |
| 前台/后台界面 | 部分实现 | `web/`、`web/admin/` | 两端独立 Vite 生产构建通过 | 浏览器响应式、键盘和 WCAG 未在本次本机运行 | 执行 Playwright 桌面/平板/移动矩阵 |
| 指标、日志、健康检查 | 部分实现 | `api/internal/platform/metrics/`、`logging/`、`api/cmd/server/main.go` | Go vet/test 通过；指标生产鉴权新增测试通过 | Trace、告警接入和线上敏感字段审计未验证 | 补 OpenTelemetry、脱敏测试和告警演练 |
| Docker、Nginx、远程部署与回滚 | 部分实现 | `deploy/`、`docs/deployment.md`、`docs/testing-environment.md` | 历史记录有远程验证，但非本轮重新执行 | 本轮未重新部署；外部凭据和线上变更尚未授权/验证 | 先执行远程只读预检，再按备份/回滚流程发布 |

## 本轮已验证

- Go 全量测试通过。
- Go `vet` 通过。
- `web` 生产构建通过。
- `admin` 生产构建通过。
- 新增 Host/Origin、可信转发 Host、指标 Bearer 鉴权测试通过。
- `git diff --check` 通过。

## 本轮未完成或阻塞

- 未执行数据库、Redis、Nginx 或远程服务操作。
- 未执行真实身份平台登录、生产域名回调和多实例并发测试。
- 未声明所有完成定义满足；矩阵中的 `部分实现` 仍是后续工作项。
