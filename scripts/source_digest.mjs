#!/usr/bin/env node

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
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

function excluded(relativePath) {
  const normalized = normalize(relativePath);
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

function collect(relativePath, entries) {
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
    for (const child of fs.readdirSync(absolutePath).sort()) {
      collect(path.join(relativePath, child), entries);
    }
    return;
  }

  const normalized = normalize(relativePath);
  if (stat.isSymbolicLink()) {
    entries.push(`L\0${normalized}\0${fs.readlinkSync(absolutePath)}\n`);
    return;
  }
  if (!stat.isFile()) {
    throw new Error(`owned source is neither a file nor a link: ${normalized}`);
  }

  const bytes = fs.readFileSync(absolutePath);
  const fileDigest = crypto.createHash("sha256").update(bytes).digest("hex");
  entries.push(`F\0${normalized}\0${bytes.length}\0${fileDigest}\n`);
}

const entries = [];
for (const ownedRoot of ownedRoots) collect(ownedRoot, entries);
entries.sort();

const digest = crypto
  .createHash("sha256")
  .update(entries.join(""), "utf8")
  .digest("hex");

process.stdout.write(`sha256:${digest}\nfiles:${entries.length}\n`);
