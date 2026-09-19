import { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import { authApi, markExplicitLogout, restoreLoginPath, sessionNeedsRefresh, setAuthEventHandler, subscribeToAuthChanges } from "./api.js";
import { loadRuntimeConfig, oidcAccountURL } from "../../platform/config/runtime.js";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [state, setState] = useState({ status: "loading" });
  const stateRef = useRef(state);
  stateRef.current = state;
  const refresh = async () => {
    try {
      const session = await authApi.session();
      if (!session.authenticated) { setState({ status: "unauthenticated" }); return; }
      const [account, authorization] = await Promise.all([authApi.account(), authApi.authorization()]);
      setState({ status: "authenticated", ...session, user: account.data, ...authorization.data, permissions: authorization.data?.permissions || [] });
      restoreLoginPath();
    } catch (error) {
      setState({ status: error?.status === 401 ? "unauthenticated" : "error", retryable: true });
    }
  };

  useEffect(() => {
    let cancelled = false;
    const stopAuthEvents = setAuthEventHandler(() => {
      if (!cancelled) setState({ status: "reauthentication_required", retryable: true });
    });
    async function initialize() {
      try {
        await loadRuntimeConfig();
        const session = await authApi.session();
        if (cancelled) return;
        if (!session.authenticated) { setState({ status: "unauthenticated" }); return; }
        const [account, authorization] = await Promise.all([authApi.account(), authApi.authorization()]);
        if (!cancelled) { setState({ status: "authenticated", ...session, user: account.data, ...authorization.data, permissions: authorization.data?.permissions || [] }); restoreLoginPath(); }
      } catch (error) { if (!cancelled) setState({ status: error?.status === 401 ? "unauthenticated" : "error", retryable: true }); }
    }
    initialize();
    const unsubscribe = subscribeToAuthChanges(() => { if (!cancelled) setState({ status: "unauthenticated" }); });
    let refreshing = false;
    const refreshSessionIfNeeded = async () => {
      const current = stateRef.current;
      if (cancelled || refreshing || current.status !== "authenticated" || !sessionNeedsRefresh(current.expiresAt)) return;
      refreshing = true;
      try {
        setState((value) => ({ ...value, status: "refreshing" }));
        await authApi.refresh();
        await refresh();
      } catch (error) {
        if (!cancelled) setState({ status: error?.status === 401 ? "reauthentication_required" : "error", retryable: true });
      }
      finally { refreshing = false; }
    };
    const refreshTimer = setInterval(refreshSessionIfNeeded, 30000);
    const onVisibilityOrFocus = () => { if (document.visibilityState !== "hidden") void refreshSessionIfNeeded(); };
    document.addEventListener("visibilitychange", onVisibilityOrFocus);
    window.addEventListener("focus", onVisibilityOrFocus);
    return () => { cancelled = true; unsubscribe(); stopAuthEvents(); clearInterval(refreshTimer); document.removeEventListener("visibilitychange", onVisibilityOrFocus); window.removeEventListener("focus", onVisibilityOrFocus); };
  }, []);

  const value = useMemo(() => ({
    state,
    user: state.user || null,
    userCenterURL: () => oidcAccountURL(),
    hasPermission: (permission) => state.permissions?.includes(permission) || false,
    login: authApi.login,
    refresh,
    refreshSession: async () => { setState((current) => ({ ...current, status: "refreshing" })); try { await authApi.refresh(); await refresh(); } catch (error) { setState({ status: error?.status === 401 ? "reauthentication_required" : "error", retryable: true }); } },
    logout: async () => { const result = await authApi.logout(); markExplicitLogout(); if (result.logoutUrl) { window.location.assign(result.logoutUrl); return; } window.location.assign("/"); },
  }), [state]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() { return useContext(AuthContext); }
