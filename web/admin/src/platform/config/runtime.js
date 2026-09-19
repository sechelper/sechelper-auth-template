let runtimeConfig = globalThis.__APP_CONFIG__ || {};
const ADMIN_BRAND_NAME = "助安社区";

export function apiOrigin() { return runtimeConfig.apiOrigin || globalThis.__APP_CONFIG__?.apiOrigin || globalThis.window?.location?.origin || ""; }

export function appName() { return String(runtimeConfig.systemName || "").trim(); }

export function formatAdminTitle(name) {
  const value = String(name || "").trim();
  return value ? `${ADMIN_BRAND_NAME} - ${value}` : ADMIN_BRAND_NAME;
}

export function adminTitle() { return formatAdminTitle(appName()); }

export function updateDocumentTitle() {
  const title = adminTitle();
  if (typeof document !== "undefined") document.title = title;
  return title;
}

export async function loadRuntimeConfig() {
  try {
    const response = await fetch(`${apiOrigin()}/v1/runtime-config`, { credentials: "same-origin", cache: "no-store" });
    if (response.ok) {
      runtimeConfig = { ...runtimeConfig, ...(await response.json()).data };
      updateDocumentTitle();
    }
  } catch { /* Runtime configuration falls back to the same-origin safe defaults. */ }
  return runtimeConfig;
}

export function oidcAccountURL() { return runtimeConfig.identity?.issuer || globalThis.window?.location?.origin || ""; }



export const adminAppConfig = {
  get systemName() { return appName(); },
};
