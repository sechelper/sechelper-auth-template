import { useMemo, useState } from "react";
import { apiOrigin } from "../../config/runtime.js";
import "./install.css";

const allFrameworkVariables = [
  ["APP_ENV", "app.environment / APP_ENV", false, "development", "development"], ["APP_NAME", "应用名称", false, "auth-template", "auth-template"],
  ["SERVER_HOST", "服务监听地址", false, "0.0.0.0", "127.0.0.1"], ["SERVER_PORT", "服务端口", false, "8080", "8080"], ["SERVER_READ_TIMEOUT", "读取超时", false, "10s", "10s"], ["SERVER_WRITE_TIMEOUT", "写入超时", false, "15s", "15s"], ["SERVER_IDLE_TIMEOUT", "空闲超时", false, "60s", "60s"], ["SERVER_SHUTDOWN_TIMEOUT", "关闭超时", false, "15s", "15s"],
  ["PUBLIC_WEB_ORIGIN", "前台域名", false, "https://example.com", ""], ["API_ORIGIN", "API 域名", false, "https://example.com", ""], ["PRIMARY_DOMAIN", "主域名", false, "example.com", ""], ["TEST_DOMAIN_SUFFIX", "测试域名后缀", false, "-test", "-test"],
  ["IDENTITY_ISSUER", "身份平台 Issuer", false, "https://passport.example.com", ""], ["IDENTITY_AUTHORIZATION_ENDPOINT", "身份授权地址", false, "https://passport.example.com/oauth2/authorize", ""], ["IDENTITY_TOKEN_ENDPOINT", "身份 Token 地址", false, "https://passport.example.com/oauth2/token", ""], ["IDENTITY_USERINFO_ENDPOINT", "身份 UserInfo 地址", false, "https://passport.example.com/userinfo", ""], ["IDENTITY_JWKS_URL", "身份 JWKS 地址", false, "https://passport.example.com/oauth2/jwks", ""], ["IDENTITY_AUDIENCE", "身份 Audience", false, "auth-template-api", ""], ["IDENTITY_CLIENT_ID", "身份 Client ID", false, "client-id", ""], ["IDENTITY_CLIENT_SECRET", "身份 Client Secret", true, "输入 Client Secret", ""], ["IDENTITY_REDIRECT_URI", "OIDC 回调地址", false, "https://example.com/v1/auth/callback", ""], ["IDENTITY_APPLICATION_CODE", "身份 Application Code", false, "app-auth-template", ""], ["IDENTITY_SCOPES", "身份 Scope", false, "openid,profile,email", "openid,profile,email"],
  ["MANIFEST_SYNC_INTERVAL", "Manifest 同步周期", false, "10m", "10m"], ["SESSION_COOKIE_NAME", "Session Cookie 名称", false, "auth_template_session", "auth_template_session"], ["SESSION_TTL", "Session 有效期", false, "8h", "8h"], ["SESSION_SECURE", "仅通过 HTTPS 发送 Cookie", false, "true", "false"], ["SESSION_SAME_SITE", "Cookie SameSite", false, "Lax", "Lax"], ["SESSION_ENCRYPTION_KEY", "Session 加密密钥", true, "RawStdEncoding 32 字节密钥", ""], ["ALLOWED_ORIGINS", "允许的浏览器 Origin", false, "https://example.com", ""], ["TRUSTED_PROXY_CIDRS", "可信代理网段", false, "127.0.0.1/32", "127.0.0.1/32"], ["METRICS_TOKEN", "Metrics Token", true, "输入 Metrics Token", ""], ["RATE_LIMIT_LOGIN_PER_MINUTE", "登录限流", false, "10", "10"], ["RATE_LIMIT_CALLBACK_PER_MINUTE", "回调限流", false, "20", "20"], ["RATE_LIMIT_REFRESH_PER_MINUTE", "刷新限流", false, "30", "30"], ["IDENTITY_REVOCATION_ENDPOINT", "身份撤销地址", false, "https://passport.example.com/oauth/revoke", ""], ["IDENTITY_END_SESSION_ENDPOINT", "身份退出地址", false, "https://passport.example.com/oauth2/logout", ""], ["LOG_MODE", "日志模式", false, "local_file", "local_file"], ["LOG_LEVEL", "日志级别", false, "info", "info"], ["LOG_OUTPUT", "日志路径", false, "/var/log/auth-template", "logs/app.log"], ["LOG_RETENTION_DAYS", "日志保留天数", false, "180", "180"], ["DEBUG", "调试开关", false, "false", "false"],
];
const frameworkVariables = allFrameworkVariables.map((entry) => entry[0] === "SESSION_COOKIE_NAME" ? [...entry.slice(0, 3), "auth", "auth"] : entry).filter(([key]) => key !== "SERVER_HOST" && key !== "SERVER_PORT");
const installFrameworkVariables = frameworkVariables;
const requiredFrameworkKeys = new Set([
  "APP_ENV", "APP_NAME", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT", "PUBLIC_WEB_ORIGIN", "API_ORIGIN", "PRIMARY_DOMAIN", "TEST_DOMAIN_SUFFIX",
  "IDENTITY_ISSUER", "IDENTITY_AUTHORIZATION_ENDPOINT", "IDENTITY_TOKEN_ENDPOINT", "IDENTITY_USERINFO_ENDPOINT", "IDENTITY_JWKS_URL", "IDENTITY_AUDIENCE", "IDENTITY_CLIENT_ID", "IDENTITY_CLIENT_SECRET", "IDENTITY_REDIRECT_URI", "IDENTITY_APPLICATION_CODE", "IDENTITY_SCOPES",
  "MANIFEST_SYNC_INTERVAL", "SESSION_COOKIE_NAME", "SESSION_TTL", "SESSION_SECURE", "SESSION_SAME_SITE", "SESSION_ENCRYPTION_KEY", "ALLOWED_ORIGINS", "TRUSTED_PROXY_CIDRS", "METRICS_TOKEN", "RATE_LIMIT_LOGIN_PER_MINUTE", "RATE_LIMIT_CALLBACK_PER_MINUTE", "RATE_LIMIT_REFRESH_PER_MINUTE", "LOG_MODE", "LOG_LEVEL", "LOG_OUTPUT", "LOG_RETENTION_DAYS", "DEBUG",
]);
const optionalFrameworkKeys = new Set(["IDENTITY_REVOCATION_ENDPOINT", "IDENTITY_END_SESSION_ENDPOINT"]);
function randomSessionEncryptionKey() {
  const bytes = new Uint8Array(32);
  window.crypto.getRandomValues(bytes);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return window.btoa(binary).replace(/=+$/, "");
}
function randomMetricsToken() {
  const bytes = new Uint8Array(32);
  window.crypto.getRandomValues(bytes);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return window.btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}
