import { Fragment, useEffect, useMemo, useRef, useState } from "react";
import { CommunityBrand } from "../platform/brand/CommunityBrand.jsx";
import { oidcAccountURL } from "../platform/config/runtime.js";
import { Link, NavIcon, navigate } from "./navigation.jsx";
import { findMenuExpansionKeys, findMenuTrail, groupMenuItems } from "./menu-model.js";
import { filterSearchableRoutes } from "./route-search.js";

export function PageHeader({ code, title, description }) { return <div className="page-header"><span className="eyebrow">{code}</span><h1>{title}</h1><p>{description}</p></div>; }
function safeAvatarURL(value) { try { const url = new URL(value); return url.protocol === "https:" || url.protocol === "http:" ? url.href : ""; } catch { return ""; } }
export function Loading({ text }) { return <div className="state-card">{text}</div>; }
export function ErrorState({ error, retry }) { return <div className="state-card error"><strong>请求失败</strong><span>{error?.message || "服务暂时不可用"}</span><button onClick={retry}>重试</button></div>; }

function expandableNodeKeys(items, parentKey) { return (items || []).flatMap((item, index) => { if (!item.children?.length) return []; const key = `${parentKey}:${index}`; return [key, ...expandableNodeKeys(item.children, key)]; }); }
function MenuNode({ item, nodeKey, topLevelExpansionKeys, depth, pathname, collapsedSections, setCollapsedSections, sidebarCollapsed }) {
  if (item.children?.length) { const collapsed = collapsedSections[nodeKey] ?? true; return <div className="menu-subgroup" data-depth={depth}><button type="button" className="menu-subgroup-toggle" aria-expanded={!collapsed} onClick={() => setCollapsedSections((current) => { const next = { ...current }; if (depth === 1) for (const key of topLevelExpansionKeys) next[key] = true; next[nodeKey] = !collapsed; return next; })}>{item.icon ? <NavIcon name={item.icon} /> : <span className="menu-subgroup-dot" aria-hidden="true">•</span>}<span>{item.label}</span><span className="menu-section-chevron" aria-hidden="true">⌄</span></button><div hidden={collapsed}>{item.children.map((child, index) => <MenuNode key={`${nodeKey}:${index}`} item={child} nodeKey={`${nodeKey}:${index}`} topLevelExpansionKeys={topLevelExpansionKeys} depth={depth + 1} pathname={pathname} collapsedSections={collapsedSections} setCollapsedSections={setCollapsedSections} sidebarCollapsed={sidebarCollapsed} />)}</div></div>; }
  return <Link path={item.path} active={pathname === item.path} icon={item.icon} collapsed={sidebarCollapsed} bullet={depth > 1 || !item.icon}>{item.label}</Link>;
}
function MenuGroup({ group, siblingGroupKeys, pathname, collapsedSections, setCollapsedSections, sidebarCollapsed }) { const collapsed = group.collapsible ? (collapsedSections[group.key] ?? true) : false; const topLevelExpansionKeys = expandableNodeKeys(group.items, group.key); const section = group.collapsible ? <button type="button" className="menu-section menu-section-toggle" aria-expanded={!collapsed} onClick={() => setCollapsedSections((current) => { const next = { ...current }; for (const siblingKey of siblingGroupKeys) if (siblingKey !== group.key && current[siblingKey] === false) next[siblingKey] = true; next[group.key] = !collapsed; return next; })}><span>{group.label}</span><span className="menu-section-chevron" aria-hidden="true">⌄</span></button> : <div className="menu-section">{group.label}</div>; return <div className="menu-group">{section}<div className="menu-group-items" hidden={collapsed}>{group.items.map((item, index) => <MenuNode key={`${group.key}:${index}`} item={item} nodeKey={`${group.key}:${index}`} topLevelExpansionKeys={topLevelExpansionKeys} depth={1} pathname={pathname} collapsedSections={collapsedSections} setCollapsedSections={setCollapsedSections} sidebarCollapsed={sidebarCollapsed} />)}</div></div>; }

