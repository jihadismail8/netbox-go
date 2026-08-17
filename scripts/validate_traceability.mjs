#!/usr/bin/env node

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  calculateSourceDigestFromEntries,
  calculateSourceManifest,
  calculateSourceManifestAtGitRevision,
} from "./source_digest.mjs";

const SCRIPT_ROOT = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const TIER_RANK = new Map([
  ["T0", 0],
  ["T1", 1],
  ["T2", 2],
  ["T3", 3],
  ["T4", 4],
]);
const ROW_KINDS = new Set([
  "scenario",
  "resource_operation",
  "resource_contract",
  "plan_rule",
]);
const CLASSIFICATIONS = new Set(["baseline", "extension", "project"]);
const PROOF_STATES = new Set([
  "covered",
  "partial",
  "pending",
  "not_applicable",
]);
const ASSESSMENT_STATES = new Set([
  "confirmed",
  "contradicted",
  "unresolved",
  "not_applicable",
]);
const IDENTITY_VERIFICATION_STATES = new Set(["partial", "complete"]);
const EXTENSION_AXIS_DIMENSIONS = new Map([
  ["contract", ["rest_extension_contract"]],
  ["parity", ["grpc_parity"]],
  ["security", ["rest_extension_contract", "browser", "cli_security"]],
]);
const PLAN_RULE_PATTERN =
  /\*\*Rule `(plan\.[a-z0-9]+(?:[.-][a-z0-9]+)+)`\.\*\*/g;
const REVIEWED_APPLICABILITY_SHA256 =
  "472317151f92ee836d3420a8fc4d18cb0bf01fb4f8168a7127c157d231d80811";
const REVIEWED_REFERENCE_AUTHORITY_SHA256 =
  "6aa38917be80e62bfad4cde89990842600ff3e1a6658f23a12e5f3a58692047b";
const REVIEWED_ROW_SEMANTICS_SHA256 =
  "e536b9d82e488b89c37abbb2d20e5d3621b0469904f18c7ad09f3aece1dd3676";
const CONTENT_SHA256_PATTERN = /^sha256:[0-9a-f]{64}$/u;
const SOURCE_DIGEST_PATTERN = /^source-v2:sha256:[0-9a-f]{64}$/u;
const EVIDENCE_ATTESTATION_PREFIX = "netbox-go-evidence-v2:";
const EVIDENCE_DIRECTORY_PREFIX = "docs/evidence/";

function addFailure(failures, code, message) {
  failures.push({ code, message });
}

function parseJSON(filePath, failures, code) {
  try {
    return JSON.parse(fs.readFileSync(filePath, "utf8"));
  } catch (error) {
    addFailure(
      failures,
      code,
      `${path.relative(SCRIPT_ROOT, filePath)}: ${error.message}`,
    );
    return null;
  }
}

function equalJSON(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function canonicalize(value) {
  if (Array.isArray(value)) return value.map(canonicalize);
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, item]) => [key, canonicalize(item)]),
    );
  }
  return value;
}

function equalCanonicalJSON(left, right) {
  return (
    JSON.stringify(canonicalize(left)) === JSON.stringify(canonicalize(right))
  );
}

function reviewedApplicabilityDigest(document) {
  const applicabilitySets = [...(document.applicability_sets ?? [])].sort(
    (left, right) => String(left?.id).localeCompare(String(right?.id)),
  );
  const rowApplicability = Object.fromEntries(
    Object.entries(document.row_applicability ?? {}).sort(([left], [right]) =>
      left.localeCompare(right),
    ),
  );
  const serialized = `${JSON.stringify(
    canonicalize({
      applicability_sets: applicabilitySets,
      row_applicability: rowApplicability,
    }),
  )}\n`;
  return crypto.createHash("sha256").update(serialized).digest("hex");
}

function reviewedReferenceAuthorityDigest(document) {
  const referenceSets = [...(document.reference_sets ?? [])].sort(
    (left, right) => String(left?.id).localeCompare(String(right?.id)),
  );
  const serialized = `${JSON.stringify(
    canonicalize({ reference_sets: referenceSets }),
  )}\n`;
  return crypto.createHash("sha256").update(serialized).digest("hex");
}

function reviewedRowSemanticsDigest(document) {
  const rows = [...(document.rows ?? [])]
    .map(
      ({
        assessment_set: _assessment,
        verification_set: _verification,
        ...row
      }) => row,
    )
    .sort((left, right) => String(left?.id).localeCompare(String(right?.id)));
  const serialized = `${JSON.stringify(
    canonicalize({ operation_catalog: document.operation_catalog, rows }),
  )}\n`;
  return crypto.createHash("sha256").update(serialized).digest("hex");
}

export function calculateRetainedClaimDigest(document, verificationSetID) {
  const verification = (document.verification_sets ?? []).find(
    (item) => item?.id === verificationSetID,
  );
  if (!verification) {
    throw new Error(`unknown verification set ${verificationSetID}`);
  }
  const byID = (items) =>
    new Map((items ?? []).map((item) => [item?.id, item]));
  const assessments = byID(document.assessment_sets);
  const applicability = byID(document.applicability_sets);
  const proofs = byID(document.proof_sets);
  const consumers = (document.rows ?? [])
    .filter((row) => row?.verification_set === verificationSetID)
    .sort((left, right) => String(left?.id).localeCompare(String(right?.id)))
    .map((row) => {
      const applicabilitySetID = document.row_applicability?.[row.id];
      const proofSetID = document.row_proofs?.[row.id];
      return {
        row,
        assessment_assignment: row.assessment_set,
        assessment: assessments.get(row.assessment_set) ?? null,
        applicability_assignment: applicabilitySetID ?? null,
        applicability: applicability.get(applicabilitySetID) ?? null,
        proof_assignment: proofSetID ?? null,
        proof: proofs.get(proofSetID) ?? null,
      };
    });
  const serialized = `${JSON.stringify(
    canonicalize({
      schema_version: 2,
      verification,
      consumers,
    }),
  )}\n`;
  return `sha256:${crypto
    .createHash("sha256")
    .update(serialized, "utf8")
    .digest("hex")}`;
}

function resolveLocalSchemaReference(rootSchema, reference) {
  if (typeof reference !== "string" || !reference.startsWith("#/")) {
    return null;
  }
  let current = rootSchema;
  for (const token of reference
    .slice(2)
    .split("/")
    .map((part) => part.replaceAll("~1", "/").replaceAll("~0", "~"))) {
    current = current?.[token];
  }
  return current ?? null;
}

function validateJSONSchema(
  value,
  schema,
  rootSchema,
  failures,
  context = "traceability",
) {
  if (schema?.$ref !== undefined) {
    const target = resolveLocalSchemaReference(rootSchema, schema.$ref);
    if (!target) {
      addFailure(
        failures,
        "schema-definition-invalid",
        `${context}: unresolved schema reference ${schema.$ref}`,
      );
      return;
    }
    validateJSONSchema(value, target, rootSchema, failures, context);
    return;
  }
  if (Array.isArray(schema?.oneOf)) {
    let matches = 0;
    for (const alternative of schema.oneOf) {
      const alternativeFailures = [];
      validateJSONSchema(
        value,
        alternative,
        rootSchema,
        alternativeFailures,
        context,
      );
      if (alternativeFailures.length === 0) matches += 1;
    }
    if (matches !== 1) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: expected exactly one schema alternative, matched ${matches}`,
      );
    }
    return;
  }
  if (Object.hasOwn(schema ?? {}, "const") && !equalJSON(value, schema.const)) {
    addFailure(
      failures,
      "schema-validation",
      `${context}: value does not match the required constant`,
    );
  }
  if (
    Array.isArray(schema?.enum) &&
    !schema.enum.some((item) => equalJSON(value, item))
  ) {
    addFailure(
      failures,
      "schema-validation",
      `${context}: value is not in the allowed enumeration`,
    );
  }

  if (schema?.type === "object") {
    if (value === null || typeof value !== "object" || Array.isArray(value)) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: expected an object`,
      );
      return;
    }
    for (const required of schema.required ?? []) {
      if (!Object.hasOwn(value, required)) {
        addFailure(
          failures,
          "schema-validation",
          `${context}: missing required property ${required}`,
        );
      }
    }
    const properties = schema.properties ?? {};
    if (schema.additionalProperties === false) {
      for (const key of Object.keys(value)) {
        if (!Object.hasOwn(properties, key)) {
          addFailure(
            failures,
            "schema-validation",
            `${context}: unsupported property ${key}`,
          );
        }
      }
    } else if (
      schema.additionalProperties &&
      typeof schema.additionalProperties === "object"
    ) {
      for (const [key, propertyValue] of Object.entries(value)) {
        if (!Object.hasOwn(properties, key)) {
          validateJSONSchema(
            propertyValue,
            schema.additionalProperties,
            rootSchema,
            failures,
            `${context}.${key}`,
          );
        }
      }
    }
    for (const [key, propertySchema] of Object.entries(properties)) {
      if (Object.hasOwn(value, key)) {
        validateJSONSchema(
          value[key],
          propertySchema,
          rootSchema,
          failures,
          `${context}.${key}`,
        );
      }
    }
  } else if (schema?.type === "array") {
    if (!Array.isArray(value)) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: expected an array`,
      );
      return;
    }
    if (value.length < (schema.minItems ?? 0)) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: expected at least ${schema.minItems} items`,
      );
    }
    if (schema.uniqueItems === true) {
      const serialized = value.map((item) => JSON.stringify(item));
      if (new Set(serialized).size !== serialized.length) {
        addFailure(
          failures,
          "schema-validation",
          `${context}: items must be unique`,
        );
      }
    }
    if (schema.items) {
      for (const [index, item] of value.entries()) {
        validateJSONSchema(
          item,
          schema.items,
          rootSchema,
          failures,
          `${context}[${index}]`,
        );
      }
    }
  } else if (schema?.type === "string") {
    if (typeof value !== "string") {
      addFailure(
        failures,
        "schema-validation",
        `${context}: expected a string`,
      );
      return;
    }
    if (value.length < (schema.minLength ?? 0)) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: string is shorter than ${schema.minLength}`,
      );
    }
    if (schema.pattern && !new RegExp(schema.pattern, "u").test(value)) {
      addFailure(
        failures,
        "schema-validation",
        `${context}: string does not match ${schema.pattern}`,
      );
    }
  }
}

function resolveGitDirectory(workTree) {
  const dotGit = path.join(workTree, ".git");
  const stat = fs.statSync(dotGit);
  if (stat.isDirectory()) return dotGit;
  if (!stat.isFile()) throw new Error(`${dotGit} is not a Git directory link`);
  const match = /^gitdir:\s*(.+)$/u.exec(
    fs.readFileSync(dotGit, "utf8").trim(),
  );
  if (!match) throw new Error(`${dotGit} has an invalid gitdir link`);
  return path.resolve(path.dirname(dotGit), match[1]);
}

function gitCommonDirectory(gitDirectory) {
  const commonDirPath = path.join(gitDirectory, "commondir");
  if (!fs.existsSync(commonDirPath)) return gitDirectory;
  return path.resolve(
    gitDirectory,
    fs.readFileSync(commonDirPath, "utf8").trim(),
  );
}

function readGitReference(gitDirectory, reference) {
  const commonDirectory = gitCommonDirectory(gitDirectory);
  for (const directory of new Set([gitDirectory, commonDirectory])) {
    const loosePath = path.join(directory, ...reference.split("/"));
    if (fs.existsSync(loosePath)) {
      return fs.readFileSync(loosePath, "utf8").trim();
    }
  }
  for (const directory of new Set([gitDirectory, commonDirectory])) {
    const packedPath = path.join(directory, "packed-refs");
    if (!fs.existsSync(packedPath)) continue;
    for (const line of fs.readFileSync(packedPath, "utf8").split(/\r?\n/u)) {
      if (line.startsWith("#") || line.startsWith("^") || line.length === 0) {
        continue;
      }
      const [sha, name] = line.split(" ", 2);
      if (name === reference) return sha;
    }
  }
  throw new Error(`unable to resolve ${reference}`);
}

function readGitHead(workTree) {
  const gitDirectory = resolveGitDirectory(workTree);
  const head = fs.readFileSync(path.join(gitDirectory, "HEAD"), "utf8").trim();
  const symbolic = /^ref:\s*(.+)$/u.exec(head);
  const sha = symbolic ? readGitReference(gitDirectory, symbolic[1]) : head;
  if (!/^[0-9a-f]{40}$/u.test(sha)) {
    throw new Error(`HEAD resolved to an invalid object ID: ${sha}`);
  }
  return sha;
}

function clone(value) {
  return structuredClone(value);
}

function indexUnique(items, keyOf, failures, code, label) {
  const index = new Map();
  if (!Array.isArray(items)) {
    addFailure(failures, code, `${label} must be an array`);
    return index;
  }
  for (const item of items) {
    const key = keyOf(item);
    if (typeof key !== "string" || key.length === 0) {
      addFailure(failures, code, `${label} contains an item without an ID`);
      continue;
    }
    if (index.has(key)) {
      addFailure(failures, code, `${label} contains duplicate ${key}`);
      continue;
    }
    index.set(key, item);
  }
  return index;
}

function countOccurrences(text, anchor) {
  if (anchor.length === 0) return 0;
  let count = 0;
  let offset = 0;
  while (true) {
    const next = text.indexOf(anchor, offset);
    if (next === -1) return count;
    count += 1;
    offset = next + anchor.length;
  }
}

function canonicalRepositoryPath(relativePath) {
  if (
    typeof relativePath !== "string" ||
    relativePath.length === 0 ||
    relativePath.includes("\\") ||
    relativePath.includes("\0") ||
    relativePath.includes("\r") ||
    relativePath.includes("\n") ||
    path.posix.isAbsolute(relativePath)
  ) {
    return null;
  }
  const normalized = path.posix.normalize(relativePath);
  if (
    normalized !== relativePath ||
    normalized === "." ||
    normalized.split("/").includes("..")
  ) {
    return null;
  }
  return normalized;
}

function pathWithin(rootPath, candidatePath) {
  return (
    candidatePath === rootPath ||
    candidatePath.startsWith(`${rootPath}${path.sep}`)
  );
}

function validateReference(root, reference, failures, context) {
  if (
    reference === null ||
    typeof reference !== "object" ||
    Array.isArray(reference)
  ) {
    addFailure(
      failures,
      "invalid-reference",
      `${context}: reference must be an object`,
    );
    return;
  }
  const referenceKeys = Object.keys(reference).sort();
  const referenceShape = equalJSON(referenceKeys, ["anchor", "path"]);
  const evidenceShape = equalJSON(referenceKeys, [
    "anchor",
    "path",
    "payload_sha256",
  ]);
  if (!referenceShape && !evidenceShape) {
    addFailure(
      failures,
      "invalid-reference",
      `${context}: reference must contain exact path/anchor fields, plus payload_sha256 only for retained evidence`,
    );
  }
  const relativePath = reference.path;
  const anchor = reference.anchor;
  const canonicalPath = canonicalRepositoryPath(relativePath);
  if (canonicalPath === null) {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: path must be canonical, POSIX-separated, and repository-relative`,
    );
    return;
  }
  const absolutePath = path.resolve(root, canonicalPath);
  if (absolutePath !== root && !absolutePath.startsWith(`${root}${path.sep}`)) {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: path escapes the repository`,
    );
    return;
  }
  let stat;
  try {
    stat = fs.statSync(absolutePath);
  } catch {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: ${relativePath} does not exist`,
    );
    return;
  }
  if (!stat.isFile()) {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: ${relativePath} is not a file`,
    );
    return;
  }
  let realRoot;
  let realPath;
  try {
    realRoot = fs.realpathSync(root);
    realPath = fs.realpathSync(absolutePath);
  } catch (error) {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: cannot resolve ${relativePath}: ${error.message}`,
    );
    return;
  }
  if (!pathWithin(realRoot, realPath)) {
    addFailure(
      failures,
      "stale-reference-path",
      `${context}: ${relativePath} resolves outside the repository`,
    );
    return;
  }
  if (typeof anchor !== "string" || anchor.length === 0) {
    addFailure(
      failures,
      "stale-reference-anchor",
      `${context}: ${relativePath} needs an exact non-empty anchor`,
    );
    return;
  }
  const source = fs.readFileSync(absolutePath, "utf8");
  const occurrences = countOccurrences(source, anchor);
  if (occurrences === 0) {
    addFailure(
      failures,
      "stale-reference-anchor",
      `${context}: anchor is absent from ${relativePath}`,
    );
  } else if (occurrences > 1) {
    addFailure(
      failures,
      "ambiguous-reference-anchor",
      `${context}: anchor occurs ${occurrences} times in ${relativePath}`,
    );
  }
}

