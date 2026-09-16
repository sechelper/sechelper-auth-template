import { useEffect, useState } from "react";
import { request } from "../auth/api.js";
import { ErrorState, Loading } from "../../app/components.jsx";
import { RefreshButton } from "../../shared/ui.jsx";

const labels = {
  database: "数据库 bootstrap 连接",
  encryptionKey: "配置中心加密密钥",
  runtimeLoaded: "运行时配置加载",
  databaseConnectivity: "数据库连通性",
  configurationTable: "配置中心数据表",
  configurationMigration: "配置中心迁移",
};
const statusLabels = { healthy: "正常", attention: "需要处理", unavailable: "不可用" };

function CheckList({ title, items = [] }) {
  const healthyCount = items.filter((item) => item.status === "healthy").length;
  return <section className="deployment-guide-card card" aria-label={title}>
    <div className="deployment-guide-card-heading"><h2>{title}</h2><span>{healthyCount}/{items.length} 正常</span></div>
    <div className="deployment-check-list">{items.map((item) => <div className="deployment-check" key={item.name}>
      <span className={`deployment-check-dot deployment-check-${item.status}`} aria-hidden="true" />
      <div><strong>{labels[item.name] || item.name}</strong><small>{item.message}</small></div>
      <b className={`deployment-check-label deployment-check-${item.status}`}>{statusLabels[item.status] || item.status}</b>
    </div>)}</div>
  </section>;
}

export function DeploymentStatusPanels() {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    setError(null);
    try {
      const response = await request("/v1/admin/deployment/guide");
      setData(response.data);
    } catch (value) {
      setError(value);
    }
  };
  useEffect(() => { void load(); }, []);
  const refresh = async () => { setBusy(true); try { await load(); } finally { setBusy(false); } };

  return <section className="deployment-overview-status" aria-label="部署状态">
    <div className="platform-section-heading deployment-overview-heading">
      <div><span className="eyebrow">DEPLOYMENT STATUS</span><h2>部署状态</h2></div>
    </div>
    {data && <>
      <section className="deployment-guide-meta card" aria-label="部署版本信息">
        <div><span>应用版本</span><strong>{data.application?.version || "—"}</strong></div>
        <div><span>运行环境</span><strong>{data.application?.environment || "—"}</strong></div>
        <div><span>Build ID</span><strong className="mono">{data.application?.buildId || "—"}</strong></div>
        <div><span>源码修订</span><strong className="mono" title={data.application?.sourceRevision || ""}>{data.application?.sourceRevision || "—"}</strong></div>
        <RefreshButton onClick={refresh} busy={busy} />
      </section>
      <div className="deployment-guide-grid">
        <CheckList title="Bootstrap 配置" items={data.bootstrap} />
        <CheckList title="配置中心" items={data.configuration} />
      </div>
      <CheckList title="框架检查" items={data.checks} />
    </>}
    {error && <ErrorState error={error} retry={refresh} />}
    {!data && !error && <Loading text="正在检查部署状态…" />}
  </section>;
}
