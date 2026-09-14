import test from "node:test";
import assert from "node:assert/strict";
import { adminBusinessModules, adminBusinessRoutes, validateAdminBusinessModules } from "./admin-business-modules.js";

test("admin business registry is isolated and valid by default", () => {
  validateAdminBusinessModules();
  assert.deepEqual(adminBusinessModules, []);
  assert.deepEqual(adminBusinessRoutes(), []);
});