const topbarIcons = { search: ["M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15z", "M16 16l5 5"], language: ["M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z", "M3 12h18", "M12 3c2.2 2.4 3.3 9 3.3 9s-1.1 6.6-3.3 9c-2.2-2.4-3.3-9-3.3-9S9.8 3 12 3z"], theme: ["M12 3v2", "M12 19v2", "M3 12h2", "M19 12h2", "M8 12a4 4 0 1 0 8 0 4 4 0 0 0-8 0z"], shortcuts: ["M4 4h6v6H4z", "M14 4h6v6h-6z", "M4 14h6v6H4z", "M14 14h6v6h-6z"], notifications: ["M18 9a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-3-3-9", "M10 21h4"], user: ["M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z", "M4 21a8 8 0 0 1 16 0"] };
function TopbarIcon({ name }) { return <svg className="topbar-icon" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">{topbarIcons[name].map((path) => <path d={path} key={path} />)}</svg>; }
const themeModes = ["auto", "night", "day"];
const themeLabels = { auto: "自动", night: "夜间", day: "白天" };
const themeModeDescriptions = { auto: "自动模式：跟随系统", night: "夜间模式", day: "白天模式" };
function ThemeModeIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2.75a9.25 9.25 0 0 0 0 18.5Z" fill="currentColor" stroke="none" /><circle cx="12" cy="12" r="9.25" /></svg>; }
function readThemeMode() {
  try { const value = window.localStorage.getItem("admin-theme-mode"); return themeModes.includes(value) ? value : "auto"; } catch { return "auto"; }
}

export function AdminLayout({ pathname, menu, state, onRefresh, onLogout, children }) {
  const [sidebarOpen, setSidebarOpen] = useState(false); const [sidebarCollapsed, setSidebarCollapsed] = useState(false); const [hoverExpanded, setHoverExpanded] = useState(false); const [profileOpen, setProfileOpen] = useState(false); const [collapsedSections, setCollapsedSections] = useState({}); const [searchOpen, setSearchOpen] = useState(false); const [searchQuery, setSearchQuery] = useState(""); const [activeSearchIndex, setActiveSearchIndex] = useState(0); const [themeMode, setThemeMode] = useState(readThemeMode); const searchInput = useRef(null); const groups = groupMenuItems(menu);
  const searchResults = useMemo(() => filterSearchableRoutes(menu, searchQuery, state.permissions).slice(0, 10), [menu, searchQuery, state.permissions]);
  useEffect(() => {
    const activeKeys = findMenuExpansionKeys(groups, pathname);
    if (!activeKeys.length) return;
    const activeGroup = groups.find((group) => group.key === activeKeys[0]);
    if (!activeGroup) return;
    setCollapsedSections((current) => {
      const next = { ...current };
      for (const group of groups) if (group.collapsible && group.key !== activeGroup.key && current[group.key] === false) next[group.key] = true;
      if (activeGroup.collapsible) next[activeGroup.key] = false;
      let items = activeGroup.items;
      let parentKey = activeGroup.key;
      for (const activeKey of activeKeys.slice(1)) {
        const siblingKeys = items.map((_, index) => `${parentKey}:${index}`);
        for (const siblingKey of siblingKeys) if (siblingKey !== activeKey && current[siblingKey] === false) next[siblingKey] = true;
        const activeIndex = Number(activeKey.slice(parentKey.length + 1));
        items = items[activeIndex]?.children || [];
        parentKey = activeKey;
      }
      for (const key of activeKeys) next[key] = false;
      const changed = Object.keys(next).length !== Object.keys(current).length || Object.keys(next).some((key) => next[key] !== current[key]);
      return changed ? next : current;
    });
  }, [pathname, menu]);
  useEffect(() => { if (searchOpen) searchInput.current?.focus(); else { setSearchQuery(""); setActiveSearchIndex(0); } }, [searchOpen]);
  useEffect(() => { const onKeyDown = (event) => { if ((event.ctrlKey || event.metaKey) && event.key === "/") { event.preventDefault(); setSearchOpen(true); } }; window.addEventListener("keydown", onKeyDown); return () => window.removeEventListener("keydown", onKeyDown); }, []);
  useEffect(() => { setActiveSearchIndex(0); }, [searchQuery]);
  useEffect(() => {
    const root = document.documentElement;
    const colorScheme = window.matchMedia?.("(prefers-color-scheme: dark)");
    const apply = () => {
      const appearance = themeMode === "auto" ? (colorScheme?.matches ? "night" : "day") : themeMode;
      root.dataset.adminThemeMode = themeMode;
      root.dataset.adminThemeAppearance = appearance;
    };
    apply();
    if (themeMode === "auto") colorScheme?.addEventListener?.("change", apply);
    return () => {
      if (themeMode === "auto") colorScheme?.removeEventListener?.("change", apply);
      delete root.dataset.adminThemeMode;
      delete root.dataset.adminThemeAppearance;
    };
  }, [themeMode]);
  const cycleThemeMode = () => setThemeMode((current) => {
    const next = themeModes[(themeModes.indexOf(current) + 1) % themeModes.length];
    try { window.localStorage.setItem("admin-theme-mode", next); } catch { /* Keep theme changes available when storage is blocked. */ }
    return next;
  });
  const openSearchResult = (path) => { setSearchOpen(false); navigate(path); };
  const visualCollapsed = sidebarCollapsed && !hoverExpanded; const shellClass = visualCollapsed ? "admin-shell sidebar-collapsed" : hoverExpanded ? "admin-shell sidebar-hover-expanded" : "admin-shell"; const menuTrail = findMenuTrail(groups, pathname); const collapsibleGroupKeys = groups.filter((group) => group.collapsible).map((group) => group.key);
  const nickname = state.nickname || state.profile?.nickname || state.profile?.name || state.profile?.preferred_username || state.email || state.subject;
  const avatarURL = safeAvatarURL(state.avatarUrl || state.profile?.picture || "");
  return <><div className={shellClass}><aside className={sidebarOpen ? "sidebar-open" : ""} aria-label="管理控制台导航" onMouseEnter={() => { if (sidebarCollapsed) setHoverExpanded(true); }} onMouseLeave={() => { if (sidebarCollapsed) setHoverExpanded(false); }}><div className="sidebar-brand"><CommunityBrand /></div><div className="sidebar-menu"><nav>{groups.map((group) => <Fragment key={group.key}>{group.key === "__root__" ? group.items.map((item) => <Link key={item.path} path={item.path} active={pathname === item.path} icon={item.icon} collapsed={visualCollapsed}>{item.label}</Link>) : <MenuGroup group={group} siblingGroupKeys={collapsibleGroupKeys} pathname={pathname} collapsedSections={collapsedSections} setCollapsedSections={setCollapsedSections} sidebarCollapsed={visualCollapsed} />}</Fragment>)}</nav></div><button className="sidebar-collapse" type="button" aria-label={sidebarCollapsed ? "展开导航栏" : "收起导航栏"} aria-expanded={!sidebarCollapsed} onClick={() => { setHoverExpanded(false); setSidebarCollapsed((value) => !value); }}><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8.5" /><path d="m13.5 8.5-3.5 3.5 3.5 3.5" /></svg></button></aside><section className="main"><header className="topbar"><button className="sidebar-toggle" type="button" onClick={() => setSidebarOpen((value) => !value)} aria-label="切换导航">☰</button><button className="topbar-search" type="button" aria-label="打开搜索" aria-expanded={searchOpen} aria-controls="admin-route-search" onClick={() => setSearchOpen(true)}><TopbarIcon name="search" /><span>搜索 (Ctrl+/)</span></button><div className="topbar-actions"><div className="topbar-theme-wrap"><button className={`topbar-theme-mode mode-${themeMode}`} type="button" data-mode={themeMode} aria-label={`当前模式：${themeLabels[themeMode]}，点击切换`} title={themeModeDescriptions[themeMode]} onClick={cycleThemeMode}><ThemeModeIcon /></button><span className="topbar-theme-hint" role="status">{themeModeDescriptions[themeMode]}</span></div><div className="topbar-user-wrap"><button className="topbar-user" type="button" aria-label={`账号：${nickname}`} title={nickname} aria-expanded={profileOpen} onClick={() => setProfileOpen((value) => !value)}><span className="topbar-avatar">{avatarURL ? <img src={avatarURL} alt="" /> : <TopbarIcon name="user" />}</span><i aria-hidden="true" /></button>{profileOpen ? <div className="topbar-user-menu"><strong>{nickname}</strong>{state.email && <span className="topbar-user-email">{state.email}</span>}<a href={oidcAccountURL()}>设置</a><button type="button" onClick={onLogout}>退出</button></div> : null}</div></div>{searchOpen ? <div className="admin-search-overlay" onMouseDown={(event) => { if (event.target === event.currentTarget) setSearchOpen(false); }}><section className="admin-search-dialog" id="admin-route-search" role="dialog" aria-modal="true" aria-label="搜索管理页面"><div className="admin-search-input-wrap"><TopbarIcon name="search" /><input ref={searchInput} value={searchQuery} onChange={(event) => setSearchQuery(event.target.value)} onKeyDown={(event) => { if (event.key === "Escape") setSearchOpen(false); if (event.key === "ArrowDown" && searchResults.length) { event.preventDefault(); setActiveSearchIndex((index) => (index + 1) % searchResults.length); } if (event.key === "ArrowUp" && searchResults.length) { event.preventDefault(); setActiveSearchIndex((index) => (index - 1 + searchResults.length) % searchResults.length); } if (event.key === "Enter" && searchResults[activeSearchIndex]) { event.preventDefault(); openSearchResult(searchResults[activeSearchIndex].path); } }} placeholder="搜索页面名称…" aria-label="搜索页面名称" aria-autocomplete="list" aria-controls="admin-search-results" aria-activedescendant={searchResults[activeSearchIndex] ? `admin-search-result-${activeSearchIndex}` : undefined} /><kbd>ESC</kbd></div><div className="admin-search-section-title">页面</div><div className="admin-search-results" id="admin-search-results" role="listbox">{searchResults.length ? searchResults.map((result, index) => <button className={`admin-search-result ${index === activeSearchIndex ? "active" : ""}`} id={`admin-search-result-${index}`} key={result.path} type="button" role="option" aria-selected={index === activeSearchIndex} onMouseEnter={() => setActiveSearchIndex(index)} onClick={() => openSearchResult(result.path)}><span className="admin-search-result-icon"><TopbarIcon name="search" /></span><span className="admin-search-result-copy"><strong>{result.label}</strong><small>{result.trail.join(" / ")}</small></span><span className="admin-search-result-enter">↵</span></button>) : <div className="admin-search-empty">没有找到可访问的页面</div>}</div><footer className="admin-search-hint"><span>↑↓ 选择</span><span>↵ 打开页面</span><span>ESC 关闭</span></footer></section></div> : null}</header><main onClick={() => { setSidebarOpen(false); setProfileOpen(false); }}>{menuTrail.length ? <nav className="current-menu-path" aria-label="当前位置"><span className="eyebrow">{menuTrail.join(" / ")}</span></nav> : null}{children}</main></section></div></>;
}
