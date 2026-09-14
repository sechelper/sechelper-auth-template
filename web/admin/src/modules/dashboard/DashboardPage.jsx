import { useEffect, useState } from "react";
import { dashboardApi } from "./api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";
import { navigate } from "../../app/navigation.jsx";
import { request } from "../auth/api.js";

const dependencyLabels = { api: "API 服务", postgres: "PostgreSQL", redis: "Redis", manifest: "权限 Manifest" };
const metricLabels = { authLoginTotal: "登录总数", authCallbackTotal: "回调总数", authRefreshTotal: "刷新总数", authLogoutTotal: "退出总数", rateLimitedTotal: "限流总数", remoteRevokeErrorsTotal: "远端撤销错误" };
const statusLabels = { healthy: "正常", degraded: "降级", unavailable: "不可用", not_configured: "未配置", not_synced: "未同步" };
const dependencyOrder = ["api", "postgres", "redis", "manifest"];
const statusLabel = (status) => statusLabels[status] || status || "未知";
const statusClass = (status) => status === "healthy" ? "health-ok" : status === "not_configured" ? "health-muted" : "health-warning";
const formatDateTime = (value) => value ? new Date(value).toLocaleString("zh-CN", { dateStyle: "medium", timeStyle: "short" }) : "—";
const formatUptime = (value) => {
  if (!value) return "—";
  const elapsed = Math.max(0, Date.now() - new Date(value).getTime());
  const minutes = Math.floor(elapsed / 60000);
  const days = Math.floor(minutes / 1440);
  const hours = Math.floor((minutes % 1440) / 60);
  const rest = minutes % 60;
  return `${days ? `${days} 天 ` : ""}${hours ? `${hours} 小时 ` : ""}${rest} 分钟`;
};

function HealthCard({ name, value }) {
  return <article className="card health-card"><div className="health-card-title"><span className={`health-dot ${statusClass(value.status)}`} /><strong>{dependencyLabels[name] || name}</strong><span className={`status-pill ${statusClass(value.status)}`}>{statusLabel(value.status)}</span></div><p>{value.message || (value.latencyMs !== undefined ? `响应 ${value.latencyMs} ms` : "当前进程可用")}</p></article>;
}

