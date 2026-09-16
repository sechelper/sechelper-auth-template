import assert from "node:assert/strict";
import test from "node:test";
import { apiOrigin, identitySettingsURL } from "./runtime.js";

test("apiOrigin uses the injected public runtime configuration", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("identity settings URL uses the shared runtime field", () => {
  globalThis.__APP_CONFIG__ = { identitySettingsUrl: "https://passport.example.test/settings" };
  assert.equal(identitySettingsURL(), "https://passport.example.test/settings");
});

test("identity settings URL rejects unsafe values", () => {
  for (const value of [undefined, "", "http://passport.example.test", "javascript:alert(1)", "https://user:pass@passport.example.test"]) {
    globalThis.__APP_CONFIG__ = { identitySettingsUrl: value };
    assert.equal(identitySettingsURL(), "");
  }
});

test("apiOrigin falls back to the browser origin", () => {
  globalThis.__APP_CONFIG__ = undefined;
  globalThis.window = { location: { origin: "http://localhost:5173" } };
  assert.equal(apiOrigin(), "http://localhost:5173");
});
