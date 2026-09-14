import { request } from "../auth/api.js";
export const adminApi = { orders: () => request("/v1/orders"), manifest: () => request("/v1/internal/authorization-manifest"), syncManifest: () => request("/v1/internal/authorization-manifest/sync", { method: "POST" }) };
