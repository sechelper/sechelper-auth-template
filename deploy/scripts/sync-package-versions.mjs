import fs from "node:fs";

const release = Object.fromEntries([...fs.readFileSync("release.yaml", "utf8").matchAll(/^([A-Za-z][A-Za-z0-9]*):\s*(\S+)\s*$/gm)].map(([, key, value]) => [key, value]));
for (const root of ["web", "web/admin"]) {
  const packagePath = `${root}/package.json`;
  const lockPath = `${root}/package-lock.json`;
  const pkg = JSON.parse(fs.readFileSync(packagePath, "utf8"));
  const lock = JSON.parse(fs.readFileSync(lockPath, "utf8"));
  pkg.version = release.releaseVersion;
  lock.version = release.releaseVersion;
  if (lock.packages?.[""]) lock.packages[""].version = release.releaseVersion;
  fs.writeFileSync(packagePath, `${JSON.stringify(pkg, null, 2)}\n`);
  fs.writeFileSync(lockPath, `${JSON.stringify(lock, null, 2)}\n`);
}
console.log(`synchronized frontend package versions to ${release.releaseVersion}`);
