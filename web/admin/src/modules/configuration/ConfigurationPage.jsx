import { useEffect, useMemo, useState } from "react";
import { configurationApi } from "./api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";
import { EmptyState, RefreshButton } from "../../shared/ui.jsx";

const initialForm = { value: "", description: "", isSecret: true };

function ConfigTypeBadge({ isSecret }) {
  return <span className={`configuration-badge ${isSecret ? "configuration-badge-secret" : "configuration-badge-plain"}`}><i aria-hidden="true" />{isSecret ? "敏感值" : "普通值"}</span>;
}

function ConfigIcon({ name }) {
  const paths = name === "shield" ? ["M12 3l7 3v5c0 4.5-2.9 8-7 10-4.1-2-7-5.5-7-10V6z", "m9 12 2 2 4-4"] : ["M12 3v18", "M3 12h18", "M5 5l14 14", "m19 5-14 14"];
  return <svg className="configuration-icon" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">{paths.map((path) => <path d={path} key={path} />)}</svg>;
}

export function ConfigurationPage() {
  const [rows, setRows] = useState(null);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [key, setKey] = useState("");
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState("all");
  const [form, setForm] = useState(initialForm);

  const load = async () => {
    setError(null);
    try {
      const result = await configurationApi.list();
      setRows(result.data || []);
    } catch (value) {
      setError(value);
    }
  };

  useEffect(() => { load(); }, []);

  const filteredRows = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return (rows || []).filter((entry) => {
      const matchesQuery = !normalized || entry.key.toLowerCase().includes(normalized) || entry.description?.toLowerCase().includes(normalized);
      const matchesType = filter === "all" || (filter === "secret" ? entry.isSecret : !entry.isSecret);
      return matchesQuery && matchesType;
    });
  }, [rows, query, filter]);

  const edit = async (entry) => {
    setKey(entry.key);
    setForm({ value: "", description: entry.description || "", isSecret: entry.isSecret });
    if (!entry.isSecret) {
      try {
        const result = await configurationApi.get(entry.key);
        setForm({ value: result.data?.value || "", description: entry.description || "", isSecret: false });
      } catch (value) {
        setError(value);
      }
    }
    document.querySelector(".configuration-editor")?.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  const resetForm = () => { setKey(""); setForm(initialForm); };

  const save = async (event) => {
    event.preventDefault();
    if (!key.trim() || !form.value) return;
    setBusy(true); setError(null);
    try { await configurationApi.save(key.trim().toUpperCase(), form); resetForm(); await load(); }
    catch (value) { setError(value); }
    finally { setBusy(false); }
  };

  const remove = async (entry) => {
    if (!window.confirm(`确认删除 ${entry.key}？`)) return;
    setBusy(true); setError(null);
    try { await configurationApi.remove(entry.key); if (key === entry.key) resetForm(); await load(); }
    catch (value) { setError(value); }
    finally { setBusy(false); }
  };

  if (error && !rows) return <><PageHeader code="PLATFORM" title="配置中心" description="集中管理应用和业务模块可读取的环境变量。" /><ErrorState error={error} retry={load} /></>;
  if (!rows) return <Loading text="正在加载配置中心…" />;

  const secretCount = rows.filter((entry) => entry.isSecret).length;
  const plainCount = rows.length - secretCount;
  const editing = Boolean(key);

  return <>
    <PageHeader code="PLATFORM" title="配置中心" description="集中管理应用和业务模块可读取的环境变量，敏感值以密文保存。" />
    <section className="configuration-hero card" aria-label="配置中心说明"><div className="configuration-hero-copy"><div className="configuration-icon-wrap"><ConfigIcon name="shield" /></div><div><span className="configuration-overline">RUNTIME CONFIGURATION</span><h2>安全地管理运行时变量</h2><p>业务模块通过服务端只读接口获取配置。敏感值不会在列表或详情中回显，基础设施连接配置修改后需按发布流程重启生效。</p></div></div><div className="configuration-hero-note"><strong>配置原则</strong><span>Key 使用大写环境变量格式</span><code>DATABASE_PASSWORD</code></div></section>
    <section className="configuration-stats" aria-label="配置统计"><article className="configuration-stat card"><span className="configuration-stat-icon configuration-stat-icon-primary"><ConfigIcon name="grid" /></span><div><strong>{rows.length}</strong><span>全部配置</span></div></article><article className="configuration-stat card"><span className="configuration-stat-icon configuration-stat-icon-warning"><ConfigIcon name="shield" /></span><div><strong>{secretCount}</strong><span>敏感值</span></div></article><article className="configuration-stat card"><span className="configuration-stat-icon configuration-stat-icon-info"><ConfigIcon name="grid" /></span><div><strong>{plainCount}</strong><span>普通值</span></div></article></section>
    {error && <div className="configuration-inline-error" role="alert"><strong>操作未完成</strong><span>{error.message || "服务暂时不可用"}</span><button type="button" onClick={() => setError(null)}>知道了</button></div>}
    <section className="configuration-editor card"><div className="configuration-section-heading"><div><span className="kicker">{editing ? "EDIT VALUE" : "ADD VALUE"}</span><h2>{editing ? `编辑 ${key}` : "新增配置"}</h2><p>{editing ? "输入新值后保存，旧值不会被回显。" : "为应用或业务模块添加一个运行时环境变量。"}</p></div>{editing && <button className="configuration-quiet-button" type="button" onClick={resetForm}>取消编辑</button>}</div><form onSubmit={save} className="configuration-form"><label><span>环境变量名</span><input value={key} onChange={(event) => setKey(event.target.value.toUpperCase())} placeholder="REDIS_URL" pattern="[A-Z][A-Z0-9_]{0,127}" autoComplete="off" required /><small>仅支持大写字母、数字和下划线，最长 128 个字符。</small></label><label><span>说明</span><input value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} placeholder="例如：业务缓存连接" /></label><label className="configuration-form-wide"><span>值</span><input type={form.isSecret ? "password" : "text"} value={form.value} onChange={(event) => setForm({ ...form, value: event.target.value })} placeholder={form.isSecret && editing ? "输入新值以替换已保存的密文" : "配置值"} autoComplete="new-password" required /><small>{form.isSecret ? "敏感值将加密保存，保存后不可回显。" : "普通值可由拥有读取权限的服务端模块读取。"}</small></label><label className="configuration-secret-toggle"><input type="checkbox" checked={form.isSecret} onChange={(event) => setForm({ ...form, isSecret: event.target.checked })} /><span><strong>敏感值</strong><small>建议对密码、Token、连接字符串保持开启</small></span></label><div className="configuration-form-actions"><button className="configuration-primary-button" type="submit" disabled={busy}>{busy ? "保存中…" : editing ? "保存修改" : "新增配置"}</button>{editing && <button className="configuration-secondary-button" type="button" onClick={resetForm}>清空表单</button>}</div></form></section>
    <section className="configuration-list card"><div className="configuration-section-heading configuration-list-heading"><div><span className="kicker">SAVED VALUES</span><h2>已保存配置</h2><p>只展示配置元数据，不在列表中暴露敏感值。</p></div><RefreshButton onClick={load} busy={busy} /></div><div className="configuration-toolbar"><label className="configuration-search"><span className="sr-only">搜索配置</span><span aria-hidden="true">⌕</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索变量名或说明…" /></label><select aria-label="筛选配置类型" value={filter} onChange={(event) => setFilter(event.target.value)}><option value="all">全部类型</option><option value="secret">仅敏感值</option><option value="plain">仅普通值</option></select><span className="configuration-result-count">显示 {filteredRows.length} / {rows.length}</span></div>{filteredRows.length ? <div className="table-wrap"><table className="configuration-table"><thead><tr><th>变量名</th><th>说明</th><th>类型</th><th>版本</th><th>更新时间</th><th><span className="sr-only">操作</span></th></tr></thead><tbody>{filteredRows.map((entry) => <tr key={entry.key}><td><code className="configuration-key">{entry.key}</code></td><td className="configuration-description">{entry.description || "暂无说明"}</td><td><ConfigTypeBadge isSecret={entry.isSecret} /></td><td><span className="configuration-version">v{entry.version}</span></td><td className="configuration-updated">{new Date(entry.updatedAt).toLocaleString()}</td><td><div className="configuration-row-actions"><button type="button" onClick={() => edit(entry)}>编辑</button><button className="configuration-delete-button" type="button" onClick={() => remove(entry)}>删除</button></div></td></tr>)}</tbody></table></div> : <EmptyState title={query || filter !== "all" ? "没有匹配配置" : "暂无配置"} description={query || filter !== "all" ? "尝试调整搜索词或筛选条件。" : "新增一个环境变量后，业务模块可以通过配置提供者读取它。"} />}</section>
  </>;
}
