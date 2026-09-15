import fs from "node:fs";

const fail = (message) => {
  console.error(message);
  process.exitCode = 1;
};

const releaseText = fs.readFileSync("release.yaml", "utf8");
const entries = [...releaseText.matchAll(/^([A-Za-z][A-Za-z0-9]*):\s*(\S+)\s*$/gm)];
const release = Object.fromEntries(entries.map(([, key, value]) => [key, value]));
if (entries.length !== 2 || Object.keys(release).sort().join(",") !== "releaseBuildId,releaseVersion") {
  fail("release.yaml must contain exactly releaseVersion and releaseBuildId");
}
if (!/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$/.test(release.releaseVersion ?? "")) {
  fail("release.yaml releaseVersion must be a valid SemVer value");
}
const buildId = release.releaseBuildId ?? "";
const buildDate = /^\d{14}$/.test(buildId)
  ? new Date(`${buildId.slice(0, 4)}-${buildId.slice(4, 6)}-${buildId.slice(6, 8)}T${buildId.slice(8, 10)}:${buildId.slice(10, 12)}:${buildId.slice(12, 14)}Z`)
  : null;
if (!buildDate || !Number.isFinite(buildDate.valueOf()) || buildDate.toISOString().replace(/[-:TZ]/g, "").slice(0, 14) !== buildId) {
  fail("release.yaml releaseBuildId must be a valid UTC YYYYMMDDHHmmss timestamp");
}

for (const root of ["web", "web/admin"]) {
  const manifest = JSON.parse(fs.readFileSync(`${root}/package.json`, "utf8"));
  const lock = JSON.parse(fs.readFileSync(`${root}/package-lock.json`, "utf8"));
  if (manifest.version !== release.releaseVersion || lock.version !== release.releaseVersion || lock.packages?.[""]?.version !== release.releaseVersion) {
    fail(`${root} package.json and package-lock.json versions must be synchronized from release.yaml (${release.releaseVersion})`);
  }
}

for (const configPath of ["config.example.yaml", "deploy/config.dev.yaml"]) {
  const config = fs.readFileSync(configPath, "utf8");
  let inApp = false;
  for (const line of config.split(/\r?\n/)) {
    if (/^app:\s*$/.test(line)) inApp = true;
    else if (line.trim() && !/^\s/.test(line)) inApp = false;
    else if (inApp && /^\s+version\s*:/.test(line)) fail(`${configPath} must not configure app.version; release.yaml is authoritative`);
  }
}
const configSource = fs.readFileSync("api/internal/platform/config/config.go", "utf8");
if (configSource.includes('"app.version"') || configSource.includes('"APP_VERSION"')) fail("application release version must not be configurable through Viper or APP_VERSION");
if (/type AppConfig struct[^\n]*\bVersion\b/.test(configSource)) fail("release version must not be part of the runtime AppConfig schema");
const makefile = fs.readFileSync("Makefile", "utf8");
for (const buildArgument of ["RELEASE_VERSION", "BUILD_ID", "SOURCE_REVISION"]) {
  if (!makefile.includes(`--build-arg ${buildArgument}=`)) fail(`canonical Make build must inject ${buildArgument} into every application component`);
}
if (!makefile.includes("write-artifact-manifest.mjs")) fail("canonical Make build must emit an artifact manifest");
if (makefile.includes("BUILD_ID=\"$GITHUB_SHA\"")) fail("Build ID must not be derived from the CI source revision");
const ciWorkflow = fs.readFileSync(".github/workflows/ci.yml", "utf8");
if (!ciWorkflow.includes("make build ENV=test") || !ciWorkflow.includes("make build ENV=production")) fail("CI must build test and production artifacts through the canonical Make entrypoint");
const apiDockerfile = fs.readFileSync("deploy/Dockerfile.api", "utf8");
const webDockerfile = fs.readFileSync("deploy/Dockerfile.web", "utf8");
if (!apiDockerfile.includes("-X main.releaseVersion=$RELEASE_VERSION") || !apiDockerfile.includes("-X main.buildID=$BUILD_ID") || !apiDockerfile.includes("-X main.buildRevision=$SOURCE_REVISION")) fail("API binary must embed canonical version, Build ID, and source revision");
if (!webDockerfile.includes("build-info.json") || !webDockerfile.includes("org.opencontainers.image.version")) fail("Web artifacts must include canonical release metadata");

if (process.exitCode) process.exit(process.exitCode);
console.log(`release metadata valid: version=${release.releaseVersion}, buildId=${release.releaseBuildId}`);
