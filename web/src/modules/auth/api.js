import { apiOrigin } from "../../platform/config/runtime.js";

let csrfToken = "";

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
  refresh: () => request("/v1/auth/refresh", { method: "POST" }),
  logout: () => request("/v1/auth/logout", { method: "POST" }),
  login: () => { window.location.assign(`${apiOrigin()}/v1/auth/login`); },
};