const defaultValues = () => {
  const origin = window.location.origin;
  const hostname = window.location.hostname;
  return Object.fromEntries(frameworkVariables.map(([key,,,, defaultValue]) => [key, {
    APP_ENV: import.meta.env.MODE === "test" || import.meta.env.MODE === "production" ? import.meta.env.MODE : defaultValue,
    PUBLIC_WEB_ORIGIN: origin,
    API_ORIGIN: origin,
    PRIMARY_DOMAIN: hostname,
    IDENTITY_REDIRECT_URI: `${origin}/v1/auth/callback`,
    IDENTITY_END_SESSION_ENDPOINT: "",
    ALLOWED_ORIGINS: origin,
    SESSION_ENCRYPTION_KEY: randomSessionEncryptionKey(),
    METRICS_TOKEN: randomMetricsToken(),
    SESSION_SECURE: "true",
  }[key] ?? defaultValue ?? ""]));
};

async function jsonRequest(path, options = {}) { const response = await fetch(`${apiOrigin()}${path}`, { credentials: "same-origin", ...options, headers: { "Content-Type": "application/json", ...(options.headers || {}) } }); const body = await response.json().catch(() => ({})); if (!response.ok) { const error = new Error(body?.error?.message || "安装请求失败"); error.details = body?.error?.details || []; throw error; } return body; }

