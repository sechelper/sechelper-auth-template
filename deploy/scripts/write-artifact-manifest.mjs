import fs from "node:fs";
import path from "node:path";

const [output, environment, componentSelection, version, buildId, sourceRevision, artifactRootArgument] = process.argv.slice(2);
const artifactRoot = artifactRootArgument ? path.resolve(artifactRootArgument) : undefined;
const components = componentSelection === "all"
  ? ["api", "migrations", "web", "admin"]
  : componentSelection === "api" ? ["api", "migrations"] : ["web", "admin"];
if (environment === "test" && !artifactRoot) {
  throw new Error("test artifact manifest requires the native build output directory");
}
const artifactPaths = artifactRoot ? {
  api: path.join(artifactRoot, "api", "auth-template"),
  migrations: path.join(artifactRoot, "api"),
  web: path.join(artifactRoot, "web"),
  admin: path.join(artifactRoot, "admin"),
} : undefined;
const manifest = {
  applicationVersion: version,
  buildId,
  sourceRevision,
  environment,
  components: components.map((component) => ({
    component,
    artifact: artifactPaths ? artifactPaths[component] : `sechelper-auth-template-${component === "admin" ? "web" : component === "migrations" ? "api" : component}:${version}-${buildId}-${environment}`,
    applicationVersion: version,
    buildId,
    sourceRevision,
    ...(artifactPaths ? { artifactPath: artifactPaths[component] } : {}),
    ...(component === "admin" ? { route: "/admin/", ...(!artifactPaths ? { artifactPath: "/admin/" } : {}) } : {}),
    ...(component === "migrations" ? {
      ...(!artifactPaths ? { artifactPath: "/app/auth-template-migrate and /app/migrations/" } : {}),
      migrationBinary: artifactRoot ? path.join(artifactRoot, "api", "auth-template-migrate") : "/app/auth-template-migrate",
      migrationPath: artifactRoot ? path.join(artifactRoot, "api", "migrations") : "/app/migrations/",
    } : {}),
  })),
};
if (!output || !["development", "test", "production"].includes(environment) || !components.length || !version || !buildId || !sourceRevision) {
  throw new Error("artifact manifest requires an output path and complete canonical build metadata");
}
if (artifactRoot) {
  const requiredPaths = {
    api: [artifactPaths.api],
    migrations: [path.join(artifactRoot, "api", "auth-template-migrate"), path.join(artifactRoot, "api", "migrations")],
    web: [path.join(artifactRoot, "web", "index.html"), path.join(artifactRoot, "web", "build-info.json")],
    admin: [path.join(artifactRoot, "admin", "index.html"), path.join(artifactRoot, "admin", "build-info.json")],
  };
  for (const component of components) {
    for (const requiredPath of requiredPaths[component]) {
      if (!fs.existsSync(requiredPath)) throw new Error(`missing ${component} build artifact: ${requiredPath}`);
    }
  }
}
fs.mkdirSync(path.dirname(output), { recursive: true });
fs.writeFileSync(output, `${JSON.stringify(manifest, null, 2)}\n`);
console.log(`artifact manifest: ${output}`);
