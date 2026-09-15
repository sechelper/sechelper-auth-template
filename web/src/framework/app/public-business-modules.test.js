import test from "node:test";
import assert from "node:assert/strict";
import { allowsFrontendExamples, enabledPublicBusinessModules, publicBusinessModules, publicBusinessRoutes, validateProductionHomepage, validatePublicBusinessModules } from "./public-business-modules.js";

test("public business registry is isolated and valid by default", () => {
  validatePublicBusinessModules();
  assert.deepEqual(publicBusinessModules, []);
  assert.deepEqual(publicBusinessRoutes(), []);
});

test("public modules are discovered deterministically and disabled examples are omitted", () => {
  const sample = { name: "sample", enabled: false };
  const catalog = { name: "catalog" };
  const profileExample = { name: "profile-example" };
  assert.deepEqual(enabledPublicBusinessModules({ "z/sample/public-module.js": sample, "a/catalog/public-module.js": catalog, "x/profile-example/example-module.js": profileExample }), [{ ...catalog, source: "a/catalog/public-module.js" }]);
});

test("public module routes must be unique and stay outside admin", () => {
  const Page = () => null;
  assert.doesNotThrow(() => validatePublicBusinessModules([{ name: "catalog", routes: [{ path: "/catalog", element: Page }] }]));
  assert.doesNotThrow(() => validatePublicBusinessModules([{ name: "community-home", routes: [{ path: "/", element: Page }] }]));
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", routes: [{ path: "/admin/catalog", element: Page }] }]));
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", routes: [{ path: "/admin", element: Page }] }]));
  assert.throws(() => validatePublicBusinessModules([{ name: "one", routes: [{ path: "/catalog", element: Page }] }, { name: "two", routes: [{ path: "/catalog", element: Page }] }]));
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", source: "business/store/public-module.js", routes: [] }]));
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", routes: [{ path: "/404", element: Page }] }]));
});

test("production requires exactly one enabled business homepage", () => {
  const Page = () => null;
  assert.throws(() => validateProductionHomepage([]), /exactly one business route at \/$/);
  assert.doesNotThrow(() => validateProductionHomepage([{ name: "community-home", routes: [{ path: "/", element: Page }] }]));
  assert.throws(() => validateProductionHomepage([{ name: "catalog", routes: [{ path: "/catalog", element: Page }] }]), /exactly one business route at \/$/);
  assert.throws(() => validateProductionHomepage([{ name: "example", enabled: false, routes: [{ path: "/", element: Page }] }]), /exactly one business route at \/$/);
});

test("examples are enabled only in development or explicit test mode", () => {
  assert.equal(allowsFrontendExamples({ DEV: true, MODE: "development" }), true);
  assert.equal(allowsFrontendExamples({ DEV: false, MODE: "test" }), true);
  assert.equal(allowsFrontendExamples({ DEV: false, MODE: "production" }), false);
});

test("public module and route descriptors must have valid collection shapes", () => {
  assert.throws(() => validatePublicBusinessModules({ name: "catalog" }), /must be an array/);
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", routes: {} }]), /Invalid public business routes/);
  assert.throws(() => validatePublicBusinessModules([{ name: "catalog", routes: [null] }]), /Invalid public business route/);
});
