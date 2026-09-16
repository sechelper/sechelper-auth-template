import "./home.css";
import { useEffect, useState } from "react";
import { apiOrigin, oidcAccountURL } from "../../config/runtime.js";

export function HomePage({ auth }) {
  const { state, login, logout, refresh, refreshSession } = auth;
  const [releaseInfo, setReleaseInfo] = useState(null);

  useEffect(() => {
    const controller = new AbortController();
    fetch(`${apiOrigin()}/v1/version`, { credentials: "include", signal: controller.signal })
      .then((response) => response.ok ? response.json() : Promise.reject(new Error("version unavailable")))
      .then((response) => setReleaseInfo(response.data || null))
      .catch(() => {});
    return () => controller.abort();
  }, []);

  return (
    <main className="public-shell">
      <div className="public-header">
        <p className="eyebrow">SECHELPER COMMUNITY</p>
        <a
          className="admin-entry"
          href="/admin/"
          aria-label="进入管理后台"
          onClick={() => {
            try {
              window.sessionStorage.setItem("sechelper:admin-entry", "root");
            } catch {
              // The explicit /admin/ href remains the fallback when storage is unavailable.
            }
          }}
        >
          进入管理后台
          <span aria-hidden="true">↗</span>
        </a>
      </div>
      <h1>统一认证业务框架</h1>
      <p className="lead">前台业务入口使用统一会话访问 Go API，后台管理端位于 /admin。</p>
      {releaseInfo && <p className="release-info" aria-label="当前部署信息">{releaseInfo.deploymentEnvironment} 环境 · {releaseInfo.releaseVersion}{releaseInfo.buildId ? ` · ${releaseInfo.buildId}` : ""}</p>}
      <section className="public-card" aria-live="polite">
        <h2>当前会话</h2>
        {state.status === "loading" && <p>正在读取会话…</p>}
        {state.status === "unauthenticated" && <><p>当前未登录。</p><button onClick={login}>登录</button></>}
        {state.status === "authenticated" && <>
          <p><a href={oidcAccountURL()}>个人设置（统一身份中心）</a></p>
          <p>Subject：{state.subject}</p>
          <p>Email：{state.email || "未获取"}</p>
          <p>Application：{state.applicationCode}</p>
          <p>有效期：{state.expiresAt}</p>
          <h3>OIDC Profile</h3>
          {state.profile && Object.keys(state.profile).length > 0 ? (
            <pre style={{ overflow: "auto", maxHeight: "min(55vh, 560px)", padding: 20, borderRadius: 12, color: "#dff8e9", background: "#142b24", font: "13px/1.7 ui-monospace, SFMono-Regular, Menlo, monospace", whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
              {JSON.stringify(state.profile, null, 2)}
            </pre>
          ) : <p>当前会话没有可显示的 Profile claims。</p>}
          <button onClick={logout}>退出登录</button>
          <button className="secondary" onClick={refreshSession}>刷新会话</button>
        </>}
        {state.status === "error" && <><p>会话服务暂时不可用。</p><button onClick={refresh}>重试</button></>}
      </section>
    </main>
  );
}
