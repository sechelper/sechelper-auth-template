import { lazy, Suspense, useMemo } from "react";
import { useAuth } from "../auth/AuthProvider.jsx";
import { GlobalErrorPage } from "../../../shared/error-pages/GlobalErrorPage.jsx";
import { publicBusinessRoutes, validatePublicBusinessModules } from "./public-business-modules.js";

validatePublicBusinessModules();

export function App() {
  const { state, login, logout, refresh, refreshSession } = useAuth();
  const DevelopmentProfileExamplePage = useMemo(() => (import.meta.env?.DEV || import.meta.env?.MODE === "test")
    ? lazy(() => import("../../business/profile-example/example-module.js"))
    : null, []);
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";
  const businessRoute = publicBusinessRoutes().find((route) => route.path === pathname);

  if (pathname === "/403") return <GlobalErrorPage code={403} />;
  if (pathname === "/404") return <GlobalErrorPage code={404} />;
  if (pathname === "/500") return <GlobalErrorPage code={500} onAction={refresh} />;
  if (businessRoute) {
    const BusinessPage = businessRoute.element;
    return <BusinessPage auth={{ state, login, logout, refresh, refreshSession }} />;
  }
  if (pathname === "/" && (import.meta.env?.DEV || import.meta.env?.MODE === "test")) {
    return <Suspense fallback={<p role="status">正在载入开发示例…</p>}><DevelopmentProfileExamplePage auth={{ state, login, logout, refresh, refreshSession }} /></Suspense>;
  }
  return <GlobalErrorPage code={404} />;
}
