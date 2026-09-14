// Admin business assembly point. Do not import public modules here.
// A module exports { name, routes, navigation }, with permission on protected entries.
export const adminBusinessModules = [];

export function adminBusinessRoutes() {
  return adminBusinessModules.flatMap((module) => module.routes || []).map((route) => ({ ...route, surface: "admin" }));
}

export function adminBusinessNavigation() {
  return adminBusinessModules.flatMap((module) => module.navigation || []);
}

export function validateAdminBusinessModules() {
  const names = new Set();
  for (const module of adminBusinessModules) {
    if (!module?.name || names.has(module.name)) throw new Error(`Invalid admin business module: ${module?.name || "unnamed"}`);
    names.add(module.name);
    for (const route of module.routes || []) {
      if (!route.path?.startsWith("/admin/") || !route.element) throw new Error(`Invalid admin business route in ${module.name}`);
    }
    for (const item of module.navigation || []) {
      if (!item.path?.startsWith("/admin/") || !item.permission) throw new Error(`Admin navigation requires permission in ${module.name}`);
    }
  }
}