function validateReferenceList(
  root,
  references,
  failures,
  context,
  { required = true } = {},
) {
  if (!Array.isArray(references)) {
    addFailure(
      failures,
      "invalid-reference-list",
      `${context}: references must be an array`,
    );
    return;
  }
  if (required && references.length === 0) {
    addFailure(
      failures,
      "missing-reference",
      `${context}: at least one exact reference is required`,
    );
  }
  const seen = new Set();
  for (const [index, reference] of references.entries()) {
    const key = `${reference?.path ?? ""}\u0000${reference?.anchor ?? ""}`;
    if (seen.has(key)) {
      addFailure(
        failures,
        "duplicate-reference",
        `${context}: duplicate reference at index ${index}`,
      );
    }
    seen.add(key);
    validateReference(root, reference, failures, `${context}[${index}]`);
  }
}

function validateEvidenceReferenceList(
  root,
  references,
  failures,
  context,
  { required = true } = {},
) {
  validateReferenceList(root, references, failures, context, { required });
  for (const [index, reference] of (Array.isArray(references)
    ? references
    : []
  ).entries()) {
    if (
      reference === null ||
      typeof reference !== "object" ||
      Array.isArray(reference) ||
      !equalJSON(Object.keys(reference).sort(), [
        "anchor",
        "path",
        "payload_sha256",
      ]) ||
      !CONTENT_SHA256_PATTERN.test(reference.payload_sha256 ?? "")
    ) {
      addFailure(
        failures,
        "invalid-evidence-payload",
        `${context}[${index}]: retained evidence requires exact path, anchor, and payload_sha256`,
      );
    }
  }
}

function validateLinkGroup(root, group, failures, context) {
  if (group === null || typeof group !== "object" || Array.isArray(group)) {
    addFailure(
      failures,
      "invalid-link-group",
      `${context}: link group must be an object`,
    );
    return;
  }
  if (group.status === "linked") {
    validateReferenceList(root, group.references, failures, context);
    if (group.reason !== undefined) {
      addFailure(
        failures,
        "invalid-link-group",
        `${context}: linked references cannot use a not-applicable reason`,
      );
    }
    return;
  }
  if (group.status === "not_applicable") {
    if (typeof group.reason !== "string" || group.reason.length === 0) {
      addFailure(
        failures,
        "invalid-link-group",
        `${context}: not_applicable requires a reason`,
      );
    }
    if (Array.isArray(group.references) && group.references.length > 0) {
      addFailure(
        failures,
        "invalid-link-group",
        `${context}: not_applicable cannot carry references`,
      );
    }
    return;
  }
  addFailure(
    failures,
    "invalid-link-group",
    `${context}: status must be linked or not_applicable`,
  );
}

function validateProof(root, proof, failures, context) {
  if (proof === null || typeof proof !== "object" || Array.isArray(proof)) {
    addFailure(
      failures,
      "invalid-proof",
      `${context}: proof must be an object`,
    );
    return;
  }
  if (!PROOF_STATES.has(proof.status)) {
    addFailure(
      failures,
      "invalid-proof",
      `${context}: invalid proof status ${proof.status ?? "<missing>"}`,
    );
    return;
  }
  const references = proof.references ?? [];
  const needsReference =
    proof.status === "covered" || proof.status === "partial";
  validateReferenceList(root, references, failures, context, {
    required: needsReference,
  });
  if (
    proof.status !== "covered" &&
    (typeof proof.reason !== "string" || proof.reason.length === 0)
  ) {
    addFailure(
      failures,
      "invalid-proof",
      `${context}: ${proof.status} requires a narrow reason`,
    );
  }
  if (
    (proof.status === "pending" || proof.status === "not_applicable") &&
    references.length > 0
  ) {
    addFailure(
      failures,
      "invalid-proof",
      `${context}: ${proof.status} cannot carry test references`,
    );
  }
}

function profileResourceMap(profile) {
  return new Map(
    (profile?.resources ?? []).map((resource) => [
      `${resource.module}.${resource.name}`,
      resource,
    ]),
  );
}

function loadResourceMetadata(root, profilePath, profile, failures) {
  const resources = new Map();
  const profileDir = path.dirname(profilePath);
  for (const relativePath of profile?.resource_metadata ?? []) {
    const metadata = parseJSON(
      path.resolve(profileDir, relativePath),
      failures,
      "resource-metadata-read",
    );
    for (const resource of metadata?.resources ?? []) {
      const key = `${metadata.module}.${resource.name}`;
      if (resources.has(key)) {
        addFailure(
          failures,
          "duplicate-resource-metadata",
          `resource metadata contains duplicate ${key}`,
        );
      } else {
        resources.set(key, resource);
      }
    }
  }
  return resources;
}

function loadAndValidateIdentityMetadata(
  profilePath,
  profile,
  suppliedDocument,
  failures,
) {
  const extension = profile.identity_extension;
  if (
    extension === null ||
    typeof extension !== "object" ||
    Array.isArray(extension)
  ) {
    addFailure(
      failures,
      "invalid-extension-verification",
      "profile identity_extension must be a closed object",
    );
    return null;
  }
  const extensionKeys = Object.keys(extension).sort();
  const expectedExtensionKeys = [
    "classification",
    "owner",
    "resource_metadata",
    "tier",
    "verification",
  ];
  const verificationKeys = Object.keys(extension.verification ?? {}).sort();
  const expectedVerificationKeys = [...EXTENSION_AXIS_DIMENSIONS.keys()].sort();
  if (
    JSON.stringify(extensionKeys) !== JSON.stringify(expectedExtensionKeys) ||
    JSON.stringify(verificationKeys) !==
      JSON.stringify(expectedVerificationKeys)
  ) {
    addFailure(
      failures,
      "invalid-extension-verification",
      "profile identity_extension and verification must contain exactly the schema-declared properties",
    );
  }
  if (
    extension.classification !== "extension" ||
    extension.tier !== "not_applicable" ||
    typeof extension.resource_metadata !== "string" ||
    extension.resource_metadata.length === 0 ||
    typeof extension.owner !== "string" ||
    extension.owner.length === 0
  ) {
    addFailure(
      failures,
      "invalid-extension-verification",
      "profile identity_extension must declare extension classification, no T tier, metadata, and an owner",
    );
  }
  const metadataPath = path.resolve(
    path.dirname(profilePath),
    extension.resource_metadata ?? "",
  );
  const metadata =
    suppliedDocument ??
    parseJSON(metadataPath, failures, "identity-metadata-read");
  for (const axis of EXTENSION_AXIS_DIMENSIONS.keys()) {
    const profileStatus = extension.verification?.[axis];
    const metadataStatus = metadata?.verification?.[axis];
    if (!IDENTITY_VERIFICATION_STATES.has(profileStatus)) {
      addFailure(
        failures,
        "invalid-extension-verification",
        `profile identity ${axis} verification must be partial or complete`,
      );
    }
    if (!IDENTITY_VERIFICATION_STATES.has(metadataStatus)) {
      addFailure(
        failures,
        "invalid-extension-verification",
        `identity metadata ${axis} verification must be partial or complete`,
      );
    }
    if (profileStatus !== metadataStatus) {
      addFailure(
        failures,
        "extension-verification-drift",
        `identity metadata ${axis} verification differs from the active profile`,
      );
    }
  }
  return metadata;
}

function loadScenarioIDs(contractRoot, failures) {
  const ids = new Set();
  const scenarioDir = path.join(contractRoot, "scenarios");
  for (const name of fs
    .readdirSync(scenarioDir)
    .filter((entry) => entry.endsWith(".yaml"))
    .sort()) {
    const document = parseJSON(
      path.join(scenarioDir, name),
      failures,
      "scenario-read",
    );
    for (const scenario of document?.scenarios ?? []) {
      if (ids.has(scenario.id)) {
        addFailure(
          failures,
          "duplicate-scenario-source",
          `scenario sources contain duplicate ${scenario.id}`,
        );
      }
      ids.add(scenario.id);
    }
  }
  return ids;
}

function loadPlanRuleIDs(root, implementationPlan, failures) {
  const ids = new Set();
  const planPath = path.resolve(root, implementationPlan);
  let source;
  try {
    source = fs.readFileSync(planPath, "utf8");
  } catch (error) {
    addFailure(
      failures,
      "plan-read",
      `${implementationPlan}: ${error.message}`,
    );
    return ids;
  }
  for (const match of source.matchAll(PLAN_RULE_PATTERN)) {
    const id = match[1];
    if (ids.has(id)) {
      addFailure(
        failures,
        "duplicate-plan-rule-source",
        `implementation plan contains duplicate ${id}`,
      );
    }
    ids.add(id);
  }
  if (ids.size === 0) {
    addFailure(
      failures,
      "missing-plan-rule-sources",
      "implementation plan has no stable plan.* rule labels",
    );
  }
  return ids;
}

function validateProfileSelector(
  selector,
  classification,
  profileResources,
  metadataResources,
  failures,
  context,
) {
  if (
    selector === null ||
    typeof selector !== "object" ||
    Array.isArray(selector)
  ) {
    addFailure(
      failures,
      "invalid-profile-selector",
      `${context}: profile selector must be an object`,
    );
    return;
  }
  if (selector.kind === "resource") {
    if (
      classification !== "baseline" ||
      !profileResources.has(selector.key) ||
      !metadataResources.has(selector.key)
    ) {
      addFailure(
        failures,
        "invalid-profile-selector",
        `${context}: unknown or misclassified resource ${selector.key ?? "<missing>"}`,
      );
    }
    return;
  }
  if (selector.kind === "identity_extension") {
    if (classification !== "extension") {
      addFailure(
        failures,
        "invalid-profile-selector",
        `${context}: identity extension must be classified as extension`,
      );
    }
    return;
  }
  if (selector.kind === "profile") {
    if (classification === "extension") {
      addFailure(
        failures,
        "invalid-profile-selector",
        `${context}: extension cannot use the whole-profile selector`,
      );
    }
    return;
  }
  addFailure(
    failures,
    "invalid-profile-selector",
    `${context}: unsupported selector kind ${selector.kind ?? "<missing>"}`,
  );
}

function validateMetadataSelector(
  selector,
  classification,
  profile,
  profilePath,
  metadataResources,
  failures,
  context,
) {
  if (
    selector === null ||
    typeof selector !== "object" ||
    Array.isArray(selector)
  ) {
    addFailure(
      failures,
      "invalid-metadata-selector",
      `${context}: metadata selector must be an object`,
    );
    return;
  }
  if (selector.kind === "resource") {
    if (classification !== "baseline" || !metadataResources.has(selector.key)) {
      addFailure(
        failures,
        "invalid-metadata-selector",
        `${context}: unknown or misclassified resource ${selector.key ?? "<missing>"}`,
      );
    }
    return;
  }
  if (selector.kind === "identity") {
    const identityPath = path.resolve(
      path.dirname(profilePath),
      profile.identity_extension?.resource_metadata ?? "",
    );
    if (classification !== "extension" || !fs.existsSync(identityPath)) {
      addFailure(
        failures,
        "invalid-metadata-selector",
        `${context}: identity metadata is missing or misclassified`,
      );
    }
    return;
  }
  if (selector.kind === "not_applicable") {
    if (
      classification !== "project" ||
      typeof selector.reason !== "string" ||
      selector.reason.length === 0
    ) {
      addFailure(
        failures,
        "invalid-metadata-selector",
        `${context}: not_applicable requires a project classification and reason`,
      );
    }
    return;
  }
  addFailure(
    failures,
    "invalid-metadata-selector",
    `${context}: unsupported metadata selector kind ${selector.kind ?? "<missing>"}`,
  );
}

