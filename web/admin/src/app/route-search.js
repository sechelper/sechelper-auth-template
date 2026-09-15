function normalize(value) {
  return String(value || "").trim().toLocaleLowerCase();
}

export function searchableRoutes(menu = [], permissions = []) {
  const routes = [];
  const seenPaths = new Set();
  const grantedPermissions = new Set(permissions);

  function visit(items, ancestors = [], parentAllowed = true) {
    for (const item of items || []) {
      const allowed = parentAllowed && (!item.permission || grantedPermissions.has(item.permission));
      if (!allowed) continue;
      const trail = [...ancestors, item.label].filter(Boolean);
      if (item.path && !seenPaths.has(item.path)) {
        seenPaths.add(item.path);
        routes.push({ path: item.path, label: item.label || item.path, trail });
      }
      if (item.children?.length) visit(item.children, trail, allowed);
    }
  }

  for (const item of menu) {
    const ancestors = item.section ? [item.section] : [];
    visit([item], ancestors);
  }
  return routes;
}

export function filterSearchableRoutes(menu, query, permissions) {
  const routes = searchableRoutes(menu, permissions);
  const normalizedQuery = normalize(query);
  if (!normalizedQuery) return routes;
  return routes.filter((route) => normalize(`${route.trail.join(" ")} ${route.path}`).includes(normalizedQuery));
}
