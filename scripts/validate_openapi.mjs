#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const specPath = path.resolve(
  ROOT,
  process.argv[2] ?? "netbox-backend/api/openapi/netbox-go-v1.yaml",
);
const profilePath = path.join(
  ROOT,
  "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const profile = JSON.parse(fs.readFileSync(profilePath, "utf8"));
const spec = JSON.parse(fs.readFileSync(specPath, "utf8"));
const identity = JSON.parse(
  fs.readFileSync(
    path.resolve(
      path.dirname(profilePath),
      profile.identity_extension.resource_metadata,
    ),
    "utf8",
  ),
);
const failures = [];

const choiceFields = {};
for (const relativePath of profile.resource_metadata) {
  const document = JSON.parse(
    fs.readFileSync(path.resolve(path.dirname(profilePath), relativePath), "utf8"),
  );
  for (const resource of document.resources) {
    choiceFields[resource.name] = resource.choice_fields ?? {};
  }
}

function assert(condition, message) {
  if (!condition) failures.push(message);
}

assert(spec.openapi === "3.1.0", "OpenAPI version must be 3.1.0");
assert(
  spec.info?.version === "1.0.0-pre",
  "OpenAPI contract must remain pre-publication",
);
assert(
  spec.components?.securitySchemes?.TokenAuth,
  "TokenAuth security scheme is missing",
);
assert(
  spec.components?.securitySchemes?.SessionCookie,
  "SessionCookie security scheme is missing",
);

const expectedPaths = new Map();
for (const resource of profile.resources) {
  expectedPaths.set(resource.rest_path, new Set(["get", "post"]));
  expectedPaths.set(
    `${resource.rest_path}{id}/`,
    new Set(["get", "put", "patch", "delete"]),
  );
  assert(
    spec.components.schemas[resource.name],
    `missing response schema ${resource.name}`,
  );
  assert(
    spec.components.schemas[`${resource.name}Write`],
    `missing write schema ${resource.name}Write`,
  );
}
for (const operation of identity.rest_operations) {
  const methods = expectedPaths.get(operation.path) ?? new Set();
  methods.add(operation.method.toLowerCase());
  expectedPaths.set(operation.path, methods);
}

function choiceObject(schema) {
  if (schema?.type === "object") return schema;
  return schema?.oneOf?.find((candidate) => candidate.type === "object");
}

for (const [resource, fields] of Object.entries(choiceFields)) {
  for (const [field, metadata] of Object.entries(fields)) {
    const responseSchema = spec.components.schemas[resource]?.properties?.[field];
    const responseChoice = choiceObject(responseSchema);
    assert(responseChoice, `${resource}.${field} response must be a choice object (or nullable choice object)`);
    assert(
      responseChoice?.additionalProperties === false,
      `${resource}.${field} choice response must reject undeclared members`,
    );
    assert(
      JSON.stringify([...(responseChoice?.required ?? [])].sort()) === JSON.stringify(["label", "value"]),
      `${resource}.${field} choice response must require value and label`,
    );
    const writeField = spec.components.schemas[`${resource}Write`]?.properties?.[field];
    if (writeField) {
      assert(!choiceObject(writeField), `${resource}Write.${field} must remain scalar`);
      assert(
        Array.isArray(writeField.type) === metadata.nullable,
        `${resource}Write.${field} nullability must match resource metadata`,
      );
    }
    assert(
      Boolean(responseSchema?.oneOf?.some?.((candidate) => candidate.type === "null")) === metadata.nullable,
      `${resource}.${field} response nullability must match resource metadata`,
    );
  }
}
assert(
  Object.values(choiceFields).reduce((total, fields) => total + Object.keys(fields).length, 0) === 19,
  "resource metadata must declare the 19 pinned first-profile choice fields",
);

assert(
  spec.components.schemas.IPAddress?.properties?.assigned_object_id?.type?.includes?.("integer") ||
    spec.components.schemas.IPAddress?.properties?.assigned_object_id?.type === "integer",
  "IPAddress.assigned_object_id must remain an integer identifier, not an object reference",
);
assert(
  choiceObject(spec.components.schemas.IPAddress?.properties?.assigned_object) === undefined &&
    spec.components.schemas.IPAddress?.properties?.assigned_object?.oneOf?.some?.(
      (candidate) => candidate.$ref === "#/components/schemas/ObjectReference",
    ),
  "IPAddress.assigned_object must be a nullable object reference",
);

const operationIDs = new Set();
for (const [apiPath, pathItem] of Object.entries(spec.paths ?? {})) {
  assert(expectedPaths.has(apiPath), `undeclared OpenAPI path ${apiPath}`);
  const expectedMethods = expectedPaths.get(apiPath) ?? new Set();
  const actualMethods = new Set(
    Object.keys(pathItem).filter((key) =>
      ["get", "post", "put", "patch", "delete"].includes(key),
    ),
  );
  for (const method of expectedMethods) {
    assert(
      actualMethods.has(method),
      `missing ${method.toUpperCase()} ${apiPath}`,
    );
  }
  for (const method of actualMethods) {
    assert(
      expectedMethods.has(method),
      `undeclared ${method.toUpperCase()} ${apiPath}`,
    );
    const operation = pathItem[method];
    assert(
      operation.operationId,
      `${method.toUpperCase()} ${apiPath}: missing operationId`,
    );
    assert(
      !operationIDs.has(operation.operationId),
      `duplicate operationId ${operation.operationId}`,
    );
    operationIDs.add(operation.operationId);
    for (const parameter of apiPath.matchAll(/\{([^}]+)\}/g)) {
      assert(
        operation.parameters?.some(
          (candidate) =>
            candidate.name === parameter[1] &&
            candidate.in === "path" &&
            candidate.required === true,
        ),
        `${method.toUpperCase()} ${apiPath}: path parameter ${parameter[1]} is not declared`,
      );
    }
  }
}
for (const expectedPath of expectedPaths.keys()) {
  assert(spec.paths?.[expectedPath], `missing OpenAPI path ${expectedPath}`);
}

const serialized = JSON.stringify(spec);
for (const reference of serialized.matchAll(
  /"\$ref":"#\/components\/schemas\/([^"]+)"/g,
)) {
  assert(
    spec.components.schemas[reference[1]],
    `unresolved schema reference ${reference[1]}`,
  );
}
assert(
  !serialized.includes("google.api.http"),
  "OpenAPI must not be inferred from gRPC annotations",
);

if (failures.length > 0) {
  console.error("OpenAPI validation failed:");
  for (const failure of failures) console.error(`  ${failure}`);
  process.exitCode = 1;
} else {
  console.log(
    `OpenAPI valid: ${expectedPaths.size} paths and ${operationIDs.size} operations`,
  );
}