export function InstallPage() {
  const [values, setValues] = useState(defaultValues); const [custom, setCustom] = useState([{ key: "", value: "", description: "", isSecret: true }]); const [installKey, setInstallKey] = useState(""); const [busy, setBusy] = useState(false); const [error, setError] = useState(null); const [fieldErrors, setFieldErrors] = useState({}); const [importMessage, setImportMessage] = useState(""); const [restartState, setRestartState] = useState("");
  const entries = useMemo(() => [...installFrameworkVariables.map(([key, description, isSecret]) => ({ key, description, isSecret, value: values[key] })).filter((item) => item.value), ...custom.filter((item) => item.key || item.value)], [values, custom]);
  const importOIDC = async (event) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    try {
      const parsed = JSON.parse(await file.text());
      const source = parsed.configuration || parsed;
      const read = (...keys) => { for (const key of keys) { const value = source[key]?.value ?? source[key]; if (value !== undefined && value !== null && value !== "") return Array.isArray(value) ? value.join(",") : String(value); } return ""; };
      const imported = {
        IDENTITY_ISSUER: read("issuer"), IDENTITY_AUTHORIZATION_ENDPOINT: read("authorization_endpoint", "authorizationEndpoint"), IDENTITY_TOKEN_ENDPOINT: read("token_endpoint", "tokenEndpoint"), IDENTITY_USERINFO_ENDPOINT: read("userinfo_endpoint", "userinfoEndpoint"), IDENTITY_JWKS_URL: read("jwks_uri", "jwksUrl"), IDENTITY_APPLICATION_CODE: read("application_code", "applicationCode"), IDENTITY_AUDIENCE: read("api_audience", "audience"), IDENTITY_CLIENT_ID: read("client_id", "clientId"), IDENTITY_CLIENT_SECRET: read("client_secret", "clientSecret"), IDENTITY_REDIRECT_URI: read("redirect_uris", "redirectUri"), IDENTITY_SCOPES: read("allowed_scopes", "scopes").replace(/[\n\r]+/g, ","),
      };
      setValues((current) => ({ ...current, ...Object.fromEntries(Object.entries(imported).filter(([, value]) => value)) }));
      setImportMessage(`已导入 ${Object.values(imported).filter(Boolean).length} 项 OIDC 配置`);
      setError(null);
    } catch (value) { setImportMessage(""); setError(new Error(`OIDC JSON 导入失败：${value.message || "文件格式无效"}`)); }
  };
  const waitForRestart = async () => {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      try {
        const ready = await fetch(`${apiOrigin()}/readyz`, { credentials: "same-origin", cache: "no-store" });
        if (!ready.ok) continue;
        const runtime = await fetch(`${apiOrigin()}/v1/runtime-config`, { credentials: "same-origin", cache: "no-store" });
        if (runtime.ok) return;
      } catch { /* The service is expected to be unavailable during restart. */ }
    }
    throw new Error("服务重启后未在规定时间内恢复，请检查服务日志");
  };
  const submit = async (event) => {
    event.preventDefault();
    const missing = installFrameworkVariables.filter(([key]) => requiredFrameworkKeys.has(key) && !String(values[key] || "").trim()).map(([key, label]) => ({ key, label }));
    if (!installKey.trim()) missing.unshift({ key: "INSTALL_KEY", label: "安装密钥" });
    if (missing.length) {
      const nextErrors = Object.fromEntries(missing.map(({ key, label }) => [key, `${label}不能为空` ]));
      setFieldErrors(nextErrors);
      const first = missing.find(({ key }) => key !== "INSTALL_KEY")?.key;
      document.querySelector(first ? `[name="${first}"]` : ".install-key-field input")?.scrollIntoView({ behavior: "smooth", block: "center" });
      document.querySelector(first ? `[name="${first}"]` : ".install-key-field input")?.focus();
      return;
    }
    setBusy(true); setError(null); setFieldErrors({});
    try { const result = await jsonRequest("/v1/install/configuration", { method: "PUT", headers: { "X-Install-Key": installKey }, body: JSON.stringify({ entries }) }); if (result?.data?.restartRequired) { setRestartState("配置已保存，正在重启服务并加载框架配置…"); await waitForRestart(); } window.location.assign("/"); }
    catch (value) { setError(value); setFieldErrors(Object.fromEntries((value.details || []).map((item) => [item.field, "不能为空"]))); }
    finally { setBusy(false); }
  };
  return <main className="install-shell"><section className="install-card"><span className="install-kicker">FIRST RUN INSTALLATION</span><h1>首次运行安装引导</h1><p className="install-lead">带“必填”标记的关键配置不能为空；域名、OIDC 凭据、密钥和 Token 等部署专属字段必须完成填写。敏感值会加密保存，安装完成后此页面自动关闭。</p><form noValidate onSubmit={submit}><label className="install-key-field"><span>安装密钥</span><input type="password" value={installKey} onChange={(event) => { setInstallKey(event.target.value); setFieldErrors((current) => { const next = { ...current }; delete next.INSTALL_KEY; return next; }); }} placeholder="输入部署提供的安装密钥" required /><small className="install-field-error">{fieldErrors.INSTALL_KEY || ""}</small></label><div className="install-oidc-import"><div><strong>OIDC 配置导入</strong><small>支持统一身份中心导出的 JSON，自动填充 Issuer、Client、Endpoint、Scope 等字段。</small></div><label className="install-import-button"><input type="file" accept="application/json,.json" onChange={importOIDC} />导入 JSON</label></div>{importMessage && <p className="install-import-success" role="status">{importMessage}</p>}<h2>框架环境变量</h2><div className="install-grid">{frameworkVariables.map(([key, label, secret, placeholder]) => <label key={key}><span>{label}<code>{key}</code>{requiredFrameworkKeys.has(key) && <em>必填</em>}</span><input name={key} type={secret ? "password" : "text"} value={values[key]} onChange={(event) => { setValues((current) => ({ ...current, [key]: event.target.value })); setFieldErrors((current) => { const next = { ...current }; delete next[key]; return next; }); }} placeholder={placeholder} required={requiredFrameworkKeys.has(key)} aria-invalid={Boolean(fieldErrors[key])} aria-describedby={`${key}-error`} autoComplete="new-password" /><small id={`${key}-error`} className="install-field-error">{fieldErrors[key] || ""}</small></label>)}</div><div className="install-custom-heading"><h2>业务自定义变量</h2><button type="button" onClick={() => setCustom([...custom, { key: "", value: "", description: "", isSecret: true }])}>添加变量</button></div>{custom.map((item, index) => <div className="install-custom-row" key={index}><input value={item.key} onChange={(event) => setCustom(custom.map((current, row) => row === index ? { ...current, key: event.target.value.toUpperCase() } : current))} placeholder="BUSINESS_FEATURE_KEY" pattern="[A-Z][A-Z0-9_]{0,127}" /><input value={item.description} onChange={(event) => setCustom(custom.map((current, row) => row === index ? { ...current, description: event.target.value } : current))} placeholder="说明" /><input type={item.isSecret ? "password" : "text"} value={item.value} onChange={(event) => setCustom(custom.map((current, row) => row === index ? { ...current, value: event.target.value } : current))} placeholder="变量值" autoComplete="new-password" /><label><input type="checkbox" checked={item.isSecret} onChange={(event) => setCustom(custom.map((current, row) => row === index ? { ...current, isSecret: event.target.checked } : current))} />敏感</label></div>)}{restartState && <p className="install-import-success" role="status">{restartState}</p>}{error && <p className="install-error">{error.message}</p>}<button className="install-submit" disabled={busy}>{busy ? (restartState ? "等待服务恢复…" : "安装中…") : "保存并完成安装"}</button></form></section></main>;
}
