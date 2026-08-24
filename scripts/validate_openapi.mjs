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

const contractedResourceShapes = {
  Manufacturer: {
    request: {
      name: { type: "string" },
      slug: { type: "string" },
      description: { type: "string" },
    },
    response: {
      name: { type: "string" },
      slug: { type: "string" },
      description: { type: "string" },
      id: { type: "integer", format: "int64" },
      url: { type: "string", format: "uri" },
      display: { type: "string" },
      created: { type: "string", format: "date-time" },
      last_updated: { type: "string", format: "date-time" },
      devicetype_count: { type: "integer", format: "int64" },
    },
  },
  Site: {
    request: {
      name: { type: "string" },
      slug: { type: "string" },
      facility: { type: "string" },
      description: { type: "string" },
      comments: { type: "string" },
    },
    response: {
      name: { type: "string" },
      slug: { type: "string" },
      facility: { type: "string" },
      description: { type: "string" },
      comments: { type: "string" },
      id: { type: "integer", format: "int64" },
      url: { type: "string", format: "uri" },
      display: { type: "string" },
      created: { type: "string", format: "date-time" },
      last_updated: { type: "string", format: "date-time" },
      device_count: { type: "integer", format: "int64" },
      prefix_count: { type: "integer", format: "int64" },
      rack_count: { type: "integer", format: "int64" },
    },
  },
  IPAddress: {
    request: {
      address: { type: "string" },
      vrf: { type: "integer", format: "int64" },
      dns_name: { type: "string" },
      description: { type: "string" },
      comments: { type: "string" },
      assigned_object_type: { type: "string" },
      assigned_object_id: { type: "integer", format: "int64" },
    },
    response: {
      address: { type: "string" },
      dns_name: { type: "string" },
      description: { type: "string" },
      comments: { type: "string" },
      assigned_object_type: { type: "string" },
      assigned_object_id: { type: "integer", format: "int64" },
      id: { type: "integer", format: "int64" },
      url: { type: "string", format: "uri" },
      display: { type: "string" },
      created: { type: "string", format: "date-time" },
      last_updated: { type: "string", format: "date-time" },
    },
    responseReferences: {
      vrf: "#/components/schemas/ObjectReference",
      assigned_object: "#/components/schemas/ObjectReference",
    },
  },
};

const resourceMetadata = {};
const choiceFields = {};
for (const relativePath of profile.resource_metadata) {
  const document = JSON.parse(
    fs.readFileSync(
      path.resolve(path.dirname(profilePath), relativePath),
      "utf8",
    ),
  );
  for (const resource of document.resources) {
    resourceMetadata[resource.name] = resource;
    choiceFields[resource.name] = resource.choice_fields ?? {};
  }
}

function assert(condition, message) {
  if (!condition) failures.push(message);
}

function schemaNullable(schema) {
  return (
    schema?.type === "null" ||
    schema?.type?.includes?.("null") ||
    schema?.oneOf?.some?.((candidate) => candidate.type === "null") === true
  );
}

function stringValuedField(resource, field) {
  const choice = resource.choice_fields?.[field];
  if (choice) return choice.value_type === "string";
  if ((resource.relationships ?? []).includes(field)) return false;
  return !field.endsWith("_id");
}

function exactMembers(actual, expected) {
  return (
    JSON.stringify([...(actual ?? [])].sort()) ===
    JSON.stringify([...expected].sort())
  );
}

function nonNullTypes(schema) {
  const declaredTypes = Array.isArray(schema?.type)
    ? schema.type
    : schema?.type === undefined
      ? []
      : [schema.type];
  return declaredTypes.filter((type) => type !== "null");
}

function assertScalarShape(schema, expected, label) {
  assert(
    exactMembers(nonNullTypes(schema), [expected.type]),
    `${label} base type must be ${expected.type}`,
  );
  assert(
    schema?.format === expected.format,
    `${label} format must be ${expected.format ?? "absent"}`,
  );
}

function assertReferenceShape(schema, expectedReference, label) {
  const concreteVariants = (schema?.oneOf ?? []).filter(
    (candidate) => candidate.type !== "null",
  );
  assert(
    concreteVariants.length === 1 &&
      concreteVariants[0]?.$ref === expectedReference,
    `${label} must reference ${expectedReference}`,
  );
}

