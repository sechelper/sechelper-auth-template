const DEFAULT_BUSINESS_SECTION_ORDER = 1000;
const FRAMEWORK_SECTIONS = new Set(["账号与安全", "监控与审计"]);

export function filterMenuByPermission(items, hasPermission) {
  return (items || []).flatMap((item) => {
    if (item.permission && !hasPermission(item.permission)) return [];

    if (item.children?.length) {
      const children = filterMenuByPermission(item.children, hasPermission);
      if (!children.length && !item.path) return [];
      return [{ ...item, children }];
    }

    return item.path ? [item] : [];
  });
}

export function validateBusinessMenuSections(items) {
  const sections = new Map();
  for (const item of items || []) {
    if (!item.section) continue;
    if (FRAMEWORK_SECTIONS.has(item.section)) {
      throw new Error(`Business navigation cannot modify the framework section ${item.section}`);
    }
    if (item.sectionOrder !== undefined && !Number.isFinite(item.sectionOrder)) {
      throw new Error(`Invalid sectionOrder for ${item.section}`);
    }
    const config = { order: item.sectionOrder ?? DEFAULT_BUSINESS_SECTION_ORDER, collapsible: item.sectionCollapsible === true };
    const existing = sections.get(item.section);
    if (existing && (existing.order !== config.order || existing.collapsible !== config.collapsible)) {
      throw new Error(`Conflicting admin section configuration: ${item.section}`);
    }
    sections.set(item.section, config);
  }
}

export function groupMenuItems(items) {
  const groups = [];
  const bySection = new Map();
  for (const [index, item] of items.entries()) {
    const key = item.section || "__root__";
    let group = bySection.get(key);
    if (!group) {
      group = {
        key,
        label: item.section || "",
        collapsible: item.sectionCollapsible === true,
        order: item.section ? item.sectionOrder ?? DEFAULT_BUSINESS_SECTION_ORDER : Number.NEGATIVE_INFINITY,
        firstIndex: index,
        items: [],
      };
      bySection.set(key, group);
      groups.push(group);
    }
    group.items.push(item);
  }
  return groups.sort((left, right) => left.order - right.order || left.firstIndex - right.firstIndex);
}

function findPathInItems(items, pathname, ancestors = []) {
  for (const item of items || []) {
    const trail = item.label ? [...ancestors, item.label] : ancestors;
    if (item.path === pathname) return trail;
    const childTrail = findPathInItems(item.children, pathname, trail);
    if (childTrail) return childTrail;
  }
  return null;
}

export function findMenuTrail(groups, pathname) {
  for (const group of groups || []) {
    const trail = findPathInItems(group.items, pathname);
    if (trail) return trail;
  }
  return [];
}

function findExpansionKeys(items, pathname, parentKey) {
  for (const [index, item] of (items || []).entries()) {
    const nodeKey = `${parentKey}:${index}`;
    if (item.path === pathname) return [];
    const childKeys = findExpansionKeys(item.children, pathname, nodeKey);
    if (childKeys) return item.children?.length ? [nodeKey, ...childKeys] : childKeys;
  }
  return null;
}

export function findMenuExpansionKeys(groups, pathname) {
  for (const group of groups || []) {
    const keys = findExpansionKeys(group.items, pathname, group.key);
    if (keys) return [group.key, ...keys];
  }
  return [];
}
