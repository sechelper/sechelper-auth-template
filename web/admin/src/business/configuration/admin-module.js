import { ConfigurationPage } from "../../modules/configuration/ConfigurationPage.jsx";
import { DeploymentGuidePage } from "../../modules/deployment/DeploymentGuidePage.jsx";

export default {
  name: "configuration",
  routes: [
    { path: "/admin/configuration/framework", element: DeploymentGuidePage, permission: "deployment:read" },
    { path: "/admin/configuration/business", element: ConfigurationPage, permission: "configuration:read" },
  ],
  navigation: [{ label: "配置中心", icon: "operations", section: "系统设置", sectionOrder: 300, order: 10, children: [
    { path: "/admin/configuration/framework", label: "框架配置", permission: "deployment:read" },
    { path: "/admin/configuration/business", label: "业务配置", permission: "configuration:read" },
  ] }],
};
