import { useEffect } from "react";
import { useAuth } from "../auth/AuthProvider.jsx";
import { appName } from "../config/runtime.js";
import { authApi, restoreLoginPath } from "../auth/api.js";
import { GlobalErrorPage } from "../../../shared/error-pages/GlobalErrorPage.jsx";
import { publicBusinessRoutes, validatePublicBusinessModules } from "./public-business-modules.js";
import { InstallPage } from "../pages/install/InstallPage.jsx";

validatePublicBusinessModules();

export function App() {
  const { state, login, silentLogin, logout, refresh, refreshSession, userCenterURL } = useAuth();
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";
  useEffect(() => { if (appName()) document.title = appName(); }, [state.status]);
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

  if (pathname === "/403") return <GlobalErrorPage code={403} brandName={appName()} />;
  if (pathname === "/404") return <GlobalErrorPage code={404} brandName={appName()} />;
  if (pathname === "/500") return <GlobalErrorPage code={500} brandName={appName()} onAction={refresh} />;
  if (pathname === "/install") return <InstallPage />;
  if (businessRoute) {
    const BusinessPage = businessRoute.element;
    return <BusinessPage auth={{ state, login, silentLogin, logout, refresh, refreshSession, userCenterURL, appName: appName() }} />;
  }
  return <GlobalErrorPage code={404} brandName={appName()} />;
}
