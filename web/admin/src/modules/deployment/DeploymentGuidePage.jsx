import { useState } from "react";
import { PageHeader } from "../../app/components.jsx";
import { configurationApi } from "../configuration/api.js";

const variableDefinitions = [
  ["APP_ENV", "app.environment / APP_ENV", false, "development"], ["APP_NAME", "应用名称", false, "auth-template"],
  ["SERVER_HOST", "服务监听地址", false, "0.0.0.0"], ["SERVER_PORT", "服务端口", false, "8080"], ["SERVER_READ_TIMEOUT", "读取超时", false, "10s"], ["SERVER_WRITE_TIMEOUT", "写入超时", false, "15s"], ["SERVER_IDLE_TIMEOUT", "空闲超时", false, "60s"], ["SERVER_SHUTDOWN_TIMEOUT", "关闭超时", false, "15s"],
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
  ["IDENTITY_REVOCATION_ENDPOINT", "身份撤销地址", false, "https://passport.example.com/oauth/revoke"], ["IDENTITY_END_SESSION_ENDPOINT", "身份退出地址", false, "https://passport.example.com/oauth2/logout"],
  ["IDENTITY_REDIRECT_URI", "OIDC 回调地址", false, "https://example.com/v1/auth/callback"],
  ["IDENTITY_APPLICATION_CODE", "身份 Application Code", false, "app-auth-template"],
  ["IDENTITY_SCOPES", "身份 Scope", false, "openid,profile,email"],
  ["MANIFEST_SYNC_INTERVAL", "Manifest 同步周期", false, "10m"],
  ["RATE_LIMIT_LOGIN_PER_MINUTE", "登录限流", false, "10"], ["RATE_LIMIT_CALLBACK_PER_MINUTE", "回调限流", false, "20"], ["RATE_LIMIT_REFRESH_PER_MINUTE", "刷新限流", false, "30"],
  ["SESSION_COOKIE_NAME", "Session Cookie 名称", false, "auth_template_session"],
  ["SESSION_TTL", "Session 有效期", false, "8h"],
  ["SESSION_SECURE", "仅通过 HTTPS 发送 Cookie", false, "true"],
  ["SESSION_SAME_SITE", "Cookie SameSite", false, "Lax"],
  ["SESSION_ENCRYPTION_KEY", "Session 加密密钥", true, "RawStdEncoding 的 32 字节密钥"],
  ["ALLOWED_ORIGINS", "允许的浏览器 Origin", false, "https://example.com"], ["TRUSTED_PROXY_CIDRS", "可信代理网段", false, "127.0.0.1/32"],
  ["METRICS_TOKEN", "Metrics Token", true, "输入 Metrics Token"],
  ["LOG_MODE", "日志模式", false, "local_file"],
  ["LOG_LEVEL", "日志级别", false, "info"],
  ["LOG_OUTPUT", "日志路径", false, "/opt/sechelper-auth-template/logs"],
  ["LOG_RETENTION_DAYS", "日志保留天数", false, "180"],
  ["DEBUG", "调试开关", false, "false"],
];
const emptyValues = () => Object.fromEntries(variableDefinitions.map(([key]) => [key, ""]));

export function DeploymentGuidePage() {
  const [values, setValues] = useState(emptyValues);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(null);
  const [saved, setSaved] = useState(null);

  const saveVariables = async (event) => {
    event.preventDefault();
    const entries = variableDefinitions.filter(([key]) => values[key]);
    if (!entries.length) return;
    setBusy(true);
    setError(null);
    setSaved(null);
    let savedCount = 0;
    try {
      for (const [key, label, isSecret] of entries) {
        await configurationApi.save(key, { value: values[key], isSecret, description: label });
        savedCount += 1;
      }
      setValues(emptyValues());
      setSaved(`已保存 ${savedCount} 项配置到配置中心`);
    } catch (value) {
      setError(`已保存 ${savedCount} 项；后续保存失败：${value.message || "服务暂时不可用"}`);
    } finally {
      setBusy(false);
    }
  };

  return <>
    <PageHeader code="PLATFORM" title="框架配置" description="填写框架环境变量并保存到配置中心。部署状态和版本信息位于概览页面。" />
    <section className="deployment-variable-card card">
      <div className="deployment-guide-card-heading"><div><h2>框架环境变量</h2><p>填写后将加密保存到配置中心；空字段跳过，已有敏感值不会回显。</p></div></div>
      {saved && <div className="deployment-save-success" role="status">{saved}</div>}
      {error && <div className="configuration-inline-error" role="alert">{error}</div>}
      <form className="deployment-variable-form" onSubmit={saveVariables}>
        <div className="deployment-variable-grid">{variableDefinitions.map(([key, label, isSecret, placeholder]) => <label key={key}>
          <span><strong>{label}</strong><code>{key}</code></span>
          <input type={isSecret ? "password" : "text"} value={values[key]} onChange={(event) => setValues({ ...values, [key]: event.target.value })} placeholder={placeholder} autoComplete="new-password" />
          <small>{isSecret ? "敏感值将加密保存，保存后不可回显。" : "保存后在下一次服务启动时生效。"}</small>
        </label>)}</div>
        <div className="deployment-variable-actions"><button className="configuration-primary-button" type="submit" disabled={busy}>{busy ? "保存中…" : "保存到配置中心"}</button><span>需要 configuration:write 权限</span></div>
      </form>
    </section>
    <section className="deployment-guide-notice card"><strong>安全边界</strong><p>数据库 bootstrap 连接和配置中心加密密钥不能同时依赖配置中心本身。修改数据库、Redis、身份平台或 Session 等基础设施配置后，需要按发布流程重启服务。</p></section>
  </>;
}
