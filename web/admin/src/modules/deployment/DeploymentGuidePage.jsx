import { useEffect, useState } from "react";
import { request } from "../auth/api.js";
import { ErrorState, Loading, PageHeader } from "../../app/components.jsx";
import { RefreshButton } from "../../shared/ui.jsx";
import { configurationApi } from "../configuration/api.js";

const labels = { database: "数据库 bootstrap 连接", encryptionKey: "配置中心加密密钥", runtimeLoaded: "运行时配置加载", databaseConnectivity: "数据库连通性", configurationTable: "配置中心数据表", configurationMigration: "配置中心迁移" };
const statusLabels = { healthy: "正常", attention: "需要处理", unavailable: "不可用" };
const variableDefinitions = [
  ["DATABASE_URL", "数据库连接", true, "postgres://user:password@host:5432/database"],
  ["REDIS_URL", "Redis 连接", true, "redis://host:6379/0"],
  ["PUBLIC_WEB_ORIGIN", "前台域名", false, "https://example.com"],
  ["API_ORIGIN", "API 域名", false, "https://example.com"],
  ["PRIMARY_DOMAIN", "主域名", false, "example.com"],
  ["TEST_DOMAIN_SUFFIX", "测试域名后缀", false, "-test"],
  ["IDENTITY_ISSUER", "身份平台 Issuer", false, "https://passport.example.com"],
  ["IDENTITY_AUTHORIZATION_ENDPOINT", "身份授权地址", false, "https://passport.example.com/oauth2/authorize"],
  ["IDENTITY_TOKEN_ENDPOINT", "身份 Token 地址", false, "https://passport.example.com/oauth2/token"],
  ["IDENTITY_USERINFO_ENDPOINT", "身份 UserInfo 地址", false, "https://passport.example.com/userinfo"],
  ["IDENTITY_JWKS_URL", "身份 JWKS 地址", false, "https://passport.example.com/oauth2/jwks"],
  ["IDENTITY_AUDIENCE", "身份 Audience", false, "auth-template-api"],
  ["IDENTITY_CLIENT_ID", "身份 Client ID", false, "client-id"],
  ["IDENTITY_CLIENT_SECRET", "身份 Client Secret", true, "输入 Client Secret"],
  ["IDENTITY_REDIRECT_URI", "OIDC 回调地址", false, "https://example.com/v1/auth/callback"],
  ["IDENTITY_APPLICATION_CODE", "身份 Application Code", false, "app-auth-template"],
  ["IDENTITY_SCOPES", "身份 Scope", false, "openid,profile,email"],
  ["MANIFEST_SYNC_INTERVAL", "Manifest 同步周期", false, "10m"],
  ["SESSION_COOKIE_NAME", "Session Cookie 名称", false, "auth_template_session"],
  ["SESSION_TTL", "Session 有效期", false, "8h"],
  ["SESSION_SECURE", "仅通过 HTTPS 发送 Cookie", false, "true"],
  ["SESSION_SAME_SITE", "Cookie SameSite", false, "Lax"],
  ["SESSION_ENCRYPTION_KEY", "Session 加密密钥", true, "RawStdEncoding 的 32 字节密钥"],
  ["ALLOWED_ORIGINS", "允许的浏览器 Origin", false, "https://example.com"],
  ["METRICS_TOKEN", "Metrics Token", true, "输入 Metrics Token"],
  ["LOG_MODE", "日志模式", false, "local_file"],
  ["LOG_LEVEL", "日志级别", false, "info"],
  ["LOG_OUTPUT", "日志路径", false, "/opt/sechelper-auth-template/logs"],
  ["LOG_RETENTION_DAYS", "日志保留天数", false, "180"],
  ["DEBUG", "调试开关", false, "false"],
];

function CheckList({ title, items }) {
  return <section className="deployment-guide-card card"><div className="deployment-guide-card-heading"><h2>{title}</h2><span>{items?.filter((item) => item.status === "healthy").length || 0}/{items?.length || 0} 正常</span></div><div className="deployment-check-list">{(items || []).map((item) => <div className="deployment-check" key={item.name}><span className={`deployment-check-dot deployment-check-${item.status}`} /><div><strong>{labels[item.name] || item.name}</strong><small>{item.message}</small></div><b className={`deployment-check-label deployment-check-${item.status}`}>{statusLabels[item.status] || item.status}</b></div>)}</div></section>;
}