function validateReferenceSet(
  root,
  referenceSet,
  dimensions,
  profile,
  profilePath,
  profileResources,
  metadataResources,
  failures,
) {
  const context = `reference set ${referenceSet.id ?? "<missing>"}`;
  if (!CLASSIFICATIONS.has(referenceSet.classification)) {
    addFailure(
      failures,
      "invalid-reference-classification",
      `${context}: invalid classification`,
    );
  }
  if (
    !Array.isArray(referenceSet.capabilities) ||
    referenceSet.capabilities.length === 0 ||
    new Set(referenceSet.capabilities).size !== referenceSet.capabilities.length
  ) {
    addFailure(
      failures,
      "invalid-reference-capabilities",
      `${context}: capabilities must be a non-empty unique array`,
    );
  }
  validateProfileSelector(
    referenceSet.profile_ref,
    referenceSet.classification,
    profileResources,
    metadataResources,
    failures,
    context,
  );
  validateMetadataSelector(
    referenceSet.metadata_ref,
    referenceSet.classification,
    profile,
    profilePath,
    metadataResources,
    failures,
    context,
  );
  validateLinkGroup(
    root,
    referenceSet.pinned_source,
    failures,
    `${context} pinned source`,
  );
  validateLinkGroup(
    root,
    referenceSet.upstream_tests,
    failures,
    `${context} upstream tests`,
  );
  const requiresPinnedUpstream =
    referenceSet.classification === "baseline" ||
    referenceSet.id === "identity" ||
    referenceSet.id === "identity-baseline-support" ||
    referenceSet.id === "identity-interface-support" ||
    referenceSet.id === "baseline-common-support" ||
    referenceSet.id === "deferred-baseline-support";
  if (
    requiresPinnedUpstream &&
    (referenceSet.pinned_source?.status !== "linked" ||
      referenceSet.upstream_tests?.status !== "linked")
  ) {
    addFailure(
      failures,
      "missing-upstream-link",
      `${context}: managed-resource and identity reference sets require linked pinned source and upstream tests`,
    );
  }
  if (requiresPinnedUpstream) {
    for (const reference of referenceSet.pinned_source?.references ?? []) {
      if (!reference.path.startsWith("netbox/")) {
        addFailure(
          failures,
          "invalid-upstream-reference",
          `${context}: pinned source must be under the pinned netbox tree`,
        );
      }
    }
    for (const reference of referenceSet.upstream_tests?.references ?? []) {
      if (
        !reference.path.startsWith("netbox/") ||
        (!reference.path.includes("/tests/") &&
          !reference.path.startsWith("netbox/netbox/utilities/testing/"))
      ) {
        addFailure(
          failures,
          "invalid-upstream-reference",
          `${context}: upstream tests must be under pinned NetBox tests or reusable testing support`,
        );
      }
    }
  }

  const proofs = referenceSet.test_inventory;
  if (proofs === null || typeof proofs !== "object" || Array.isArray(proofs)) {
    addFailure(
      failures,
      "missing-proof-dimension",
      `${context}: test_inventory must be an object`,
    );
    return;
  }
  const proofKeys = Object.keys(proofs).sort();
  const expectedKeys = [...dimensions].sort();
  if (JSON.stringify(proofKeys) !== JSON.stringify(expectedKeys)) {
    addFailure(
      failures,
      "missing-proof-dimension",
      `${context}: expected exactly ${expectedKeys.join(", ")}`,
    );
  }
  for (const dimension of dimensions) {
    if (proofs[dimension] !== undefined) {
      validateProof(
        root,
        proofs[dimension],
        failures,
        `${context} ${dimension}`,
      );
    }
  }
}

function validateProofSet(root, proofSet, dimensions, failures) {
  const context = `proof set ${proofSet.id ?? "<missing>"}`;
  const proofs = proofSet.proofs;
  if (proofs === null || typeof proofs !== "object" || Array.isArray(proofs)) {
    addFailure(
      failures,
      "missing-proof-dimension",
      `${context}: proofs must be an object`,
    );
    return;
  }
  const proofKeys = Object.keys(proofs).sort();
  const expectedKeys = [...dimensions].sort();
  if (JSON.stringify(proofKeys) !== JSON.stringify(expectedKeys)) {
    addFailure(
      failures,
      "missing-proof-dimension",
      `${context}: expected exactly ${expectedKeys.join(", ")}`,
    );
  }
  for (const dimension of dimensions) {
    if (proofs[dimension] !== undefined) {
      validateProof(
        root,
        proofs[dimension],
        failures,
        `${context} ${dimension}`,
      );
    }
  }
}

function validateApplicabilitySet(applicabilitySet, dimensions, failures) {
  const context = `applicability set ${applicabilitySet.id ?? "<missing>"}`;
  const declarations = applicabilitySet.dimensions;
  if (
    declarations === null ||
    typeof declarations !== "object" ||
    Array.isArray(declarations)
  ) {
    addFailure(
      failures,
      "missing-applicability-dimension",
      `${context}: dimensions must be an object`,
    );
    return;
  }
  const keys = Object.keys(declarations).sort();
  const expectedKeys = [...dimensions].sort();
  if (JSON.stringify(keys) !== JSON.stringify(expectedKeys)) {
    addFailure(
      failures,
      "missing-applicability-dimension",
      `${context}: expected exactly ${expectedKeys.join(", ")}`,
    );
  }
  for (const dimension of dimensions) {
    const declaration = declarations[dimension];
    if (
      declaration === null ||
      typeof declaration !== "object" ||
      Array.isArray(declaration) ||
      !["applicable", "not_applicable"].includes(declaration.status)
    ) {
      addFailure(
        failures,
        "invalid-applicability",
        `${context} ${dimension}: status must be applicable or not_applicable`,
      );
      continue;
    }
    if (
      declaration.status === "not_applicable" &&
      (typeof declaration.reason !== "string" ||
        declaration.reason.length === 0)
    ) {
      addFailure(
        failures,
        "invalid-applicability",
        `${context} ${dimension}: not_applicable requires a narrow reason`,
      );
    }
    if (
      declaration.status === "applicable" &&
      declaration.reason !== undefined
    ) {
      addFailure(
        failures,
        "invalid-applicability",
        `${context} ${dimension}: applicable cannot carry a reason`,
      );
    }
  }
}

function validateAssessmentSet(root, assessment, failures) {
  const context = `assessment set ${assessment.id ?? "<missing>"}`;
  if (!ASSESSMENT_STATES.has(assessment.status)) {
    addFailure(
      failures,
      "invalid-assessment",
      `${context}: invalid status ${assessment.status ?? "<missing>"}`,
    );
    return;
  }
  if (assessment.status !== "confirmed") {
    if (
      typeof assessment.reason !== "string" ||
      assessment.reason.length === 0
    ) {
      addFailure(
        failures,
        "invalid-assessment",
        `${context}: ${assessment.status} requires a reason`,
      );
    }
  }
  if (
    assessment.status === "contradicted" ||
    assessment.status === "unresolved"
  ) {
    if (
      typeof assessment.resolution_goal !== "string" ||
      assessment.resolution_goal.length === 0
    ) {
      addFailure(
        failures,
        "invalid-assessment",
        `${context}: ${assessment.status} requires a resolution_goal`,
      );
    }
  }
  if (assessment.status === "contradicted") {
    validateReferenceList(
      root,
      assessment.conflict_references,
      failures,
      `${context} conflicts`,
    );
  } else if (
    Array.isArray(assessment.conflict_references) &&
    assessment.conflict_references.length > 0
  ) {
    addFailure(
      failures,
      "invalid-assessment",
      `${context}: only contradicted assessments carry conflict references`,
    );
  }
}

function evidenceDocumentBytes(root, relativePath, evidenceDocuments) {
  if (evidenceDocuments instanceof Map && evidenceDocuments.has(relativePath)) {
    const source = evidenceDocuments.get(relativePath);
    if (Buffer.isBuffer(source) || source instanceof Uint8Array) {
      return Buffer.from(source);
    }
    if (typeof source !== "string") {
      throw new Error("injected evidence document must be bytes or a string");
    }
    return Buffer.from(source, "utf8");
  }
  return fs.readFileSync(path.resolve(root, relativePath));
}

function decodeEvidenceUTF8(source) {
  return new TextDecoder("utf-8", { fatal: true }).decode(source);
}

function parseStrictJSON(source) {
  let offset = 0;
  const fail = (message) => {
    throw new Error(`${message} at offset ${offset}`);
  };
  const skipWhitespace = () => {
    while (
      source[offset] === " " ||
      source[offset] === "\t" ||
      source[offset] === "\r" ||
      source[offset] === "\n"
    ) {
      offset += 1;
    }
  };
  const parseString = () => {
    if (source[offset] !== '"') fail("expected JSON string");
    const start = offset;
    offset += 1;
    while (offset < source.length) {
      if (source[offset] === '"') {
        offset += 1;
        return JSON.parse(source.slice(start, offset));
      }
      if (source[offset] === "\\") {
        offset += 2;
      } else {
        offset += 1;
      }
    }
    fail("unterminated JSON string");
  };
  const parseValue = () => {
    skipWhitespace();
    const token = source[offset];
    if (token === '"') return parseString();
    if (token === "{") {
      offset += 1;
      skipWhitespace();
      const value = Object.create(null);
      const keys = new Set();
      if (source[offset] === "}") {
        offset += 1;
        return value;
      }
      while (offset < source.length) {
        const key = parseString();
        if (keys.has(key)) fail(`duplicate JSON member ${JSON.stringify(key)}`);
        keys.add(key);
        skipWhitespace();
        if (source[offset] !== ":") fail("expected colon after JSON member");
        offset += 1;
        Object.defineProperty(value, key, {
          value: parseValue(),
          enumerable: true,
          configurable: true,
          writable: true,
        });
        skipWhitespace();
        if (source[offset] === "}") {
          offset += 1;
          return value;
        }
        if (source[offset] !== ",") fail("expected comma in JSON object");
        offset += 1;
        skipWhitespace();
      }
      fail("unterminated JSON object");
    }
    if (token === "[") {
      offset += 1;
      skipWhitespace();
      const value = [];
      if (source[offset] === "]") {
        offset += 1;
        return value;
      }
      while (offset < source.length) {
        value.push(parseValue());
        skipWhitespace();
        if (source[offset] === "]") {
          offset += 1;
          return value;
        }
        if (source[offset] !== ",") fail("expected comma in JSON array");
        offset += 1;
      }
      fail("unterminated JSON array");
    }
    for (const [literal, value] of [
      ["true", true],
      ["false", false],
      ["null", null],
    ]) {
      if (source.startsWith(literal, offset)) {
        offset += literal.length;
        return value;
      }
    }
    const number =
      /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?/u.exec(
        source.slice(offset),
      );
    if (number) {
      offset += number[0].length;
      return Number(number[0]);
    }
    fail("invalid JSON value");
  };

  const value = parseValue();
  skipWhitespace();
  if (offset !== source.length) fail("unexpected trailing JSON content");
  return value;
}

function parseEvidenceAttestations(source, failures, context) {
  const prefixPattern = /<!--[\t ]*netbox-go-evidence-v2:/gu;
  const markerPattern =
    /^[\t ]*<!--[\t ]*netbox-go-evidence-v2:[\t ]*([^\r\n]*?)[\t ]*-->[\t ]*(?:\r?\n|$)/gmu;
  const prefixCount = [...source.matchAll(prefixPattern)].length;
  const markerMatches = [...source.matchAll(markerPattern)];
  if (prefixCount === 0) {
    addFailure(
      failures,
      "missing-evidence-attestation",
      `${context}: result artifact has no ${EVIDENCE_ATTESTATION_PREFIX} marker`,
    );
    return [];
  }
  if (markerMatches.length !== prefixCount) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: every ${EVIDENCE_ATTESTATION_PREFIX} marker must occupy one complete line and contain single-line JSON`,
    );
  }
  const attestations = [];
  for (const [index, match] of markerMatches.entries()) {
    try {
      const attestation = parseStrictJSON(match[1]);
      if (
        attestation === null ||
        typeof attestation !== "object" ||
        Array.isArray(attestation)
      ) {
        throw new Error("payload must be a JSON object");
      }
      attestations.push(attestation);
    } catch (error) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${context}: marker ${index + 1} is invalid: ${error.message}`,
      );
    }
  }
  return attestations;
}

