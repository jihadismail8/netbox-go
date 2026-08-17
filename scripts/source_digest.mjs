#!/usr/bin/env node

import crypto from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const DEFAULT_REPOSITORY_ROOT = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);

// The NetBox oracle submodule is identified independently by its pinned Git
// commit. The digest below covers every owned runtime, contract, harness, and
// documentation input while deliberately excluding build products, dependency
// caches, credentials, local pre-Git scratch fragments, and the evidence
// directory which cites the digest.
const ownedRoots = [
  ".dockerignore",
  ".editorconfig",
  ".gitattributes",
  ".github",
  ".gitignore",
  ".gitmodules",
  ".go-version",
  ".node-version",
  ".nvmrc",
  "AGENTS.md",
  "CONTEXT.md",
  "CONTRIBUTING.md",
  "Dockerfile",
  "LICENSE",
  "Makefile",
  "README.md",
  "REWRITE_PLAN.md",
  "THIRD_PARTY_NOTICES.md",
  "contracts",
  "docker-compose.yml",
  "docs",
  "netbox-backend",
  "netbox-frontend",
  "scripts",
  "tests",
];

const excludedPrefixes = [
  "docs/evidence/",
  "netbox-frontend/coverage/",
  "netbox-frontend/dist/",
  "netbox-frontend/node_modules/",
  "tests/deployment/.artifacts/",
];
const excludedExactPaths = new Set([
  "docs/business-",
  "docs/entities/dc",
  "netbox-backend/cover.out",
  "netbox-backend/netbox_go",
  "netbox-frontend/.vscode/settings.json",
]);

function normalize(relativePath) {
  return relativePath.split(path.sep).join("/");
}

function assertUTF8RoundTrip(value, label) {
  const bytes = Buffer.from(value, "utf8");
  const decoded = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  if (decoded !== value) throw new Error(`${label} is not canonical UTF-8`);
}

function assertSafeDigestToken(value, label) {
  if (value.includes("\0") || value.includes("\r") || value.includes("\n")) {
    throw new Error(`${label} contains a source-digest delimiter`);
  }
}

function excluded(relativePath) {
  const normalized = normalize(relativePath);
  assertUTF8RoundTrip(normalized, `owned source path ${normalized}`);
  const base = path.posix.basename(normalized);
  if (base === ".env" || base.startsWith(".env.")) return true;
  return (
    excludedExactPaths.has(normalized) ||
    excludedPrefixes.some(
      (prefix) =>
        normalized === prefix.replace(/\/$/, "") ||
        normalized.startsWith(prefix),
    )
  );
}

function owned(relativePath) {
  const normalized = normalize(relativePath);
  return (
    ownedRoots.some(
      (ownedRoot) =>
        normalized === ownedRoot || normalized.startsWith(`${ownedRoot}/`),
    ) && !excluded(normalized)
  );
}

function collect(repositoryRoot, relativePath, entries) {
  if (excluded(relativePath)) return;

  const absolutePath = path.join(repositoryRoot, relativePath);
  let stat;
  try {
    stat = fs.lstatSync(absolutePath);
  } catch (error) {
    throw new Error(
      `cannot inspect owned source ${relativePath}: ${error.message}`,
    );
  }

  if (stat.isDirectory()) {
    const childNames = fs
      .readdirSync(absolutePath, { encoding: "buffer" })
      .sort(Buffer.compare);
    for (const childBytes of childNames) {
      const child = new TextDecoder("utf-8", { fatal: true }).decode(
        childBytes,
      );
      if (!Buffer.from(child, "utf8").equals(childBytes)) {
        throw new Error(
          `owned source directory entry is not canonical UTF-8: ${relativePath}`,
        );
      }
      collect(repositoryRoot, path.join(relativePath, child), entries);
    }
    return;
  }

  const normalized = normalize(relativePath);
  assertSafeDigestToken(normalized, `owned source path ${normalized}`);
  if (stat.isSymbolicLink()) {
    const targetBytes = fs.readlinkSync(absolutePath, { encoding: "buffer" });
    const target = new TextDecoder("utf-8", { fatal: true }).decode(
      targetBytes,
    );
    if (!Buffer.from(target, "utf8").equals(targetBytes)) {
      throw new Error(
        `owned source link target is not canonical UTF-8: ${normalized}`,
      );
    }
    assertSafeDigestToken(target, `owned source link ${normalized}`);
    entries.push({ kind: "symlink", path: normalized, target });
    return;
  }
  if (!stat.isFile()) {
    throw new Error(`owned source is neither a file nor a link: ${normalized}`);
  }

  const bytes = fs.readFileSync(absolutePath);
  const fileDigest = crypto.createHash("sha256").update(bytes).digest("hex");
  entries.push({
    kind: "file",
    mode: stat.mode & 0o111 ? "100755" : "100644",
    path: normalized,
    size: bytes.length,
    sha256: `sha256:${fileDigest}`,
  });
}

