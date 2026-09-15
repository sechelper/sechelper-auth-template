import { useEffect, useState } from "react";
import { useAuth } from "../modules/auth/AuthProvider.jsx";
import { initialAdminPath, menu, navigate, normalizePathname } from "./navigation.jsx";
import { AdminLayout, Loading } from "./components.jsx";
import { DashboardPage } from "../modules/dashboard/DashboardPage.jsx";
import { ManifestPage } from "../modules/permissions/ManifestPage.jsx";
import { SessionsPage } from "../modules/account/SessionsPage.jsx";
import { PermissionsPage } from "../modules/permissions/PermissionsPage.jsx";
import { AuditPage } from "../modules/audit/AuditPage.jsx";
import { ResourcesPage } from "../modules/access/ResourcesPage.jsx";
import { GlobalErrorPage } from "../../../shared/error-pages/GlobalErrorPage.jsx";
import { adminBusinessRoutes, adminBusinessNavigation, validateAdminBusinessModules } from "./admin-business-modules.js";
import { filterMenuByPermission, validateBusinessMenuSections } from "./menu-model.js";

validateAdminBusinessModules();
const businessNavigation = adminBusinessNavigation();
validateBusinessMenuSections(businessNavigation);

function ErrorPage({ code, title, description, action = "返回管理概览", onAction = () => navigate("/admin") }) { return <GlobalErrorPage code={code} title={title} description={description} action={action} onAction={onAction} />; }
function ForbiddenPage() { return <ErrorPage code="403" title="暂无访问权限" description="当前账号没有访问这个管理页面所需的权限。如需开通，请联系系统管理员。" />; }
function NotFoundPage() { return <ErrorPage code="404" title="页面不存在" description="你访问的管理页面可能已被移动、删除，或地址输入有误。" />; }
function ServerErrorPage({ retry }) { return <ErrorPage code="500" title="服务暂时不可用" description="管理端遇到了一点问题，请稍后重试。如果问题持续存在，请联系系统管理员。" action="重新加载" onAction={retry || (() => window.location.reload())} />; }

const frameworkRoutePermissions = {
  "/admin/manifest": "auth:manifest:read",
  "/admin/sessions": "auth:session",
  "/admin/permissions": "auth:manifest:read",
  "/admin/audit-events": "audit:read",
  "/admin/resources": "admin:access",
};

function Page({ pathname, session, hasPermission }) {
  const businessRoute = adminBusinessRoutes().find((route) => route.path === pathname);
  if (businessRoute) {
    if (!hasPermission(businessRoute.permission)) return <ForbiddenPage />;
    const BusinessPage = businessRoute.element;
    return <BusinessPage />;
  }
  const requiredPermission = frameworkRoutePermissions[pathname];
  if (requiredPermission && !hasPermission(requiredPermission)) return <ForbiddenPage />;
  if (pathname === "/admin") return <DashboardPage session={session} />;
  if (pathname === "/admin/manifest") return <ManifestPage />;
  if (pathname === "/admin/sessions") return <SessionsPage />;
  if (pathname === "/admin/permissions") return <PermissionsPage />;
  if (pathname === "/admin/audit-events") return <AuditPage />;
  if (pathname === "/admin/resources") return <ResourcesPage />;
  return <NotFoundPage />;
}

export function AdminApp() {
  const { state, hasPermission, login, logout, refreshSession } = useAuth();
  const [signedOut, setSignedOut] = useState(false);
  const [pathname, setPathname] = useState(() => initialAdminPath());
  useEffect(() => { const handler = () => setPathname(normalizePathname()); window.addEventListener("popstate", handler); return () => window.removeEventListener("popstate", handler); }, []);
  useEffect(() => { if (state.status === "unauthenticated" && !signedOut) login(); }, [state.status, signedOut, login]);
  if (state.status === "loading") return <Loading text="正在验证管理员会话…" />;
  if (signedOut && state.status !== "authenticated") return null;
  if (state.status === "error") return <ServerErrorPage retry={login} />;
  if (state.status !== "authenticated") return null;
  if (!hasPermission("admin:access")) return <ForbiddenPage />;
  const visibleMenu = filterMenuByPermission([...menu, ...businessNavigation], hasPermission);
  return <AdminLayout pathname={pathname} menu={visibleMenu} state={state} onRefresh={refreshSession} onLogout={async () => { setSignedOut(true); await logout(); window.location.assign("/"); }}><Page pathname={pathname} session={state} hasPermission={hasPermission} /></AdminLayout>;
}
