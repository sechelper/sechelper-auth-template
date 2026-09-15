// Public business assembly point. Do not import admin modules here.
// Every business module owns a business/<name>/public-module.js descriptor.
const discoveredPublicModules = import.meta.env?.MODE !== undefined
  ? import.meta.glob("../../business/*/public-module.js", { eager: true, import: "default" })
  : {};

export function enabledPublicBusinessModules(discoveredModules) {
  const modules = Object.entries(discoveredModules)
    .filter(([source]) => source.endsWith("/public-module.js"))
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([source, module]) => ({ ...module, source }));
  validatePublicBusinessModules(modules);
  return modules.filter((module) => module.enabled !== false);
}

export const publicBusinessModules = enabledPublicBusinessModules(discoveredPublicModules);

export function publicBusinessRoutes() {
  return publicBusinessModules.flatMap((module) => module.routes || []).map((route) => ({ ...route, surface: "public" }));
}

export function validateProductionHomepage(modules) {
  const homepageCount = modules
    .filter((module) => module.enabled !== false)
    .flatMap((module) => module.routes || [])
    .filter((route) => route.path === "/").length;
  if (homepageCount !== 1) throw new Error("Production public app requires exactly one business route at /");
}

export function allowsFrontendExamples(environment) {
  return Boolean(environment?.DEV || environment?.MODE === "test");
}

if (import.meta.env?.PROD && import.meta.env?.MODE !== "test") validateProductionHomepage(publicBusinessModules);

export function validatePublicBusinessModules(modules = publicBusinessModules) {
  if (!Array.isArray(modules)) throw new Error("Public business modules must be an array");
  const reservedRoutes = new Set(["/403", "/404", "/500"]);
  const names = new Set();
  const routes = new Set();
  for (const module of modules) {
    if (!module?.name || names.has(module.name)) throw new Error(`Invalid public business module: ${module?.name || "unnamed"}`);
    if (module.routes !== undefined && !Array.isArray(module.routes)) throw new Error(`Invalid public business routes in ${module.name}`);
    const directoryName = module.source?.split("/").at(-2);
    if (directoryName && directoryName !== module.name) throw new Error(`Public business module name does not match directory: ${module.name}`);
    names.add(module.name);
    for (const route of module.routes || []) {
      if (!route || !route.path?.startsWith("/") || !route.element || route.path === "/admin" || route.path.startsWith("/admin/") || routes.has(route.path) || reservedRoutes.has(route.path)) throw new Error(`Invalid public business route in ${module.name}`);
      routes.add(route.path);
    }
  }
}