function validateRawEvidenceMarkers(source, failures, context) {
  const prefix = Buffer.from(EVIDENCE_ATTESTATION_PREFIX, "ascii");
  let start = 0;
  while (start < source.length) {
    const newline = source.indexOf(0x0a, start);
    const end = newline < 0 ? source.length : newline + 1;
    const line = source.subarray(start, end);
    if (line.indexOf(prefix) >= 0 && line.some((byte) => byte >= 0x80)) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${context}: evidence attestation marker lines must be ASCII`,
      );
    }
    start = end;
  }
}

function evidencePayloadSHA256(source) {
  const retainedLines = [];
  let start = 0;
  while (start < source.length) {
    const newline = source.indexOf(0x0a, start);
    const end = newline < 0 ? source.length : newline + 1;
    const line = source.subarray(start, end);
    const ascii = line.every((byte) => byte < 0x80);
    const lineText = ascii ? line.toString("ascii") : "";
    if (
      !/^[\t ]*<!--[\t ]*netbox-go-evidence-v2:[\t ]*[^\r\n]*?[\t ]*-->[\t ]*(?:\r?\n|$)$/u.test(
        lineText,
      )
    ) {
      retainedLines.push(line);
    }
    start = end;
  }
  const payload = Buffer.concat(retainedLines);
  return `sha256:${crypto.createHash("sha256").update(payload).digest("hex")}`;
}

function validateRetainedEvidencePath(root, relativePath, failures, context) {
  const canonicalPath = canonicalRepositoryPath(relativePath);
  if (
    canonicalPath === null ||
    !canonicalPath.startsWith(EVIDENCE_DIRECTORY_PREFIX)
  ) {
    addFailure(
      failures,
      "invalid-retained-evidence",
      `${context}: retained evidence path must be canonical and lexically under docs/evidence/`,
    );
    return null;
  }
  try {
    const repositoryRoot = fs.realpathSync(root);
    const expectedEvidenceRoot = path.resolve(repositoryRoot, "docs/evidence");
    const evidenceRoot = fs.realpathSync(expectedEvidenceRoot);
    const expectedArtifactPath = path.resolve(repositoryRoot, canonicalPath);
    const artifactPath = fs.realpathSync(expectedArtifactPath);
    const ledgerPath = path.resolve(evidenceRoot, "README.md");
    const artifactStat = fs.lstatSync(expectedArtifactPath);
    const ledgerStat = fs.statSync(ledgerPath);
    if (
      evidenceRoot !== expectedEvidenceRoot ||
      artifactPath !== expectedArtifactPath ||
      !pathWithin(evidenceRoot, artifactPath) ||
      artifactPath === ledgerPath ||
      !artifactStat.isFile() ||
      artifactStat.nlink !== 1 ||
      (artifactStat.dev === ledgerStat.dev &&
        artifactStat.ino === ledgerStat.ino)
    ) {
      addFailure(
        failures,
        "invalid-retained-evidence",
        `${context}: retained evidence must resolve to a result artifact under docs/evidence, not the ledger index`,
      );
      return null;
    }
  } catch (error) {
    addFailure(
      failures,
      "invalid-retained-evidence",
      `${context}: cannot resolve retained evidence path: ${error.message}`,
    );
    return null;
  }
  return canonicalPath;
}

function contentSHA256(source) {
  return `sha256:${crypto.createHash("sha256").update(source).digest("hex")}`;
}

function validateTestedSourceManifest(
  root,
  reference,
  expectedDigest,
  revisionManifest,
  currentManifest,
  evidenceDocuments,
  failures,
  context,
) {
  const failureStart = failures.length;
  if (
    reference === null ||
    typeof reference !== "object" ||
    Array.isArray(reference) ||
    !equalJSON(Object.keys(reference).sort(), ["path", "sha256"])
  ) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested_source_manifest must contain exactly path and sha256`,
    );
    return null;
  }
  if (!CONTENT_SHA256_PATTERN.test(reference.sha256 ?? "")) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested_source_manifest sha256 is invalid`,
    );
    return null;
  }
  const manifestPath = validateRetainedEvidencePath(
    root,
    reference.path,
    failures,
    `${context} tested_source_manifest`,
  );
  if (manifestPath === null) return null;

  let sourceBytes;
  let source;
  try {
    sourceBytes = evidenceDocumentBytes(root, manifestPath, evidenceDocuments);
    source = decodeEvidenceUTF8(sourceBytes);
  } catch (error) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: cannot read tested source manifest: ${error.message}`,
    );
    return null;
  }
  if (contentSHA256(sourceBytes) !== reference.sha256) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested source manifest bytes do not match the committed sha256`,
    );
    return null;
  }

  let manifest;
  try {
    manifest = parseStrictJSON(source);
  } catch (error) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested source manifest is not strict JSON: ${error.message}`,
    );
    return null;
  }
  if (
    manifest === null ||
    typeof manifest !== "object" ||
    Array.isArray(manifest) ||
    !equalJSON(Object.keys(manifest).sort(), [
      "digest",
      "entries",
      "files",
      "schema_version",
    ]) ||
    manifest.schema_version !== 2 ||
    !Number.isSafeInteger(manifest.files) ||
    manifest.files < 1 ||
    !Array.isArray(manifest.entries) ||
    manifest.files !== manifest.entries.length
  ) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested source manifest has an invalid closed top-level shape`,
    );
    return null;
  }

  const entries = [];
  const testedByPath = new Map();
  let previousPath = "";
  for (const [index, entry] of manifest.entries.entries()) {
    const entryContext = `${context} tested source entry[${index}]`;
    if (entry === null || typeof entry !== "object" || Array.isArray(entry)) {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${entryContext}: entry must be an object`,
      );
      continue;
    }
    const canonicalPath = canonicalRepositoryPath(entry.path);
    if (
      canonicalPath === null ||
      canonicalPath <= previousPath ||
      testedByPath.has(canonicalPath)
    ) {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${entryContext}: paths must be canonical, unique, and strictly sorted`,
      );
      continue;
    }
    previousPath = canonicalPath;

    if (entry.kind === "file") {
      if (
        !equalJSON(Object.keys(entry).sort(), [
          "kind",
          "mode",
          "path",
          "sha256",
          "size",
        ]) ||
        !["100644", "100755"].includes(entry.mode) ||
        !Number.isSafeInteger(entry.size) ||
        entry.size < 0 ||
        !CONTENT_SHA256_PATTERN.test(entry.sha256 ?? "")
      ) {
        addFailure(
          failures,
          "invalid-tested-source-manifest",
          `${entryContext}: file entry must contain exact kind/path/size/sha256 data`,
        );
        continue;
      }
    } else if (entry.kind === "symlink") {
      if (
        !equalJSON(Object.keys(entry).sort(), ["kind", "path", "target"]) ||
        typeof entry.target !== "string" ||
        entry.target.length === 0 ||
        entry.target.includes("\0") ||
        entry.target.includes("\r") ||
        entry.target.includes("\n")
      ) {
        addFailure(
          failures,
          "invalid-tested-source-manifest",
          `${entryContext}: symlink entry must contain exact kind/path/target data`,
        );
        continue;
      }
    } else {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${entryContext}: kind must be file or symlink`,
      );
      continue;
    }
    const normalizedEntry = { ...entry, path: canonicalPath };
    entries.push(normalizedEntry);
    testedByPath.set(canonicalPath, normalizedEntry);
  }
  if (entries.length !== manifest.entries.length) return null;

  let calculatedDigest;
  try {
    calculatedDigest = calculateSourceDigestFromEntries(entries);
  } catch (error) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested source manifest cannot be serialized: ${error.message}`,
    );
    return null;
  }
  if (
    manifest.digest !== calculatedDigest ||
    calculatedDigest !== expectedDigest
  ) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested source manifest digest does not match tested_digest`,
    );
    return null;
  }
  if (
    revisionManifest === null ||
    !equalCanonicalJSON(manifest, revisionManifest)
  ) {
    addFailure(
      failures,
      "invalid-tested-revision",
      `${context}: tested source manifest must exactly match the owned source reconstructed from tested_revision`,
    );
    return null;
  }
  if (!currentManifest || !Array.isArray(currentManifest.entries)) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: current owned-source manifest is unavailable`,
    );
    return null;
  }

  const currentByPath = new Map(
    currentManifest.entries.map((entry) => [entry.path, entry]),
  );
  if (testedByPath.size !== currentByPath.size) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: tested/current owned-source path sets differ; additions and deletions require rerunning evidence`,
    );
    return null;
  }

  const changes = [];
  for (const [entryPath, testedEntry] of testedByPath) {
    const currentEntry = currentByPath.get(entryPath);
    if (!currentEntry || currentEntry.kind !== testedEntry.kind) {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${context}: tested/current owned-source path or kind differs at ${entryPath}; additions, deletions, and type changes require rerunning evidence`,
      );
      continue;
    }
    if (testedEntry.kind === "symlink") {
      if (testedEntry.target !== currentEntry.target) {
        addFailure(
          failures,
          "invalid-tested-source-manifest",
          `${context}: symlink change at ${entryPath} requires rerunning evidence`,
        );
      }
      continue;
    }
    if (
      testedEntry.mode !== currentEntry.mode ||
      testedEntry.size !== currentEntry.size ||
      testedEntry.sha256 !== currentEntry.sha256
    ) {
      changes.push({
        path: entryPath,
        tested_mode: testedEntry.mode,
        tested_size: testedEntry.size,
        tested_sha256: testedEntry.sha256,
        attestation_mode: currentEntry.mode,
        attestation_size: currentEntry.size,
        attestation_sha256: currentEntry.sha256,
      });
    }
  }
  if (
    failures
      .slice(failureStart)
      .some((failure) => failure.code === "invalid-tested-source-manifest")
  ) {
    return null;
  }
  return changes;
}

function isClaimOnlyPath(relativePath, traceabilityPath) {
  return (
    relativePath === traceabilityPath ||
    relativePath === "docs/STATUS.md" ||
    relativePath === "docs/COMPATIBILITY.md" ||
    /^contracts\/netbox\/[^/]+\/profiles\/[^/]+\.yaml$/u.test(relativePath) ||
    /^contracts\/netbox\/[^/]+\/resources\/identity\.yaml$/u.test(
      relativePath,
    ) ||
    relativePath === "docs/contracts/core-workflow-v1.md" ||
    relativePath === "docs/contracts/grpc-v1.md" ||
    relativePath === "netbox-backend/api/openapi/netbox-go-v1.yaml"
  );
}

function validateClaimManifest(
  root,
  claimManifest,
  verification,
  attestationDigest,
  traceabilityPath,
  testedSourceManifest,
  testedRevisionManifest,
  currentSourceManifest,
  evidenceDocuments,
  failures,
  context,
) {
  if (
    testedRevisionManifest === null ||
    testedRevisionManifest?.digest !== verification.tested_digest
  ) {
    addFailure(
      failures,
      "invalid-tested-revision",
      `${context}: tested_digest must be reconstructed from the reachable tested_revision`,
    );
  }
  if (!Array.isArray(claimManifest)) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: claim_manifest must be an array`,
    );
    return;
  }
  if (verification.tested_digest === attestationDigest) {
    if (!equalCanonicalJSON(testedRevisionManifest, currentSourceManifest)) {
      addFailure(
        failures,
        "invalid-tested-revision",
        `${context}: a same-digest claim requires the current owned source to exactly match tested_revision`,
      );
    }
    if (verification.tested_source_manifest !== undefined) {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${context}: an unchanged tested/attestation digest must not carry a tested source manifest`,
      );
    }
    if (claimManifest.length !== 0) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${context}: an unchanged tested/attestation digest requires an empty claim_manifest`,
      );
    }
    return;
  }
  if (claimManifest.length === 0) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: a two-digest claim requires a non-empty claim_manifest`,
    );
    return;
  }

  if (
    verification.tested_source_manifest === undefined ||
    !equalCanonicalJSON(
      testedSourceManifest,
      verification.tested_source_manifest,
    )
  ) {
    addFailure(
      failures,
      "invalid-tested-source-manifest",
      `${context}: a two-digest claim must bind the verification set's committed tested source manifest`,
    );
    return;
  }
  const derivedChanges = validateTestedSourceManifest(
    root,
    verification.tested_source_manifest,
    verification.tested_digest,
    testedRevisionManifest,
    currentSourceManifest,
    evidenceDocuments,
    failures,
    context,
  );
  if (derivedChanges === null) return;
  if (derivedChanges.length === 0) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: differing digests require at least one derived source change`,
    );
    return;
  }

  for (const [index, entry] of claimManifest.entries()) {
    const entryContext = `${context} claim_manifest[${index}]`;
    if (
      entry === null ||
      typeof entry !== "object" ||
      Array.isArray(entry) ||
      !equalJSON(Object.keys(entry).sort(), [
        "attestation_mode",
        "attestation_sha256",
        "attestation_size",
        "path",
        "tested_mode",
        "tested_sha256",
        "tested_size",
      ])
    ) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${entryContext}: expected exact path, tested/attestation mode, size, and sha256`,
      );
      continue;
    }
    const expected = derivedChanges[index];
    if (!expected || !equalCanonicalJSON(entry, expected)) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${entryContext}: entry does not match the complete derived tested/current source diff`,
      );
      continue;
    }
    if (!isClaimOnlyPath(expected.path, traceabilityPath)) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${entryContext}: ${expected.path} is outside the reviewed claim-only source boundary`,
      );
    }
  }
  if (claimManifest.length !== derivedChanges.length) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: claim_manifest must enumerate every and only derived source change`,
    );
  }
  if (!derivedChanges.some((entry) => entry.path === traceabilityPath)) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: a two-digest verification claim must derive a traceability-document change`,
    );
  }
}

function requiredAttestationDimensions(
  verification,
  consumers,
  applicabilitySets,
  rowApplicability,
  proofDimensions,
) {
  const required = new Set();
  const appliesToConsumer = (dimension) =>
    consumers.some(
      (row) =>
        applicabilityForRow(row, dimension, applicabilitySets, rowApplicability)
          ?.status === "applicable",
    );
  if (verification.classification === "baseline") {
    if (
      TIER_RANK.get(verification.tier) >= TIER_RANK.get("T2") &&
      appliesToConsumer("rest_differential")
    ) {
      required.add("rest_differential");
    }
    if (
      TIER_RANK.get(verification.tier) >= TIER_RANK.get("T3") &&
      appliesToConsumer("grpc_parity")
    ) {
      required.add("grpc_parity");
    }
    if (
      TIER_RANK.get(verification.tier) >= TIER_RANK.get("T4") &&
      appliesToConsumer("browser")
    ) {
      required.add("browser");
    }
  }
  if (verification.classification === "extension") {
    for (const [axis, dimensions] of EXTENSION_AXIS_DIMENSIONS) {
      if (verification.extension_verification?.[axis] !== "complete") continue;
      for (const dimension of dimensions) {
        if (appliesToConsumer(dimension)) required.add(dimension);
      }
    }
  }
  if (verification.classification === "project") {
    for (const dimension of proofDimensions) {
      if (appliesToConsumer(dimension)) {
        required.add(dimension);
      }
    }
  }
  return required;
}

