import { request } from "../auth/api.js";
export const dashboardApi = { overview: () => request("/v1/admin/dashboard/overview") };
