import { OrdersPage } from "./OrdersPage.jsx";

// Development/test example only. Production Vite module discovery excludes
// example-module.js files before they enter the module graph.
export default {
  name: "orders",
  routes: [{ path: "/admin/orders", element: OrdersPage, permission: "order:read" }],
  navigation: [
    { label: "订单", icon: "orders", section: "业务运营", sectionOrder: 50, sectionCollapsible: false, children: [{ label: "订单1", children: [{ path: "/admin/orders", label: "订单2", permission: "order:read" }] }] },
  ],
};
