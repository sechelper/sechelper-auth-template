import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = path.dirname(fileURLToPath(import.meta.url));

function adminModuleProfile(command, mode) {
  const buildProfile = mode.toLowerCase();
  return {
    name: "admin-module-build-profile",
    enforce: "pre",
    transform(source, id) {
      // Development and test builds include example modules; production and
      // custom modes only discover production business modules.
      if (!id.endsWith("/src/app/admin-business-modules.js") || (command === "build" && !["development", "test"].includes(buildProfile))) return null;
      const patterns = ["../business/*/admin-module.js", "../business/*/example-module.js"];
      if (buildProfile === "test") patterns.push("../business/*/test-module.js");
      const eagerGlob = 'import.meta.glob("../business/*/admin-module.js", { eager: true, import: "default" })';
      const profileGlob = `import.meta.glob(${JSON.stringify(patterns)}, { eager: true, import: "default" })`;
      if (!source.includes(eagerGlob)) throw new Error("Admin business module discovery contract changed");
      return source.replace(eagerGlob, profileGlob);
    },
  };
}

export default defineConfig(({ command, mode }) => ({ base: "/admin/", plugins: [adminModuleProfile(command, mode), react()], resolve: { alias: { "react/jsx-runtime": path.resolve(projectRoot, "node_modules/react/jsx-runtime.js"), react: path.resolve(projectRoot, "node_modules/react") } } }));
