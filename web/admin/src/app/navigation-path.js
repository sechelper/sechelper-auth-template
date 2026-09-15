export function normalizePathname(pathname = window.location.pathname) {
  return pathname.replace(/\/$/, "") || "/admin";
}

export function initialAdminPath({ pathname = window.location.pathname, storage = window.sessionStorage, history = window.history } = {}) {
  try {
    if (storage.getItem("sechelper:admin-entry") === "root") {
      storage.removeItem("sechelper:admin-entry");
      history.replaceState({}, "", "/admin/");
      return "/admin";
    }
  } catch {
    // Fall back to the address bar when browser storage is unavailable.
  }
  return normalizePathname(pathname);
}