function requestSchemaReference(resource, operation) {
  const operationDetails = {
    create: { method: "post", path: resource.rest_path },
    replace: { method: "put", path: `${resource.rest_path}{id}/` },
    update: { method: "patch", path: `${resource.rest_path}{id}/` },
  }[operation];
  return {
    ...operationDetails,
    reference:
      spec.paths?.[operationDetails.path]?.[operationDetails.method]
        ?.requestBody?.content?.["application/json"]?.schema?.$ref,
  };
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
  const metadata = resourceMetadata[resource.name];
  const requestSchemaNames = metadata?.field_contracts
    ? {
        create: `${resource.name}Create`,
        replace: `${resource.name}Replace`,
        update: `${resource.name}Update`,
      }
    : {
        create: `${resource.name}Write`,
        replace: `${resource.name}Write`,
        update: `${resource.name}Write`,
      };
  for (const operation of ["create", "replace", "update"]) {
    const schemaName = requestSchemaNames[operation];
    assert(
      spec.components.schemas[schemaName],
      `missing request schema ${schemaName}`,
    );
    const request = requestSchemaReference(resource, operation);
    const expectedReference = `#/components/schemas/${schemaName}`;
    assert(
      request.reference === expectedReference,
      `${request.method.toUpperCase()} ${request.path} must reference ${expectedReference}`,
    );
  }
  if (metadata?.field_contracts) {
    assert(
      spec.components.schemas[`${resource.name}Write`] === undefined,
      `${resource.name} must not retain a conflated shared write schema`,
    );
  }
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
  const metadataDocument = resourceMetadata[resource];
  const requestSchemaNames = metadataDocument.field_contracts
    ? [`${resource}Create`, `${resource}Replace`, `${resource}Update`]
    : [`${resource}Write`];
  for (const [field, metadata] of Object.entries(fields)) {
    const responseSchema =
      spec.components.schemas[resource]?.properties?.[field];
    const responseChoice = choiceObject(responseSchema);
    assert(
      responseChoice,
      `${resource}.${field} response must be a choice object (or nullable choice object)`,
    );
    assert(
      responseChoice?.additionalProperties === false,
      `${resource}.${field} choice response must reject undeclared members`,
    );
    assert(
      JSON.stringify([...(responseChoice?.required ?? [])].sort()) ===
        JSON.stringify(["label", "value"]),
      `${resource}.${field} choice response must require value and label`,
    );
    const expectedChoiceShape =
      metadata.value_type === "integer"
        ? { type: "integer", format: "int64" }
        : { type: "string" };
    assertScalarShape(
      responseChoice?.properties?.value,
      expectedChoiceShape,
      `${resource}.${field}.value`,
    );
    assertScalarShape(
      responseChoice?.properties?.label,
      { type: "string" },
      `${resource}.${field}.label`,
    );
    for (const requestSchemaName of requestSchemaNames) {
      const writeField =
        spec.components.schemas[requestSchemaName]?.properties?.[field];
      if (!writeField) continue;
      assert(
        !choiceObject(writeField),
        `${requestSchemaName}.${field} must remain scalar`,
      );
      const operation = requestSchemaName.endsWith("Create")
        ? "create"
        : requestSchemaName.endsWith("Replace")
          ? "replace"
          : requestSchemaName.endsWith("Update")
            ? "update"
            : null;
      const expectedNullable = operation
        ? metadataDocument.field_contracts[operation].nullable_fields.includes(
            field,
          )
        : metadata.nullable;
      assert(
        schemaNullable(writeField) === expectedNullable,
        `${requestSchemaName}.${field} nullability must match resource metadata`,
      );
      assertScalarShape(
        writeField,
        expectedChoiceShape,
        `${requestSchemaName}.${field}`,
      );
    }
    assert(
      schemaNullable(responseSchema) ===
        (metadataDocument.field_contracts
          ? metadataDocument.field_contracts.response.nullable_fields.includes(
              field,
            )
          : metadata.nullable),
      `${resource}.${field} response nullability must match resource metadata`,
    );
  }
}
assert(
  Object.values(choiceFields).reduce(
    (total, fields) => total + Object.keys(fields).length,
    0,
  ) === 19,
  "resource metadata must declare the 19 pinned first-profile choice fields",
);

