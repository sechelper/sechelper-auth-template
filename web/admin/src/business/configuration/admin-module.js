import { ConfigurationPage } from "../../modules/configuration/ConfigurationPage.jsx";
export default {
  name: "configuration",
  routes: [{ path: "/admin/configuration", element: ConfigurationPage, permission: "configuration:read" }],
  navigation: [{ path: "/admin/configuration", label: "配置中心", icon: "operations", section: "系统设置", sectionOrder: 300, permission: "configuration:read", order: 10 }],
};