export function DashboardPage({ session }) {
  const [data, setData] = useState(null); const [operations, setOperations] = useState(null); const [error, setError] = useState(null); const [operationsError, setOperationsError] = useState(null); const [busy, setBusy] = useState(false);
  const [accessForm, setAccessForm] = useState({ resourceType: "", resourceId: "", action: "read" });
  const [accessResult, setAccessResult] = useState(null); const [accessError, setAccessError] = useState(null); const [accessBusy, setAccessBusy] = useState(false);
  const load = async () => {
    setError(null); setOperationsError(null);
    const [dashboardResult, operationsResult] = await Promise.allSettled([dashboardApi.overview(), request("/v1/admin/operations/overview")]);
    if (dashboardResult.status === "fulfilled") setData(dashboardResult.value.data); else setError(dashboardResult.reason);
    if (operationsResult.status === "fulfilled") setOperations(operationsResult.value.data); else setOperationsError(operationsResult.reason);
  };
  useEffect(() => { load(); }, []);
  const refresh = async () => { setBusy(true); try { await load(); } finally { setBusy(false); } };
  const checkAccess = async (event) => { event.preventDefault(); setAccessBusy(true); setAccessError(null); setAccessResult(null); try { const value = await request("/v1/admin/access-decisions/check", { method: "POST", body: JSON.stringify(accessForm) }); setAccessResult(value.data); } catch (value) { setAccessError(value); } finally { setAccessBusy(false); } };
  if (error) return <><PageHeader code="PLATFORM" title="平台总览" description="查看当前应用的认证、授权和运行状态。" /><ErrorState error={error} retry={() => { setError(null); load(); }} /></>;
  if (!data) return <Loading text="正在加载平台状态…" />;

  const dependencies = dependencyOrder.map((name) => [name, data.dependencies?.[name] || { status: "unknown" }]);
  const app = data.application || {}; const manifest = data.manifest || {};
  const currentSession = session?.status === "authenticated" ? session : data.currentSession || {};
  const permissions = session?.permissions || [];
  const operationsDependencies = Object.entries(operations?.dependencies || {});
  const unhealthy = dependencies.filter(([, value]) => value.status !== "healthy" && value.status !== "not_configured").length;

  return <><PageHeader code="PLATFORM OVERVIEW" title="平台总览" description="这里集中展示认证业务框架的会话、访问控制和系统运行状态。" /><section className="platform-hero card"><div><span className="eyebrow">APPLICATION</span><h2>{app.name || "统一认证业务框架"}</h2><p>{app.environment || "unknown"} 环境 · 当前版本 {app.version || operations?.application?.version || "—"}</p></div><div className={unhealthy ? "overview-status warning" : "overview-status"}><span className="health-dot" />{unhealthy ? `${unhealthy} 项需要关注` : "系统运行正常"}</div></section>

    <div className="dashboard-grid platform-grid platform-summary-grid"><section className="card platform-card"><div className="card-heading"><div><span className="eyebrow">AUTHENTICATION</span><h2>当前会话</h2></div><span className="card-mark">01</span></div><div className="detail"><div className="detail-row"><span>Subject</span><strong className="mono">{currentSession.subject || "—"}</strong></div><div className="detail-row"><span>Email</span><strong>{currentSession.email || "未获取"}</strong></div><div className="detail-row"><span>Application</span><strong>{currentSession.applicationCode || "—"}</strong></div><div className="detail-row"><span>有效期</span><strong>{formatDateTime(currentSession.expiresAt)}</strong></div></div><div className="permissions"><span>当前权限</span><div>{permissions.length ? permissions.map((item) => <code key={item}>{item}</code>) : <span>暂无权限</span>}</div></div></section><section className="card platform-card"><div className="card-heading"><div><span className="eyebrow">SYSTEM STATUS</span><h2>系统状态</h2></div><span className={`status-pill ${unhealthy ? "health-warning" : "health-ok"}`}>{unhealthy ? "需关注" : "正常"}</span></div><div className="detail"><div className="detail-row"><span>运行时长</span><strong>{formatUptime(app.startedAt)}</strong></div><div className="detail-row"><span>健康依赖</span><strong>{dependencies.length - unhealthy}/{dependencies.length} 项</strong></div><div className="detail-row"><span>应用版本</span><strong>{operations?.application?.version || app.version || "—"}</strong></div><div className="detail-row"><span>最近检查</span><strong>{formatDateTime(new Date().toISOString())}</strong></div></div></section><section className="card platform-card"><div className="card-heading"><div><span className="eyebrow">AUTHORIZATION</span><h2>Manifest 状态</h2></div><span className="card-mark">03</span></div><div className="detail"><div className="detail-row"><span>同步状态</span><strong>{statusLabel(manifest.status)}</strong></div><div className="detail-row"><span>Manifest 版本</span><strong>{manifest.version || "—"}</strong></div><div className="detail-row"><span>服务端修订</span><strong>{manifest.serverRevision || "—"}</strong></div></div><button className="button-quiet platform-action" onClick={() => navigate("/admin/manifest")}>查看权限清单 →</button></section></div>

    <section className="overview-section"><div className="section-heading"><div><span className="eyebrow">RUNTIME HEALTH</span><h2>系统依赖与运行指标</h2></div><button className="text-button" onClick={refresh} disabled={busy}>{busy ? "检查中…" : "重新检查"}</button></div>{operationsError ? <div className="card overview-inline-error"><ErrorState error={operationsError} retry={load} /></div> : <><section className="health-grid">{(operationsDependencies.length ? operationsDependencies : dependencies).map(([name, value]) => <HealthCard key={name} name={name} value={value} />)}</section><section className="card detail overview-metrics"><div className="detail-row"><span>系统运行状态</span><strong>{operations?.application?.environment || app.environment || "—"}</strong></div>{Object.entries(operations?.metrics || {}).map(([key, value]) => <div className="detail-row" key={key}><span>{metricLabels[key] || key}</span><strong>{value}</strong></div>)}</section></>}</section>

    <section className="card overview-access-card"><div className="card-heading"><div><span className="eyebrow">ACCESS CONTROL</span><h2>访问控制诊断</h2><p>检查当前管理员对指定资源执行动作时的授权决策，不会绕过业务 API。</p></div><span className="card-mark">04</span></div><form onSubmit={checkAccess}><label>资源类型<input value={accessForm.resourceType} onChange={(event) => setAccessForm({ ...accessForm, resourceType: event.target.value })} placeholder="例如 document" required /></label><label>资源标识<input value={accessForm.resourceId} onChange={(event) => setAccessForm({ ...accessForm, resourceId: event.target.value })} placeholder="例如 resource-001" required /></label><label>动作<input value={accessForm.action} onChange={(event) => setAccessForm({ ...accessForm, action: event.target.value })} placeholder="例如 read" required /></label><button disabled={accessBusy}>{accessBusy ? "检查中…" : "检查访问权限"}</button></form>{accessError && <ErrorState error={accessError} retry={() => setAccessError(null)} />}{accessResult && <div className={`decision-result ${accessResult.allowed ? "allowed" : "denied"}`}><strong>{accessResult.allowed ? "允许访问" : "拒绝访问"}</strong><span>原因：{accessResult.reasonCode}</span>{accessResult.permission && <span>所需权限：{accessResult.permission}</span>}</div>}</section><section className="card platform-notice"><div><span className="eyebrow">PLATFORM NOTE</span><h2>业务数据位于业务模块</h2><p>平台总览集中展示平台状态；订单等业务指标请从左侧业务模块进入。</p></div><button onClick={() => navigate("/admin/orders")}>进入订单模块</button></section></>;
}
