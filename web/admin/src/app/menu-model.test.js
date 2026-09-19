import test from "node:test";
import assert from "node:assert/strict";
import { filterMenuByPermission, findMenuExpansionKeys, findMenuTrail, groupMenuItems, validateBusinessMenuSections } from "./menu-model.js";

test("menu permission filtering recursively removes inaccessible routes and empty parents", () => {
  const menu = [{
    label: "目录",
    section: "业务运营",
    children: [
      { label: "查看", path: "/admin/catalog", permission: "resource:read" },
      { label: "配置", path: "/admin/catalog/settings", permission: "resource:configure" },
    ],
  }];
  const visible = filterMenuByPermission(menu, (permission) => permission === "resource:read");
  assert.deepEqual(visible, [{
    label: "目录",
    section: "业务运营",
    children: [{ label: "查看", path: "/admin/catalog", permission: "resource:read" }],
  }]);
  assert.deepEqual(filterMenuByPermission(menu, () => false), []);
});

test("a permitted parent permission gates every nested menu item", () => {
  const menu = [{
    label: "财务",
    permission: "finance:access",
    children: [{ label: "报表", path: "/admin/reports", permission: "report:read" }],
  }];
  assert.deepEqual(filterMenuByPermission(menu, (permission) => permission === "report:read"), []);
  assert.equal(filterMenuByPermission(menu, () => true).length, 1);
});

test("menu groups preserve order and keep unsectioned entries at root", () => {
  assert.deepEqual(groupMenuItems([
    { path: "/admin", label: "概览" },
    { path: "/admin/catalog", label: "目录", section: "业务运营" },
    { path: "/admin/stock", label: "库存", section: "业务运营" },
    { path: "/admin/audit", label: "审计", section: "监控与审计" },
  ]), [
    { key: "__root__", label: "", collapsible: false, order: Number.NEGATIVE_INFINITY, firstIndex: 0, items: [{ path: "/admin", label: "概览" }] },
    { key: "业务运营", label: "业务运营", collapsible: false, order: 1000, firstIndex: 1, items: [{ path: "/admin/catalog", label: "目录", section: "业务运营" }, { path: "/admin/stock", label: "库存", section: "业务运营" }] },
    { key: "监控与审计", label: "监控与审计", collapsible: false, order: 1000, firstIndex: 3, items: [{ path: "/admin/audit", label: "审计", section: "监控与审计" }] },
  ]);
});

test("business sections can move around fixed framework sections using declaration order", () => {
  const sections = groupMenuItems([
    { path: "/admin", label: "概览" },
    { path: "/admin/security", label: "安全", section: "账号与安全", sectionOrder: 100 },
    { path: "/admin/audit", label: "审计", section: "监控与审计", sectionOrder: 200 },
    { path: "/admin/catalog", label: "目录", section: "业务运营", sectionOrder: -10 },
    { path: "/admin/reports", label: "报表", section: "业务分析", sectionOrder: 1000 },
  ]);
  assert.deepEqual(sections.map((section) => section.key), ["__root__", "业务运营", "账号与安全", "监控与审计", "业务分析"]);
});

test("business section declarations cannot override reserved groups or conflict", () => {
  assert.doesNotThrow(() => validateBusinessMenuSections([
    { section: "业务运营", sectionOrder: -10 },
    { section: "业务运营", sectionOrder: -10 },
  ]));
  assert.throws(() => validateBusinessMenuSections([{ section: "账号与安全" }]), /framework section/);
  assert.throws(() => validateBusinessMenuSections([
    { section: "业务运营", sectionOrder: -10 },
    { section: "业务运营", sectionOrder: 10 },
  ]), /Conflicting admin section configuration/);
  assert.throws(() => validateBusinessMenuSections([{ section: "业务运营", sectionOrder: Number.NaN }]), /Invalid sectionOrder/);
});

test("menu trail resolves nested business paths without the peer section title", () => {
  const groups = groupMenuItems([
    { label: "目录", section: "业务运营", children: [{ label: "项目1", children: [{ path: "/admin/catalog", label: "项目2" }] }] },
    { path: "/admin/audit-events", label: "操作审计", section: "监控与审计" },
  ]);
  assert.deepEqual(findMenuTrail(groups, "/admin/catalog"), ["目录", "项目1", "项目2"]);
  assert.deepEqual(findMenuTrail(groups, "/admin/audit-events"), ["操作审计"]);
  assert.deepEqual(findMenuTrail(groups, "/admin/unknown"), []);
  assert.deepEqual(findMenuExpansionKeys(groups, "/admin/catalog"), ["业务运营", "业务运营:0", "业务运营:0:0"]);
});
