import fs from "node:fs";

export const readProjectConfig = (filePath = "project.yaml") => {
  const result = {};
  let section;
  for (const [index, rawLine] of fs.readFileSync(filePath, "utf8").split(/\r?\n/).entries()) {
    const line = rawLine.replace(/\s+#.*$/, "");
    if (!line.trim()) continue;
    const match = line.match(/^( {0}| {2})([A-Za-z][A-Za-z0-9]*):\s*(\S.*)?$/);
    if (!match) throw new Error(`${filePath}:${index + 1} must use the supported project.yaml format`);
    const [, indent, key, value] = match;
    if (indent === "") {
      if (value) throw new Error(`${filePath}:${index + 1} top-level values must be sections`);
      if (Object.prototype.hasOwnProperty.call(result, key)) throw new Error(`${filePath}:${index + 1} duplicates section ${key}`);
      section = key;
      result[section] = {};
    } else if (!section || !value) {
      throw new Error(`${filePath}:${index + 1} must contain a scalar value under a section`);
    } else {
      if (Object.prototype.hasOwnProperty.call(result[section], key)) throw new Error(`${filePath}:${index + 1} duplicates ${section}.${key}`);
      result[section][key] = value.replace(/^['"]|['"]$/g, "");
    }
  }
  return result;
};

if (process.argv[1]?.endsWith("read-project-config.mjs")) {
  const [section, key] = process.argv.slice(2);
  const value = readProjectConfig()[section]?.[key];
  if (value === undefined) process.exitCode = 1;
  else console.log(value);
}
