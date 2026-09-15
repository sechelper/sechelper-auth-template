import fs from "node:fs";
import path from "node:path";

const [output, environment, componentSelection, version, buildId, sourceRevision] = process.argv.slice(2);
const components = componentSelection === "all"
  ? ["api", "migrations", "web", "admin"]
  : componentSelection === "api" ? ["api", "migrations"] : ["web", "admin"];
const manifest = {
  applicationVersion: version,
  buildId,
  sourceRevision,
  environment,
  components: components.map((component) => ({
    component,
    artifact: `sechelper-auth-template-${component === "admin" ? "web" : component === "migrations" ? "api" : component}:${version}-${buildId}-${environment}`,
    applicationVersion: version,
    buildId,
    sourceRevision,
    ...(component === "admin" ? { artifactPath: "/admin/" } : {}),
    ...(component === "migrations" ? { artifactPath: "/app/auth-template-migrate and /app/migrations/" } : {}),
  })),
};
if (!output || !["development", "test", "production"].includes(environment) || !components.length || !version || !buildId || !sourceRevision) {
  throw new Error("artifact manifest requires an output path and complete canonical build metadata");
}
fs.mkdirSync(path.dirname(output), { recursive: true });
fs.writeFileSync(output, `${JSON.stringify(manifest, null, 2)}\n`);
console.log(`artifact manifest: ${output}`);
