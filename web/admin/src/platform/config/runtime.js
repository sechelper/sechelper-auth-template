let runtimeConfig = globalThis.__APP_CONFIG__ || {};

export function apiOrigin() { return runtimeConfig.apiOrigin || globalThis.__APP_CONFIG__?.apiOrigin || globalThis.window?.location?.origin || ""; }

export async function loadRuntimeConfig() {
  try {
    const response = await fetch(`${apiOrigin()}/v1/runtime-config`, { credentials: "same-origin", cache: "no-store" });
    if (response.ok) runtimeConfig = { ...runtimeConfig, ...(await response.json()).data };
  } catch { /* Runtime configuration falls back to the same-origin safe defaults. */ }
  return runtimeConfig;
}

export function oidcAccountURL() { return runtimeConfig.identity?.issuer || globalThis.window?.location?.origin || ""; }



export const adminAppConfig = {
  get systemName() { return runtimeConfig.systemName || "模版演示"; },
};