function validateSourceManifestEntry(entry) {
  if (
    entry === null ||
    typeof entry !== "object" ||
    Array.isArray(entry) ||
    typeof entry.path !== "string" ||
    entry.path.length === 0 ||
    entry.path.includes("\\") ||
    path.posix.isAbsolute(entry.path) ||
    path.posix.normalize(entry.path) !== entry.path ||
    entry.path === "." ||
    entry.path.split("/").includes("..")
  ) {
    throw new Error("source-manifest entry has an invalid canonical path");
  }
  assertSafeDigestToken(entry.path, `owned source path ${entry.path}`);
  if (entry.kind === "file") {
    if (
      Object.keys(entry).sort().join("\0") !==
        ["kind", "mode", "path", "sha256", "size"].join("\0") ||
      !["100644", "100755"].includes(entry.mode) ||
      !Number.isSafeInteger(entry.size) ||
      entry.size < 0 ||
      !/^sha256:[0-9a-f]{64}$/u.test(entry.sha256 ?? "")
    ) {
      throw new Error(`source-manifest file entry is invalid: ${entry.path}`);
    }
    return;
  }
  if (
    entry.kind !== "symlink" ||
    Object.keys(entry).sort().join("\0") !==
      ["kind", "path", "target"].join("\0") ||
    typeof entry.target !== "string" ||
    entry.target.length === 0
  ) {
    throw new Error(`source-manifest link entry is invalid: ${entry.path}`);
  }
  assertSafeDigestToken(entry.target, `owned source link ${entry.path}`);
}

function validateSymlinkClosure(entries) {
  const byPath = new Map(entries.map((entry) => [entry.path, entry]));
  for (const entry of entries) {
    if (entry.kind !== "symlink") continue;
    if (path.posix.isAbsolute(entry.target)) {
      throw new Error(`owned source link is absolute: ${entry.path}`);
    }
    const targetPath = path.posix.normalize(
      path.posix.join(path.posix.dirname(entry.path), entry.target),
    );
    if (
      targetPath === "." ||
      targetPath.startsWith("../") ||
      targetPath.includes("/../") ||
      !owned(targetPath)
    ) {
      throw new Error(
        `owned source link leaves the owned source boundary: ${entry.path}`,
      );
    }
    const exactTarget = byPath.get(targetPath);
    if (exactTarget?.kind === "symlink") {
      throw new Error(`owned source link chains through a link: ${entry.path}`);
    }
    const hasOwnedTarget =
      exactTarget?.kind === "file" ||
      entries.some(
        (candidate) =>
          candidate.kind === "file" &&
          candidate.path.startsWith(`${targetPath}/`),
      );
    if (!hasOwnedTarget) {
      throw new Error(
        `owned source link target has no independently hashed content: ${entry.path}`,
      );
    }
  }
}

export function serializeSourceManifestEntry(entry) {
  validateSourceManifestEntry(entry);
  if (entry.kind === "file") {
    return `F\0${entry.mode}\0${entry.path}\0${entry.size}\0${entry.sha256.slice("sha256:".length)}\n`;
  }
  return `L\0${entry.path}\0${entry.target}\n`;
}

