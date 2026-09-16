import { DeploymentGuidePage } from "../../modules/deployment/DeploymentGuidePage.jsx";
export default {
  name: "configuration",
  routes: [{ path: "/admin/configuration", element: DeploymentGuidePage, permission: "deployment:read" }],
  navigation: [{ path: "/admin/configuration", label: "配置中心", icon: "operations", section: "系统设置", sectionOrder: 300, permission: "deployment:read", order: 10 }],
};
