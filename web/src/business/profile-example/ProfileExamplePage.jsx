import { useEffect, useState } from "react";
import "./profile-example.css";

export function ProfileExamplePage({ auth }) {
  const { state, login, logout, refreshSession, refresh } = auth;
  const [pending, setPending] = useState("");
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    document.title = "SECHELPER COMMUNITY · OIDC Profile 示例";
  }, []);

  async function runAction(name, action) {
    setPending(name);
    setActionError("");
    try {
      await action();
    } catch {
      setActionError("操作未完成，请重试。");
    } finally {
      setPending("");
    }
  }

  return (
    <main className="profile-example">
      <header className="profile-example__header">
        <a className="profile-example__brand" href="/" aria-label="SECHELPER COMMUNITY 示例首页">SECHELPER COMMUNITY · 示例</a>
        <a className="profile-example__admin" href="/admin/">进入管理后台 <span aria-hidden="true">↗</span></a>
      </header>

      <section className="profile-example__intro">
        <p className="profile-example__eyebrow">开发 / 测试示例</p>
        <h1>OIDC Profile 展示</h1>
        <p>此页仅演示如何通过统一会话显示身份平台返回的 Profile，不代表正式业务首页。</p>
      </section>

      <section className="profile-example__card" aria-live="polite" aria-busy={state.status === "loading"}>
        <div className="profile-example__card-heading">
          <div>
            <p className="profile-example__eyebrow">当前会话</p>
            <h2>Profile JSON</h2>
          </div>
          {state.status === "authenticated" && <span className="profile-example__status">已登录</span>}
        </div>

        {state.status === "loading" && <p role="status">正在读取会话…</p>}
        {state.status === "unauthenticated" && (
          <div className="profile-example__empty">
            <p>登录后即可查看 OIDC Profile 示例数据。</p>
            <button onClick={login}>统一认证登录</button>
          </div>
        )}
        {state.status === "authenticated" && (
          <>
            <dl className="profile-example__identity">
              <div><dt>Subject</dt><dd>{state.subject || "未提供"}</dd></div>
              <div><dt>Email</dt><dd>{state.email || "未获取"}</dd></div>
              <div><dt>Application</dt><dd>{state.applicationCode || "未提供"}</dd></div>
              <div><dt>会话有效期</dt><dd>{state.expiresAt || "未提供"}</dd></div>
            </dl>
            {state.profile && Object.keys(state.profile).length > 0 ? (
              <pre className="profile-example__json">{JSON.stringify(state.profile, null, 2)}</pre>
            ) : (
              <p className="profile-example__empty">当前会话没有可显示的 Profile claims。</p>
            )}
            {actionError && <p className="profile-example__error" role="alert">{actionError}</p>}
            <div className="profile-example__actions">
              <button className="profile-example__secondary" disabled={Boolean(pending)} onClick={() => runAction("refresh", refreshSession)}>
                {pending === "refresh" ? "正在刷新…" : "刷新会话"}
              </button>
              <button className="profile-example__secondary" disabled={Boolean(pending)} onClick={() => runAction("logout", logout)}>
                {pending === "logout" ? "正在退出…" : "退出登录"}
              </button>
            </div>
          </>
        )}
        {state.status === "error" && (
          <div className="profile-example__empty">
            <p>会话服务暂时不可用。</p>
            <button onClick={refresh}>重试</button>
          </div>
        )}
      </section>
    </main>
  );
}
