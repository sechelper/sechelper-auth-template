import assert from "node:assert/strict";
import test from "node:test";
import { adminAppConfig, adminTitle, apiOrigin, formatAdminTitle, oidcAccountURL } from "./runtime.js";

test("admin runtime config exposes the injected API origin", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("admin runtime config does not invent an application name before APP_NAME loads", () => {
  assert.equal(adminAppConfig.systemName, "");
});

test("admin browser title uses the community brand and APP_NAME", () => {
  assert.equal(formatAdminTitle("auth-template"), "助安社区 - auth-template");
  assert.equal(adminTitle(), "助安社区");
});

test("OIDC account URL falls back to the current origin before runtime config loads", () => {
  globalThis.window = { location: { origin: "http://localhost:5173" } };
  assert.equal(oidcAccountURL(), "http://localhost:5173");
});
