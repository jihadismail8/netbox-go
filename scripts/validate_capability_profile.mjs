#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { validateTraceability } from "./validate_traceability.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const profilePath = path.resolve(
  ROOT,
  process.argv[2] ??
    "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const baseDir = path.dirname(profilePath);
const contractRoot = path.dirname(baseDir);
const failures = [];
let traceabilityCounts = null;
const IDENTITY_VERIFICATION_STATES = new Set(["partial", "complete"]);
const FIELD_CONTRACT_KEYS = ["create", "replace", "response", "update"];
const WRITE_FIELD_CONTRACT_KEYS = [
  "blank_fields",
  "nullable_fields",
  "required_fields",
];

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

function exactKeys(value, expected) {
  return (
    value !== null &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    JSON.stringify(Object.keys(value).sort()) === JSON.stringify(expected)
  );
}

function validateFieldList({
  resourceKey,
  contractName,
  listName,
  value,
  allowedFields,
  isStringField,
}) {
  const label = `${resourceKey}: field_contracts.${contractName}.${listName}`;
  if (!Array.isArray(value)) {
    assert(false, `${label} must be an array`);
    return;
  }
  const seen = new Set();
  for (const field of value) {
    if (typeof field !== "string" || field.length === 0) {
      assert(false, `${label} contains an invalid field name`);
      continue;
    }
    assert(!seen.has(field), `${label} contains duplicate field ${field}`);
    seen.add(field);
    assert(
      allowedFields.has(field),
      `${label} contains undeclared field ${field}`,
    );
    if (listName === "blank_fields" && allowedFields.has(field)) {
      assert(
        isStringField(field),
        `${label} contains non-string field ${field}`,
      );
    }
  }
}

function validateFieldContracts(resourceKey, resource) {
  const contracts = resource.field_contracts;
  assert(
    exactKeys(contracts, FIELD_CONTRACT_KEYS),
    `${resourceKey}: field_contracts must declare exactly response, create, replace, and update`,
  );
  if (!contracts || typeof contracts !== "object" || Array.isArray(contracts)) {
    return;
  }

  const writableFields = new Set(resource.writable_fields ?? []);
  const responseFields = new Set([
    ...(resource.writable_fields ?? []),
    ...(resource.response_only_fields ?? []),
  ]);
  const isStringField = (field) => {
    const choice = resource.choice_fields?.[field];
    if (choice) return choice.value_type === "string";
    if ((resource.relationships ?? []).includes(field)) return false;
    return !field.endsWith("_id");
  };

  const response = contracts.response;
  assert(
    exactKeys(response, ["nullable_fields"]),
    `${resourceKey}: field_contracts.response must declare exactly nullable_fields`,
  );
  if (response && typeof response === "object" && !Array.isArray(response)) {
    validateFieldList({
      resourceKey,
      contractName: "response",
      listName: "nullable_fields",
      value: response.nullable_fields,
      allowedFields: responseFields,
      isStringField,
    });
  }

  for (const operation of ["create", "replace", "update"]) {
    const contract = contracts[operation];
    assert(
      exactKeys(contract, WRITE_FIELD_CONTRACT_KEYS),
      `${resourceKey}: field_contracts.${operation} must declare exactly required_fields, nullable_fields, and blank_fields`,
    );
    if (!contract || typeof contract !== "object" || Array.isArray(contract)) {
      continue;
    }
    for (const listName of WRITE_FIELD_CONTRACT_KEYS) {
      validateFieldList({
        resourceKey,
        contractName: operation,
        listName,
        value: contract[listName],
        allowedFields: writableFields,
        isStringField,
      });
    }
  }
  assert(
    Array.isArray(contracts.update?.required_fields) &&
      contracts.update.required_fields.length === 0,
    `${resourceKey}: field_contracts.update.required_fields must be empty for PATCH`,
  );
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
      if (resource.field_contracts !== undefined) {
        validateFieldContracts(key, resource);
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
  assert(
    metadataResources.get("ipam.IPAddress")?.field_contracts !== undefined,
    "ipam.IPAddress: field_contracts must declare operation-specific presence semantics",
  );
  assert(
    metadataResources.get("dcim.Site")?.field_contracts !== undefined,
    "dcim.Site: field_contracts must declare operation-specific presence semantics",
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
    JSON.stringify(Object.keys(profile.identity_extension).sort()) ===
      JSON.stringify(
        [
          "classification",
          "owner",
          "resource_metadata",
          "tier",
          "verification",
        ].sort(),
      ),
    "identity: extension contains missing or unsupported properties",
  );
  assert(
    JSON.stringify(
      Object.keys(profile.identity_extension.verification ?? {}).sort(),
    ) === JSON.stringify(["contract", "parity", "security"]),
    "identity: verification contains missing or unsupported properties",
  );
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
      IDENTITY_VERIFICATION_STATES.has(
        profile.identity_extension.verification?.[status],
      ),
      `identity: ${status} verification must be partial or complete`,
    );
    assert(
      IDENTITY_VERIFICATION_STATES.has(identity?.verification?.[status]),
      `identity metadata: ${status} verification must be partial or complete`,
    );
    assert(
      profile.identity_extension.verification?.[status] ===
        identity?.verification?.[status],
      `identity metadata: ${status} verification differs from the active profile`,
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
    schema.properties?.traceability,
    "profile schema: traceability definition is missing",
  );
  assert(
    typeof profile.traceability === "string" && profile.traceability.length > 0,
    "profile: traceability matrix reference is missing",
  );
  assert(
    profile.deferred.some((item) => item === "GraphQL"),
    "profile: GraphQL must be explicitly deferred",
  );
}

if (profile) {
  const traceability = validateTraceability({
    root: ROOT,
    profilePath,
  });
  traceabilityCounts = traceability.counts;
  for (const failure of traceability.failures) {
    failures.push(`traceability [${failure.code}]: ${failure.message}`);
  }
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
    `Capability profile valid: ${resourceCount} resources, ${interfaceCount} interfaces, ${scenarioCount} scenarios, ${traceabilityCounts?.rows ?? 0} traceability rows`,
  );
}
