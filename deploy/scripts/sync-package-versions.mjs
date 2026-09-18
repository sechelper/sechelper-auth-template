import fs from "node:fs";
import { readProjectConfig } from "./read-project-config.mjs";

const release = readProjectConfig().release;
if (!release.version) throw new Error("project.yaml release.version is required");
const releaseVersion = release.version;
for (const root of ["web", "web/admin"]) {
  const packagePath = `${root}/package.json`;
  const lockPath = `${root}/package-lock.json`;
  const pkg = JSON.parse(fs.readFileSync(packagePath, "utf8"));
  const lock = JSON.parse(fs.readFileSync(lockPath, "utf8"));
  pkg.version = releaseVersion;
  lock.version = releaseVersion;
  if (lock.packages?.[""]) lock.packages[""].version = releaseVersion;
  fs.writeFileSync(packagePath, `${JSON.stringify(pkg, null, 2)}\n`);
  fs.writeFileSync(lockPath, `${JSON.stringify(lock, null, 2)}\n`);
}
console.log(`synchronized generated package metadata from project.yaml: version=${releaseVersion}`);
