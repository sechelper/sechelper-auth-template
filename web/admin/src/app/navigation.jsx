import React from "react";
export { initialAdminPath, normalizePathname } from "./navigation-path.js";
import { normalizePathname } from "./navigation-path.js";

export const menu = [
  { path: "/admin", label: "概览", icon: "dashboard" },
  { path: "/admin/permissions", label: "权限清单", icon: "permissions", section: "账号与安全", sectionOrder: 100, permission: "auth:manifest:read" },
  { path: "/admin/resources", label: "资源目录", icon: "resources", section: "账号与安全", sectionOrder: 100, permission: "admin:access" },
  { path: "/admin/audit-events", label: "操作审计", icon: "audit", section: "监控与审计", sectionOrder: 200, permission: "audit:read" },
];

const iconPaths = {
  dashboard: ["M4 4h6v6H4z", "M14 4h6v6h-6z", "M4 14h6v6H4z", "M14 14h6v6h-6z"],
  manifest: ["M6 3h9l3 3v15H6z", "M14 3v4h4", "M9 12h6", "M9 16h6"],
  permissions: ["M12 3l8 3v5c0 5-3.4 8.4-8 10-4.6-1.6-8-10V6z", "M9 12l2 2 4-4"],
  resources: ["M12 3l8 4-8 4-8-4z", "M4 12l8 4 8-4", "M4 17l8 4 8-4"],
  diagnostic: ["M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15z", "M16 16l5 5", "M7.5 10.5h6", "M10.5 7.5v6"],
  profile: ["M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z", "M4 21a8 8 0 0 1 16 0"],
  audit: ["M8 4h8", "M9 2h6v4H9z", "M6 4H4v17h16V4h-2", "M8 11h8", "M8 16h5"],
  operations: ["M3 12h4l2-5 4 10 2-5h6", "M12 3a9 9 0 1 1-8.5 6"],
};

export function NavIcon({ name }) {
  return <svg className="nav-icon" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">{(iconPaths[name] || iconPaths.dashboard).map((path) => <path d={path} key={path} />)}</svg>;
}

export function navigate(path) { window.history.pushState({}, "", path); window.dispatchEvent(new PopStateEvent("popstate")); }
export function Link({ path, children, active, icon, collapsed, bullet = false }) { return <a href={path} className={active ? "nav-link active" : "nav-link"} aria-current={active ? "page" : undefined} title={collapsed ? children : undefined} onClick={(event) => { event.preventDefault(); navigate(path); }}>{bullet ? <span className="nav-bullet" aria-hidden="true">•</span> : <NavIcon name={icon} />}<span className="nav-label">{children}</span></a>; }
