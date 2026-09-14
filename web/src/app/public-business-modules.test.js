import test from "node:test";
import assert from "node:assert/strict";
import { publicBusinessModules, publicBusinessRoutes, validatePublicBusinessModules } from "./public-business-modules.js";

test("public business registry is isolated and valid by default", () => {
  validatePublicBusinessModules();
  assert.deepEqual(publicBusinessModules, []);
  assert.deepEqual(publicBusinessRoutes(), []);
});
