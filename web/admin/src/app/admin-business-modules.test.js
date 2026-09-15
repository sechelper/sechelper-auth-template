import test from "node:test";
import assert from "node:assert/strict";
import { adminBusinessModules, adminBusinessRoutes, enabledAdminBusinessModules, validateAdminBusinessModules } from "./admin-business-modules.js";

test("admin business registry is isolated and valid by default", () => {
  validateAdminBusinessModules();
  assert.deepEqual(adminBusinessModules, []);
  assert.deepEqual(adminBusinessRoutes(), []);
});

test("admin modules are discovered deterministically and disabled examples are omitted", () => {
  const sample = { name: "sample", enabled: false };
  const billing = { name: "billing" };
  assert.deepEqual(enabledAdminBusinessModules({ "z/sample/admin-module.js": sample, "a/billing/admin-module.js": billing }), [{ ...billing, source: "a/billing/admin-module.js" }]);
});

test("admin module routes and navigation require permissions and unique paths", () => {
  const Page = () => null;
  assert.doesNotThrow(() => validateAdminBusinessModules([{ name: "billing", routes: [{ path: "/admin/billing", element: Page, permission: "billing:read" }], navigation: [{ path: "/admin/billing", label: "Billing", permission: "billing:read" }] }]));
  assert.throws(() => validateAdminBusinessModules([{ name: "billing", routes: [{ path: "/admin/billing", element: Page }] }]));
  assert.throws(() => validateAdminBusinessModules([{ name: "billing", routes: [{ path: "/admin/billing", element: Page, permission: "billing:read" }], navigation: [{ path: "/admin/missing", permission: "billing:read" }] }]));
  assert.throws(() => validateAdminBusinessModules([{ name: "billing", source: "business/invoices/admin-module.js", routes: [] }]));
  assert.throws(() => validateAdminBusinessModules([{ name: "billing", routes: [{ path: "/admin/manifest", element: Page, permission: "billing:read" }] }]));
  assert.throws(() => validateAdminBusinessModules([{ name: "billing", routes: [{ path: "/admin/billing", element: Page, permission: "billing:read" }], navigation: [{ path: "/admin/billing", permission: "billing:write" }] }]));
  assert.doesNotThrow(() => validateAdminBusinessModules([{ name: "nested", routes: [{ path: "/admin/nested", element: Page, permission: "nested:read" }], navigation: [{ label: "订单", children: [{ label: "订单1", children: [{ path: "/admin/nested", label: "订单2", permission: "nested:read" }] }] }] }]));
});