export function DeploymentGuidePage() {
  const [data, setData] = useState(null); const [error, setError] = useState(null); const [busy, setBusy] = useState(false); const [saved, setSaved] = useState(null); const [values, setValues] = useState(() => Object.fromEntries(variableDefinitions.map(([key]) => [key, ""])));
  const load = () => request("/v1/admin/deployment/guide").then((value) => setData(value.data)).catch(setError);
  useEffect(() => { load(); }, []);
  const refresh = async () => { setBusy(true); setError(null); try { await load(); } finally { setBusy(false); } };
  const saveVariables = async (event) => { event.preventDefault(); const entries = variableDefinitions.filter(([key]) => values[key]).map(([key, , isSecret]) => [key, values[key], isSecret]); if (!entries.length) return; setBusy(true); setError(null); setSaved(null); try { for (const [key, value, isSecret] of entries) await configurationApi.save(key, { value, isSecret, description: variableDefinitions.find(([candidate]) => candidate === key)?.[1] || "框架部署配置" }); setValues(Object.fromEntries(variableDefinitions.map(([key]) => [key, ""]))); setSaved(`已保存 ${entries.length} 项配置到配置中心`); await load(); } catch (value) { setError(value); } finally { setBusy(false); } };
  if (error) return <><PageHeader code="PLATFORM" title="部署引导" description="检查框架启动所需的 bootstrap、迁移和配置中心状态。" /><ErrorState error={error} retry={() => { setError(null); load(); }} /></>;
  if (!data) return <Loading text="正在检查部署状态…" />;
  const app = data.application || {};
  return <><PageHeader code="PLATFORM" title="部署引导" description="检查框架启动所需的 bootstrap、迁移和配置中心状态。页面只展示状态，不返回任何密钥、密码或连接串。" /><section className={`deployment-guide-hero card ${data.ready ? "deployment-guide-ready" : "deployment-guide-attention"}`}><div><span className="kicker">DEPLOYMENT READINESS</span><h2>{data.ready ? "当前环境已准备就绪" : "当前环境需要处理"}</h2><p>{data.ready ? "框架可以按当前配置正常启动，配置中心将在启动阶段参与配置解析。" : "请根据下方检查结果补充 bootstrap 配置或执行缺失迁移。"}</p></div><div className="deployment-guide-status"><i />{data.ready ? "READY" : "ACTION REQUIRED"}</div></section><section className="deployment-guide-meta card"><div><span>应用版本</span><strong>{app.version || "—"}</strong></div><div><span>运行环境</span><strong>{app.environment || "—"}</strong></div><div><span>Build ID</span><strong className="mono">{app.buildId || "—"}</strong></div><div><span>源码修订</span><strong className="mono">{app.sourceRevision || "—"}</strong></div><RefreshButton onClick={refresh} busy={busy} /></section><div className="deployment-guide-grid"><CheckList title="Bootstrap 配置" items={data.bootstrap} /><CheckList title="配置中心" items={data.configuration} /></div><CheckList title="框架检查" items={data.checks} /><section className="deployment-variable-card card"><div className="deployment-guide-card-heading"><div><h2>框架环境变量</h2><p>填写后将加密保存到配置中心，空字段会跳过，已有敏感值不会回显。</p></div></div>{saved && <div className="deployment-save-success" role="status">{saved}</div>}<form className="deployment-variable-form" onSubmit={saveVariables}><div className="deployment-variable-grid">{variableDefinitions.map(([key, label, isSecret, placeholder]) => <label key={key}><span><strong>{label}</strong><code>{key}</code></span><input type={isSecret ? "password" : "text"} value={values[key]} onChange={(event) => setValues({ ...values, [key]: event.target.value })} placeholder={placeholder} autoComplete="new-password" /><small>{isSecret ? "敏感值将加密保存，保存后不可回显。" : "保存后在下一次服务启动时生效。"}</small></label>)}</div><div className="deployment-variable-actions"><button className="configuration-primary-button" type="submit" disabled={busy}>{busy ? "保存中…" : "保存到配置中心"}</button><span>需要 `configuration:write` 权限</span></div></form></section><section className="deployment-guide-notice card"><strong>安全边界</strong><p>数据库 bootstrap 连接和配置中心加密密钥不能同时依赖配置中心本身。修改数据库、Redis、身份平台或 Session 等基础设施配置后，需要按发布流程重启服务。</p></section></>;
}
