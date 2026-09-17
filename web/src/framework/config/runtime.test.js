import assert from "node:assert/strict";
import test from "node:test";
import { apiOrigin, oidcAccountURL } from "./runtime.js";

test("apiOrigin uses the injected public runtime configuration", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("OIDC account URL falls back to the current origin before runtime config loads", () => {
  globalThis.window = { location: { origin: "http://localhost:5173" } };
  assert.equal(oidcAccountURL(), "http://localhost:5173");
});

test("apiOrigin falls back to the browser origin", () => {
  globalThis.__APP_CONFIG__ = undefined;
  globalThis.window = { location: { origin: "http://localhost:5173" } };
  assert.equal(apiOrigin(), "http://localhost:5173");
});