export function calculateSourceDigestFromEntries(entries) {
  if (!Array.isArray(entries)) {
    throw new Error("source-manifest entries must be an array");
  }
  const paths = entries.map((entry) => entry?.path);
  if (new Set(paths).size !== paths.length) {
    throw new Error("source-manifest entries contain duplicate paths");
  }
  const serialized = entries.map(serializeSourceManifestEntry).sort();
  return `source-v2:sha256:${crypto
    .createHash("sha256")
    .update(serialized.join(""), "utf8")
    .digest("hex")}`;
}

export function calculateSourceManifest(
  repositoryRoot = DEFAULT_REPOSITORY_ROOT,
) {
  const resolvedRoot = path.resolve(repositoryRoot);
  const entries = [];
  for (const ownedRoot of ownedRoots) {
    collect(resolvedRoot, ownedRoot, entries);
  }
  entries.sort((left, right) =>
    left.path < right.path ? -1 : left.path > right.path ? 1 : 0,
  );
  validateSymlinkClosure(entries);

  return {
    schema_version: 2,
    digest: calculateSourceDigestFromEntries(entries),
    files: entries.length,
    entries,
  };
}

function git(repositoryRoot, arguments_, options = {}) {
  const { env: optionEnvironment = {}, ...executionOptions } = options;
  return execFileSync(
    "git",
    [
      "--no-replace-objects",
      "-c",
      "advice.graftFileDeprecated=false",
      "-C",
      repositoryRoot,
      ...arguments_,
    ],
    {
      encoding: "utf8",
      env: {
        ...process.env,
        ...optionEnvironment,
        GIT_GRAFT_FILE: "/dev/null",
        GIT_NO_REPLACE_OBJECTS: "1",
      },
      maxBuffer: 256 * 1024 * 1024,
      stdio: ["ignore", "pipe", "pipe"],
      ...executionOptions,
    },
  );
}

function gitBlobBytes(repositoryRoot, objectIDs) {
  const uniqueObjectIDs = [...new Set(objectIDs)];
  if (uniqueObjectIDs.length === 0) return new Map();
  const result = spawnSync(
    "git",
    [
      "--no-replace-objects",
      "-c",
      "advice.graftFileDeprecated=false",
      "-C",
      repositoryRoot,
      "cat-file",
      "--batch",
    ],
    {
      input: `${uniqueObjectIDs.join("\n")}\n`,
      encoding: null,
      env: {
        ...process.env,
        GIT_GRAFT_FILE: "/dev/null",
        GIT_NO_REPLACE_OBJECTS: "1",
      },
      maxBuffer: 256 * 1024 * 1024,
      stdio: ["pipe", "pipe", "pipe"],
    },
  );
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(
      `git cat-file failed: ${String(result.stderr ?? "").trim()}`,
    );
  }

  const output = result.stdout;
  const blobs = new Map();
  let offset = 0;
  for (const expectedObjectID of uniqueObjectIDs) {
    const headerEnd = output.indexOf(0x0a, offset);
    if (headerEnd < 0) {
      throw new Error("git cat-file returned a truncated object header");
    }
    const header = output.subarray(offset, headerEnd).toString("ascii");
    const match = /^([0-9a-f]{40,64}) blob ([0-9]+)$/u.exec(header);
    if (!match || match[1] !== expectedObjectID) {
      throw new Error(`git cat-file returned an unexpected header: ${header}`);
    }
    const size = Number(match[2]);
    if (!Number.isSafeInteger(size) || size < 0) {
      throw new Error(`git cat-file returned an invalid blob size: ${header}`);
    }
    const bodyStart = headerEnd + 1;
    const bodyEnd = bodyStart + size;
    if (bodyEnd >= output.length || output[bodyEnd] !== 0x0a) {
      throw new Error(`git cat-file returned a truncated blob: ${match[1]}`);
    }
    blobs.set(match[1], output.subarray(bodyStart, bodyEnd));
    offset = bodyEnd + 1;
  }
  if (offset !== output.length) {
    throw new Error("git cat-file returned unrequested trailing data");
  }
  return blobs;
}

