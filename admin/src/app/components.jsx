import { Fragment, useState } from "react";
import { CommunityBrand } from "../platform/brand/CommunityBrand.jsx";
import { Link } from "./navigation.jsx";

export function PageHeader({ code, title, description }) { return <div className="page-header"><span className="eyebrow">{code}</span><h1>{title}</h1><p>{description}</p></div>; }
export function Loading({ text }) { return <div className="state-card">{text}</div>; }
export function ErrorState({ error, retry }) { return <div className="state-card error"><strong>请求失败</strong><span>{error?.message || "服务暂时不可用"}</span><button onClick={retry}>重试</button></div>; }

const topbarIcons = {
  search: ["M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15z", "M16 16l5 5"],
  language: ["M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z", "M3 12h18", "M12 3c2.2 2.4 3.3 5.4 3.3 9s-1.1 6.6-3.3 9c-2.2-2.4-3.3-5.4-3.3-9S9.8 5.4 12 3z"],
  theme: ["M12 3v2", "M12 19v2", "M3 12h2", "M19 12h2", "m5.6 5.6 1.4-1.4", "m17 7 1.4-1.4", "m5.6 5.6 1.4 1.4", "m17 17 1.4 1.4", "M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8z"],
  shortcuts: ["M4 4h6v6H4z", "M14 4h6v6h-6z", "M4 14h6v6H4z", "M14 14h6v6h-6z"],
  notifications: ["M18 9a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9", "M10 21h4"],
  user: ["M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z", "M4 21a8 8 0 0 1 16 0"],
};

function TopbarIcon({ name }) {
  return <svg className="topbar-icon" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">{topbarIcons[name].map((path) => <path d={path} key={path} />)}</svg>;
}

export function AdminLayout({ pathname, menu, state, onRefresh, onLogout, children }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const shellClass = sidebarCollapsed ? "admin-shell sidebar-collapsed" : "admin-shell";
  return <div className={shellClass}><aside className={sidebarOpen ? "sidebar-open" : ""} aria-label="管理控制台导航"><div className="sidebar-brand"><CommunityBrand /><button className="sidebar-collapse" type="button" aria-label={sidebarCollapsed ? "展开导航栏" : "收起导航栏"} aria-expanded={!sidebarCollapsed} onClick={() => setSidebarCollapsed((value) => !value)}><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8.5" /><path d="m13.5 8.5-3.5 3.5 3.5 3.5" /></svg></button></div><div className="sidebar-menu"><nav>{menu.map((item, index) => <Fragment key={item.path}>{item.section && item.section !== menu[index - 1]?.section ? <div className="menu-section">{item.section}</div> : null}<Link path={item.path} active={pathname === item.path} icon={item.icon} collapsed={sidebarCollapsed}>{item.label}</Link></Fragment>)}</nav></div></aside><section className="main"><header className="topbar"><button className="sidebar-toggle" type="button" onClick={() => setSidebarOpen((value) => !value)} aria-label="切换导航">☰</button><button className="topbar-search" type="button" aria-label="打开搜索"><TopbarIcon name="search" /><span>搜索 (Ctrl+/)</span></button><div className="topbar-actions"><button className="topbar-action" type="button" aria-label="切换语言" title="切换语言"><TopbarIcon name="language" /></button><button className="topbar-action" type="button" aria-label="切换主题" title="切换主题"><TopbarIcon name="theme" /></button><button className="topbar-action" type="button" aria-label="快捷入口" title="快捷入口"><TopbarIcon name="shortcuts" /></button><button className="topbar-action topbar-notification" type="button" aria-label="通知" title="通知"><TopbarIcon name="notifications" /><span>5</span></button><div className="topbar-user-wrap"><button className="topbar-user" type="button" aria-label="打开账号菜单" aria-expanded={profileOpen} onClick={() => setProfileOpen((value) => !value)}><span className="topbar-avatar"><TopbarIcon name="user" /></span><i aria-hidden="true" /></button>{profileOpen ? <div className="topbar-user-menu"><strong>{state.email || state.subject}</strong><button type="button" onClick={() => { setProfileOpen(false); onRefresh(); }}>刷新会话</button><button type="button" onClick={onLogout}>退出</button></div> : null}</div></div></header><main onClick={() => { setSidebarOpen(false); setProfileOpen(false); }}>{children}</main></section></div>;
}
