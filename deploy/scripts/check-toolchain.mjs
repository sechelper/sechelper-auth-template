import fs from "node:fs";
import { execFileSync } from "node:child_process";
import { readProjectConfig } from "./read-project-config.mjs";

const fail = (message) => {
  console.error(message);
  process.exitCode = 1;
};
const versions = readProjectConfig().toolchain;
const actual = (command, args) => {
  try {
    const version = execFileSync(command, args, { encoding: "utf8" }).trim();
    const location = execFileSync("which", [command], { encoding: "utf8" }).trim();
    return { version, location };
  } catch {
    return { version: "missing", location: "missing" };
  }
};
const requireExact = (name, command, args, extract = (value) => value) => {
  const found = actual(command, args);
  const value = extract(found.version);
  if (value !== versions[name]) fail(`${name} mismatch: expected ${versions[name]}, actual ${found.version} at ${found.location}`);
};

requireExact("go", "go", ["version"], (value) => value.split(" ")[2]?.replace(/^go/, "") ?? value);
requireExact("node", "node", ["--version"], (value) => value.replace(/^v/, ""));
requireExact("npm", "npm", ["--version"]);
const goMod = fs.readFileSync("api/go.mod", "utf8");
if (!goMod.includes(`go ${versions.go}\n`)) fail("api/go.mod minimum Go version must match the verified project baseline");
const goToolchain = goMod.match(/^toolchain\s+(\S+)$/m)?.[1];
if (goToolchain && goToolchain !== `go${versions.go}`) fail(`api/go.mod toolchain mismatch: expected go${versions.go}, actual ${goToolchain}`);

for (const root of ["web", "web/admin"]) {
  const pkg = JSON.parse(fs.readFileSync(`${root}/package.json`, "utf8"));
  const lock = JSON.parse(fs.readFileSync(`${root}/package-lock.json`, "utf8"));
  for (const [dependency, name] of [["react", "react"], ["react-dom", "react"], ["vite", "vite"]]) {
    if (pkg.dependencies?.[dependency] !== versions[name] || lock.packages?.[""]?.dependencies?.[dependency] !== versions[name]) {
      fail(`${root} ${dependency} must match project.yaml toolchain.${name}=${versions[name]}`);
    }
  }
  if (pkg.packageManager !== `npm@${versions.npm}` || pkg.engines?.node !== `=${versions.node}` || pkg.engines?.npm !== `=${versions.npm}`) {
    fail(`${root} must declare the pinned Node.js and npm versions`);
  }
}
if (!fs.readFileSync("deploy/Dockerfile.api", "utf8").includes(`FROM golang:${versions.go} AS build`)) fail("API Docker toolchain does not match project.yaml");
if (!fs.readFileSync("deploy/Dockerfile.web", "utf8").includes(`FROM node:${versions.node}-alpine AS build`)) fail("Web Docker toolchain does not match project.yaml");
if (!fs.readFileSync("deploy/Dockerfile.web", "utf8").includes(`ARG NPM_VERSION=${versions.npm}`)) fail("Web Docker npm default must match project.yaml");
if (!fs.readFileSync("deploy/Dockerfile.web", "utf8").includes(`test \"$(node --version)\" = \"v${versions.node}\"`)) fail("Web Docker build must verify Node.js before package-manager installation");
if (!fs.readFileSync("deploy/Dockerfile.web", "utf8").includes('npm install --global "npm@$NPM_VERSION"')) fail("Web Docker build must install the pinned npm version before dependency installation");
if (!fs.readFileSync("deploy/Dockerfile.web", "utf8").includes("test \"$(npm --version)\" = \"$NPM_VERSION\"")) fail("Web Docker build must verify its effective npm version");
if (!fs.readFileSync(".github/workflows/ci.yml", "utf8").includes(`npm install --global npm@${versions.npm}`)) fail("CI must install the pinned npm version before dependency installation");
if (!fs.readFileSync(".github/workflows/ci.yml", "utf8").includes(`test "$(node --version)" = "v${versions.node}"`)) fail("CI must verify Node.js before changing the package-manager installation");
if (!fs.readFileSync(".github/workflows/ci.yml", "utf8").includes(`go-version: '${versions.go}'`)) fail("CI Go version must match project.yaml");
for (const workflow of [".github/workflows/ci.yml", ".github/workflows/e2e.yml"]) {
  if (fs.existsSync(workflow) && !fs.readFileSync(workflow, "utf8").includes(`node-version: ${versions.node}`)) fail(`${workflow} must select Node.js ${versions.node} from project.yaml`);
}
if (process.exitCode) process.exit(process.exitCode);
console.log(`toolchain versions and paths verified: Go ${versions.go}, Node.js ${versions.node}, npm ${versions.npm}`);
