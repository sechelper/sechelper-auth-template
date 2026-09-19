import test from "node:test";
import assert from "node:assert/strict";
import { filterSearchableRoutes, searchableRoutes } from "./route-search.js";

const visibleMenu = [
  {
    label: "目录",
    section: "业务运营",
    children: [
      { label: "目录管理", children: [{ path: "/admin/catalog", label: "目录列表", permission: "resource:read" }] },
      { path: "/admin/resources", label: "资源记录", permission: "resource:read" },
    ],
  },
  { path: "/admin/audit-events", label: "操作审计", section: "监控与审计", permission: "audit:read" },
];

test("search matches nested labels, section names, and the concrete route path", () => {
  const permissions = ["resource:read", "audit:read"];
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "目录", permissions).map((route) => route.path), ["/admin/catalog", "/admin/resources"]);
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "/admin/audit", permissions).map((route) => route.path), ["/admin/audit-events"]);
  assert.equal(filterSearchableRoutes(visibleMenu, "业务运营", permissions).length, 2);
});

test("nested routes are omitted when the current session lacks their permission", () => {
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "目录", ["audit:read"]), []);
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "审计", ["resource:read"]), []);
  assert.deepEqual(searchableRoutes(visibleMenu, ["resource:read"]).map((route) => route.path), ["/admin/catalog", "/admin/resources"]);
});

test("empty search lists visible paths in menu order and duplicate paths are omitted", () => {
  const menu = [...visibleMenu, { path: "/admin/catalog", label: "重复目录入口" }];
  assert.deepEqual(filterSearchableRoutes(menu, "", ["resource:read", "audit:read"]).map((route) => route.path), ["/admin/catalog", "/admin/resources", "/admin/audit-events"]);
});
