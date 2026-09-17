import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { authApi, clearAuthStatus, consumeInteractiveLoginAttempt, consumeSilentLoginAttempt, hasExplicitLogout, hasSilentLoginFailure, markExplicitLogout, markSilentLoginAttempt } from "./api.js";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [state, setState] = useState({ status: "loading" });
  const refresh = async () => {
    try {
      const value = await authApi.session();
      if (!value.authenticated) { setState({ status: "unauthenticated" }); return; }
      setState({ status: "authenticated", ...value });
    } catch { setState({ status: "error", retryable: true }); }
  };
  useEffect(() => {
    if (window.location.pathname.replace(/\/$/, "") === "/install") {
      setState({ status: "install" });
      return undefined;
    }
    let cancelled = false;
    async function initialize() {
      try {
        const value = await authApi.session();
        if (cancelled) return;
        if (value.authenticated) {
          consumeInteractiveLoginAttempt();
          consumeSilentLoginAttempt();
          setState({ status: "authenticated", ...value });
          return;
        }
        if (hasExplicitLogout()) {
          clearAuthStatus();
          setState({ status: "unauthenticated" });
          return;
        }
        if (consumeInteractiveLoginAttempt()) {
          clearAuthStatus();
          setState({ status: "unauthenticated" });
          return;
        }
        const silentLoginFailed = hasSilentLoginFailure();
        const silentLoginWasAttempted = consumeSilentLoginAttempt();
        if (!silentLoginFailed && !silentLoginWasAttempted) {
          markSilentLoginAttempt();
          authApi.login({ prompt: "none" });
          return;
        }
        clearAuthStatus();
        setState({ status: "unauthenticated" });
      } catch {
        if (!cancelled) setState({ status: "error", retryable: true });
      }
    }
    initialize();
    return () => { cancelled = true; };
  }, []);
  const value = useMemo(() => ({ state, login: authApi.login, silentLogin: () => { markSilentLoginAttempt(); authApi.login({ prompt: "none" }); }, refreshSession: async () => { await authApi.refresh(); await refresh(); }, logout: async () => { const result = await authApi.logout(); markExplicitLogout(); if (result.logoutUrl) { window.location.assign(result.logoutUrl); return; } setState({ status: "unauthenticated" }); }, refresh }), [state]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() { return useContext(AuthContext); }
