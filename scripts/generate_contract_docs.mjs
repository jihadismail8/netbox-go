#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const CHECK = process.argv.includes("--check");
const positional = process.argv
  .slice(2)
  .filter((argument) => argument !== "--check");
const profilePath = path.resolve(
  ROOT,
  positional[0] ??
    "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const descriptorPath = positional[1] ? path.resolve(ROOT, positional[1]) : null;
const profile = JSON.parse(fs.readFileSync(profilePath, "utf8"));
const profileDir = path.dirname(profilePath);
const contractRoot = path.dirname(profileDir);

if (descriptorPath) {
  const descriptor = fs.readFileSync(descriptorPath);
  if (!descriptor.includes(Buffer.from("netbox/dcim/v1/dcim_service.proto"))) {
    throw new Error("descriptor does not contain the netbox v1 module");
  }
}

const metadata = new Map();
const choiceMetadata = new Map();
for (const relativePath of profile.resource_metadata) {
  const document = JSON.parse(
    fs.readFileSync(path.resolve(profileDir, relativePath), "utf8"),
  );
  for (const resource of document.resources) {
    const resourceKey = `${document.module}.${resource.name}`;
    metadata.set(resourceKey, resource);
    for (const [field, choice] of Object.entries(resource.choice_fields ?? {})) {
      choiceMetadata.set(`${resourceKey}.${field}`, choice);
    }
  }
}
const identity = JSON.parse(
  fs.readFileSync(
    path.resolve(profileDir, profile.identity_extension.resource_metadata),
    "utf8",
  ),
);

const booleanFields = new Set([
  "desc_units",
  "enabled",
  "enforce_unique",
  "exclude_from_utilization",
  "is_full_depth",
  "is_pool",
  "mark_utilized",
  "mgmt_only",
  "vm_role",
  "write_enabled",
]);
const integerFields = new Set([
  "assigned_object_id",
  "children",
  "count_ipaddresses",
  "depth",
  "_depth",
  "device_count",
  "devicetype_count",
  "family",
  "id",
  "interface_count",
  "interface_template_count",
  "ipaddress_count",
  "mtu",
  "prefix_count",
  "rack_count",
  "speed",
  "starting_unit",
  "u_height",
  "width",
]);
const relationFields = new Set([
  "assigned_object",
  "device",
  "device_type",
  "manufacturer",
  "parent",
  "rack",
  "rack_type",
  "role",
  "site",
  "vrf",
]);
const decimalFields = new Set(["dcim.DeviceType.u_height", "dcim.Device.position"]);
const nullableDecimalFields = new Set(["dcim.Device.position"]);
const nullableIntegerFields = new Set([
  "dcim.Interface.mtu",
  "dcim.Interface.speed",
  "ipam.IPAddress.assigned_object_id",
]);

function scalarFieldSchema(resourceKey, field, response) {
  if (
    field === "created" ||
    field === "last_updated" ||
    field === "expires" ||
    field === "last_used"
  ) {
    return { type: ["string", "null"], format: "date-time" };
  }
  if (booleanFields.has(field)) return { type: "boolean" };
  if (decimalFields.has(`${resourceKey}.${field}`))
    return {
      type: nullableDecimalFields.has(`${resourceKey}.${field}`)
        ? ["number", "null"]
        : "number",
      format: "double",
    };
  if (integerFields.has(field) || field.endsWith("_count"))
    return {
      type: nullableIntegerFields.has(`${resourceKey}.${field}`)
        ? ["integer", "null"]
        : "integer",
      format: "int64",
    };
  if (relationFields.has(field)) {
    return response
      ? {
          oneOf: [
            { $ref: "#/components/schemas/ObjectReference" },
            { type: "null" },
          ],
        }
      : { type: ["integer", "null"], format: "int64" };
  }
  if (field === "url") return { type: "string", format: "uri" };
  return { type: ["string", "null"] };
}

function fieldSchema(resourceKey, field, response) {
  const qualifiedField = `${resourceKey}.${field}`;
  const choiceField = choiceMetadata.get(qualifiedField);
  if (choiceField) {
    const scalar = choiceField.value_type === "integer"
      ? { type: "integer", format: "int64" }
      : { type: "string" };
    if (!response) {
      return choiceField.nullable
        ? { ...scalar, type: [scalar.type, "null"] }
        : scalar;
    }
    const choice = {
      type: "object",
      additionalProperties: false,
      required: ["value", "label"],
      properties: {
        value: scalar,
        label: { type: "string" },
      },
    };
    return choiceField.nullable
      ? { oneOf: [choice, { type: "null" }] }
      : choice;
  }
  return scalarFieldSchema(resourceKey, field, response);
}

function operationID(verb, name) {
  return `${verb}${name}`;
}

const schemas = {
  ObjectReference: {
    type: "object",
    additionalProperties: false,
    required: ["id", "url", "display"],
    properties: {
      id: { type: "integer", format: "int64" },
      url: { type: "string", format: "uri" },
      display: { type: "string" },
    },
  },
  Error: {
    type: "object",
    additionalProperties: false,
    required: ["detail"],
    properties: {
      detail: { type: "string" },
      errors: { type: "object", additionalProperties: true },
      reason: { type: "string" },
      request_id: { type: "string" },
    },
  },
};
const paths = {};

for (const profileResource of profile.resources) {
  const key = `${profileResource.module}.${profileResource.name}`;
  const resource = metadata.get(key);
  const writeName = `${profileResource.name}Write`;
  const listName = `${profileResource.name}List`;
  const responseFields = [
    ...new Set([...resource.writable_fields, ...resource.response_only_fields]),
  ];
  schemas[writeName] = {
    type: "object",
    additionalProperties: false,
    properties: Object.fromEntries(
      resource.writable_fields.map((field) => [
        field,
        fieldSchema(key, field, false),
      ]),
    ),
  };
  schemas[profileResource.name] = {
    type: "object",
    additionalProperties: false,
    properties: Object.fromEntries(
      responseFields.map((field) => [field, fieldSchema(key, field, true)]),
    ),
  };
  schemas[listName] = {
    type: "object",
    additionalProperties: false,
    required: ["count", "results"],
    properties: {
      count: { type: "integer", format: "int64" },
      next: { type: ["string", "null"], format: "uri" },
      previous: { type: ["string", "null"], format: "uri" },
      results: {
        type: "array",
        items: { $ref: `#/components/schemas/${profileResource.name}` },
      },
    },
  };

  const parameters = [
    { name: "limit", in: "query", schema: { type: "integer", minimum: 0 } },
    { name: "offset", in: "query", schema: { type: "integer", minimum: 0 } },
    { name: "q", in: "query", schema: { type: "string" } },
    { name: "ordering", in: "query", schema: { type: "string" } },
    ...resource.filters.map((filter) => ({
      name: filter,
      in: "query",
      schema: { type: "string" },
    })),
  ];
  paths[profileResource.rest_path] = {
    get: {
      operationId: operationID("list", `${profileResource.name}s`),
      tags: [profileResource.module],
      parameters,
      responses: {
        200: {
          description: "Success",
          content: {
            "application/json": {
              schema: { $ref: `#/components/schemas/${listName}` },
            },
          },
        },
      },
    },
    post: {
      operationId: operationID("create", profileResource.name),
      tags: [profileResource.module],
      requestBody: {
        required: true,
        content: {
          "application/json": {
            schema: { $ref: `#/components/schemas/${writeName}` },
          },
        },
      },
      responses: {
        201: {
          description: "Created",
          content: {
            "application/json": {
              schema: { $ref: `#/components/schemas/${profileResource.name}` },
            },
          },
        },
      },
    },
  };
  const detailPath = `${profileResource.rest_path}{id}/`;
  const detailParameters = [
    {
      name: "id",
      in: "path",
      required: true,
      schema: { type: "integer", format: "int64" },
    },
  ];
  paths[detailPath] = {
    get: {
      operationId: operationID("get", profileResource.name),
      tags: [profileResource.module],
      parameters: detailParameters,
      responses: {
        200: {
          description: "Success",
          content: {
            "application/json": {
              schema: { $ref: `#/components/schemas/${profileResource.name}` },
            },
          },
        },
      },
    },
    put: {
      operationId: operationID("replace", profileResource.name),
      tags: [profileResource.module],
      parameters: detailParameters,
      requestBody: {
        required: true,
        content: {
          "application/json": {
            schema: { $ref: `#/components/schemas/${writeName}` },
          },
        },
      },
      responses: {
        200: {
          description: "Replaced",
          content: {
            "application/json": {
              schema: { $ref: `#/components/schemas/${profileResource.name}` },
            },
          },
        },
      },
    },
    patch: {
      operationId: operationID("update", profileResource.name),
      tags: [profileResource.module],
      parameters: detailParameters,
      requestBody: {
        required: true,
        content: {
          "application/json": {
            schema: { $ref: `#/components/schemas/${writeName}` },
          },
        },
      },
      responses: {
        200: {
          description: "Updated",
          content: {
            "application/json": {
              schema: { $ref: `#/components/schemas/${profileResource.name}` },
            },
          },
        },
      },
    },
    delete: {
      operationId: operationID("delete", profileResource.name),
      tags: [profileResource.module],
      parameters: detailParameters,
      responses: { 204: { description: "Deleted" } },
    },
  };
}

for (const operation of identity.rest_operations) {
  const pathItem = (paths[operation.path] ??= {});
  const identityOperationID = `${operation.method.toLowerCase()}${operation.path
    .split("/")
    .filter((part) => part && !part.startsWith("{"))
    .map((part) => `${part[0].toUpperCase()}${part.slice(1)}`)
    .join("")}`;
  pathItem[operation.method.toLowerCase()] = {
    operationId: operation.grpc_mapping ?? identityOperationID,
    tags: ["identity"],
    parameters: operation.path.includes("{id}")
      ? [
          {
            name: "id",
            in: "path",
            required: true,
            schema: { type: "integer", format: "int64" },
          },
        ]
      : undefined,
    security: operation.public ? [] : undefined,
    responses: {
      200: { description: "Success" },
      201: { description: "Created" },
      204: { description: "No content" },
    },
  };
}

const openAPI = {
  openapi: "3.1.0",
  info: {
    title: "NetBox Go Core Workflow API",
    version: "1.0.0-pre",
    description:
      "Pre-publication exact REST contract for the core-workflow-v1 capability profile.",
  },
  servers: [{ url: "/" }],
  security: [{ TokenAuth: [] }, { SessionCookie: [] }],
  tags: [{ name: "identity" }, { name: "dcim" }, { name: "ipam" }],
  paths: Object.fromEntries(
    Object.entries(paths).sort(([a], [b]) => a.localeCompare(b)),
  ),
  components: {
    securitySchemes: {
      TokenAuth: {
        type: "apiKey",
        in: "header",
        name: "Authorization",
        description: "Token <key>",
      },
      SessionCookie: { type: "apiKey", in: "cookie", name: "sessionid" },
    },
    schemas: Object.fromEntries(
      Object.entries(schemas).sort(([a], [b]) => a.localeCompare(b)),
    ),
  },
};

function protobufInventory() {
  const protoRoot = path.join(ROOT, "netbox-backend", "api", "proto");
  const services = [];
  for (const file of fs.readdirSync(path.join(protoRoot, "netbox")).sort()) {
    const moduleDir = path.join(protoRoot, "netbox", file, "v1");
    if (!fs.existsSync(moduleDir)) continue;
    for (const proto of fs
      .readdirSync(moduleDir)
      .filter((name) => name.endsWith(".proto"))
      .sort()) {
      const content = fs.readFileSync(path.join(moduleDir, proto), "utf8");
      const packageName = content.match(/^package\s+([^;]+);/m)?.[1];
      for (const service of content.matchAll(
        /^service\s+(\w+)\s*\{([\s\S]*?)^\}/gm,
      )) {
        services.push({
          package: packageName,
          service: service[1],
          rpcs: [...service[2].matchAll(/^\s*rpc\s+(\w+)/gm)].map(
            (rpc) => rpc[1],
          ),
          source: path.relative(ROOT, path.join(moduleDir, proto)),
        });
      }
    }
  }
  return services;
}

let capabilityDoc = `# Core workflow capability contract\n\n`;
capabilityDoc += `> Generated from \`${path.relative(ROOT, profilePath)}\`. Contract state: **${profile.contract_state}**.\n\n`;
capabilityDoc += `| Module | Resource | REST | gRPC service | Tier | Owner |\n`;
capabilityDoc += `| --- | --- | --- | --- | --- | --- |\n`;
for (const resource of profile.resources) {
  capabilityDoc += `| ${resource.module} | ${resource.name} | \`${resource.rest_path}\` | \`${resource.grpc_service}\` | ${resource.tier} | ${resource.owner} |\n`;
}
capabilityDoc += `\nThe identity surface is an extension with tier \`${profile.identity_extension.tier}\`; it cannot be counted as baseline T2 compatibility.\n`;

let grpcDoc = `# gRPC v1 contract inventory\n\n`;
grpcDoc += `> Generated from the canonical handwritten protobuf module. HTTP annotations are intentionally absent.\n\n`;
for (const service of protobufInventory()) {
  grpcDoc += `## ${service.package}.${service.service}\n\n`;
  grpcDoc += `Source: \`${service.source}\`\n\n`;
  for (const rpc of service.rpcs) grpcDoc += `- \`${rpc}\`\n`;
  grpcDoc += `\n`;
}
grpcDoc = `${grpcDoc.trimEnd()}\n`;

const outputs = new Map([
  [
    path.join(ROOT, "netbox-backend", "api", "openapi", "netbox-go-v1.yaml"),
    `${JSON.stringify(openAPI, null, 2)}\n`,
  ],
  [path.join(ROOT, "docs", "contracts", "core-workflow-v1.md"), capabilityDoc],
  [path.join(ROOT, "docs", "contracts", "grpc-v1.md"), grpcDoc],
]);
const drift = [];
for (const [filePath, content] of outputs) {
  if (CHECK) {
    if (
      !fs.existsSync(filePath) ||
      fs.readFileSync(filePath, "utf8") !== content
    ) {
      drift.push(path.relative(ROOT, filePath));
    }
  } else {
    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    fs.writeFileSync(filePath, content);
  }
}

if (drift.length > 0) {
  console.error("Generated contract documentation is out of date:");
  for (const file of drift) console.error(`  ${file}`);
  process.exitCode = 1;
} else {
  console.log(
    `${CHECK ? "Checked" : "Generated"} OpenAPI and ${outputs.size - 1} contract documents`,
  );
}
