import { request } from "../auth/api.js";
export const adminApi = { manifest: () => request("/v1/internal/authorization-manifest"), syncManifest: () => request("/v1/internal/authorization-manifest/sync", { method: "POST" }) };
