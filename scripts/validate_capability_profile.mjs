#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const profilePath = path.resolve(
  ROOT,
  process.argv[2] ??
    "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const baseDir = path.dirname(profilePath);
const contractRoot = path.dirname(baseDir);
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

const profile = readJSON(profilePath);
const baseline = readJSON(path.join(contractRoot, "baseline.yaml"));
const oracle = readJSON(path.join(contractRoot, "oracle-profile.yaml"));
const schema = readJSON(
  path.join(contractRoot, "schema", "profile.schema.json"),
);

if (profile && baseline && oracle && schema) {
  assert(profile.schema_version === 1, "profile: schema_version must be 1");
  assert(
    profile.id === "core-workflow-v1",
    "profile: id must be core-workflow-v1",
  );
  assert(
    profile.compatibility_baseline === baseline.id,
    "profile: baseline reference mismatch",
  );
  assert(
    baseline.git_sha === oracle.asserted_git_sha,
    "oracle-profile: asserted SHA must equal baseline SHA",
  );
  assert(
    oracle.refuse_on_sha_mismatch === true,
    "oracle-profile: SHA mismatch must fail closed",
  );
  assert(
    oracle.refuse_on_configuration_mismatch === true,
    "oracle-profile: configuration mismatch must fail closed",
  );
  assert(
    JSON.stringify(profile.operations) ===
      JSON.stringify(["list", "get", "create", "replace", "update", "delete"]),
    "profile: operations must be the agreed six single-object operations",
  );
  assert(
    profile.bulk_operations === false,
    "profile: bulk operations must remain deferred",
  );
  assert(
    profile.interfaces?.rest?.compatibility === "exact-in-profile" &&
      profile.interfaces?.grpc?.compatibility === "semantic-parity" &&
      profile.interfaces?.vue?.compatibility === "workflow-parity",
    "profile: REST, gRPC, and Vue interface commitments are incomplete",
  );
  assert(
    Array.isArray(profile.resources),
    "profile: resources must be an array",
  );
  assert(
    profile.resources?.length === 13,
    `profile: expected 13 resources, found ${profile.resources?.length ?? 0}`,
  );

  const resourceKeys = new Set();
  const resourcePaths = new Set();
  for (const resource of profile.resources ?? []) {
    const key = `${resource.module}.${resource.name}`;
    assert(!resourceKeys.has(key), `profile: duplicate resource ${key}`);
    assert(
      !resourcePaths.has(resource.rest_path),
      `profile: duplicate path ${resource.rest_path}`,
    );
    resourceKeys.add(key);
    resourcePaths.add(resource.rest_path);
    assert(
      resource.tier === "T1",
      `profile:${key}: runtime scaffold must remain T1 until differential REST evidence exists`,
    );
    assert(
      profile.owners.includes(resource.owner),
      `profile:${key}: undeclared owner ${resource.owner}`,
    );
    assert(
      resource.rest_path.endsWith("/"),
      `profile:${key}: REST path needs trailing slash`,
    );
    assert(
      resource.grpc_service ===
        `netbox.${resource.module}.v1.${resource.module === "dcim" ? "DCIMService" : "IPAMService"}`,
      `profile:${key}: wrong bounded-module gRPC service`,
    );
  }

  const metadataResources = new Map();
  let choiceFieldCount = 0;
  for (const relativePath of profile.resource_metadata ?? []) {
    const metadata = readJSON(path.resolve(baseDir, relativePath));
    for (const resource of metadata?.resources ?? []) {
      const key = `${metadata.module}.${resource.name}`;
      assert(
        !metadataResources.has(key),
        `resource metadata: duplicate ${key}`,
      );
      metadataResources.set(key, resource);
      for (const field of [
        "writable_fields",
        "response_only_fields",
        "filters",
        "ordering",
      ]) {
        assert(
          Array.isArray(resource[field]) && resource[field].length > 0,
          `${key}: ${field} is empty`,
        );
      }
      for (const [field, choice] of Object.entries(
        resource.choice_fields ?? {},
      )) {
        choiceFieldCount += 1;
        assert(
          resource.writable_fields.includes(field) ||
            resource.response_only_fields.includes(field),
          `${key}: choice field ${field} is not part of the response contract`,
        );
        assert(
          choice?.value_type === "string" || choice?.value_type === "integer",
          `${key}: choice field ${field} has an invalid value_type`,
        );
        assert(
          typeof choice?.nullable === "boolean",
          `${key}: choice field ${field} must declare nullability`,
        );
      }
    }
  }
  assert(
    metadataResources.size === 13,
    `resource metadata: expected 13 resources, found ${metadataResources.size}`,
  );
  assert(
    choiceFieldCount === 19,
    `resource metadata: expected 19 choice fields, found ${choiceFieldCount}`,
  );
  for (const resource of profile.resources ?? []) {
    const key = `${resource.module}.${resource.name}`;
    const metadata = metadataResources.get(key);
    assert(metadata, `resource metadata: missing ${key}`);
    assert(
      metadata?.rest_path === resource.rest_path,
      `resource metadata: path mismatch for ${key}`,
    );
    assert(
      metadata?.grpc_prefix === resource.grpc_prefix,
      `resource metadata: gRPC mapping mismatch for ${key}`,
    );
  }

  const identityPath = path.resolve(
    baseDir,
    profile.identity_extension.resource_metadata,
  );
  const identity = readJSON(identityPath);
  assert(
    profile.identity_extension.classification === "extension",
    "identity: must be an extension",
  );
  assert(
    profile.identity_extension.tier === "not_applicable",
    "identity: extension cannot use a T tier",
  );
  assert(
    identity?.tier === "not_applicable",
    "identity metadata: extension cannot use a T tier",
  );
  for (const status of ["contract", "parity", "security"]) {
    assert(
      profile.identity_extension.verification?.[status],
      `identity: missing ${status} verification`,
    );
    assert(
      identity?.verification?.[status],
      `identity metadata: missing ${status} verification`,
    );
  }
  assert(
    identity?.rest_operations?.length === 8,
    "identity: expected eight declared REST operations",
  );
  assert(
    identity?.grpc_operations?.length === 5,
    "identity: expected five declared gRPC operations",
  );

  const scenarioIDs = new Set();
  const scenarioDir = path.join(contractRoot, "scenarios");
  for (const file of fs
    .readdirSync(scenarioDir)
    .filter((name) => name.endsWith(".yaml"))
    .sort()) {
    const scenarioFile = readJSON(path.join(scenarioDir, file));
    for (const scenario of scenarioFile?.scenarios ?? []) {
      assert(
        !scenarioIDs.has(scenario.id),
        `scenarios: duplicate id ${scenario.id}`,
      );
      scenarioIDs.add(scenario.id);
      assert(
        scenario.given?.length > 0,
        `scenario ${scenario.id}: missing given`,
      );
      assert(
        scenario.when?.length > 0,
        `scenario ${scenario.id}: missing when`,
      );
      assert(
        scenario.then?.length > 0,
        `scenario ${scenario.id}: missing then`,
      );
    }
  }
  for (const scenario of profile.scenarios ?? []) {
    assert(scenarioIDs.has(scenario), `profile: unknown scenario ${scenario}`);
  }
  for (const scenario of scenarioIDs) {
    assert(
      profile.scenarios.includes(scenario),
      `profile: unreferenced scenario ${scenario}`,
    );
  }
  assert(
    schema.properties?.resources,
    "profile schema: resources definition is missing",
  );
  assert(
    profile.deferred.some((item) => item === "GraphQL"),
    "profile: GraphQL must be explicitly deferred",
  );
}

if (failures.length > 0) {
  console.error("Capability profile validation failed:");
  for (const failure of failures) console.error(`  ${failure}`);
  process.exitCode = 1;
} else {
  const resourceCount = profile.resources?.length ?? 0;
  const interfaceCount = Object.keys(profile.interfaces ?? {}).length;
  const scenarioCount = profile.scenarios?.length ?? 0;
  console.log(
    `Capability profile valid: ${resourceCount} resources, ${interfaceCount} interfaces, ${scenarioCount} scenarios`,
  );
}
