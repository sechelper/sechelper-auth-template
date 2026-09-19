import { request } from "../auth/api.js";
export const configurationApi = {
  list: () => request("/v1/admin/configuration"),
  get: (key) => request(`/v1/admin/configuration/${encodeURIComponent(key)}`),
  saveFramework: (key, value) => request(`/v1/admin/configuration/${encodeURIComponent(key)}`, { method: "PUT", body: JSON.stringify(value) }),
  restartFramework: () => request("/v1/admin/configuration/restart", { method: "POST" }),
  save: (key, value) => request(`/v1/admin/business-configuration/${encodeURIComponent(key)}`, { method: "PUT", body: JSON.stringify(value) }),
  remove: (key) => request(`/v1/admin/business-configuration/${encodeURIComponent(key)}`, { method: "DELETE" }),
};