function validateEvidenceAttestation(
  root,
  attestation,
  verification,
  profile,
  compatibilityBaseline,
  dimensions,
  currentSourceDigest,
  currentSourceManifest,
  testedRevisionManifest,
  retainedClaimDigest,
  traceabilityPath,
  evidenceDocuments,
  failures,
  context,
) {
  const expectedKeys = [
    "schema_version",
    "profile",
    "compatibility_baseline",
    "verification_set",
    "classification",
    "tier",
    "proof_dimensions",
    "tested_digest",
    "tested_revision",
    "attestation_digest",
    "claim_digest",
    "claim_manifest",
    "result",
    ...(verification.tested_source_manifest !== undefined
      ? ["tested_source_manifest"]
      : []),
    ...(verification.classification === "extension"
      ? ["extension_verification"]
      : []),
  ].sort();
  const actualKeys = Object.keys(attestation).sort();
  if (!equalJSON(actualKeys, expectedKeys)) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: expected exactly ${expectedKeys.join(", ")}`,
    );
  }
  if (
    attestation.schema_version !== 2 ||
    attestation.profile !== profile.id ||
    !equalCanonicalJSON(
      attestation.compatibility_baseline,
      compatibilityBaseline,
    ) ||
    attestation.verification_set !== verification.id ||
    attestation.classification !== verification.classification ||
    attestation.tier !== verification.tier ||
    attestation.tested_digest !== verification.tested_digest ||
    attestation.tested_revision !== verification.tested_revision ||
    attestation.attestation_digest !== currentSourceDigest ||
    attestation.claim_digest !== retainedClaimDigest ||
    !equalCanonicalJSON(
      attestation.tested_source_manifest,
      verification.tested_source_manifest,
    ) ||
    attestation.result !== "pass"
  ) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: profile, baseline, verification set, classification, tier, tested revision/digests, exact consumer claim digest, and pass result must match the retained claim`,
    );
  }
  validateClaimManifest(
    root,
    attestation.claim_manifest,
    verification,
    attestation.attestation_digest,
    traceabilityPath,
    attestation.tested_source_manifest,
    testedRevisionManifest,
    currentSourceManifest,
    evidenceDocuments,
    failures,
    context,
  );
  if (
    verification.classification === "extension" &&
    !equalCanonicalJSON(
      attestation.extension_verification,
      verification.extension_verification,
    )
  ) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: extension axes must exactly match the retained verification set`,
    );
  }
  const attestedDimensions = attestation.proof_dimensions;
  const uniqueDimensions = new Set(
    Array.isArray(attestedDimensions) ? attestedDimensions : [],
  );
  const canonicalDimensions = dimensions.filter((dimension) =>
    uniqueDimensions.has(dimension),
  );
  if (
    !Array.isArray(attestedDimensions) ||
    attestedDimensions.length === 0 ||
    uniqueDimensions.size !== attestedDimensions.length ||
    !equalJSON(attestedDimensions, canonicalDimensions)
  ) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: proof_dimensions must be a non-empty unique subset in canonical matrix order`,
    );
    return new Set();
  }
  return uniqueDimensions;
}

function validateRetainedEvidenceAttestations(
  root,
  verification,
  evidence,
  profile,
  compatibilityBaseline,
  dimensions,
  evidenceDocuments,
  consumers,
  applicabilitySets,
  rowApplicability,
  proofSets,
  rowProofs,
  currentSourceDigest,
  currentSourceManifest,
  testedRevisionManifest,
  retainedClaimDigest,
  traceabilityPath,
  retainedEvidenceOwners,
  failures,
) {
  const context = `verification set ${verification.id}`;
  const allAttestedDimensions = new Set();
  for (const [index, reference] of evidence.entries()) {
    const evidenceContext = `${context} evidence[${index}]`;
    const evidencePath = validateRetainedEvidencePath(
      root,
      reference?.path,
      failures,
      evidenceContext,
    );
    if (evidencePath === null) continue;
    let sourceBytes;
    try {
      sourceBytes = evidenceDocumentBytes(
        root,
        evidencePath,
        evidenceDocuments,
      );
    } catch (error) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${evidenceContext}: cannot read attestation: ${error.message}`,
      );
      continue;
    }
    validateRawEvidenceMarkers(sourceBytes, failures, evidenceContext);
    if (
      !CONTENT_SHA256_PATTERN.test(reference?.payload_sha256 ?? "") ||
      evidencePayloadSHA256(sourceBytes) !== reference.payload_sha256
    ) {
      addFailure(
        failures,
        "invalid-evidence-payload",
        `${evidenceContext}: non-attestation result payload does not match the matrix commitment`,
      );
    }
    let source;
    try {
      source = decodeEvidenceUTF8(sourceBytes);
    } catch (error) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${evidenceContext}: cannot decode attestation as UTF-8: ${error.message}`,
      );
      continue;
    }
    const attestations = parseEvidenceAttestations(
      source,
      failures,
      evidenceContext,
    );
    const expectedOwners =
      retainedEvidenceOwners.get(evidencePath) ?? new Set();
    const actualOwners = attestations.map(
      (attestation) => attestation.verification_set,
    );
    if (
      actualOwners.length !== expectedOwners.size ||
      new Set(actualOwners).size !== actualOwners.length ||
      actualOwners.some((owner) => !expectedOwners.has(owner))
    ) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${evidenceContext}: attestation markers must match every and only the retained verification sets that cite this artifact`,
      );
    }
    const matching = attestations.filter(
      (attestation) => attestation.verification_set === verification.id,
    );
    if (matching.length !== 1) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${evidenceContext}: expected exactly one attestation for ${verification.id}, found ${matching.length}`,
      );
      continue;
    }
    const attestedDimensions = validateEvidenceAttestation(
      root,
      matching[0],
      verification,
      profile,
      compatibilityBaseline,
      dimensions,
      currentSourceDigest,
      currentSourceManifest,
      testedRevisionManifest,
      retainedClaimDigest,
      traceabilityPath,
      evidenceDocuments,
      failures,
      evidenceContext,
    );
    for (const dimension of attestedDimensions) {
      allAttestedDimensions.add(dimension);
    }
  }

  const requiredDimensions = requiredAttestationDimensions(
    verification,
    consumers,
    applicabilitySets,
    rowApplicability,
    dimensions,
  );
  if (
    requiredDimensions.size === 0 &&
    verification.classification !== "extension"
  ) {
    addFailure(
      failures,
      "invalid-evidence-attestation",
      `${context}: retained verification has no declared compatibility, extension, or project-support boundary`,
    );
  }
  const extensionDimensions = new Set(
    [...EXTENSION_AXIS_DIMENSIONS.entries()]
      .filter(([axis]) =>
        ["partial", "complete"].includes(
          verification.extension_verification?.[axis],
        ),
      )
      .flatMap(([, axisDimensions]) => axisDimensions),
  );
  for (const required of requiredDimensions) {
    if (!allAttestedDimensions.has(required)) {
      addFailure(
        failures,
        "invalid-evidence-attestation",
        `${context}: retained claim is missing required ${required} evidence`,
      );
    }
  }
  for (const dimension of allAttestedDimensions) {
    if (
      verification.classification === "extension" &&
      !extensionDimensions.has(dimension)
    ) {
      addFailure(
        failures,
        "evidence-boundary-mismatch",
        `${context}: attested ${dimension} is not an identity-extension proof dimension`,
      );
      continue;
    }
    const relevantConsumers = consumers.filter(
      (row) =>
        applicabilityForRow(row, dimension, applicabilitySets, rowApplicability)
          ?.status === "applicable",
    );
    if (relevantConsumers.length === 0) {
      addFailure(
        failures,
        "evidence-boundary-mismatch",
        `${context}: attested ${dimension} has no applicable consumer row`,
      );
      continue;
    }
    const uncovered = relevantConsumers.filter((row) => {
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      return proofSet?.proofs?.[dimension]?.status !== "covered";
    });
    if (uncovered.length > 0) {
      addFailure(
        failures,
        "evidence-boundary-mismatch",
        `${context}: attested ${dimension} is not covered for ${uncovered.map((row) => row.id).join(", ")}`,
      );
    }
  }
}

function validateVerificationSet(root, verification, profile, failures) {
  const context = `verification set ${verification.id ?? "<missing>"}`;
  if (!CLASSIFICATIONS.has(verification.classification)) {
    addFailure(
      failures,
      "invalid-verification",
      `${context}: invalid classification`,
    );
    return;
  }
  if (!["pending", "retained"].includes(verification.state)) {
    addFailure(
      failures,
      "invalid-verification",
      `${context}: invalid state ${verification.state ?? "<missing>"}`,
    );
  }
  const evidence = verification.evidence ?? [];
  validateEvidenceReferenceList(
    root,
    evidence,
    failures,
    `${context} evidence`,
    {
      required: verification.state === "retained",
    },
  );
  if (
    verification.classification === "extension" ||
    verification.classification === "project"
  ) {
    if (verification.tier !== "not_applicable") {
      addFailure(
        failures,
        verification.classification === "extension"
          ? "extension-tier-claim"
          : "project-tier-claim",
        `${context}: ${verification.classification} rows must use tier not_applicable`,
      );
    }
  }
  if (
    verification.classification !== "extension" &&
    verification.extension_verification !== undefined
  ) {
    addFailure(
      failures,
      "invalid-verification",
      `${context}: extension_verification is valid only for extension sets`,
    );
  }
  if (verification.classification === "extension") {
    if (verification.state !== "pending" && verification.state !== "retained") {
      addFailure(
        failures,
        "invalid-verification",
        `${context}: extension verification must be pending or retained`,
      );
    }
    for (const key of ["contract", "parity", "security"]) {
      if (
        verification.extension_verification?.[key] !==
        profile.identity_extension?.verification?.[key]
      ) {
        addFailure(
          failures,
          "extension-verification-drift",
          `${context}: ${key} does not match the active profile`,
        );
      }
    }
  } else if (
    verification.classification === "baseline" &&
    !TIER_RANK.has(verification.tier)
  ) {
    addFailure(
      failures,
      "invalid-verification",
      `${context}: baseline verification needs a T0-T4 tier`,
    );
  }
  if (verification.state !== "retained") {
    if (
      typeof verification.pending_reason !== "string" ||
      verification.pending_reason.length === 0
    ) {
      addFailure(
        failures,
        "invalid-verification",
        `${context}: non-retained verification requires a pending_reason`,
      );
    }
    if (evidence.length > 0) {
      addFailure(
        failures,
        "invalid-verification",
        `${context}: non-retained verification cannot cite evidence`,
      );
    }
    if (verification.tested_digest !== undefined) {
      addFailure(
        failures,
        "invalid-evidence-source-digest",
        `${context}: non-retained verification cannot claim a tested digest`,
      );
    }
    if (verification.tested_revision !== undefined) {
      addFailure(
        failures,
        "invalid-tested-revision",
        `${context}: non-retained verification cannot claim a tested revision`,
      );
    }
    if (verification.tested_source_manifest !== undefined) {
      addFailure(
        failures,
        "invalid-tested-source-manifest",
        `${context}: non-retained verification cannot commit a tested source manifest`,
      );
    }
  } else {
    if (verification.pending_reason !== undefined) {
      addFailure(
        failures,
        "invalid-verification",
        `${context}: retained verification cannot carry a stale pending_reason`,
      );
    }
    if (!SOURCE_DIGEST_PATTERN.test(verification.tested_digest ?? "")) {
      addFailure(
        failures,
        "invalid-evidence-source-digest",
        `${context}: retained verification requires tested_digest in source-v2:sha256:<64 lowercase hex> form`,
      );
    }
    if (!/^[0-9a-f]{40}$/u.test(verification.tested_revision ?? "")) {
      addFailure(
        failures,
        "invalid-tested-revision",
        `${context}: retained verification requires a full lowercase tested Git revision`,
      );
    }
    for (const [index, reference] of evidence.entries()) {
      validateRetainedEvidencePath(
        root,
        reference?.path,
        failures,
        `${context} evidence[${index}]`,
      );
    }
    if (
      verification.classification === "baseline" &&
      TIER_RANK.get(verification.tier) < TIER_RANK.get("T2")
    ) {
      addFailure(
        failures,
        "invalid-verification",
        `${context}: retained baseline verification must claim T2 or higher`,
      );
    }
  }
  if (TIER_RANK.get(verification.tier) >= TIER_RANK.get("T2")) {
    if (verification.state !== "retained" || evidence.length === 0) {
      addFailure(
        failures,
        "tier-inflation",
        `${context}: T2+ requires retained evidence`,
      );
    }
  }
}

function expectedOperationSources(profile) {
  const expected = new Map();
  for (const resource of profile?.resources ?? []) {
    const capability = `${resource.module}.${resource.name}`;
    for (const operation of profile.operations ?? []) {
      expected.set(`${capability}.${operation}`, {
        capability,
        operation,
      });
    }
    for (const action of resource.extra_rpcs ?? []) {
      const operation =
        action === "AssignIPAddress"
          ? "assign"
          : action === "UnassignIPAddress"
            ? "unassign"
            : action;
      expected.set(`${capability}.${action}`, { capability, operation });
    }
  }
  return expected;
}

function affectedCapabilities(row) {
  return new Set(row.affected_capabilities ?? [row.capability]);
}

function applicabilityForRow(
  row,
  dimension,
  applicabilitySets,
  rowApplicability,
) {
  const set = applicabilitySets.get(rowApplicability?.[row.id]);
  return set?.dimensions?.[dimension];
}

function boundaryProofSatisfied(
  row,
  proofSet,
  dimension,
  applicabilitySets,
  rowApplicability,
) {
  const applicability = applicabilityForRow(
    row,
    dimension,
    applicabilitySets,
    rowApplicability,
  );
  if (applicability?.status === "not_applicable") return true;
  return (
    applicability?.status === "applicable" &&
    proofSet?.proofs?.[dimension]?.status === "covered"
  );
}

