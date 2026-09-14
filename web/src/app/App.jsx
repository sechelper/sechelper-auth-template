import { useAuth } from "../modules/auth/AuthProvider.jsx";
import { GlobalErrorPage } from "../../shared/error-pages/GlobalErrorPage.jsx";
import { publicBusinessRoutes, validatePublicBusinessModules } from "./public-business-modules.js";

validatePublicBusinessModules();

export function App() {
  const { state, login, logout, refresh, refreshSession } = useAuth();
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";
  const businessRoute = publicBusinessRoutes().find((route) => route.path === pathname);

  if (pathname === "/403") return <GlobalErrorPage code={403} />;
  if (pathname === "/404") return <GlobalErrorPage code={404} />;
  if (pathname === "/500" || state.status === "error") return <GlobalErrorPage code={500} onAction={refresh} />;
  if (businessRoute) { const BusinessPage = businessRoute.element; return <BusinessPage />; }
  if (pathname !== "/") return <GlobalErrorPage code={404} />;

  return (
    <main className="public-shell">
      <div className="public-header">
        <p className="eyebrow">SECHELPER COMMUNITY</p>
        <a className="admin-entry" href="/admin/" aria-label="进入管理后台">
          进入管理后台
          <span aria-hidden="true">↗</span>
        </a>
      </div>
      <h1>统一认证业务框架</h1>
      <p className="lead">前台业务入口使用统一会话访问 Go API，后台管理端位于 /admin。</p>
      <section className="public-card" aria-live="polite">
        <h2>当前会话</h2>
        {state.status === "loading" && <p>正在读取会话…</p>}
        {state.status === "unauthenticated" && <><p>当前未登录。</p><button onClick={login}>登录</button></>}
        {state.status === "authenticated" && <>
          <p>Subject：{state.subject}</p>
          <p>Email：{state.email || "未获取"}</p>
          <p>Application：{state.applicationCode}</p>
          <p>有效期：{state.expiresAt}</p>
          <button onClick={logout}>退出登录</button>
          <button className="secondary" onClick={refreshSession}>刷新会话</button>
        </>}
        {state.status === "error" && <><p>会话服务暂时不可用。</p><button onClick={refresh}>重试</button></>}
      </section>
    </main>
  );
}
