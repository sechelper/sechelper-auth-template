export function apiOrigin() { return globalThis.__APP_CONFIG__?.apiOrigin || window.location.origin; }

export const adminAppConfig = {
  systemName: globalThis.__APP_CONFIG__?.systemName || "模版演示",
};
