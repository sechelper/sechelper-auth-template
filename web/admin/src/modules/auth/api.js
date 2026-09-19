import { apiOrigin } from "../../platform/config/runtime.js";

let csrfToken = "";
let refreshInFlight = null;
let authEventHandler = null;
export const SESSION_REFRESH_LEEWAY_MS = 120000;
const loginPathKey = "auth-template.admin-login-path";
const silentLoginAttemptKey = "auth-template.silent-login-attempt";
const explicitLogoutKey = "auth-template.explicit-logout";

async function rawRequest(path, options = {}) {
  const response = await fetch(`${apiOrigin()}${path}`, { credentials: "include", ...options, headers: { "Content-Type": "application/json", ...(csrfToken ? { "X-CSRF-Token": csrfToken } : {}), ...(options.headers || {}) } });
  csrfToken = response.headers.get("X-CSRF-Token") || csrfToken;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw Object.assign(new Error(body?.error?.message || "请求失败"), { code: body?.error?.code, status: response.status });
  return body;
}

export async function request(path, options = {}, { retry = true } = {}) {
  try {
    return await rawRequest(path, options);
  } catch (error) {
    if (error?.status === 401 && retry && path !== "/v1/auth/admin/refresh" && path !== "/v1/auth/admin/logout") {
      try {
        await authApi.refresh();
        return await request(path, options, { retry: false });
      } catch (refreshError) {
        authEventHandler?.({ type: "reauthentication_required", error: refreshError });
        throw refreshError;
      }
    }
    if (error?.status === 403 && error?.code === "CSRF_FAILED" && retry && path !== "/v1/auth/admin/session") {
      await rawRequest("/v1/auth/admin/session");
      return request(path, options, { retry: false });
    }
    throw error;
  }
}

export function setAuthEventHandler(handler) { authEventHandler = handler; return () => { if (authEventHandler === handler) authEventHandler = null; }; }

export function sessionNeedsRefresh(expiresAt, now = Date.now()) {
  const expires = Date.parse(expiresAt || "");
  return Number.isFinite(expires) && expires - now <= SESSION_REFRESH_LEEWAY_MS;
}

export const authApi = {
  session: () => request("/v1/auth/admin/session"),
  account: () => request("/v1/admin/account"),
  authorization: () => request("/v1/admin/authorization/me"),
  refresh: () => {
    if (!refreshInFlight) {
      refreshInFlight = rawRequest("/v1/auth/admin/refresh", { method: "POST" }).catch(async (error) => {
        if (error?.status !== 403 || error?.code !== "CSRF_FAILED") throw error;
        await rawRequest("/v1/auth/admin/session");
        return rawRequest("/v1/auth/admin/refresh", { method: "POST" });
      }).finally(() => { refreshInFlight = null; });
    }
    return refreshInFlight;
  },
  logout: () => request("/v1/auth/admin/logout", { method: "POST" }),
  login: ({ prompt = "login", preservePath = true } = {}) => {
    if (prompt === "login") clearExplicitLogout();
    if (preservePath) { try { window.sessionStorage.setItem(loginPathKey, `${window.location.pathname}${window.location.search}${window.location.hash}`); } catch { /* storage is optional */ } }
    const returnTo = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    window.location.assign(`${apiOrigin()}/v1/auth/admin/login?prompt=${encodeURIComponent(prompt)}&return_to=${encodeURIComponent(returnTo)}`);
  },
};

export function markExplicitLogout() { try { window.sessionStorage.setItem(explicitLogoutKey, "1"); window.localStorage.setItem(explicitLogoutKey, String(Date.now())); } catch { /* storage is optional */ } }
export function clearExplicitLogout() { try { window.sessionStorage.removeItem(explicitLogoutKey); window.localStorage.removeItem(explicitLogoutKey); } catch { /* storage is optional */ } }
export function hasExplicitLogout() { try { return window.sessionStorage.getItem(explicitLogoutKey) === "1"; } catch { return false; } }
export function markSilentLoginAttempt() { try { window.sessionStorage.setItem(silentLoginAttemptKey, "1"); } catch { /* storage is optional */ } }
export function consumeSilentLoginAttempt() { try { const attempted = window.sessionStorage.getItem(silentLoginAttemptKey) === "1"; window.sessionStorage.removeItem(silentLoginAttemptKey); return attempted; } catch { return false; } }
export function hasSilentLoginFailure() { return new URLSearchParams(window.location.search).get("auth") === "login-required"; }
export function clearAuthStatus() { const url = new URL(window.location.href); if (!url.searchParams.has("auth")) return; url.searchParams.delete("auth"); window.history.replaceState({}, "", `${url.pathname}${url.search}${url.hash}`); }
export function subscribeToAuthChanges(onChange) { const handler = (event) => { if (event.key === explicitLogoutKey && event.newValue) onChange({ type: "logout" }); }; window.addEventListener("storage", handler); return () => window.removeEventListener("storage", handler); }
export function restoreLoginPath() { try { const path = window.sessionStorage.getItem(loginPathKey); window.sessionStorage.removeItem(loginPathKey); if (path && path.startsWith("/admin") && !path.startsWith("//")) { window.history.replaceState({}, "", path); window.dispatchEvent(new PopStateEvent("popstate")); } } catch { /* storage is optional */ } }
