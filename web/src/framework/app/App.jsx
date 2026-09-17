import { lazy, Suspense, useEffect, useMemo } from "react";
import { useAuth } from "../auth/AuthProvider.jsx";
import { authApi, restoreLoginPath } from "../auth/api.js";
import { GlobalErrorPage } from "../../../shared/error-pages/GlobalErrorPage.jsx";
import { publicBusinessRoutes, validatePublicBusinessModules } from "./public-business-modules.js";
import { InstallPage } from "../pages/install/InstallPage.jsx";

validatePublicBusinessModules();

export function App() {
  const { state, login, silentLogin, logout, refresh, refreshSession, userCenterURL } = useAuth();
  const DevelopmentProfileExamplePage = useMemo(() => (import.meta.env?.DEV || import.meta.env?.MODE === "test")
    ? lazy(() => import("../../business/profile-example/example-module.js"))
    : null, []);
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";
  const businessRoute = publicBusinessRoutes().find((route) => route.path === pathname);

  const logoutState = new URLSearchParams(window.location.search).get("state");
  useEffect(() => {
	if (state.status === "authenticated" && pathname === "/") restoreLoginPath();
  }, [state.status, pathname]);
  useEffect(() => {
    if (pathname !== "/" || !logoutState) return;
    authApi.logoutCallback(logoutState).finally(() => {
      window.history.replaceState({}, "", "/");
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
  }, [pathname, logoutState]);

  if (pathname === "/403") return <GlobalErrorPage code={403} />;
  if (pathname === "/404") return <GlobalErrorPage code={404} />;
  if (pathname === "/500") return <GlobalErrorPage code={500} onAction={refresh} />;
  if (pathname === "/install") return <InstallPage />;
  if (businessRoute) {
    const BusinessPage = businessRoute.element;
    return <BusinessPage auth={{ state, login, silentLogin, logout, refresh, refreshSession, userCenterURL }} />;
  }
  if (pathname === "/" && (import.meta.env?.DEV || import.meta.env?.MODE === "test")) {
    return <Suspense fallback={<p role="status">正在载入开发示例…</p>}><DevelopmentProfileExamplePage auth={{ state, login, silentLogin, logout, refresh, refreshSession, userCenterURL }} /></Suspense>;
  }
  return <GlobalErrorPage code={404} />;
}
