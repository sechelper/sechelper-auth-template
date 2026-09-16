const OIDC_DISCOVERY_ORIGIN = "https://passport-test.sechelper.com";

export function apiOrigin() {
  return globalThis.__APP_CONFIG__?.apiOrigin || window.location.origin;
}

export function oidcAccountURL() {
  return OIDC_DISCOVERY_ORIGIN;
}
