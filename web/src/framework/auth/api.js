import { apiOrigin } from "../config/runtime.js";

let csrfToken = "";
let refreshInFlight = null;
let authEventHandler = null;
const loginPathKey = "auth-template.login-path";
const silentLoginAttemptKey = "auth-template.silent-login-attempt";
const explicitLogoutKey = "auth-template.explicit-logout";
const interactiveLoginAttemptKey = "auth-template.interactive-login-attempt";

async function rawRequest(path, options = {}) {
  const response = await fetch(`${apiOrigin()}${path}`, {
    credentials: "include",
    ...options,
    headers: { "Content-Type": "application/json", ...(csrfToken ? { "X-CSRF-Token": csrfToken } : {}), ...(options.headers || {}) },
  });
  csrfToken = response.headers.get("X-CSRF-Token") || csrfToken;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw Object.assign(new Error(body?.error?.message || "请求失败"), { code: body?.error?.code, status: response.status });
  return body;
}

export async function request(path, options = {}, { retry = true } = {}) {
  try {
    return await rawRequest(path, options);
  } catch (error) {
    if (error?.status === 401 && retry && path !== "/v1/auth/refresh" && path !== "/v1/auth/logout") {
      try {
        await authApi.refresh();
        return await request(path, options, { retry: false });
      } catch (refreshError) {
        authEventHandler?.({ type: "reauthentication_required", error: refreshError });
        throw refreshError;
      }
    }
    if (error?.status === 403 && error?.code === "CSRF_FAILED" && retry && path !== "/v1/auth/session") {
      await rawRequest("/v1/auth/session");
      return request(path, options, { retry: false });
    }
    throw error;
  }
}

export function setAuthEventHandler(handler) { authEventHandler = handler; return () => { if (authEventHandler === handler) authEventHandler = null; }; }

export const authApi = {
  session: () => request("/v1/auth/session"),
  account: () => request("/v1/account/me"),
  refresh: () => { if (!refreshInFlight) refreshInFlight = rawRequest("/v1/auth/refresh", { method: "POST" }).finally(() => { refreshInFlight = null; }); return refreshInFlight; },
  logout: () => request("/v1/auth/logout", { method: "POST" }),
  logoutCallback: (state) => request(`/v1/auth/logout/callback?state=${encodeURIComponent(state)}`),
	login: ({ prompt = "login", preservePath = true } = {}) => {
		if (prompt === "login") {
			try {
				window.sessionStorage.removeItem(silentLoginAttemptKey);
				clearExplicitLogout();
				window.sessionStorage.setItem(interactiveLoginAttemptKey, "1");
			} catch { /* storage is optional */ }
		}
		if (preservePath) {
			const path = `${window.location.pathname}${window.location.search}${window.location.hash}`;
			if (path.startsWith("/") && !path.startsWith("//")) {
				try {
					const existingPath = window.sessionStorage.getItem(loginPathKey);
					if (path !== "/" || !existingPath || existingPath === "/") window.sessionStorage.setItem(loginPathKey, path);
				} catch { /* storage is optional */ }
			}
		}
		window.location.assign(`${apiOrigin()}/v1/auth/login?prompt=${encodeURIComponent(prompt)}`);
	},
};

export function markSilentLoginAttempt() {
	try { window.sessionStorage.setItem(silentLoginAttemptKey, "1"); } catch { /* storage is optional */ }
}

export function markExplicitLogout() {
	try { window.sessionStorage.setItem(explicitLogoutKey, "1"); window.localStorage.setItem(explicitLogoutKey, String(Date.now())); } catch { /* storage is optional */ }
}

export function clearExplicitLogout() {
	try { window.sessionStorage.removeItem(explicitLogoutKey); window.localStorage.removeItem(explicitLogoutKey); } catch { /* storage is optional */ }
}

export function subscribeToAuthChanges(onChange) {
	const handler = (event) => { if (event.key === explicitLogoutKey && event.newValue) onChange({ type: "logout" }); };
	window.addEventListener("storage", handler);
	return () => window.removeEventListener("storage", handler);
}

export function hasExplicitLogout() {
	try { return window.sessionStorage.getItem(explicitLogoutKey) === "1"; } catch { return false; }
}

export function consumeInteractiveLoginAttempt() {
	try {
		const attempted = window.sessionStorage.getItem(interactiveLoginAttemptKey) === "1";
		window.sessionStorage.removeItem(interactiveLoginAttemptKey);
		return attempted;
	} catch {
		return false;
	}
}

export function consumeSilentLoginAttempt() {
	try {
		const attempted = window.sessionStorage.getItem(silentLoginAttemptKey) === "1";
		window.sessionStorage.removeItem(silentLoginAttemptKey);
		return attempted;
	} catch {
		return false;
	}
}

export function hasSilentLoginFailure() {
	return new URLSearchParams(window.location.search).get("auth") === "login-required";
}

export function clearAuthStatus() {
	const url = new URL(window.location.href);
	if (!url.searchParams.has("auth")) return;
	url.searchParams.delete("auth");
	window.history.replaceState({}, "", `${url.pathname}${url.search}${url.hash}`);
}

export function restoreLoginPath() {
	try {
		const path = window.sessionStorage.getItem(loginPathKey);
		window.sessionStorage.removeItem(loginPathKey);
		if (path && path.startsWith("/") && !path.startsWith("//") && path !== "/") {
			if (/^\/admin(?:\/|$)/.test(path)) {
				window.location.assign(path);
				return;
			}
			window.history.replaceState({}, "", path);
			window.dispatchEvent(new PopStateEvent("popstate"));
		}
	} catch { /* storage is optional */ }
}
