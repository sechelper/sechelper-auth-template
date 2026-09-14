import { useEffect, useState } from "react";
import { request } from "../auth/api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";
import { DataTable, EmptyState, RefreshButton } from "../../shared/ui.jsx";

const eventLabels = { AUTH_LOGIN_SUCCESS: "登录成功", AUTH_LOGIN_FAILED: "登录失败", AUTH_LOGOUT: "退出登录", AUTH_SESSION_REVOKED: "会话撤销", RESOURCE_ACCESS_DENIED: "资源访问拒绝", RESOURCE_SCOPE_DENIED: "资源范围拒绝", ACCESS_DECISION_CHECKED: "访问决策诊断", MANIFEST_SYNC_SUCCEEDED: "Manifest 同步成功", MANIFEST_SYNC_FAILED: "Manifest 同步失败", MANIFEST_CHANGE_DETECTED: "Manifest 发生变化" };

export function AuditPage() {
  const [rows, setRows] = useState(null); const [error, setError] = useState(null); const [busy, setBusy] = useState(false); const [filters, setFilters] = useState({ eventType: "", outcome: "" });
  const load = () => { const params = new URLSearchParams({ "page[size]": "50" }); if (filters.eventType) params.set("filter[eventType]", filters.eventType); if (filters.outcome) params.set("filter[outcome]", filters.outcome); return request(`/v1/admin/audit-events?${params}`).then((value) => setRows(value.data || [])).catch(setError); };
  useEffect(() => { load(); }, [filters.eventType, filters.outcome]);
  const refresh = async () => { setBusy(true); setError(null); try { await load(); } finally { setBusy(false); } };
  if (error && !rows) return <><PageHeader code="SECURITY" title="操作审计" description="只查看关键认证、授权、资源和平台操作。" /><ErrorState error={error} retry={() => { setError(null); load(); }} /></>;
  if (!rows) return <Loading text="正在加载操作审计…" />;
  return <><PageHeader code="SECURITY" title="操作审计" description="这里只记录关键安全和平台操作，不记录普通 HTTP 请求或页面访问。" /><div className="audit-filters"><select aria-label="事件类型" value={filters.eventType} onChange={(event) => setFilters({ ...filters, eventType: event.target.value })}><option value="">全部事件</option>{Object.entries(eventLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select><select aria-label="操作结果" value={filters.outcome} onChange={(event) => setFilters({ ...filters, outcome: event.target.value })}><option value="">全部结果</option><option value="success">成功</option><option value="failure">失败</option><option value="denied">拒绝</option></select><RefreshButton onClick={refresh} busy={busy} /></div><section className="card">{rows.length ? <DataTable rows={rows} columns={[{ key: "occurredAt", label: "时间", render: (row) => new Date(row.occurredAt).toLocaleString() }, { key: "eventType", label: "事件", render: (row) => eventLabels[row.eventType] || row.eventType }, { key: "actorSubject", label: "操作者", render: (row) => row.actorSubject || "系统" }, { key: "resourceId", label: "资源", render: (row) => row.resourceId || "—" }, { key: "outcome", label: "结果", render: (row) => <span className={`tag audit-${row.outcome}`}>{row.outcome}</span> }, { key: "reasonCode", label: "原因", render: (row) => row.reasonCode || "—" }, { key: "requestId", label: "Request ID", render: (row) => <span className="mono">{row.requestId || "—"}</span> }]} /> : <EmptyState title="暂无关键操作" description="当前筛选条件下没有审计事件。" />}</section></>;
}
