import assert from "node:assert/strict";
import test from "node:test";
import { adminAppConfig, apiOrigin } from "./runtime.js";

test("admin runtime config exposes the injected API origin", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("admin runtime config has a safe default system name", () => {
  assert.equal(adminAppConfig.systemName, "模版演示");
});
