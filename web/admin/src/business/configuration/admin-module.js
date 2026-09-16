import { DeploymentGuidePage } from "../../modules/deployment/DeploymentGuidePage.jsx";
export default {
  name: "configuration",
  routes: [{ path: "/admin/deployment", element: DeploymentGuidePage, permission: "deployment:read" }],
  navigation: [{ path: "/admin/deployment", label: "部署与配置中心", icon: "deployment", section: "系统设置", sectionOrder: 300, permission: "deployment:read", order: 10 }],
};
