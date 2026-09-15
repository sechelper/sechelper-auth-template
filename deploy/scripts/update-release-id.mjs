import fs from "node:fs";
import os from "node:os";
import path from "node:path";

const releasePath = path.resolve(process.argv[2] ?? "release.yaml");
const lockPath = path.join(os.tmpdir(), "sechelper-auth-template-release-id.lock");
let locked = false;
for (let attempt = 0; attempt < 400; attempt += 1) {
  try {
    fs.mkdirSync(lockPath);
    fs.writeFileSync(path.join(lockPath, "owner"), String(process.pid));
    locked = true;
    break;
  } catch (error) {
    if (error.code !== "EEXIST") throw error;
    try {
      const owner = Number(fs.readFileSync(path.join(lockPath, "owner"), "utf8"));
      if (owner) {
        try { process.kill(owner, 0); }
        catch (ownerError) { if (ownerError.code === "ESRCH") fs.rmSync(lockPath, { recursive: true, force: true }); }
      }
    } catch (ownerError) {
      if (ownerError.code === "ESRCH" || ownerError.code === "ENOENT") fs.rmSync(lockPath, { recursive: true, force: true });
    }
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 25);
  }
}
if (!locked) throw new Error("could not acquire release Build ID allocation lock");

try {
  const original = fs.readFileSync(releasePath, "utf8");
  const matches = [...original.matchAll(/^releaseBuildId:\s*(\d{14})\s*$/gm)];
  if (matches.length !== 1) throw new Error("release.yaml must contain exactly one releaseBuildId before allocation");
  const previous = matches[0][1];
  const previousDate = new Date(`${previous.slice(0, 4)}-${previous.slice(4, 6)}-${previous.slice(6, 8)}T${previous.slice(8, 10)}:${previous.slice(10, 12)}:${previous.slice(12, 14)}Z`);
  const now = new Date();
  const next = new Date(Math.max(now.valueOf(), previousDate.valueOf() + 1000));
  const id = next.toISOString().replace(/[-:TZ]/g, "").slice(0, 14);
  const updated = original.replace(/^releaseBuildId:\s*\d{14}\s*$/m, `releaseBuildId: ${id}`);
  fs.writeFileSync(releasePath, updated);
  console.log(id);
} finally {
  fs.rmSync(lockPath, { recursive: true, force: true });
}