export function calculateSourceManifestAtGitRevision(repositoryRoot, revision) {
  const resolvedRoot = path.resolve(repositoryRoot);
  if (!/^[0-9a-f]{40}$/u.test(revision ?? "")) {
    throw new Error("tested Git revision must be a full lowercase SHA-1");
  }
  const resolvedRevision = git(resolvedRoot, [
    "rev-parse",
    "--verify",
    `${revision}^{commit}`,
  ]).trim();
  if (resolvedRevision !== revision) {
    throw new Error(
      `tested Git revision resolves to ${resolvedRevision}, expected ${revision}`,
    );
  }
  try {
    git(resolvedRoot, ["merge-base", "--is-ancestor", revision, "HEAD"]);
  } catch {
    throw new Error(
      `tested Git revision ${revision} is not reachable from the current HEAD`,
    );
  }

  const treeBytes = git(
    resolvedRoot,
    ["ls-tree", "-r", "-z", "--full-tree", revision],
    { encoding: null },
  );
  const tree = new TextDecoder("utf-8", { fatal: true }).decode(treeBytes);
  if (!Buffer.from(tree, "utf8").equals(treeBytes)) {
    throw new Error("git ls-tree returned a non-round-trippable UTF-8 path");
  }
  const treeEntries = [];
  for (const record of tree.split("\0")) {
    if (record.length === 0) continue;
    const separator = record.indexOf("\t");
    const header = separator < 0 ? "" : record.slice(0, separator);
    const relativePath = separator < 0 ? "" : record.slice(separator + 1);
    const match = /^([0-7]{6}) ([a-z]+) ([0-9a-f]{40,64})$/u.exec(header);
    if (!match || relativePath.length === 0) {
      throw new Error(`git ls-tree returned an invalid entry: ${record}`);
    }
    if (!owned(relativePath)) continue;
    assertUTF8RoundTrip(relativePath, `owned Git source path ${relativePath}`);
    if (
      match[2] !== "blob" ||
      !["100644", "100755", "120000"].includes(match[1])
    ) {
      throw new Error(
        `owned Git source has unsupported mode/type at ${relativePath}`,
      );
    }
    assertSafeDigestToken(
      relativePath,
      `owned Git source path ${relativePath}`,
    );
    treeEntries.push({
      mode: match[1],
      objectID: match[3],
      path: relativePath,
    });
  }
  treeEntries.sort((left, right) =>
    left.path < right.path ? -1 : left.path > right.path ? 1 : 0,
  );
  if (
    new Set(treeEntries.map((entry) => entry.path)).size !== treeEntries.length
  ) {
    throw new Error("owned Git source contains duplicate canonical paths");
  }

  const blobs = gitBlobBytes(
    resolvedRoot,
    treeEntries.map((entry) => entry.objectID),
  );
  const entries = treeEntries.map((entry) => {
    const bytes = blobs.get(entry.objectID);
    if (!bytes) {
      throw new Error(`missing Git blob for owned source ${entry.path}`);
    }
    if (entry.mode === "120000") {
      const target = bytes.toString("utf8");
      if (!Buffer.from(target, "utf8").equals(bytes)) {
        throw new Error(`owned Git symlink target is not UTF-8: ${entry.path}`);
      }
      assertSafeDigestToken(target, `owned Git source link ${entry.path}`);
      return { kind: "symlink", path: entry.path, target };
    }
    return {
      kind: "file",
      mode: entry.mode,
      path: entry.path,
      size: bytes.length,
      sha256: `sha256:${crypto.createHash("sha256").update(bytes).digest("hex")}`,
    };
  });
  validateSymlinkClosure(entries);
  return {
    schema_version: 2,
    digest: calculateSourceDigestFromEntries(entries),
    files: entries.length,
    entries,
  };
}

export function calculateSourceDigest(
  repositoryRoot = DEFAULT_REPOSITORY_ROOT,
) {
  const manifest = calculateSourceManifest(repositoryRoot);
  return { digest: manifest.digest, files: manifest.files };
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href
) {
  if (process.argv[2] === "--manifest") {
    process.stdout.write(
      `${JSON.stringify(calculateSourceManifest(), null, 2)}\n`,
    );
  } else {
    const result = calculateSourceDigest();
    process.stdout.write(`${result.digest}\nfiles:${result.files}\n`);
  }
}