for (const resource of profile.resources) {
  const metadata = resourceMetadata[resource.name];
  if (!metadata?.field_contracts) continue;

  const resourceShapes = contractedResourceShapes[resource.name];
  assert(
    resourceShapes,
    `${resource.name} must have independently pinned OpenAPI field shapes`,
  );

  const responseFields = [
    ...new Set([...metadata.writable_fields, ...metadata.response_only_fields]),
  ];
  const responseSchema = spec.components.schemas[resource.name];
  assert(
    responseSchema?.additionalProperties === false,
    `${resource.name} response must reject undeclared members`,
  );
  assert(
    exactMembers(Object.keys(responseSchema?.properties ?? {}), responseFields),
    `${resource.name} response fields must match resource metadata`,
  );
  const responseChoiceFields = new Set(
    Object.keys(metadata.choice_fields ?? {}),
  );
  const responseReferenceFields = new Set(
    Object.keys(resourceShapes?.responseReferences ?? {}),
  );
  const scalarResponseFields = responseFields.filter(
    (field) =>
      !responseChoiceFields.has(field) && !responseReferenceFields.has(field),
  );
  assert(
    exactMembers(
      Object.keys(resourceShapes?.response ?? {}),
      scalarResponseFields,
    ),
    `${resource.name} response scalar shape pins must cover every non-choice field`,
  );
  for (const field of responseFields) {
    const fieldSchema = responseSchema?.properties?.[field];
    const expectedNullable =
      metadata.field_contracts.response.nullable_fields.includes(field);
    assert(
      schemaNullable(fieldSchema) === expectedNullable,
      `${resource.name}.${field} response nullability must match field_contracts.response`,
    );
    const expectedScalarShape = resourceShapes?.response?.[field];
    if (expectedScalarShape) {
      assertScalarShape(
        fieldSchema,
        expectedScalarShape,
        `${resource.name}.${field}`,
      );
    }
    const expectedReference = resourceShapes?.responseReferences?.[field];
    if (expectedReference) {
      assertReferenceShape(
        fieldSchema,
        expectedReference,
        `${resource.name}.${field}`,
      );
    }
  }

  const requestChoiceFields = new Set(
    Object.keys(metadata.choice_fields ?? {}).filter((field) =>
      metadata.writable_fields.includes(field),
    ),
  );
  const scalarRequestFields = metadata.writable_fields.filter(
    (field) => !requestChoiceFields.has(field),
  );
  assert(
    exactMembers(
      Object.keys(resourceShapes?.request ?? {}),
      scalarRequestFields,
    ),
    `${resource.name} request scalar shape pins must cover every non-choice field`,
  );

  for (const operation of ["create", "replace", "update"]) {
    const schemaName = `${resource.name}${operation[0].toUpperCase()}${operation.slice(1)}`;
    const requestSchema = spec.components.schemas[schemaName];
    const contract = metadata.field_contracts[operation];
    assert(
      requestSchema?.additionalProperties === false,
      `${schemaName} must reject undeclared members`,
    );
    assert(
      exactMembers(
        Object.keys(requestSchema?.properties ?? {}),
        metadata.writable_fields,
      ),
      `${schemaName} fields must match writable_fields`,
    );
    assert(
      exactMembers(requestSchema?.required, contract.required_fields),
      `${schemaName} required fields must match field_contracts.${operation}`,
    );
    for (const field of metadata.writable_fields) {
      const fieldSchema = requestSchema?.properties?.[field];
      const expectedNullable = contract.nullable_fields.includes(field);
      assert(
        schemaNullable(fieldSchema) === expectedNullable,
        `${schemaName}.${field} nullability must match field_contracts.${operation}`,
      );
      const expectedScalarShape = resourceShapes?.request?.[field];
      if (expectedScalarShape) {
        assertScalarShape(
          fieldSchema,
          expectedScalarShape,
          `${schemaName}.${field}`,
        );
      }
      if (stringValuedField(metadata, field)) {
        const expectedBlank = contract.blank_fields.includes(field);
        const minLength = fieldSchema?.minLength;
        assert(
          expectedBlank
            ? minLength === undefined || minLength === 0
            : minLength === 1,
          `${schemaName}.${field} blankability must match field_contracts.${operation}`,
        );
      }
    }
  }
}

assert(
  spec.components.schemas.IPAddress?.properties?.assigned_object_id?.type?.includes?.(
    "integer",
  ) ||
    spec.components.schemas.IPAddress?.properties?.assigned_object_id?.type ===
      "integer",
  "IPAddress.assigned_object_id must remain an integer identifier, not an object reference",
);
assert(
  choiceObject(
    spec.components.schemas.IPAddress?.properties?.assigned_object,
  ) === undefined &&
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
