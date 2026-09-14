// Public business assembly point. Do not import admin modules here.
// A public module exports { name, routes } and may use only framework clients.
export const publicBusinessModules = [];

export function publicBusinessRoutes() {
  return publicBusinessModules.flatMap((module) => module.routes || []).map((route) => ({ ...route, surface: "public" }));
}

export function validatePublicBusinessModules() {
  const names = new Set();
  for (const module of publicBusinessModules) {
    if (!module?.name || names.has(module.name)) throw new Error(`Invalid public business module: ${module?.name || "unnamed"}`);
    names.add(module.name);
    for (const route of module.routes || []) {
      if (!route.path || !route.element || route.path.startsWith("/admin")) throw new Error(`Invalid public business route in ${module.name}`);
    }
  }
}
