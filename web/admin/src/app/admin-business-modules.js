// Admin business assembly point. Do not import public modules here.
// Every business module owns a business/<name>/admin-module.js descriptor.
const discoveredAdminModules = import.meta.env?.MODE !== undefined
  ? import.meta.glob("../business/*/admin-module.js", { eager: true, import: "default" })
  : {};

export function enabledAdminBusinessModules(discoveredModules) {
  const modules = Object.entries(discoveredModules)
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([source, module]) => ({ ...module, source }));
  validateAdminBusinessModules(modules);
  return modules.filter((module) => module.enabled !== false);
}

export const adminBusinessModules = enabledAdminBusinessModules(discoveredAdminModules);

export function adminBusinessRoutes() {
  return adminBusinessModules.flatMap((module) => module.routes || []).map((route) => ({ ...route, surface: "admin" }));
}

export function adminBusinessNavigation() {
  return adminBusinessModules
    .flatMap((module) => module.navigation || [])
    .sort((left, right) => (left.order ?? 0) - (right.order ?? 0));
}

export function validateAdminBusinessModules(modules = adminBusinessModules) {
  const reservedRoutes = new Set(["/admin/manifest", "/admin/sessions", "/admin/permissions", "/admin/audit-events", "/admin/resources"]);
  const names = new Set();
  const routes = new Set();
  const navigation = new Set();
  for (const module of modules) {
    if (!module?.name || names.has(module.name)) throw new Error(`Invalid admin business module: ${module?.name || "unnamed"}`);
    const directoryName = module.source?.split("/").at(-2);
    if (directoryName && directoryName !== module.name) throw new Error(`Admin business module name does not match directory: ${module.name}`);
    names.add(module.name);
    for (const route of module.routes || []) {
      if (!route.path?.startsWith("/admin/") || !route.element || !route.permission || routes.has(route.path) || reservedRoutes.has(route.path)) throw new Error(`Invalid admin business route in ${module.name}`);
      routes.add(route.path);
    }
    const validateNavigation = (items) => { for (const item of items || []) { if (!item?.label) throw new Error(`Invalid admin navigation in ${module.name}`); if (item.children?.length) validateNavigation(item.children); else { if (!item.path?.startsWith("/admin/") || !item.permission || navigation.has(item.path)) throw new Error(`Invalid admin navigation in ${module.name}`); navigation.add(item.path); if (!routes.has(item.path)) throw new Error(`Admin navigation has no route in ${module.name}`); if (module.routes.find((route) => route.path === item.path)?.permission !== item.permission) throw new Error(`Admin navigation permission does not match route in ${module.name}`); } } };
    validateNavigation(module.navigation);
  }
}
