import { useEffect, useState } from "react";
import { request } from "../auth/api.js";
import { DataTable, EmptyState, RefreshButton } from "../../shared/ui.jsx";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";

export function ResourcesPage() {
  const [rows, setRows] = useState(null); const [error, setError] = useState(null); const [busy, setBusy] = useState(false);
  const load = () => request("/v1/admin/resources").then((value) => setRows(value.data || [])).catch(setError);
  useEffect(() => { load(); }, []);
  const refresh = async () => { setBusy(true); setError(null); try { await load(); } finally { setBusy(false); } };
  if (error && !rows) return <><PageHeader code="ACCESS CONTROL" title="资源目录" description="查看业务模块注册的资源类型、动作和所需权限。" /><ErrorState error={error} retry={() => { setError(null); load(); }} /></>;
  if (!rows) return <Loading text="正在加载资源目录…" />;
  return <><PageHeader code="ACCESS CONTROL" title="资源目录" description="查看业务模块注册的资源类型、动作和所需权限。资源数据和角色分配仍由业务模块与统一认证平台负责。" /><div className="section-heading"><span>{rows.length} 类已注册资源</span><RefreshButton onClick={refresh} busy={busy} /></div><section className="card">{rows.length ? <DataTable rows={rows} columns={[{ key: "type", label: "资源类型", render: (row) => <code>{row.type}</code> }, { key: "name", label: "名称" }, { key: "description", label: "说明" }, { key: "actions", label: "动作", render: (row) => row.actions?.map((action) => <code className="resource-action" key={action.name}>{action.name} · {action.permission}</code>) }]} /> : <EmptyState title="暂无资源定义" description="业务模块尚未注册资源访问定义。" />}</section></>;
}
