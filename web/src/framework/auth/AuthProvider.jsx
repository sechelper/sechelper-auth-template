import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { authApi, clearAuthStatus, consumeInteractiveLoginAttempt, consumeSilentLoginAttempt, hasExplicitLogout, hasSilentLoginFailure, markExplicitLogout, markSilentLoginAttempt, subscribeToAuthChanges } from "./api.js";
import { oidcAccountURL } from "../config/runtime.js";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [state, setState] = useState({ status: "loading" });
  const loadAuthenticated = async () => {
    const session = await authApi.session();
    if (!session.authenticated) return { status: "unauthenticated" };
    const account = await authApi.account();
    return { status: "authenticated", ...session, user: account.data };
  };
  const refresh = async () => {
    try { setState(await loadAuthenticated()); } catch { setState({ status: "error", retryable: true }); }
  };

  useEffect(() => {
    if (window.location.pathname.replace(/\/$/, "") === "/install") { setState({ status: "install" }); return undefined; }
    let cancelled = false;
    async function initialize() {
      try {
        const value = await authApi.session();
        if (cancelled) return;
        if (value.authenticated) {
          consumeInteractiveLoginAttempt(); consumeSilentLoginAttempt();
          const account = await authApi.account();
          if (!cancelled) setState({ status: "authenticated", ...value, user: account.data });
          return;
        }
        if (hasExplicitLogout() || consumeInteractiveLoginAttempt()) { clearAuthStatus(); setState({ status: "unauthenticated" }); return; }
        const silentAttempted = consumeSilentLoginAttempt();
        if (!hasSilentLoginFailure() && !silentAttempted) { markSilentLoginAttempt(); authApi.login({ prompt: "none" }); return; }
        clearAuthStatus(); setState({ status: "unauthenticated" });
      } catch { if (!cancelled) setState({ status: "error", retryable: true }); }
    }
    initialize();
    const unsubscribe = subscribeToAuthChanges(() => { if (!cancelled) setState({ status: "unauthenticated" }); });
    return () => { cancelled = true; unsubscribe(); };
  }, []);

  const value = useMemo(() => ({
    state,
    user: state.user || null,
    login: authApi.login,
    silentLogin: () => { markSilentLoginAttempt(); authApi.login({ prompt: "none" }); },
    refresh,
    refreshSession: async () => { setState((current) => ({ ...current, status: "refreshing" })); try { await authApi.refresh(); await refresh(); } catch (error) { setState({ status: error?.status === 401 ? "reauthentication_required" : "error", retryable: true }); } },
    logout: async () => { const result = await authApi.logout(); markExplicitLogout(); if (result.logoutUrl) { window.location.assign(result.logoutUrl); return; } setState({ status: "unauthenticated" }); },
    userCenterURL: () => oidcAccountURL(),
  }), [state]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() { return useContext(AuthContext); }
