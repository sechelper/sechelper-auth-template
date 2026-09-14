import { useState } from "react";
import { request } from "../auth/api.js";
import { ErrorState, PageHeader } from "../../app/components.jsx";

export function AccessCheckPage() {
  const [form, setForm] = useState({ resourceType: "", resourceId: "", action: "read" });
  const [result, setResult] = useState(null); const [error, setError] = useState(null); const [busy, setBusy] = useState(false);
  const submit = async (event) => { event.preventDefault(); setBusy(true); setError(null); setResult(null); try { const value = await request("/v1/admin/access-decisions/check", { method: "POST", body: JSON.stringify(form) }); setResult(value.data); } catch (value) { setError(value); } finally { setBusy(false); } };
  return <><PageHeader code="ACCESS CONTROL" title="访问控制诊断" description="检查当前管理员对指定资源执行某个动作时的授权决策。该工具只用于诊断，不会绕过业务 API。" /><section className="card access-check-card"><form onSubmit={submit}><label>资源类型<input value={form.resourceType} onChange={(event) => setForm({ ...form, resourceType: event.target.value })} placeholder="例如 document" required /></label><label>资源标识<input value={form.resourceId} onChange={(event) => setForm({ ...form, resourceId: event.target.value })} placeholder="例如 resource-001" required /></label><label>动作<input value={form.action} onChange={(event) => setForm({ ...form, action: event.target.value })} placeholder="例如 read" required /></label><button disabled={busy}>{busy ? "检查中…" : "检查访问权限"}</button></form>{error && <ErrorState error={error} retry={() => setError(null)} />}{result && <div className={`decision-result ${result.allowed ? "allowed" : "denied"}`}><strong>{result.allowed ? "允许访问" : "拒绝访问"}</strong><span>原因：{result.reasonCode}</span>{result.permission && <span>所需权限：{result.permission}</span>}</div>}</section></>;
}
