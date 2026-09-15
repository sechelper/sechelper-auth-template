import { ComponentReferencePage } from "./ComponentReferencePage.jsx";

// Test-only documentation surface. It is not discovered in development or
// production builds.
export default {
  name: "component-reference",
  routes: [{ path: "/admin/component-reference", element: ComponentReferencePage, permission: "admin:access" }],
  navigation: [
    { path: "/admin/component-reference", label: "组件参考", icon: "resources", section: "开发参考", sectionOrder: 90, order: 1, permission: "admin:access" },
  ],
};
