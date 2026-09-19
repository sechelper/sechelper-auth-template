import { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import { authApi, clearAuthStatus, consumeInteractiveLoginAttempt, consumeSilentLoginAttempt, hasExplicitLogout, hasSilentLoginFailure, markExplicitLogout, markSilentLoginAttempt, setAuthEventHandler, subscribeToAuthChanges } from "./api.js";
import { loadRuntimeConfig, oidcAccountURL } from "../config/runtime.js";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [state, setState] = useState({ status: "loading" });
  const stateRef = useRef(state);
  stateRef.current = state;
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
    const stopAuthEvents = setAuthEventHandler(() => {
      if (hasExplicitLogout()) return;
      setState({ status: "reauthentication_required", retryable: true });
    });
    async function initialize() {
      try {
        await loadRuntimeConfig();
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
      } catch (error) {
        if (cancelled) return;
        if (error?.status === 401 && !hasSilentLoginFailure() && !hasExplicitLogout()) {
          markSilentLoginAttempt();
          authApi.login({ prompt: "none" });
          return;
        }
        setState({ status: "error", retryable: true });
      }
    }
    initialize();
    const unsubscribe = subscribeToAuthChanges(() => { if (!cancelled) setState({ status: "unauthenticated" }); });
    const refreshTimer = setInterval(async () => {
      const current = stateRef.current;
      if (cancelled || current.status !== "authenticated" || !current.expiresAt) return;
      const remaining = Date.parse(current.expiresAt) - Date.now();
      if (remaining > 120000) return;
      try {
        setState((current) => ({ ...current, status: "refreshing" }));
        await authApi.refresh();
        const next = await loadAuthenticated();
        if (!cancelled) setState(next);
      } catch (error) {
        if (cancelled) return;
        if (error?.status === 401 && !hasExplicitLogout()) {
          setState({ status: "reauthentication_required", retryable: true });
        } else setState({ status: "error", retryable: true });
      }
    }, 30000);
    return () => { cancelled = true; unsubscribe(); stopAuthEvents(); clearInterval(refreshTimer); };
  }, []);

  useEffect(() => {
    if (state.status !== "reauthentication_required" || hasExplicitLogout()) return undefined;
    markSilentLoginAttempt();
    authApi.login({ prompt: "none" });
    return undefined;
  }, [state.status]);

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
