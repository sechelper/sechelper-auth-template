export function apiOrigin() {
  return globalThis.__APP_CONFIG__?.apiOrigin || window.location.origin;
}