function validateRows({
  root,
  rows,
  referenceSets,
  applicabilitySets,
  rowApplicability,
  proofSets,
  rowProofs,
  verificationSets,
  assessmentSets,
  profile,
  compatibilityBaseline,
  proofDimensions,
  evidenceDocuments,
  currentSourceDigest,
  currentSourceManifest,
  testedRevisionManifests,
  retainedClaimDigests,
  traceabilityPath,
  scenarioIDs,
  planRuleIDs,
  operationCatalog,
  failures,
}) {
  const rowIndex = indexUnique(
    rows,
    (row) => row?.id,
    failures,
    "duplicate-row",
    "rows",
  );
  const profileResources = profileResourceMap(profile);
  const expectedOperations = expectedOperationSources(profile);
  const sourceRows = new Map();
  const proofSetConsumers = new Map();
  const applicabilitySetConsumers = new Map();
  const proofSetApplicability = new Map();
  const verificationSetConsumers = new Map();
  const assessmentSetConsumers = new Map();

  for (const row of rowIndex.values()) {
    const context = `row ${row.id}`;
    const proofSetID = rowProofs?.[row.id];
    const proofSet = proofSets.get(proofSetID);
    if (!proofSet) {
      addFailure(
        failures,
        "missing-row-proof",
        `${context}: unknown or missing proof set ${proofSetID ?? "<missing>"}`,
      );
    } else {
      const consumers = proofSetConsumers.get(proofSetID) ?? [];
      consumers.push(row.id);
      proofSetConsumers.set(proofSetID, consumers);
    }
    const applicabilitySetID = rowApplicability?.[row.id];
    const applicabilitySet = applicabilitySets.get(applicabilitySetID);
    if (!applicabilitySet) {
      addFailure(
        failures,
        "missing-row-applicability",
        `${context}: unknown or missing applicability set ${applicabilitySetID ?? "<missing>"}`,
      );
    } else {
      const consumers = applicabilitySetConsumers.get(applicabilitySetID) ?? [];
      consumers.push(row.id);
      applicabilitySetConsumers.set(applicabilitySetID, consumers);
      if (proofSet) {
        const selected = proofSetApplicability.get(proofSetID) ?? new Set();
        selected.add(applicabilitySetID);
        proofSetApplicability.set(proofSetID, selected);
        for (const dimension of Object.keys(applicabilitySet.dimensions)) {
          const applicability = applicabilitySet.dimensions[dimension];
          const proof = proofSet.proofs?.[dimension];
          const matches =
            applicability.status === "applicable"
              ? proof?.status !== "not_applicable"
              : proof?.status === "not_applicable" &&
                proof.reason === applicability.reason &&
                (proof.references?.length ?? 0) === 0;
          if (!matches) {
            addFailure(
              failures,
              "proof-applicability-mismatch",
              `${context}: ${dimension} proof does not match reviewed applicability set ${applicabilitySetID}`,
            );
          }
        }
      }
    }
    if (!ROW_KINDS.has(row.kind)) {
      addFailure(
        failures,
        "invalid-row",
        `${context}: invalid kind ${row.kind ?? "<missing>"}`,
      );
      continue;
    }
    if (
      typeof row.source_id !== "string" ||
      typeof row.capability !== "string" ||
      typeof row.operation !== "string"
    ) {
      addFailure(
        failures,
        "invalid-row",
        `${context}: source_id, capability, and operation are required`,
      );
      continue;
    }
    if (!operationCatalog.has(row.operation)) {
      addFailure(
        failures,
        "unknown-operation",
        `${context}: operation ${row.operation} is not in the matrix catalogue`,
      );
    }
    const affected = affectedCapabilities(row);
    if (!affected.has(row.capability)) {
      addFailure(
        failures,
        "invalid-affected-capability",
        `${context}: affected_capabilities must include primary capability ${row.capability}`,
      );
    }
    const allowedCapabilities = new Set([
      "profile",
      "contract",
      "identity",
      ...profileResources.keys(),
    ]);
    for (const capability of affected) {
      if (!allowedCapabilities.has(capability)) {
        addFailure(
          failures,
          "invalid-affected-capability",
          `${context}: unknown affected capability ${capability}`,
        );
      }
    }
    const sourceKey = `${row.kind}\u0000${row.source_id}`;
    if (sourceRows.has(sourceKey)) {
      addFailure(
        failures,
        "duplicate-source-row",
        `${context}: duplicate ${row.kind} source ${row.source_id}`,
      );
    }
    sourceRows.set(sourceKey, row);

    const referenceSet = referenceSets.get(row.reference_set);
    if (!referenceSet) {
      addFailure(
        failures,
        "invalid-row-reference-set",
        `${context}: unknown reference set ${row.reference_set ?? "<missing>"}`,
      );
    } else if (!referenceSet.capabilities?.includes(row.capability)) {
      addFailure(
        failures,
        "invalid-row-reference-set",
        `${context}: ${row.capability} is not owned by ${row.reference_set}`,
      );
    }
    const relatedReferenceSetIDs = new Set(
      row.related_reference_sets ?? [row.reference_set],
    );
    if (!relatedReferenceSetIDs.has(row.reference_set)) {
      addFailure(
        failures,
        "invalid-related-reference-set",
        `${context}: related_reference_sets must include primary set ${row.reference_set}`,
      );
    }
    const relatedReferenceSets = [];
    for (const referenceSetID of relatedReferenceSetIDs) {
      const related = referenceSets.get(referenceSetID);
      if (!related) {
        addFailure(
          failures,
          "invalid-related-reference-set",
          `${context}: unknown related reference set ${referenceSetID}`,
        );
      } else {
        relatedReferenceSets.push(related);
      }
    }
    for (const capability of affected) {
      if (
        !relatedReferenceSets.some((related) =>
          related.capabilities?.includes(capability),
        )
      ) {
        addFailure(
          failures,
          "missing-affected-reference-set",
          `${context}: no related reference set owns ${capability}`,
        );
      }
    }
    for (const related of relatedReferenceSets) {
      if (
        !related.capabilities?.some((capability) => affected.has(capability))
      ) {
        addFailure(
          failures,
          "invalid-related-reference-set",
          `${context}: related reference set ${related.id} owns no affected capability`,
        );
      }
    }
    if (proofSet) {
      for (const [dimension, proof] of Object.entries(proofSet.proofs ?? {})) {
        if (!["covered", "partial"].includes(proof.status)) continue;
        const allowedReferences = new Set(
          relatedReferenceSets.flatMap((related) =>
            (related.test_inventory?.[dimension]?.references ?? []).map(
              (reference) => JSON.stringify([reference.path, reference.anchor]),
            ),
          ),
        );
        for (const reference of proof.references ?? []) {
          const key = JSON.stringify([reference.path, reference.anchor]);
          if (!allowedReferences.has(key)) {
            addFailure(
              failures,
              "unowned-proof-reference",
              `${context}: ${dimension} proof reference is not declared by an affected capability inventory`,
            );
          }
        }
      }
    }

    const verification = verificationSets.get(row.verification_set);
    if (!verification) {
      addFailure(
        failures,
        "invalid-row-verification-set",
        `${context}: unknown verification set ${row.verification_set ?? "<missing>"}`,
      );
    } else {
      const consumers = verificationSetConsumers.get(verification.id) ?? [];
      consumers.push(row);
      verificationSetConsumers.set(verification.id, consumers);
      if (
        referenceSet &&
        verification.classification !== referenceSet.classification
      ) {
        addFailure(
          failures,
          "classification-mismatch",
          `${context}: reference and verification classifications differ`,
        );
      }
    }
    const assessment = assessmentSets.get(row.assessment_set);
    if (!assessment) {
      addFailure(
        failures,
        "invalid-row-assessment-set",
        `${context}: unknown assessment set ${row.assessment_set ?? "<missing>"}`,
      );
    } else {
      const consumers = assessmentSetConsumers.get(assessment.id) ?? [];
      consumers.push(row);
      assessmentSetConsumers.set(assessment.id, consumers);
      if (
        assessment.status === "contradicted" &&
        verification &&
        (verification.state === "retained" ||
          TIER_RANK.get(verification.tier) > TIER_RANK.get("T1"))
      ) {
        addFailure(
          failures,
          "contradicted-tier-claim",
          `${context}: contradicted behavior cannot be retained or exceed T1`,
        );
      }
    }
    if (
      assessment &&
      (assessment.status === "contradicted" ||
        assessment.status === "unresolved") &&
      proofSet
    ) {
      for (const [dimension, proof] of Object.entries(proofSet.proofs ?? {})) {
        if (proof.status === "covered") {
          addFailure(
            failures,
            "unresolved-covered-proof",
            `${context}: ${assessment.status} row cannot inherit covered ${dimension} proof`,
          );
        }
      }
    }

    if (row.covered_operations !== undefined) {
      if (
        !Array.isArray(row.covered_operations) ||
        row.covered_operations.length === 0
      ) {
        addFailure(
          failures,
          "invalid-covered-operations",
          `${context}: covered_operations must be a non-empty array`,
        );
      } else {
        for (const operation of row.covered_operations) {
          if (!operationCatalog.has(operation)) {
            addFailure(
              failures,
              "unknown-covered-operation",
              `${context}: covered operation ${operation} is not in the matrix catalogue`,
            );
          }
        }
      }
    }

    if (row.kind === "scenario") {
      if (!scenarioIDs.has(row.source_id) || row.operation !== "workflow") {
        addFailure(
          failures,
          "invalid-scenario-row",
          `${context}: invalid scenario source or operation`,
        );
      }
      if (
        !Array.isArray(row.covered_operations) ||
        row.covered_operations.length === 0
      ) {
        addFailure(
          failures,
          "invalid-scenario-row",
          `${context}: scenario needs covered_operations`,
        );
      }
    } else if (row.kind === "resource_operation") {
      const expected = expectedOperations.get(row.source_id);
      if (
        !expected ||
        expected.capability !== row.capability ||
        expected.operation !== row.operation
      ) {
        addFailure(
          failures,
          "invalid-operation-row",
          `${context}: operation does not match the active profile`,
        );
      }
      if (row.reference_set !== row.capability) {
        addFailure(
          failures,
          "invalid-operation-row",
          `${context}: resource operation must use its resource reference set`,
        );
      }
    } else if (row.kind === "resource_contract") {
      if (
        !profileResources.has(row.source_id) ||
        row.source_id !== row.capability ||
        row.operation !== "contract" ||
        row.reference_set !== row.capability
      ) {
        addFailure(
          failures,
          "invalid-resource-contract-row",
          `${context}: resource contract does not match the active profile`,
        );
      }
    } else if (row.kind === "plan_rule") {
      if (!planRuleIDs.has(row.source_id)) {
        addFailure(
          failures,
          "invalid-plan-rule-row",
          `${context}: unknown plan rule ${row.source_id}`,
        );
      }
    }

    if (
      verification &&
      verification.classification === "baseline" &&
      TIER_RANK.has(verification.tier)
    ) {
      const owner = profileResources.get(row.capability);
      if (
        owner &&
        TIER_RANK.get(verification.tier) > TIER_RANK.get(owner.tier)
      ) {
        addFailure(
          failures,
          "tier-inflation",
          `${context}: ${verification.tier} exceeds owner ${row.capability} profile tier ${owner.tier}`,
        );
      }
      const requiredProofs = [];
      if (TIER_RANK.get(verification.tier) >= TIER_RANK.get("T2")) {
        requiredProofs.push("rest_differential");
      }
      if (TIER_RANK.get(verification.tier) >= TIER_RANK.get("T3")) {
        requiredProofs.push("grpc_parity");
      }
      if (TIER_RANK.get(verification.tier) >= TIER_RANK.get("T4")) {
        requiredProofs.push("browser");
      }
      for (const dimension of requiredProofs) {
        if (
          !boundaryProofSatisfied(
            row,
            proofSet,
            dimension,
            applicabilitySets,
            rowApplicability,
          )
        ) {
          addFailure(
            failures,
            "tier-boundary-missing",
            `${context}: ${verification.tier} requires covered ${dimension} proof`,
          );
        }
      }
    }
  }

  for (const scenarioID of scenarioIDs) {
    if (!sourceRows.has(`scenario\u0000${scenarioID}`)) {
      addFailure(
        failures,
        "missing-scenario-row",
        `missing traceability row for scenario ${scenarioID}`,
      );
    }
  }
  for (const sourceID of expectedOperations.keys()) {
    if (!sourceRows.has(`resource_operation\u0000${sourceID}`)) {
      addFailure(
        failures,
        "missing-operation-row",
        `missing traceability row for operation ${sourceID}`,
      );
    }
  }
  for (const resourceKey of profileResources.keys()) {
    if (!sourceRows.has(`resource_contract\u0000${resourceKey}`)) {
      addFailure(
        failures,
        "missing-resource-contract-row",
        `missing traceability row for resource contract ${resourceKey}`,
      );
    }
  }
  for (const ruleID of planRuleIDs) {
    if (!sourceRows.has(`plan_rule\u0000${ruleID}`)) {
      addFailure(
        failures,
        "missing-plan-rule-row",
        `missing traceability row for plan rule ${ruleID}`,
      );
    }
  }

  for (const row of sourceRows.values()) {
    if (row.kind === "scenario" && !scenarioIDs.has(row.source_id)) {
      addFailure(
        failures,
        "extra-source-row",
        `extra scenario row ${row.source_id}`,
      );
    } else if (
      row.kind === "resource_operation" &&
      !expectedOperations.has(row.source_id)
    ) {
      addFailure(
        failures,
        "extra-source-row",
        `extra operation row ${row.source_id}`,
      );
    } else if (
      row.kind === "resource_contract" &&
      !profileResources.has(row.source_id)
    ) {
      addFailure(
        failures,
        "extra-source-row",
        `extra resource contract row ${row.source_id}`,
      );
    } else if (row.kind === "plan_rule" && !planRuleIDs.has(row.source_id)) {
      addFailure(
        failures,
        "extra-source-row",
        `extra plan rule row ${row.source_id}`,
      );
    }
  }

  for (const rowID of rowIndex.keys()) {
    if (!Object.hasOwn(rowProofs ?? {}, rowID)) {
      addFailure(
        failures,
        "missing-row-proof",
        `missing proof mapping for row ${rowID}`,
      );
    }
  }
  for (const rowID of Object.keys(rowProofs ?? {})) {
    if (!rowIndex.has(rowID)) {
      addFailure(
        failures,
        "extra-row-proof",
        `proof mapping names unknown row ${rowID}`,
      );
    }
  }
  for (const rowID of rowIndex.keys()) {
    if (!Object.hasOwn(rowApplicability ?? {}, rowID)) {
      addFailure(
        failures,
        "missing-row-applicability",
        `missing applicability mapping for row ${rowID}`,
      );
    }
  }
  for (const rowID of Object.keys(rowApplicability ?? {})) {
    if (!rowIndex.has(rowID)) {
      addFailure(
        failures,
        "extra-row-applicability",
        `applicability mapping names unknown row ${rowID}`,
      );
    }
  }
  for (const [proofSetID, proofSet] of proofSets) {
    const consumers = proofSetConsumers.get(proofSetID) ?? [];
    if (consumers.length === 0) {
      addFailure(
        failures,
        "unused-proof-set",
        `proof set ${proofSetID} is not selected by any row`,
      );
      continue;
    }
    const hasLiveProof = Object.values(proofSet.proofs ?? {}).some(
      (proof) => proof.status === "partial" || proof.status === "covered",
    );
    if (hasLiveProof && consumers.length !== 1) {
      addFailure(
        failures,
        "shared-live-proof-set",
        `proof set ${proofSetID} has live references but is shared by ${consumers.length} rows`,
      );
    }
    const selectedApplicability = proofSetApplicability.get(proofSetID);
    if (selectedApplicability && selectedApplicability.size > 1) {
      addFailure(
        failures,
        "shared-proof-applicability",
        `proof set ${proofSetID} is shared across reviewed applicability sets ${[...selectedApplicability].join(", ")}`,
      );
    }
  }
  for (const applicabilitySetID of applicabilitySets.keys()) {
    if (
      (applicabilitySetConsumers.get(applicabilitySetID) ?? []).length === 0
    ) {
      addFailure(
        failures,
        "unused-applicability-set",
        `applicability set ${applicabilitySetID} is not selected by any row`,
      );
    }
  }
  const retainedEvidenceOwners = new Map();
  const retainedEvidencePayloads = new Map();
  for (const verification of verificationSets.values()) {
    if (verification.state !== "retained") continue;
    for (const reference of verification.evidence ?? []) {
      const evidencePath = canonicalRepositoryPath(reference?.path);
      if (evidencePath === null) continue;
      const owners = retainedEvidenceOwners.get(evidencePath) ?? new Set();
      owners.add(verification.id);
      retainedEvidenceOwners.set(evidencePath, owners);
      const priorPayload = retainedEvidencePayloads.get(evidencePath);
      if (
        priorPayload !== undefined &&
        priorPayload !== reference.payload_sha256
      ) {
        addFailure(
          failures,
          "invalid-evidence-payload",
          `${evidencePath}: retained verification sets commit different result payload hashes`,
        );
      } else {
        retainedEvidencePayloads.set(evidencePath, reference.payload_sha256);
      }
    }
  }
  for (const [verificationSetID, verification] of verificationSets) {
    const consumers = verificationSetConsumers.get(verificationSetID) ?? [];
    if (consumers.length === 0) {
      addFailure(
        failures,
        "unused-verification-set",
        `verification set ${verificationSetID} is not selected by any row`,
      );
      continue;
    }
    if (verification.state === "retained") {
      validateRetainedEvidenceAttestations(
        root,
        verification,
        verification.evidence ?? [],
        profile,
        compatibilityBaseline,
        proofDimensions,
        evidenceDocuments,
        consumers,
        applicabilitySets,
        rowApplicability,
        proofSets,
        rowProofs,
        currentSourceDigest,
        currentSourceManifest,
        testedRevisionManifests.get(verification.tested_revision) ?? null,
        retainedClaimDigests.get(verification.id) ?? null,
        traceabilityPath,
        retainedEvidenceOwners,
        failures,
      );
    }
  }
  for (const assessmentSetID of assessmentSets.keys()) {
    if ((assessmentSetConsumers.get(assessmentSetID) ?? []).length === 0) {
      addFailure(
        failures,
        "unused-assessment-set",
        `assessment set ${assessmentSetID} is not selected by any row`,
      );
    }
  }

  const promotedBaselineRank = [...rowIndex.values()].reduce((rank, row) => {
    const verification = verificationSets.get(row.verification_set);
    if (verification?.classification !== "baseline") return rank;
    return Math.max(rank, TIER_RANK.get(verification.tier) ?? -1);
  }, -1);
  let promotedProfileRank = -1;
  for (const [resourceKey, resource] of profileResources) {
    const resourceRank = TIER_RANK.get(resource.tier) ?? -1;
    promotedProfileRank = Math.max(promotedProfileRank, resourceRank);
    if (resourceRank < TIER_RANK.get("T2")) continue;
    const requiredDimensions = ["rest_differential"];
    if (resourceRank >= TIER_RANK.get("T3")) {
      requiredDimensions.push("grpc_parity");
    }
    if (resourceRank >= TIER_RANK.get("T4")) {
      requiredDimensions.push("browser");
    }
    const resourceRows = [...rowIndex.values()].filter(
      (row) => row.capability === resourceKey,
    );
    const vacuousDimensions = requiredDimensions.filter(
      (dimension) =>
        !resourceRows.some(
          (row) =>
            applicabilityForRow(
              row,
              dimension,
              applicabilitySets,
              rowApplicability,
            )?.status === "applicable",
        ),
    );
    const blockers = resourceRows.filter((row) => {
      const verification = verificationSets.get(row.verification_set);
      const assessment = assessmentSets.get(row.assessment_set);
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      const boundaryMissing = requiredDimensions.some((dimension) => {
        return !boundaryProofSatisfied(
          row,
          proofSet,
          dimension,
          applicabilitySets,
          rowApplicability,
        );
      });
      return (
        verification?.classification !== "baseline" ||
        TIER_RANK.get(verification.tier) !== resourceRank ||
        verification?.state !== "retained" ||
        !["confirmed", "not_applicable"].includes(assessment?.status) ||
        boundaryMissing
      );
    });
    if (
      resourceRows.length === 0 ||
      vacuousDimensions.length > 0 ||
      blockers.length > 0
    ) {
      addFailure(
        failures,
        "profile-tier-unproven",
        `${resourceKey} ${resource.tier} is not closed by all ${resourceRows.length} traceability rows; vacuous boundaries: ${vacuousDimensions.join(", ") || "none"}; blockers: ${blockers.map((row) => row.id).join(", ") || "none"}`,
      );
    }
  }

  const claimedBoundaryRank = Math.max(
    promotedBaselineRank,
    promotedProfileRank,
  );
  if (claimedBoundaryRank >= TIER_RANK.get("T2")) {
    const supportDimensions = ["rest_differential"];
    if (claimedBoundaryRank >= TIER_RANK.get("T3")) {
      supportDimensions.push("grpc_parity");
    }
    if (claimedBoundaryRank >= TIER_RANK.get("T4")) {
      supportDimensions.push("browser");
    }
    const supportingBlockers = [...rowIndex.values()].filter((row) => {
      const verification = verificationSets.get(row.verification_set);
      if (verification?.classification !== "project") return false;
      const assessment = assessmentSets.get(row.assessment_set);
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      const relevantDimensions = supportDimensions.filter(
        (dimension) =>
          applicabilityForRow(
            row,
            dimension,
            applicabilitySets,
            rowApplicability,
          )?.status === "applicable",
      );
      if (relevantDimensions.length === 0) return false;
      const proofBlocked = relevantDimensions.some(
        (dimension) => proofSet?.proofs?.[dimension]?.status !== "covered",
      );
      return (
        verification.state !== "retained" ||
        !["confirmed", "not_applicable"].includes(assessment?.status) ||
        proofBlocked
      );
    });
    if (supportingBlockers.length > 0) {
      addFailure(
        failures,
        "tier-support-pending",
        `T2+ is blocked by ${supportingBlockers.length} unresolved project support rows: ${supportingBlockers.map((row) => row.id).join(", ")}`,
      );
    }
  }

  const extensionRows = [...rowIndex.values()].filter((row) => {
    const verification = verificationSets.get(row.verification_set);
    return verification?.classification === "extension";
  });
  const completedAxes = [...EXTENSION_AXIS_DIMENSIONS].filter(
    ([axis]) => profile.identity_extension?.verification?.[axis] === "complete",
  );
  for (const [axis, dimensions] of completedAxes) {
    for (const dimension of dimensions) {
      if (
        !extensionRows.some(
          (row) =>
            applicabilityForRow(
              row,
              dimension,
              applicabilitySets,
              rowApplicability,
            )?.status === "applicable",
        )
      ) {
        addFailure(
          failures,
          "extension-applicability-missing",
          `complete identity ${axis} verification has no applicable ${dimension} row`,
        );
      }
    }
    for (const row of extensionRows) {
      const applicableDimensions = dimensions.filter(
        (dimension) =>
          applicabilityForRow(
            row,
            dimension,
            applicabilitySets,
            rowApplicability,
          )?.status === "applicable",
      );
      if (applicableDimensions.length === 0) continue;
      const verification = verificationSets.get(row.verification_set);
      const assessment = assessmentSets.get(row.assessment_set);
      if (
        verification?.state !== "retained" ||
        !["confirmed", "not_applicable"].includes(assessment?.status)
      ) {
        addFailure(
          failures,
          "extension-verification-pending",
          `${row.id}: complete identity ${axis} verification requires retained verification and a closed assessment`,
        );
      }
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      for (const dimension of applicableDimensions) {
        if (proofSet?.proofs?.[dimension]?.status !== "covered") {
          addFailure(
            failures,
            "extension-boundary-missing",
            `${row.id}: complete identity ${axis} verification requires covered ${dimension} proof`,
          );
        }
      }
    }
  }

  const identityVerificationComplete = completedAxes.length === 3;
  if (identityVerificationComplete) {
    for (const row of extensionRows) {
      const verification = verificationSets.get(row.verification_set);
      const assessment = assessmentSets.get(row.assessment_set);
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      const applicableDimensions = Object.keys(
        applicabilitySets.get(rowApplicability?.[row.id])?.dimensions ?? {},
      ).filter(
        (dimension) =>
          applicabilityForRow(
            row,
            dimension,
            applicabilitySets,
            rowApplicability,
          )?.status === "applicable",
      );
      if (
        verification?.state !== "retained" ||
        !["confirmed", "not_applicable"].includes(assessment?.status) ||
        applicableDimensions.some(
          (dimension) => proofSet?.proofs?.[dimension]?.status !== "covered",
        )
      ) {
        addFailure(
          failures,
          "extension-verification-pending",
          `${row.id}: complete identity verification requires retained, closed proof for every applicable dimension`,
        );
      }
    }
    const supportBlockers = [...rowIndex.values()].filter((row) => {
      if (!affectedCapabilities(row).has("identity")) return false;
      const verification = verificationSets.get(row.verification_set);
      if (verification?.classification !== "project") return false;
      const assessment = assessmentSets.get(row.assessment_set);
      const proofSet = proofSets.get(rowProofs?.[row.id]);
      const proofBlocked = Object.keys(
        applicabilitySets.get(rowApplicability?.[row.id])?.dimensions ?? {},
      ).some(
        (dimension) =>
          applicabilityForRow(
            row,
            dimension,
            applicabilitySets,
            rowApplicability,
          )?.status === "applicable" &&
          proofSet?.proofs?.[dimension]?.status !== "covered",
      );
      return (
        verification?.state !== "retained" ||
        !["confirmed", "not_applicable"].includes(assessment?.status) ||
        proofBlocked
      );
    });
    if (supportBlockers.length > 0) {
      addFailure(
        failures,
        "extension-support-pending",
        `complete identity verification is blocked by ${supportBlockers.length} project-support rows: ${supportBlockers.map((row) => row.id).join(", ")}`,
      );
    }
  }
}

