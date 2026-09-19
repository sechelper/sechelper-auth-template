import { useEffect, useState } from "react";
import { dashboardApi } from "./api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";
import { request } from "../auth/api.js";
import { DeploymentStatusPanels } from "../deployment/DeploymentStatusPanels.jsx";
import { apiOrigin } from "../../platform/config/runtime.js";

const dependencyLabels = { api: "API 服务", postgres: "PostgreSQL", redis: "Redis", manifest: "权限 Manifest" };
const statusLabels = { healthy: "正常", degraded: "降级", unavailable: "不可用", not_configured: "未配置", not_synced: "未同步", consistent: "一致", applied: "已应用" };
const dependencyOrder = ["api", "postgres", "redis", "manifest"];
const statusLabel = (status) => statusLabels[status] || status || "未知";
const statusClass = (status) => ["healthy", "consistent", "applied"].includes(status) ? "health-ok" : status === "not_configured" ? "health-muted" : "health-warning";
const formatDateTime = (value) => {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime()) || date.getUTCFullYear() <= 1) return "—";
  return date.toLocaleString("zh-CN", { dateStyle: "medium", timeStyle: "short" });
};
const formatUptime = (value) => {
  if (!value) return "—";
  const elapsed = Math.max(0, Date.now() - new Date(value).getTime());
  const minutes = Math.floor(elapsed / 60000);
  const days = Math.floor(minutes / 1440);
  const hours = Math.floor((minutes % 1440) / 60);
  const rest = minutes % 60;
  return `${days ? `${days} 天 ` : ""}${hours ? `${hours} 小时 ` : ""}${rest} 分钟`;
};
const formatSampleTime = (value) => {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleTimeString("zh-CN", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });
};
const boundedPercent = (value) => typeof value === "number" && Number.isFinite(value) ? Math.min(100, Math.max(0, value)) : null;
const formatPercent = (value) => {
  const percent = boundedPercent(value);
  return percent === null ? "—%" : `${percent.toFixed(1)}%`;
};
const formatBytes = (value) => {
  if (!Number.isFinite(value) || value < 0) return "—";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let amount = value;
  let unit = 0;
  while (amount >= 1024 && unit < units.length - 1) {
    amount /= 1024;
    unit += 1;
  }
  return `${amount.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
};
const resourceReconnectDelayMs = 1000;
const resourceHistoryLimit = 60;
const emptyResourceHistory = () => ({ sampledAt: null, cpu: [], memory: [] });
const resourceSeries = [
  ["cpu", "cpuPercent"],
  ["memory", "memoryPercent"],
];

function appendResourceSnapshot(history, snapshot) {
  if (!snapshot?.sampledAt || snapshot.sampledAt === history.sampledAt) return history;
  const next = { ...history, sampledAt: snapshot.sampledAt };
  let appended = false;
  for (const [series, field] of resourceSeries) {
    const value = snapshot[field];
    if (typeof value !== "number" || !Number.isFinite(value)) continue;
    next[series] = [...history[series], { sampledAt: snapshot.sampledAt, value }].slice(-resourceHistoryLimit);
    appended = true;
  }
  return appended ? next : history;
}

function ResourceTrend({ label, points, maximum = 100, color }) {
  const width = 1000;
  const height = 42;
  const inset = 4;
  const plotWidth = width - inset * 2;
  const values = points.map((point) => point.value);
  const scale = maximum || Math.max(1, ...values);
  const coordinates = points.map((point, index) => ({
    x: points.length < 2 ? width - inset : inset + index / (points.length - 1) * plotWidth,
    y: height - inset - Math.min(scale, Math.max(0, point.value)) / scale * (height - inset * 2),
  }));
  const polyline = coordinates.map((point) => `${point.x},${point.y}`).join(" ");
  const last = coordinates.at(-1);
  const current = points.at(-1)?.value;
  const currentLabel = current === undefined ? "数据采集中" : maximum ? formatPercent(current) : current.toLocaleString("zh-CN");
  return <svg className="platform-resource-trend" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={`${label}近 30 秒趋势，当前${currentLabel}`} style={{ display: "block", width: "100%", height: "2.35rem", minWidth: 0, overflow: "visible" }}>
    <title>{label}趋势 · {currentLabel}</title>
    <line x1={inset} y1={height - inset} x2={width - inset} y2={height - inset} stroke="var(--sneat-border)" strokeDasharray="3 5" strokeWidth="1" />
    {coordinates.length > 0 && <polyline points={polyline} fill="none" stroke={color} strokeWidth="2.25" strokeLinecap="round" strokeLinejoin="round" />}
    {last && <><circle cx={last.x} cy={last.y} r="5.5" fill={color} opacity=".18" /><circle cx={last.x} cy={last.y} r="2.8" fill={color} /></>}
  </svg>;
}

function HealthCard({ name, value }) {
  const detail = value.message || (value.latencyMs !== undefined ? `响应 ${value.latencyMs} ms` : "当前进程可用");
  return <div className="platform-dependency">
    <div className="platform-dependency-main">
      <div className="platform-dependency-name"><span className={`health-dot ${statusClass(value.status)}`} /><strong>{dependencyLabels[name] || name}</strong></div>
      <span className={`status-pill ${statusClass(value.status)}`}>{statusLabel(value.status)}</span>
    </div>
    <small>{detail}</small>
  </div>;
}

const gaugeTicks = [
  { label: "0%", x1: 51, y1: 168, x2: 42, y2: 168, tx: 24, ty: 173 },
  { label: "10%", x1: 57, y1: 128, x2: 49, y2: 125, tx: 34, ty: 124 },
  { label: "20%", x1: 76, y1: 92, x2: 68, y2: 87, tx: 57, ty: 80 },
  { label: "30%", x1: 104, y1: 64, x2: 99, y2: 56, tx: 91, ty: 47 },
  { label: "40%", x1: 140, y1: 45, x2: 137, y2: 37, tx: 133, ty: 25 },
  { label: "50%", x1: 180, y1: 39, x2: 180, y2: 30, tx: 180, ty: 17 },
  { label: "60%", x1: 220, y1: 45, x2: 223, y2: 37, tx: 227, ty: 25 },
  { label: "70%", x1: 256, y1: 64, x2: 261, y2: 56, tx: 269, ty: 47 },
  { label: "80%", x1: 284, y1: 92, x2: 292, y2: 87, tx: 303, ty: 80 },
  { label: "90%", x1: 303, y1: 128, x2: 311, y2: 125, tx: 326, ty: 124 },
  { label: "100%", x1: 309, y1: 168, x2: 318, y2: 168, tx: 336, ty: 173 },
];

function ResourceGauge({ label, percent, detail, trend, color }) {
  const value = boundedPercent(percent);
  const firstSegment = value === null ? 0 : Math.min(value, 50);
  const secondSegment = value === null ? 0 : Math.max(value - 50, 0);
  const displayedValue = formatPercent(value);
  return <article className="platform-resource-gauge">
    <h3>{label}</h3>
    <p>{detail || "Linux 主机资源"}</p>
    <svg className="platform-resource-gauge__dial" role="img" aria-label={value === null ? `${label}暂无数据` : `${label}${displayedValue}`} viewBox="0 0 360 210">
      <path className="platform-gauge-track" d="M 60 168 A 120 120 0 0 1 300 168" />
      <path className="platform-gauge-low" pathLength="100" strokeDasharray={`${firstSegment} ${100 - firstSegment}`} d="M 60 168 A 120 120 0 0 1 300 168" style={{ transition: "stroke-dasharray 450ms ease" }} />
      <path className="platform-gauge-high" pathLength="100" strokeDasharray={`${secondSegment} ${100 - secondSegment}`} strokeDashoffset="-50" d="M 60 168 A 120 120 0 0 1 300 168" style={{ transition: "stroke-dasharray 450ms ease" }} />
      {gaugeTicks.map((tick) => <g key={tick.label}>
        <line className="platform-gauge-tick" x1={tick.x1} y1={tick.y1} x2={tick.x2} y2={tick.y2} />
        <text className="platform-gauge-label" x={tick.tx} y={tick.ty}>{tick.label}</text>
      </g>)}
      <text className="platform-gauge-value" x="180" y="163">{displayedValue}</text>
      <text className="platform-gauge-note" x="180" y="181">{value === null ? "数据暂不可用" : "实时使用率"}</text>
    </svg>
    <ResourceTrend label={label} points={trend} color={color} />
  </article>;
}

export function DashboardPage({ session }) {
  const [data, setData] = useState(null);
  const [resourceData, setResourceData] = useState(null);
  const [resourceHistory, setResourceHistory] = useState(emptyResourceHistory);
  const [resourceError, setResourceError] = useState(false);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [checkedAt, setCheckedAt] = useState(null);
  const [accessForm, setAccessForm] = useState({ resourceType: "", resourceId: "", action: "read" });
  const [accessResult, setAccessResult] = useState(null);
  const [accessError, setAccessError] = useState(null);
  const [accessBusy, setAccessBusy] = useState(false);

  const load = async () => {
    setError(null);
    try {
      const value = await dashboardApi.overview();
      setData(value.data);
      if (value.data.resources) {
        setResourceData(value.data.resources);
        setResourceHistory((current) => appendResourceSnapshot(current, value.data.resources));
      }
    } catch (value) {
      setError(value);
    } finally {
      setCheckedAt(new Date());
    }
  };

  useEffect(() => { load(); }, []);

  useEffect(() => {
    if (!data) return undefined;
    let active = true;
    let reconnectTimer;
    let socket;
    const connect = () => {
      if (!active) return;
      const origin = apiOrigin() || window.location.origin;
      const protocol = origin.startsWith("https:") ? "wss:" : "ws:";
      const wsOrigin = origin.replace(/^https?:/, protocol);
      socket = new WebSocket(`${wsOrigin}/v1/admin/dashboard/resources/ws`);
      socket.onopen = () => { if (active) setResourceError(false); };
      socket.onmessage = (event) => {
        if (!active) return;
        try {
          const value = JSON.parse(event.data)?.data;
          if (!value) return;
          setResourceData(value);
          setResourceHistory((current) => appendResourceSnapshot(current, value));
        } catch {
          setResourceError(true);
        }
      };
      socket.onerror = () => { if (active) setResourceError(true); };
      socket.onclose = () => {
        if (active) reconnectTimer = window.setTimeout(connect, resourceReconnectDelayMs);
      };
    };
    connect();
    return () => {
      active = false;
      window.clearTimeout(reconnectTimer);
      socket?.close();
    };
  }, [data]);

  const refresh = async () => {
    setBusy(true);
    try { await load(); } finally { setBusy(false); }
  };

  const checkAccess = async (event) => {
    event.preventDefault();
    setAccessBusy(true);
    setAccessError(null);
    setAccessResult(null);
    try {
      const value = await request("/v1/admin/access-decisions/check", { method: "POST", body: JSON.stringify(accessForm) });
      setAccessResult(value.data);
    } catch (value) {
      setAccessError(value);
    } finally {
      setAccessBusy(false);
    }
  };

  if (error) return <><PageHeader code="PLATFORM" title="平台总览" description="查看当前应用的系统状态、依赖与授权同步。" /><ErrorState error={error} retry={load} /></>;
  if (!data) return <Loading text="正在加载平台状态…" />;

  const app = data.application || {};
  const manifest = data.manifest || {};
  const resources = resourceData || data.resources || {};
  const dependencies = dependencyOrder.map((name) => [name, data.dependencies?.[name] || { status: "unknown" }]);
  const unhealthy = dependencies.some(([, value]) => !["healthy", "not_configured"].includes(value.status));
  const manifestReady = ["consistent", "applied"].includes(manifest.status);

  return <div className="platform-overview-page">
    <PageHeader code="PLATFORM OVERVIEW" title="平台总览" description="查看应用健康、依赖服务、服务器资源与授权同步情况。" />

    <section className="platform-app-banner card" aria-label="应用信息">
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "2rem", flexWrap: "wrap" }}>
        <div className="platform-app-copy">
          <span className="eyebrow">APPLICATION</span>
          <h2>{app.name || "—"}</h2>
        </div>
        <div className={`platform-app-state ${unhealthy ? "warning" : ""}`} role="status">
          <span className="health-dot" />{unhealthy ? "系统需关注" : "系统运行正常"}
        </div>
      </div>
      <div className="platform-system-facts">
        <div><span>运行时长</span><strong>{formatUptime(app.startedAt)}</strong></div>
        <div><span>最近检查</span><strong>{formatDateTime(checkedAt)}</strong></div>
      </div>
    </section>

    <section className="platform-dependencies-status" aria-label="依赖服务">
      <div className="platform-section-heading">
        <div><span className="eyebrow">DEPENDENCIES</span><h2>依赖服务</h2><p>查看 API、数据库、缓存和权限同步服务状态。</p></div>
        <span className={`status-pill ${unhealthy ? "health-warning" : "health-ok"}`}>{unhealthy ? "需要关注" : "全部正常"}</span>
      </div>
      <div className="platform-dependency-grid">
        {dependencies.map(([name, value]) => <HealthCard key={name} name={name} value={value} />)}
      </div>
    </section>

    <DeploymentStatusPanels />

    <div className="platform-lower-grid">
      <section className="card platform-resource-card" aria-labelledby="platform-resource-title">
        <div className="platform-section-heading">
          <div><span className="eyebrow">SERVER RESOURCES</span><h2 id="platform-resource-title">服务器资源</h2><p>{resources.sampledAt ? `最近采样 · ${formatSampleTime(resources.sampledAt)} · 保留近 30 秒趋势` : "等待首个资源快照"}</p></div>
          {resourceError && <span className="status-pill health-warning" role="status"><span className="health-dot health-warning" />正在重试</span>}
        </div>
        <div className="platform-resource-gauges">
          <ResourceGauge label="CPU 使用率" percent={resources.cpuPercent} trend={resourceHistory.cpu} color="var(--sneat-info)" />
          <ResourceGauge label="内存占用" percent={resources.memoryPercent} trend={resourceHistory.memory} color="var(--sneat-primary)" detail={resources.memoryUsedBytes != null && resources.memoryTotalBytes != null ? `${formatBytes(resources.memoryUsedBytes)} / ${formatBytes(resources.memoryTotalBytes)}` : undefined} />
        </div>
        <div className="platform-resource-rows">
          <div className="platform-resource-row">
            <span className="platform-resource-icon">▤</span>
            <div className="platform-resource-row-main"><span>磁盘使用率</span><strong>{formatPercent(resources.diskPercent)}</strong><small>项目所在盘 · {resources.diskUsedBytes != null && resources.diskTotalBytes != null ? `${formatBytes(resources.diskUsedBytes)} / ${formatBytes(resources.diskTotalBytes)}` : "数据暂不可用"}</small></div>
          </div>
          <div className="platform-resource-row">
            <span className="platform-resource-icon green">Go</span>
            <div className="platform-resource-row-main"><span>Goroutines</span><strong>{Number.isInteger(resources.goroutines) ? resources.goroutines.toLocaleString("zh-CN") : "—"}</strong><small>{resources.sampledAt ? "Go 进程协程数" : "等待资源采样"}</small></div>
          </div>
        </div>
      </section>

      <section className="card platform-manifest-access-card" aria-label="Manifest 状态与访问控制诊断">
        <div className="platform-section-heading platform-manifest-heading">
          <div><span className="eyebrow">AUTHORIZATION</span><h2>Manifest 状态</h2></div>
          <span className={`status-pill ${manifestReady ? "health-ok" : "health-warning"}`}>{statusLabel(manifest.status)}</span>
        </div>
        <dl className="platform-manifest-facts">
          <div><dt>Manifest 版本</dt><dd>{manifest.version || "—"}</dd></div>
          <div><dt>服务端修订</dt><dd>{manifest.serverRevision || "—"}</dd></div>
          <div><dt>最近更新时间</dt><dd>{formatDateTime(manifest.updatedAt)}</dd></div>
        </dl>
        <div className="platform-combined-divider" />
        <div className="platform-diagnostic-heading"><h3>访问控制诊断</h3><p>检查指定资源的授权决策，不会绕过业务 API。</p></div>
        <form className="platform-access-form" onSubmit={checkAccess}>
          <label>资源类型<input value={accessForm.resourceType} onChange={(event) => setAccessForm({ ...accessForm, resourceType: event.target.value })} placeholder="例如 document" required /></label>
          <label>资源标识<input value={accessForm.resourceId} onChange={(event) => setAccessForm({ ...accessForm, resourceId: event.target.value })} placeholder="例如 resource-001" required /></label>
          <label>动作<select value={accessForm.action} onChange={(event) => setAccessForm({ ...accessForm, action: event.target.value })}><option value="read">读取 read</option><option value="write">写入 write</option><option value="delete">删除 delete</option></select></label>
          <button disabled={accessBusy}>{accessBusy ? "检查中…" : "检查权限"}</button>
        </form>
        {accessError && <ErrorState error={accessError} retry={() => setAccessError(null)} />}
        {accessResult && <div className={`decision-result ${accessResult.allowed ? "allowed" : "denied"}`}><strong>{accessResult.allowed ? "允许访问" : "拒绝访问"}</strong><span>原因：{accessResult.reasonCode}</span>{accessResult.permission && <span>所需权限：{accessResult.permission}</span>}</div>}
      </section>
    </div>
  </div>;
}
