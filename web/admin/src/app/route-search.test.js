import test from "node:test";
import assert from "node:assert/strict";
import { filterSearchableRoutes, searchableRoutes } from "./route-search.js";

const visibleMenu = [
  {
    label: "订单",
    section: "业务运营",
    children: [
      { label: "订单管理", children: [{ path: "/admin/orders", label: "订单列表", permission: "order:read" }] },
      { path: "/admin/refunds", label: "退款记录", permission: "order:read" },
    ],
  },
  { path: "/admin/audit-events", label: "操作审计", section: "监控与审计", permission: "audit:read" },
];

test("search matches nested labels, section names, and the concrete route path", () => {
  const permissions = ["order:read", "audit:read"];
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "订单", permissions).map((route) => route.path), ["/admin/orders", "/admin/refunds"]);
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "/admin/audit", permissions).map((route) => route.path), ["/admin/audit-events"]);
  assert.equal(filterSearchableRoutes(visibleMenu, "业务运营", permissions).length, 2);
});

test("nested routes are omitted when the current session lacks their permission", () => {
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "订单", ["audit:read"]), []);
  assert.deepEqual(filterSearchableRoutes(visibleMenu, "审计", ["order:read"]), []);
  assert.deepEqual(searchableRoutes(visibleMenu, ["order:read"]).map((route) => route.path), ["/admin/orders", "/admin/refunds"]);
});

test("empty search lists visible paths in menu order and duplicate paths are omitted", () => {
  const menu = [...visibleMenu, { path: "/admin/orders", label: "重复订单入口" }];
  assert.deepEqual(filterSearchableRoutes(menu, "", ["order:read", "audit:read"]).map((route) => route.path), ["/admin/orders", "/admin/refunds", "/admin/audit-events"]);
});
