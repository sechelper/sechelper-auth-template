import assert from "node:assert/strict";
import test from "node:test";
import { initialAdminPath, normalizePathname } from "./navigation-path.js";

test("normalizePathname keeps admin routes and removes the trailing slash", () => {
  assert.equal(normalizePathname("/admin/"), "/admin");
  assert.equal(normalizePathname("/admin/sessions"), "/admin/sessions");
});

test("a public entry intent always opens the admin overview", () => {
  const values = new Map([["sechelper:admin-entry", "root"]]);
  const storage = {
    getItem: (key) => values.get(key) ?? null,
    removeItem: (key) => values.delete(key),
  };
  const replacements = [];
  const history = { replaceState: (...args) => replacements.push(args) };

  assert.equal(initialAdminPath({ pathname: "/admin/sessions", storage, history }), "/admin");
  assert.equal(values.has("sechelper:admin-entry"), false);
  assert.deepEqual(replacements, [[{}, "", "/admin/"]]);
});

test("ordinary admin navigation preserves the requested page", () => {
  const storage = { getItem: () => null, removeItem: () => assert.fail("must not consume storage") };
  const history = { replaceState: () => assert.fail("must not replace history") };

  assert.equal(initialAdminPath({ pathname: "/admin/sessions", storage, history }), "/admin/sessions");
});
