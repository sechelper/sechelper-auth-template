export function apiOrigin() { return globalThis.__APP_CONFIG__?.apiOrigin || window.location.origin; }

export function identitySettingsURL() {
  const configured = globalThis.__APP_CONFIG__?.identitySettingsUrl;
  if (typeof configured !== "string" || !configured.trim()) return "";
  try {
    const url = new URL(configured.trim());
    if (url.protocol !== "https:" || !url.hostname || url.username || url.password) return "";
    return url.href;
  } catch {
    return "";
  }
}

export const identityCenterURL = identitySettingsURL;

export const adminAppConfig = {
  systemName: globalThis.__APP_CONFIG__?.systemName || "模版演示",
};
