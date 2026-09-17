# 授权与 Manifest

本文档只描述运行时授权行为和 Manifest 同步语义。框架边界、维护流程、变更门槛、测试门禁和回滚规则统一维护在[授权框架维护](authorization-framework-maintenance.md)，不要在业务模块文档中复制这些规则。

业务模块通过 Manifest Registry 声明 `permission_code`、风险等级和 API 绑定。权限码使用业务命名空间，例如 `order:read`。

## 当前执行链

`GET /v1/authorization/me` 已经接入认证会话和授权模块。授权模块先读取 Session，再读取授权缓存；缓存未命中时从 Session 中恢复登录时取得的 Application 权限集合。业务接口应通过 `RequirePermission("<permission>")` 中间件执行权限检查。

缺少会话返回 `401 UNAUTHORIZED`，已登录但缺少权限返回 `403 FORBIDDEN`。授权上下文通过模块内部 Context 注入 Handler，业务模块不应读取 Session Store 或认证客户端。

统一认证平台的授权响应是 Application 级授权。租户、项目、资源和状态规则仍由业务模块在 use case 中判断。

本项目不管理用户或角色。平台授权模块提供当前 Subject、Application Code 和 Application Permission；每个业务模块负责自己的资源策略和数据范围。通用资源注册与访问诊断接口为 `POST /v1/admin/access-decisions/check`，只解释当前管理员对已注册资源动作的决策，不返回资源内容，也不绕过业务 API。

授权失败默认拒绝。授权查询失败、响应结构无效、Application Code 不匹配或权限缓存不可用时，不应执行业务 Handler。

Manifest 同步接口为：

- `GET /v1/internal/authorization-manifest`，需要 `auth:manifest:read`；
- `POST /v1/internal/authorization-manifest/sync`，需要 `auth:manifest:sync`。

服务启动时先执行一次同步，运行期间按 `MANIFEST_SYNC_INTERVAL` 定时同步。同步使用内容哈希和 `Idempotency-Key`，只有成功保存远端回执后才更新本地状态。Manifest 内容变化时会清理对应 Application 的授权缓存。

开发环境默认使用内存授权缓存；部署环境通过 `.env` 注入 `REDIS_URL` 后使用 Redis 授权缓存。生产环境必须配置 Redis。Redis 使用按 Session 的缓存 Key 和按 Application 的索引集合，Manifest 变化时可以批量清理对应 Application 的授权缓存。

## 相关文档

- [授权框架维护](authorization-framework-maintenance.md)：框架代码地图、注册契约、维护授权、验证和回滚。
- [业务模块开发模板](business-module-template.md)：业务模块如何声明权限、资源、审计和受保护路由。
- [OpenAPI 合同](contracts/openapi.yaml)：路由、字段、状态码和错误响应的机器可读定义。
