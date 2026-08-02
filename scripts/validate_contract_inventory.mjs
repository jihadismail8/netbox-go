#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const inventoryDir = path.resolve(
  ROOT,
  process.argv[2] ?? "contracts/netbox/v4.4.6-post7/inventory",
);
const profilePath = path.join(
  ROOT,
  "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const failures = [];

function readJSON(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, "utf8"));
  } catch (error) {
    failures.push(`${path.relative(ROOT, filePath)}: ${error.message}`);
    return null;
  }
}

function assert(condition, message) {
  if (!condition) failures.push(message);
}

function validateDocument(name, expectedCount) {
  const filePath = path.join(inventoryDir, name);
  const value = readJSON(filePath);
  if (!value) return null;
  assert(value.schema_version === 1, `${name}: schema_version must be 1`);
  assert(
    value.compatibility_baseline === "v4.4.6-post7",
    `${name}: wrong baseline`,
  );
  assert(Array.isArray(value.entries), `${name}: entries must be an array`);
  if (!Array.isArray(value.entries)) return null;
  assert(
    value.entries.length === expectedCount,
    `${name}: expected ${expectedCount} entries, found ${value.entries.length}`,
  );

  const ids = new Set();
  for (const entry of value.entries) {
    assert(
      typeof entry.id === "string" && entry.id !== "",
      `${name}: entry is missing id`,
    );
    assert(!ids.has(entry.id), `${name}: duplicate id ${entry.id}`);
    ids.add(entry.id);
    assert(
      ["in_profile", "extension", "deferred", "out_of_scope"].includes(
        entry.classification,
      ),
      `${name}:${entry.id}: invalid classification ${entry.classification}`,
    );
    assert(
      typeof entry.owner === "string" && entry.owner !== "",
      `${name}:${entry.id}: missing owner`,
    );
    if (entry.classification === "extension") {
      assert(
        entry.tier === "not_applicable",
        `${name}:${entry.id}: extension cannot use a T tier`,
      );
      for (const key of ["contract", "parity", "security"]) {
        assert(
          entry.verification?.[key],
          `${name}:${entry.id}: extension lacks ${key} status`,
        );
      }
    } else {
      assert(
        /^T[0-4]$/.test(entry.tier),
        `${name}:${entry.id}: baseline capability needs T0-T4`,
      );
    }
  }
  return value;
}

const baseline = validateDocument("baseline-rest.yaml", 155);
const currentREST = validateDocument("current-rest.yaml", 123);
const currentGRPC = validateDocument("current-grpc.yaml", 179);
const currentVue = validateDocument("current-vue.yaml", 13);
const profile = readJSON(profilePath);

if (baseline && currentREST && currentGRPC && currentVue && profile) {
  const profilePaths = new Set(
    profile.resources.map((resource) => resource.rest_path),
  );
  assert(
    profilePaths.size === 13,
    `profile: expected 13 unique resource paths, found ${profilePaths.size}`,
  );

  for (const [name, document, pathField, kind] of [
    ["baseline-rest.yaml", baseline, "path", "resource"],
    ["current-rest.yaml", currentREST, "path", "canonical-resource"],
    ["current-vue.yaml", currentVue, "api_path", "resource-page"],
  ]) {
    const actualProfilePaths = new Set(
      document.entries
        .filter(
          (entry) =>
            entry.classification === "in_profile" && entry.kind === kind,
        )
        .map((entry) => entry[pathField]),
    );
    for (const requiredPath of profilePaths) {
      assert(
        actualProfilePaths.has(requiredPath),
        `${name}: missing in-profile ${requiredPath}`,
      );
    }
    for (const actualPath of actualProfilePaths) {
      assert(
        profilePaths.has(actualPath),
        `${name}: undeclared in-profile path ${actualPath}`,
      );
    }
  }

  const legacyREST = currentREST.entries.filter(
    (entry) => entry.kind === "legacy-resource",
  );
  const canonicalREST = currentREST.entries.filter(
    (entry) => entry.kind === "canonical-resource",
  );
  const identityREST = currentREST.entries.filter(
    (entry) => entry.kind === "canonical-operation",
  );
  assert(
    legacyREST.length === 102,
    "current-rest.yaml: expected 102 frozen legacy resources",
  );
  assert(
    legacyREST.every(
      (entry) =>
        entry.runtime_enabled === false &&
        entry.lifecycle === "frozen-unpublished",
    ),
    "current-rest.yaml: direct-GORM legacy resources must be frozen and disabled",
  );
  assert(
    canonicalREST.length === 13,
    "current-rest.yaml: expected 13 canonical resources",
  );
  assert(
    canonicalREST.every(
      (entry) =>
        entry.runtime_enabled === true &&
        entry.lifecycle === "pre-publication" &&
        entry.tier === "T1",
    ),
    "current-rest.yaml: canonical resources must be runtime-enabled pre-publication T1 scaffolds",
  );
  assert(
    identityREST.length === 8,
    "current-rest.yaml: expected eight identity extension operations",
  );
  assert(
    identityREST.every(
      (entry) =>
        entry.runtime_enabled === true && entry.classification === "extension",
    ),
    "current-rest.yaml: identity extension operations must be runtime enabled and explicit",
  );
  const legacyGRPC = currentGRPC.entries.filter(
    (entry) => entry.kind === "legacy-service",
  );
  const canonicalGRPC = currentGRPC.entries.filter(
    (entry) => entry.kind === "canonical-service",
  );
  assert(
    legacyGRPC.length === 176,
    "current-grpc.yaml: expected 176 frozen legacy services",
  );
  assert(
    legacyGRPC.every(
      (entry) =>
        entry.runtime_enabled === false &&
        entry.lifecycle === "frozen-unpublished",
    ),
    "current-grpc.yaml: legacy services must be frozen and disabled",
  );
  assert(
    canonicalGRPC.length === 3,
    "current-grpc.yaml: expected three canonical services",
  );
  assert(
    canonicalGRPC.every(
      (entry) =>
        entry.runtime_enabled === true &&
        entry.lifecycle === "pre-publication" &&
        (entry.classification === "extension" || entry.tier === "T1"),
    ),
    "current-grpc.yaml: canonical services must be runtime enabled and pre-publication without a T2 claim",
  );
  assert(
    currentVue.entries.every(
      (entry) =>
        entry.runtime_enabled === true &&
        entry.backend_capability_enabled === true &&
        entry.classification === "in_profile" &&
        entry.tier === "T1",
    ),
    "current-vue.yaml: only the 13 runtime profile resources may be published",
  );
  assert(
    baseline.entries.some(
      (entry) =>
        entry.path === "/api/users/tokens/provision/" &&
        entry.classification === "out_of_scope",
    ),
    "baseline-rest.yaml: anonymous token provision must be explicitly out_of_scope",
  );
  assert(
    baseline.entries
      .filter((entry) => entry.kind === "action")
      .every((entry) => entry.classification !== "in_profile"),
    "baseline-rest.yaml: first profile must not silently include custom actions",
  );
}

if (failures.length > 0) {
  console.error("Contract inventory validation failed:");
  for (const failure of failures) console.error(`  ${failure}`);
  process.exitCode = 1;
} else {
  console.log(
    "Contract inventory valid: 155 baseline, 123 current REST, 179 current gRPC, 13 current Vue entries",
  );
}
