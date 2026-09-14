import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = path.dirname(fileURLToPath(import.meta.url));
export default defineConfig({ base: "/admin/", plugins: [react()], resolve: { alias: { "react/jsx-runtime": path.resolve(projectRoot, "node_modules/react/jsx-runtime.js"), react: path.resolve(projectRoot, "node_modules/react") } } });
