import { apiOrigin } from "../config/runtime.js";

let csrfToken = "";
const loginPathKey = "auth-template.login-path";
const silentLoginAttemptKey = "auth-template.silent-login-attempt";
const explicitLogoutKey = "auth-template.explicit-logout";
const interactiveLoginAttemptKey = "auth-template.interactive-login-attempt";

export async function request(path, options = {}) {
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

export const authApi = {
  session: () => request("/v1/auth/session"),
  account: () => request("/v1/account/me"),
  refresh: () => request("/v1/auth/refresh", { method: "POST" }),
  logout: () => request("/v1/auth/logout", { method: "POST" }),
  logoutCallback: (state) => request(`/v1/auth/logout/callback?state=${encodeURIComponent(state)}`),
	login: ({ prompt = "login", preservePath = true } = {}) => {
		if (prompt === "login") {
			try {
				window.sessionStorage.removeItem(silentLoginAttemptKey);
				window.sessionStorage.removeItem(explicitLogoutKey);
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
	try { window.sessionStorage.setItem(explicitLogoutKey, "1"); } catch { /* storage is optional */ }
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