export function validateTraceability({
  root = SCRIPT_ROOT,
  profilePath,
  profileDocument,
  identityMetadataDocument,
  traceabilityDocument,
  evidenceDocuments,
} = {}) {
  const failures = [];
  const resolvedProfilePath = path.resolve(
    root,
    profilePath ??
      "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
  );
  const profile =
    profileDocument ?? parseJSON(resolvedProfilePath, failures, "profile-read");
  if (!profile) return { failures, counts: {} };
  const contractRoot = path.dirname(path.dirname(resolvedProfilePath));
  const traceabilityPath = path.resolve(
    path.dirname(resolvedProfilePath),
    profile.traceability ?? "",
  );
  const document =
    traceabilityDocument ??
    parseJSON(traceabilityPath, failures, "traceability-read");
  if (!document) return { failures, counts: {} };

  const expectedSchemaReference = "../schema/traceability.schema.json";
  const traceabilitySchema = parseJSON(
    path.resolve(path.dirname(traceabilityPath), expectedSchemaReference),
    failures,
    "traceability-schema-read",
  );
  if (document.$schema !== expectedSchemaReference) {
    addFailure(
      failures,
      "traceability-schema-mismatch",
      `traceability must use ${expectedSchemaReference}`,
    );
  }
  if (traceabilitySchema) {
    if (
      traceabilitySchema.$id !==
      "https://netbox-go.local/schemas/traceability-v1.json"
    ) {
      addFailure(
        failures,
        "schema-definition-invalid",
        "traceability schema has an unexpected $id",
      );
    }
    validateJSONSchema(
      document,
      traceabilitySchema,
      traceabilitySchema,
      failures,
    );
  }
  const applicabilityDigest = reviewedApplicabilityDigest(document);
  if (applicabilityDigest !== REVIEWED_APPLICABILITY_SHA256) {
    addFailure(
      failures,
      "reviewed-applicability-drift",
      `applicability authority digest ${applicabilityDigest} does not match the reviewed 293-row mapping`,
    );
  }
  const referenceAuthorityDigest = reviewedReferenceAuthorityDigest(document);
  if (referenceAuthorityDigest !== REVIEWED_REFERENCE_AUTHORITY_SHA256) {
    addFailure(
      failures,
      "reviewed-reference-authority-drift",
      `reference authority digest ${referenceAuthorityDigest} does not match the reviewed capability ownership/source/test/inventory mapping`,
    );
  }
  const rowSemanticsDigest = reviewedRowSemanticsDigest(document);
  if (rowSemanticsDigest !== REVIEWED_ROW_SEMANTICS_SHA256) {
    addFailure(
      failures,
      "reviewed-row-semantics-drift",
      `row semantics digest ${rowSemanticsDigest} does not match the reviewed scenario/operation/rule bindings`,
    );
  }

  const baseline = parseJSON(
    path.join(contractRoot, "baseline.yaml"),
    failures,
    "baseline-read",
  );
  if (
    document.schema_version !== 1 ||
    document.profile !== profile.id ||
    document.compatibility_baseline?.id !== baseline?.id ||
    document.compatibility_baseline?.git_sha !== baseline?.git_sha
  ) {
    addFailure(
      failures,
      "traceability-authority-mismatch",
      "traceability profile/baseline/SHA does not match the active contract",
    );
  }
  if (
    typeof document.implementation_plan !== "string" ||
    document.implementation_plan.length === 0
  ) {
    addFailure(
      failures,
      "traceability-authority-mismatch",
      "traceability must name the implementation plan",
    );
  }
  const dimensions = document.proof_dimensions;
  const expectedDimensions = [
    "domain",
    "application",
    "postgresql",
    "rest_differential",
    "grpc_parity",
    "browser",
    "rest_extension_contract",
    "cli_security",
  ];
  if (
    !Array.isArray(dimensions) ||
    JSON.stringify(dimensions) !== JSON.stringify(expectedDimensions)
  ) {
    addFailure(
      failures,
      "invalid-proof-dimensions",
      `proof_dimensions must be exactly ${expectedDimensions.join(", ")}`,
    );
  }
  const operationCatalog = new Set(document.operation_catalog ?? []);
  if (
    !Array.isArray(document.operation_catalog) ||
    operationCatalog.size !== document.operation_catalog.length ||
    operationCatalog.size === 0
  ) {
    addFailure(
      failures,
      "invalid-operation-catalogue",
      "operation_catalog must be a non-empty unique array",
    );
  }
  for (const required of [
    ...(profile.operations ?? []),
    "assign",
    "unassign",
    "contract",
    "workflow",
  ]) {
    if (!operationCatalog.has(required)) {
      addFailure(
        failures,
        "invalid-operation-catalogue",
        `operation_catalog is missing ${required}`,
      );
    }
  }

  try {
    const actualSHA = readGitHead(path.join(root, "netbox"));
    if (actualSHA !== baseline?.git_sha) {
      addFailure(
        failures,
        "upstream-sha-mismatch",
        `pinned upstream checkout is ${actualSHA}, expected ${baseline?.git_sha}`,
      );
    }
  } catch (error) {
    addFailure(
      failures,
      "upstream-check-failed",
      `unable to verify pinned upstream checkout: ${error.message}`,
    );
  }

  const resources = profileResourceMap(profile);
  const metadata = loadResourceMetadata(
    root,
    resolvedProfilePath,
    profile,
    failures,
  );
  loadAndValidateIdentityMetadata(
    resolvedProfilePath,
    profile,
    identityMetadataDocument,
    failures,
  );
  const scenarioIDs = loadScenarioIDs(contractRoot, failures);
  const planRuleIDs = loadPlanRuleIDs(
    root,
    document.implementation_plan,
    failures,
  );

  const referenceSets = indexUnique(
    document.reference_sets,
    (item) => item?.id,
    failures,
    "duplicate-reference-set",
    "reference_sets",
  );
  for (const referenceSet of referenceSets.values()) {
    validateReferenceSet(
      root,
      referenceSet,
      expectedDimensions,
      profile,
      resolvedProfilePath,
      resources,
      metadata,
      failures,
    );
  }

  const applicabilitySets = indexUnique(
    document.applicability_sets,
    (item) => item?.id,
    failures,
    "duplicate-applicability-set",
    "applicability_sets",
  );
  for (const applicabilitySet of applicabilitySets.values()) {
    validateApplicabilitySet(applicabilitySet, expectedDimensions, failures);
  }

  for (const resourceKey of resources.keys()) {
    const referenceSet = referenceSets.get(resourceKey);
    if (!referenceSet) {
      addFailure(
        failures,
        "missing-resource-reference-set",
        `missing resource reference set ${resourceKey}`,
      );
    } else if (
      referenceSet.profile_ref?.kind !== "resource" ||
      referenceSet.profile_ref?.key !== resourceKey
    ) {
      addFailure(
        failures,
        "invalid-profile-selector",
        `${resourceKey} reference set must select its exact resource`,
      );
    }
    if (
      referenceSet &&
      (referenceSet.metadata_ref?.kind !== "resource" ||
        referenceSet.metadata_ref?.key !== resourceKey)
    ) {
      addFailure(
        failures,
        "invalid-metadata-selector",
        `${resourceKey} reference set must select its exact resource metadata`,
      );
    }
  }
  const expectedReferenceSets = new Set([
    "project",
    "identity",
    "identity-baseline-support",
    "identity-interface-support",
    "baseline-common-support",
    "deferred-baseline-support",
    ...resources.keys(),
  ]);
  const expectedSpecialReferenceSets = new Map([
    [
      "project",
      { classification: "project", capabilities: ["profile", "contract"] },
    ],
    ["identity", { classification: "extension", capabilities: ["identity"] }],
    [
      "identity-baseline-support",
      { classification: "project", capabilities: ["identity", "profile"] },
    ],
    [
      "identity-interface-support",
      { classification: "project", capabilities: ["identity"] },
    ],
    [
      "baseline-common-support",
      { classification: "project", capabilities: ["profile"] },
    ],
    [
      "deferred-baseline-support",
      { classification: "project", capabilities: ["profile"] },
    ],
  ]);
  for (const [referenceSetID, expected] of expectedSpecialReferenceSets) {
    const referenceSet = referenceSets.get(referenceSetID);
    if (!referenceSet) continue;
    if (
      referenceSet.classification !== expected.classification ||
      JSON.stringify(referenceSet.capabilities) !==
        JSON.stringify(expected.capabilities)
    ) {
      addFailure(
        failures,
        "invalid-reference-capabilities",
        `${referenceSetID} must be ${expected.classification} and own exactly ${expected.capabilities.join(", ")}`,
      );
    }
  }
  for (const referenceSetID of referenceSets.keys()) {
    if (!expectedReferenceSets.has(referenceSetID)) {
      addFailure(
        failures,
        "extra-reference-set",
        `unexpected reference set ${referenceSetID}`,
      );
    }
  }
  for (const referenceSetID of [
    "project",
    "identity",
    "identity-baseline-support",
    "identity-interface-support",
    "baseline-common-support",
    "deferred-baseline-support",
  ]) {
    if (!referenceSets.has(referenceSetID)) {
      addFailure(
        failures,
        "missing-reference-set",
        `missing ${referenceSetID} reference set`,
      );
    }
  }

  const assessmentSets = indexUnique(
    document.assessment_sets,
    (item) => item?.id,
    failures,
    "duplicate-assessment-set",
    "assessment_sets",
  );
  for (const assessment of assessmentSets.values()) {
    validateAssessmentSet(root, assessment, failures);
  }

  const verificationSets = indexUnique(
    document.verification_sets,
    (item) => item?.id,
    failures,
    "duplicate-verification-set",
    "verification_sets",
  );
  for (const verification of verificationSets.values()) {
    validateVerificationSet(root, verification, profile, failures);
  }

  const proofSets = indexUnique(
    document.proof_sets,
    (item) => item?.id,
    failures,
    "duplicate-proof-set",
    "proof_sets",
  );
  for (const proofSet of proofSets.values()) {
    validateProofSet(root, proofSet, expectedDimensions, failures);
  }
  if (
    document.row_proofs === null ||
    typeof document.row_proofs !== "object" ||
    Array.isArray(document.row_proofs)
  ) {
    addFailure(
      failures,
      "invalid-row-proofs",
      "row_proofs must be an object keyed by exact row ID",
    );
  }
  if (
    document.row_applicability === null ||
    typeof document.row_applicability !== "object" ||
    Array.isArray(document.row_applicability)
  ) {
    addFailure(
      failures,
      "invalid-row-applicability",
      "row_applicability must be an object keyed by exact row ID",
    );
  }

  let currentSourceDigest = null;
  let currentSourceManifest = null;
  const testedRevisionManifests = new Map();
  const retainedClaimDigests = new Map();
  if (
    [...verificationSets.values()].some(
      (verification) => verification.state === "retained",
    )
  ) {
    try {
      currentSourceManifest = calculateSourceManifest(root);
      currentSourceDigest = currentSourceManifest.digest;
    } catch (error) {
      addFailure(
        failures,
        "evidence-source-digest-failed",
        `cannot calculate current attestation digest: ${error.message}`,
      );
    }
    for (const verification of verificationSets.values()) {
      if (verification.state !== "retained") continue;
      try {
        retainedClaimDigests.set(
          verification.id,
          calculateRetainedClaimDigest(document, verification.id),
        );
      } catch (error) {
        addFailure(
          failures,
          "invalid-evidence-attestation",
          `verification set ${verification.id}: cannot calculate exact retained claim digest: ${error.message}`,
        );
      }
      if (!/^[0-9a-f]{40}$/u.test(verification.tested_revision ?? "")) {
        continue;
      }
      if (!testedRevisionManifests.has(verification.tested_revision)) {
        try {
          testedRevisionManifests.set(
            verification.tested_revision,
            calculateSourceManifestAtGitRevision(
              root,
              verification.tested_revision,
            ),
          );
        } catch (error) {
          addFailure(
            failures,
            "invalid-tested-revision",
            `verification set ${verification.id}: cannot reconstruct tested_revision: ${error.message}`,
          );
        }
      }
      const testedRevisionManifest = testedRevisionManifests.get(
        verification.tested_revision,
      );
      if (
        testedRevisionManifest &&
        testedRevisionManifest.digest !== verification.tested_digest
      ) {
        addFailure(
          failures,
          "invalid-tested-revision",
          `verification set ${verification.id}: tested_digest does not match the owned source reconstructed from tested_revision`,
        );
      }
    }
  }
  const relativeTraceabilityPath = path
    .relative(root, traceabilityPath)
    .split(path.sep)
    .join("/");

  validateRows({
    root,
    rows: document.rows,
    referenceSets,
    applicabilitySets,
    rowApplicability: document.row_applicability,
    proofSets,
    rowProofs: document.row_proofs,
    verificationSets,
    assessmentSets,
    profile,
    compatibilityBaseline: document.compatibility_baseline,
    proofDimensions: expectedDimensions,
    evidenceDocuments,
    currentSourceDigest,
    currentSourceManifest,
    testedRevisionManifests,
    retainedClaimDigests,
    traceabilityPath: relativeTraceabilityPath,
    scenarioIDs,
    planRuleIDs,
    operationCatalog,
    failures,
  });

  const metadataTotals = [...metadata.values()].reduce(
    (totals, resource) => {
      totals.writable_fields += resource.writable_fields?.length ?? 0;
      totals.response_only_fields += resource.response_only_fields?.length ?? 0;
      totals.filters += resource.filters?.length ?? 0;
      totals.ordering += resource.ordering?.length ?? 0;
      totals.relationships += resource.relationships?.length ?? 0;
      totals.choices += Object.keys(resource.choice_fields ?? {}).length;
      return totals;
    },
    {
      writable_fields: 0,
      response_only_fields: 0,
      filters: 0,
      ordering: 0,
      relationships: 0,
      choices: 0,
    },
  );
  return {
    failures,
    document: clone(document),
    counts: {
      scenarios: scenarioIDs.size,
      resource_operations: expectedOperationSources(profile).size,
      resource_contracts: resources.size,
      plan_rules: planRuleIDs.size,
      rows: Array.isArray(document.rows) ? document.rows.length : 0,
      reference_sets: referenceSets.size,
      applicability_sets: applicabilitySets.size,
      proof_sets: proofSets.size,
      ...metadataTotals,
    },
  };
}

function formatFailure(failure) {
  return `  [${failure.code}] ${failure.message}`;
}

function runCLI() {
  const profilePath =
    process.argv[2] ??
    "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml";
  const result = validateTraceability({ profilePath });
  if (result.failures.length > 0) {
    console.error("Traceability validation failed:");
    for (const failure of result.failures) {
      console.error(formatFailure(failure));
    }
    process.exitCode = 1;
    return;
  }
  const counts = result.counts;
  console.log(
    `Traceability valid: ${counts.rows} rows (${counts.scenarios} scenarios, ${counts.resource_operations} operations, ${counts.resource_contracts} resource contracts, ${counts.plan_rules} plan rules); ` +
      `${counts.writable_fields} writable fields, ${counts.response_only_fields} response-only fields, ${counts.filters} filters, ${counts.ordering} ordering keys, ${counts.relationships} relationships, ${counts.choices} choices`,
  );
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href
) {
  runCLI();
}
