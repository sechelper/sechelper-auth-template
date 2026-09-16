import assert from "node:assert/strict";
import test from "node:test";
import { adminAppConfig, apiOrigin, identitySettingsURL } from "./runtime.js";

test("admin runtime config exposes the injected API origin", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("admin runtime config has a safe default system name", () => {
  assert.equal(adminAppConfig.systemName, "模版演示");
});

test("identity settings URL comes from the shared runtime config", () => {
  globalThis.__APP_CONFIG__ = { identitySettingsUrl: "https://identity.example.test/account/settings" };
  assert.equal(identitySettingsURL(), "https://identity.example.test/account/settings");
});

test("identity settings URL rejects missing, insecure, and credentialed URLs", () => {
  for (const value of [undefined, "", "http://identity.example.test", "javascript:alert(1)", "https://user:pass@identity.example.test"]) {
    globalThis.__APP_CONFIG__ = { identitySettingsUrl: value };
    assert.equal(identitySettingsURL(), "");
  }
});
