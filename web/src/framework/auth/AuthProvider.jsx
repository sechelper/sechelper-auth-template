import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { authApi } from "./api.js";

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
  useEffect(() => { refresh(); }, []);
  const value = useMemo(() => ({ state, login: authApi.login, refreshSession: async () => { await authApi.refresh(); await refresh(); }, logout: async () => { await authApi.logout(); await refresh(); }, refresh }), [state]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() { return useContext(AuthContext); }
