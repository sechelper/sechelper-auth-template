# 关键操作审计

## 定位

`audit_events` 是安全和业务操作审计，不是 HTTP 访问日志。普通请求、页面访问、Dashboard 加载、健康检查和普通资源读取不写入审计表；它们由 Zap 技术日志和 Metrics 负责。

## 事件范围

当前事件类型由 `api/internal/modules/audit/application/service.go` 的白名单维护，覆盖认证、Session、授权决策、Manifest、运维命令和资源导出。新增事件必须先进入白名单，并明确分类和严重级别。

关键事件包括：

- `AUTH_LOGIN_SUCCESS`、`AUTH_LOGIN_FAILED`、`AUTH_LOGOUT`；
- `AUTH_SESSION_REVOKED`、`AUTH_SESSION_REVOKE_FAILED`、`AUTH_SESSION_REFRESH_FAILED`；
- `RESOURCE_ACCESS_DENIED`、`RESOURCE_SCOPE_DENIED`、`ACCESS_DECISION_CHECKED`；
- `MANIFEST_SYNC_SUCCEEDED`、`MANIFEST_SYNC_FAILED`、`MANIFEST_CHANGE_DETECTED`；
- `SECURITY_CONFIGURATION_CHANGED`、`BACKGROUND_JOB_RETRIED`、`BACKGROUND_JOB_CANCELLED`；
- `RESOURCE_EXPORT_STARTED`、`RESOURCE_EXPORT_COMPLETED`、`RESOURCE_EXPORT_FAILED`。

事件包含事件类型、分类、严重级别、结果、原因、操作者、Application、资源引用、动作、Request ID、来源和受控 Metadata。不得写入密码、Client Secret、Access Token、Refresh Token、Cookie 或数据库凭据。

## 写入原则

关键 Use Case 通过 `audit/application.Recorder` 显式写入审计。高频的普通成功读取不记录；越权、范围拒绝、敏感资源访问、导出和高风险管理命令才记录。审计写入和业务状态变更需要在后续高一致性场景中通过 Outbox 统一提交；当前小规模实现使用 PostgreSQL 同步写入。

## 查询

`GET /v1/admin/audit-events` 需要 `audit:read`，支持按事件类型、分类、严重级别、结果、操作者、资源和 UTC 时间范围筛选。后台页面 `/admin/audit-events` 只展示关键事件，审计记录本身不可通过业务接口修改或删除。
