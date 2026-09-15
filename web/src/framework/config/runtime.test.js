import assert from "node:assert/strict";
import test from "node:test";
import { apiOrigin } from "./runtime.js";

test("apiOrigin uses the injected public runtime configuration", () => {
  globalThis.__APP_CONFIG__ = { apiOrigin: "https://api.example.test" };
  assert.equal(apiOrigin(), "https://api.example.test");
});

test("apiOrigin falls back to the browser origin", () => {
  globalThis.__APP_CONFIG__ = undefined;
  globalThis.window = { location: { origin: "http://localhost:5173" } };
  assert.equal(apiOrigin(), "http://localhost:5173");
});
