import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const CONTRACT_ROOT = path.join(ROOT, "contracts/netbox/v4.4.6-post7");
const PROFILE_RELATIVE_PATH = "profiles/core-workflow-v1.yaml";
const DCIM_RELATIVE_PATH = "resources/dcim.yaml";
const IPAM_RELATIVE_PATH = "resources/ipam.yaml";
const OPENAPI_PATH = path.join(
  ROOT,
  "netbox-backend/api/openapi/netbox-go-v1.yaml",
);

const IP_ADDRESS_FIELD_CONTRACTS = {
  response: {
    nullable_fields: [
      "vrf",
      "role",
      "assigned_object_type",
      "assigned_object_id",
      "assigned_object",
      "created",
      "last_updated",
    ],
  },
  create: {
    required_fields: ["address"],
    nullable_fields: [
      "vrf",
      "role",
      "assigned_object_type",
      "assigned_object_id",
    ],
    blank_fields: ["role", "dns_name", "description", "comments"],
  },
  replace: {
    required_fields: ["address"],
    nullable_fields: [
      "vrf",
      "role",
      "assigned_object_type",
      "assigned_object_id",
    ],
    blank_fields: ["role", "dns_name", "description", "comments"],
  },
  update: {
    required_fields: [],
    nullable_fields: [
      "vrf",
      "role",
      "assigned_object_type",
      "assigned_object_id",
    ],
    blank_fields: ["role", "dns_name", "description", "comments"],
  },
};

const SITE_FIELD_CONTRACTS = {
  response: {
    nullable_fields: ["created", "last_updated"],
  },
  create: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["facility", "description", "comments"],
  },
  replace: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["facility", "description", "comments"],
  },
  update: {
    required_fields: [],
    nullable_fields: [],
    blank_fields: ["facility", "description", "comments"],
  },
};

const MANUFACTURER_FIELD_CONTRACTS = {
  response: {
    nullable_fields: ["created", "last_updated"],
  },
  create: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["description"],
  },
  replace: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["description"],
  },
  update: {
    required_fields: [],
    nullable_fields: [],
    blank_fields: ["description"],
  },
};

const RACK_ROLE_FIELD_CONTRACTS = {
  response: {
    nullable_fields: ["created", "last_updated"],
  },
  create: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["description"],
  },
  replace: {
    required_fields: ["name", "slug"],
    nullable_fields: [],
    blank_fields: ["description"],
  },
  update: {
    required_fields: [],
    nullable_fields: [],
    blank_fields: ["description"],
  },
};

const RACK_TYPE_FIELD_CONTRACTS = {
  response: {
    nullable_fields: ["created", "last_updated"],
  },
  create: {
    required_fields: ["manufacturer", "model", "slug", "form_factor"],
    nullable_fields: [],
    blank_fields: ["description", "comments"],
  },
  replace: {
    required_fields: ["manufacturer", "model", "slug"],
    nullable_fields: [],
    blank_fields: ["description", "comments"],
  },
  update: {
    required_fields: [],
    nullable_fields: [],
    blank_fields: ["description", "comments"],
  },
};

function runNode(script, argument) {
  return spawnSync(process.execPath, [path.join(ROOT, script), argument], {
    cwd: ROOT,
    encoding: "utf8",
  });
}

function diagnostics(result) {
  return `${result.stdout ?? ""}${result.stderr ?? ""}${
    result.error ? `spawn error: ${result.error.message}\n` : ""
  }`;
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function withResourceContractFixture(
  relativePath,
  resourceName,
  fieldContracts,
  mutator,
  callback,
) {
  const fixtureRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), "netbox-go-capability-profile-"),
  );
  const contractRoot = path.join(fixtureRoot, "v4.4.6-post7");
  fs.cpSync(CONTRACT_ROOT, contractRoot, { recursive: true });
  try {
    const metadataPath = path.join(contractRoot, relativePath);
    const metadata = JSON.parse(fs.readFileSync(metadataPath, "utf8"));
    const resource = metadata.resources.find(
      (candidate) => candidate.name === resourceName,
    );
    assert.ok(resource, `fixture must contain ${resourceName} metadata`);
    resource.field_contracts = structuredClone(fieldContracts);
    mutator(resource);
    fs.writeFileSync(metadataPath, `${JSON.stringify(metadata, null, 2)}\n`);
    callback(path.join(contractRoot, PROFILE_RELATIVE_PATH));
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function withContractFixture(mutator, callback) {
  withResourceContractFixture(
    IPAM_RELATIVE_PATH,
    "IPAddress",
    IP_ADDRESS_FIELD_CONTRACTS,
    (resource) => mutator(resource.field_contracts),
    callback,
  );
}

function withSiteResourceFixture(mutator, callback) {
  withResourceContractFixture(
    DCIM_RELATIVE_PATH,
    "Site",
    SITE_FIELD_CONTRACTS,
    mutator,
    callback,
  );
}

function withManufacturerResourceFixture(mutator, callback) {
  withResourceContractFixture(
    DCIM_RELATIVE_PATH,
    "Manufacturer",
    MANUFACTURER_FIELD_CONTRACTS,
    mutator,
    callback,
  );
}

function withRackRoleResourceFixture(mutator, callback) {
  withResourceContractFixture(
    DCIM_RELATIVE_PATH,
    "RackRole",
    RACK_ROLE_FIELD_CONTRACTS,
    mutator,
    callback,
  );
}

function withRackTypeResourceFixture(mutator, callback) {
  withResourceContractFixture(
    DCIM_RELATIVE_PATH,
    "RackType",
    RACK_TYPE_FIELD_CONTRACTS,
    mutator,
    callback,
  );
}

function assertProfileRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    withContractFixture(mutator, (profilePath) => {
      const result = runNode(
        "scripts/validate_capability_profile.mjs",
        profilePath,
      );
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the malformed field contract",
      );
    });
  });
}

function assertSiteProfileRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    withSiteResourceFixture(mutator, (profilePath) => {
      const result = runNode(
        "scripts/validate_capability_profile.mjs",
        profilePath,
      );
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the malformed Site field contract",
      );
    });
  });
}

function assertManufacturerProfileRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    withManufacturerResourceFixture(mutator, (profilePath) => {
      const result = runNode(
        "scripts/validate_capability_profile.mjs",
        profilePath,
      );
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the malformed Manufacturer field contract",
      );
    });
  });
}

function assertRackRoleProfileRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    withRackRoleResourceFixture(mutator, (profilePath) => {
      const result = runNode(
        "scripts/validate_capability_profile.mjs",
        profilePath,
      );
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the malformed RackRole field contract",
      );
    });
  });
}

function assertRackTypeProfileRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    withRackTypeResourceFixture(mutator, (profilePath) => {
      const result = runNode(
        "scripts/validate_capability_profile.mjs",
        profilePath,
      );
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the malformed RackType field contract",
      );
    });
  });
}

function assertOpenAPIRejected(name, mutator, expectedDiagnostic) {
  test(name, () => {
    const spec = JSON.parse(fs.readFileSync(OPENAPI_PATH, "utf8"));
    mutator(spec);
    const fixtureRoot = fs.mkdtempSync(
      path.join(os.tmpdir(), "netbox-go-openapi-"),
    );
    const specPath = path.join(fixtureRoot, "openapi.json");
    try {
      fs.writeFileSync(specPath, `${JSON.stringify(spec, null, 2)}\n`);
      const result = runNode("scripts/validate_openapi.mjs", specPath);
      assert.notEqual(
        result.status,
        0,
        `expected rejection; ${diagnostics(result)}`,
      );
      assert.match(
        diagnostics(result),
        expectedDiagnostic,
        "validator must identify the stale generated API contract",
      );
    } finally {
      fs.rmSync(fixtureRoot, { recursive: true, force: true });
    }
  });
}

test("current capability metadata validates", () => {
  const metadata = JSON.parse(
    fs.readFileSync(path.join(CONTRACT_ROOT, IPAM_RELATIVE_PATH), "utf8"),
  );
  const ipAddress = metadata.resources.find(
    (resource) => resource.name === "IPAddress",
  );
  assert.ok(ipAddress, "current metadata must contain IPAddress");
  assert.deepStrictEqual(
    ipAddress.field_contracts,
    IP_ADDRESS_FIELD_CONTRACTS,
    "current IPAddress presence metadata must match the independently pinned contract",
  );
  const dcimMetadata = JSON.parse(
    fs.readFileSync(path.join(CONTRACT_ROOT, DCIM_RELATIVE_PATH), "utf8"),
  );
  const site = dcimMetadata.resources.find(
    (resource) => resource.name === "Site",
  );
  assert.ok(site, "current metadata must contain Site");
  assert.deepStrictEqual(
    site.field_contracts,
    SITE_FIELD_CONTRACTS,
    "current Site presence metadata must match the independently pinned contract",
  );
  const manufacturer = dcimMetadata.resources.find(
    (resource) => resource.name === "Manufacturer",
  );
  assert.ok(manufacturer, "current metadata must contain Manufacturer");
  assert.deepStrictEqual(
    manufacturer.field_contracts,
    MANUFACTURER_FIELD_CONTRACTS,
    "current Manufacturer presence metadata must match the independently pinned contract",
  );
  const rackRole = dcimMetadata.resources.find(
    (resource) => resource.name === "RackRole",
  );
  assert.ok(rackRole, "current metadata must contain RackRole");
  assert.deepStrictEqual(
    rackRole.field_contracts,
    RACK_ROLE_FIELD_CONTRACTS,
    "current RackRole presence metadata must match the independently pinned contract",
  );
  const rackType = dcimMetadata.resources.find(
    (resource) => resource.name === "RackType",
  );
  assert.ok(rackType, "current metadata must contain RackType");
  assert.deepStrictEqual(
    rackType.field_contracts,
    RACK_TYPE_FIELD_CONTRACTS,
    "current RackType presence metadata must match the independently pinned contract",
  );

  const result = runNode(
    "scripts/validate_capability_profile.mjs",
    path.join(CONTRACT_ROOT, PROFILE_RELATIVE_PATH),
  );
  assert.equal(result.status, 0, diagnostics(result));
});

assertSiteProfileRejected(
  "Site field contracts are mandatory",
  (site) => {
    delete site.field_contracts;
  },
  /dcim\.Site: field_contracts must declare operation-specific presence semantics/u,
);

assertManufacturerProfileRejected(
  "Manufacturer field contracts are mandatory",
  (manufacturer) => {
    delete manufacturer.field_contracts;
  },
  /dcim\.Manufacturer: field_contracts must declare operation-specific presence semantics/u,
);

assertManufacturerProfileRejected(
  "Manufacturer field contracts require every operation",
  (manufacturer) => {
    delete manufacturer.field_contracts.replace;
  },
  /dcim\.Manufacturer: field_contracts must declare exactly response, create, replace, and update/u,
);

assertManufacturerProfileRejected(
  "Manufacturer operation fields reject undeclared fields",
  (manufacturer) => {
    manufacturer.field_contracts.create.required_fields.push("tags");
  },
  /dcim\.Manufacturer: field_contracts\.create\.required_fields contains undeclared field tags/u,
);

assertManufacturerProfileRejected(
  "Manufacturer nullable fields remain a writable-field subset",
  (manufacturer) => {
    manufacturer.field_contracts.replace.nullable_fields.push("display");
  },
  /dcim\.Manufacturer: field_contracts\.replace\.nullable_fields contains undeclared field display/u,
);

assertManufacturerProfileRejected(
  "Manufacturer blank fields remain a writable-field subset",
  (manufacturer) => {
    manufacturer.field_contracts.update.blank_fields.push("devicetype_count");
  },
  /dcim\.Manufacturer: field_contracts\.update\.blank_fields contains undeclared field devicetype_count/u,
);

assertManufacturerProfileRejected(
  "Manufacturer operation fields reject duplicates",
  (manufacturer) => {
    manufacturer.field_contracts.create.blank_fields.push("description");
  },
  /dcim\.Manufacturer: field_contracts\.create\.blank_fields contains duplicate field description/u,
);

assertManufacturerProfileRejected(
  "Manufacturer response and write contracts cannot be conflated",
  (manufacturer) => {
    manufacturer.field_contracts.response = structuredClone(
      manufacturer.field_contracts.create,
    );
  },
  /dcim\.Manufacturer: field_contracts\.response must declare exactly nullable_fields/u,
);

assertManufacturerProfileRejected(
  "Manufacturer PATCH contracts cannot require fields",
  (manufacturer) => {
    manufacturer.field_contracts.update.required_fields.push("name");
  },
  /dcim\.Manufacturer: field_contracts\.update\.required_fields must be empty for PATCH/u,
);

assertRackRoleProfileRejected(
  "RackRole field contracts are mandatory",
  (rackRole) => {
    delete rackRole.field_contracts;
  },
  /dcim\.RackRole: field_contracts must declare operation-specific presence semantics/u,
);

assertRackRoleProfileRejected(
  "RackRole field contracts require every operation",
  (rackRole) => {
    delete rackRole.field_contracts.replace;
  },
  /dcim\.RackRole: field_contracts must declare exactly response, create, replace, and update/u,
);

assertRackRoleProfileRejected(
  "RackRole operation fields reject undeclared fields",
  (rackRole) => {
    rackRole.field_contracts.create.required_fields.push("tags");
  },
  /dcim\.RackRole: field_contracts\.create\.required_fields contains undeclared field tags/u,
);

assertRackRoleProfileRejected(
  "RackRole nullable fields remain a writable-field subset",
  (rackRole) => {
    rackRole.field_contracts.replace.nullable_fields.push("display");
  },
  /dcim\.RackRole: field_contracts\.replace\.nullable_fields contains undeclared field display/u,
);

assertRackRoleProfileRejected(
  "RackRole blank fields remain a writable-field subset",
  (rackRole) => {
    rackRole.field_contracts.update.blank_fields.push("rack_count");
  },
  /dcim\.RackRole: field_contracts\.update\.blank_fields contains undeclared field rack_count/u,
);

assertRackRoleProfileRejected(
  "RackRole operation fields reject duplicates",
  (rackRole) => {
    rackRole.field_contracts.create.blank_fields.push("description");
  },
  /dcim\.RackRole: field_contracts\.create\.blank_fields contains duplicate field description/u,
);

assertRackRoleProfileRejected(
  "RackRole response and write contracts cannot be conflated",
  (rackRole) => {
    rackRole.field_contracts.response = structuredClone(
      rackRole.field_contracts.create,
    );
  },
  /dcim\.RackRole: field_contracts\.response must declare exactly nullable_fields/u,
);

assertRackRoleProfileRejected(
  "RackRole PATCH contracts cannot require fields",
  (rackRole) => {
    rackRole.field_contracts.update.required_fields.push("name");
  },
  /dcim\.RackRole: field_contracts\.update\.required_fields must be empty for PATCH/u,
);

assertRackTypeProfileRejected(
  "RackType field contracts are mandatory",
  (rackType) => {
    delete rackType.field_contracts;
  },
  /dcim\.RackType: field_contracts must declare operation-specific presence semantics/u,
);

assertRackTypeProfileRejected(
  "RackType field contracts require every operation",
  (rackType) => {
    delete rackType.field_contracts.replace;
  },
  /dcim\.RackType: field_contracts must declare exactly response, create, replace, and update/u,
);

assertRackTypeProfileRejected(
  "RackType operation fields reject undeclared fields",
  (rackType) => {
    rackType.field_contracts.create.required_fields.push("outer_width");
  },
  /dcim\.RackType: field_contracts\.create\.required_fields contains undeclared field outer_width/u,
);

assertRackTypeProfileRejected(
  "RackType nullable fields remain a writable-field subset",
  (rackType) => {
    rackType.field_contracts.replace.nullable_fields.push("display");
  },
  /dcim\.RackType: field_contracts\.replace\.nullable_fields contains undeclared field display/u,
);

assertRackTypeProfileRejected(
  "RackType blank fields reject non-choice integer height fields",
  (rackType) => {
    rackType.field_contracts.update.blank_fields.push("u_height");
  },
  /dcim\.RackType: field_contracts\.update\.blank_fields contains non-string field u_height/u,
);

assertRackTypeProfileRejected(
  "RackType blank fields reject non-choice integer starting-unit fields",
  (rackType) => {
    rackType.field_contracts.create.blank_fields.push("starting_unit");
  },
  /dcim\.RackType: field_contracts\.create\.blank_fields contains non-string field starting_unit/u,
);

assertRackTypeProfileRejected(
  "RackType blank fields reject non-choice boolean fields",
  (rackType) => {
    rackType.field_contracts.replace.blank_fields.push("desc_units");
  },
  /dcim\.RackType: field_contracts\.replace\.blank_fields contains non-string field desc_units/u,
);

assertRackTypeProfileRejected(
  "RackType operation fields reject duplicates",
  (rackType) => {
    rackType.field_contracts.create.blank_fields.push("description");
  },
  /dcim\.RackType: field_contracts\.create\.blank_fields contains duplicate field description/u,
);

assertRackTypeProfileRejected(
  "RackType response and write contracts cannot be conflated",
  (rackType) => {
    rackType.field_contracts.response = structuredClone(
      rackType.field_contracts.create,
    );
  },
  /dcim\.RackType: field_contracts\.response must declare exactly nullable_fields/u,
);

assertRackTypeProfileRejected(
  "RackType PATCH contracts cannot require fields",
  (rackType) => {
    rackType.field_contracts.update.required_fields.push("manufacturer");
  },
  /dcim\.RackType: field_contracts\.update\.required_fields must be empty for PATCH/u,
);

assertSiteProfileRejected(
  "Site field contracts require every operation",
  (site) => {
    delete site.field_contracts.replace;
  },
  /dcim\.Site: field_contracts must declare exactly response, create, replace, and update/u,
);

assertSiteProfileRejected(
  "Site operation fields reject undeclared fields",
  (site) => {
    site.field_contracts.create.required_fields.push("region");
  },
  /dcim\.Site: field_contracts\.create\.required_fields contains undeclared field region/u,
);

assertSiteProfileRejected(
  "Site nullable fields remain a writable-field subset",
  (site) => {
    site.field_contracts.replace.nullable_fields.push("display");
  },
  /dcim\.Site: field_contracts\.replace\.nullable_fields contains undeclared field display/u,
);

assertSiteProfileRejected(
  "Site blank fields remain a writable-field subset",
  (site) => {
    site.field_contracts.update.blank_fields.push("device_count");
  },
  /dcim\.Site: field_contracts\.update\.blank_fields contains undeclared field device_count/u,
);

assertSiteProfileRejected(
  "Site operation fields reject duplicates",
  (site) => {
    site.field_contracts.create.blank_fields.push("facility");
  },
  /dcim\.Site: field_contracts\.create\.blank_fields contains duplicate field facility/u,
);

assertSiteProfileRejected(
  "Site response and write contracts cannot be conflated",
  (site) => {
    site.field_contracts.response = structuredClone(
      site.field_contracts.create,
    );
  },
  /dcim\.Site: field_contracts\.response must declare exactly nullable_fields/u,
);

assertSiteProfileRejected(
  "Site PATCH contracts cannot require fields",
  (site) => {
    site.field_contracts.update.required_fields.push("name");
  },
  /dcim\.Site: field_contracts\.update\.required_fields must be empty for PATCH/u,
);

assertProfileRejected(
  "field contracts require every operation",
  (contracts) => {
    delete contracts.update;
  },
  /field_contracts must declare exactly response, create, replace, and update/u,
);

assertProfileRejected(
  "operation field lists reject undeclared fields",
  (contracts) => {
    contracts.create.required_fields.push("family");
  },
  /create\.required_fields contains undeclared field family/u,
);

assertProfileRejected(
  "nullable fields must be a writable-field subset",
  (contracts) => {
    contracts.create.nullable_fields.push("display");
  },
  /create\.nullable_fields contains undeclared field display/u,
);

assertProfileRejected(
  "blank fields must be a writable-field subset",
  (contracts) => {
    contracts.replace.blank_fields.push("display");
  },
  /replace\.blank_fields contains undeclared field display/u,
);

assertProfileRejected(
  "operation field lists reject duplicates",
  (contracts) => {
    contracts.create.nullable_fields.push("role");
  },
  /create\.nullable_fields contains duplicate field role/u,
);

assertProfileRejected(
  "blank fields must be string-valued",
  (contracts) => {
    contracts.create.blank_fields.push("assigned_object_id");
  },
  /create\.blank_fields contains non-string field assigned_object_id/u,
);

assertProfileRejected(
  "response and write contracts cannot be conflated",
  (contracts) => {
    contracts.response = structuredClone(contracts.create);
  },
  /response must declare exactly nullable_fields/u,
);

assertProfileRejected(
  "PATCH contracts cannot require fields",
  (contracts) => {
    contracts.update.required_fields.push("address");
  },
  /field_contracts\.update\.required_fields must be empty for PATCH/u,
);

for (const [name, pathName, method, expectedSchema] of [
  ["POST", "/api/ipam/ip-addresses/", "post", "IPAddressCreate"],
  ["PUT", "/api/ipam/ip-addresses/{id}/", "put", "IPAddressReplace"],
  ["PATCH", "/api/ipam/ip-addresses/{id}/", "patch", "IPAddressUpdate"],
]) {
  assertOpenAPIRejected(
    `OpenAPI validation rejects a stale ${name} request reference`,
    (spec) => {
      spec.paths[pathName][method].requestBody.content[
        "application/json"
      ].schema = { $ref: "#/components/schemas/IPAddressWrite" };
    },
    new RegExp(
      escapeRegExp(
        `${name} ${pathName} must reference #/components/schemas/${expectedSchema}`,
      ),
      "u",
    ),
  );
}

for (const [name, pathName, method, expectedSchema] of [
  ["POST", "/api/dcim/manufacturers/", "post", "ManufacturerCreate"],
  ["PUT", "/api/dcim/manufacturers/{id}/", "put", "ManufacturerReplace"],
  ["PATCH", "/api/dcim/manufacturers/{id}/", "patch", "ManufacturerUpdate"],
]) {
  assertOpenAPIRejected(
    `OpenAPI validation rejects a stale Manufacturer ${name} request reference`,
    (spec) => {
      spec.paths[pathName][method].requestBody.content[
        "application/json"
      ].schema = { $ref: "#/components/schemas/ManufacturerWrite" };
    },
    new RegExp(
      escapeRegExp(
        `${name} ${pathName} must reference #/components/schemas/${expectedSchema}`,
      ),
      "u",
    ),
  );
}

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer create requiredness",
  (spec) => {
    const schema =
      spec.components.schemas.ManufacturerCreate ??
      structuredClone(spec.components.schemas.ManufacturerWrite);
    schema.required = [];
    spec.components.schemas.ManufacturerCreate = schema;
  },
  /ManufacturerCreate required fields must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer request nullability",
  (spec) => {
    const schema =
      spec.components.schemas.ManufacturerReplace ??
      structuredClone(spec.components.schemas.ManufacturerWrite);
    schema.properties.description.type = ["string", "null"];
    spec.components.schemas.ManufacturerReplace = schema;
  },
  /ManufacturerReplace\.description nullability must match field_contracts\.replace/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer request blankability",
  (spec) => {
    const schema =
      spec.components.schemas.ManufacturerUpdate ??
      structuredClone(spec.components.schemas.ManufacturerWrite);
    schema.properties.description.minLength = 1;
    spec.components.schemas.ManufacturerUpdate = schema;
  },
  /ManufacturerUpdate\.description blankability must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer request scalar types",
  (spec) => {
    const schema =
      spec.components.schemas.ManufacturerCreate ??
      structuredClone(spec.components.schemas.ManufacturerWrite);
    schema.properties.name.type = "integer";
    spec.components.schemas.ManufacturerCreate = schema;
  },
  /ManufacturerCreate\.name base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer response timestamp nullability",
  (spec) => {
    spec.components.schemas.Manufacturer.properties.created.type = "string";
  },
  /Manufacturer\.created response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Manufacturer response scalar types",
  (spec) => {
    spec.components.schemas.Manufacturer.properties.devicetype_count.type =
      "string";
  },
  /Manufacturer\.devicetype_count base type must be integer/u,
);

for (const [name, pathName, method, expectedSchema] of [
  ["POST", "/api/dcim/rack-roles/", "post", "RackRoleCreate"],
  ["PUT", "/api/dcim/rack-roles/{id}/", "put", "RackRoleReplace"],
  ["PATCH", "/api/dcim/rack-roles/{id}/", "patch", "RackRoleUpdate"],
]) {
  assertOpenAPIRejected(
    `OpenAPI validation rejects a stale RackRole ${name} request reference`,
    (spec) => {
      spec.paths[pathName][method].requestBody.content[
        "application/json"
      ].schema = { $ref: "#/components/schemas/RackRoleWrite" };
    },
    new RegExp(
      escapeRegExp(
        `${name} ${pathName} must reference #/components/schemas/${expectedSchema}`,
      ),
      "u",
    ),
  );
}

assertOpenAPIRejected(
  "OpenAPI validation rejects a stale shared RackRole write schema",
  (spec) => {
    spec.components.schemas.RackRoleWrite = structuredClone(
      spec.components.schemas.RackRoleCreate ??
        spec.components.schemas.RackRoleWrite,
    );
  },
  /RackRole must not retain a conflated shared write schema/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole create requiredness",
  (spec) => {
    const schema =
      spec.components.schemas.RackRoleCreate ??
      structuredClone(spec.components.schemas.RackRoleWrite);
    schema.required = [];
    spec.components.schemas.RackRoleCreate = schema;
  },
  /RackRoleCreate required fields must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole request nullability",
  (spec) => {
    const schema =
      spec.components.schemas.RackRoleReplace ??
      structuredClone(spec.components.schemas.RackRoleWrite);
    schema.properties.color.type = ["string", "null"];
    spec.components.schemas.RackRoleReplace = schema;
  },
  /RackRoleReplace\.color nullability must match field_contracts\.replace/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole request blankability",
  (spec) => {
    const schema =
      spec.components.schemas.RackRoleUpdate ??
      structuredClone(spec.components.schemas.RackRoleWrite);
    schema.properties.description.minLength = 1;
    spec.components.schemas.RackRoleUpdate = schema;
  },
  /RackRoleUpdate\.description blankability must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole request scalar types",
  (spec) => {
    const schema =
      spec.components.schemas.RackRoleCreate ??
      structuredClone(spec.components.schemas.RackRoleWrite);
    schema.properties.color.type = "integer";
    spec.components.schemas.RackRoleCreate = schema;
  },
  /RackRoleCreate\.color base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole response timestamp nullability",
  (spec) => {
    spec.components.schemas.RackRole.properties.created.type = "string";
  },
  /RackRole\.created response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackRole response scalar types",
  (spec) => {
    spec.components.schemas.RackRole.properties.rack_count.type = "string";
  },
  /RackRole\.rack_count base type must be integer/u,
);

for (const [name, pathName, method, expectedSchema] of [
  ["POST", "/api/dcim/rack-types/", "post", "RackTypeCreate"],
  ["PUT", "/api/dcim/rack-types/{id}/", "put", "RackTypeReplace"],
  ["PATCH", "/api/dcim/rack-types/{id}/", "patch", "RackTypeUpdate"],
]) {
  assertOpenAPIRejected(
    `OpenAPI validation rejects a stale RackType ${name} request reference`,
    (spec) => {
      spec.paths[pathName][method].requestBody.content[
        "application/json"
      ].schema = { $ref: "#/components/schemas/RackTypeWrite" };
    },
    new RegExp(
      escapeRegExp(
        `${name} ${pathName} must reference #/components/schemas/${expectedSchema}`,
      ),
      "u",
    ),
  );
}

assertOpenAPIRejected(
  "OpenAPI validation rejects a stale shared RackType write schema",
  (spec) => {
    spec.components.schemas.RackTypeWrite = structuredClone(
      spec.components.schemas.RackTypeCreate,
    );
  },
  /RackType must not retain a conflated shared write schema/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType create requiredness",
  (spec) => {
    spec.components.schemas.RackTypeCreate.required = [
      "manufacturer",
      "model",
      "slug",
    ];
  },
  /RackTypeCreate required fields must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType replace requiredness",
  (spec) => {
    spec.components.schemas.RackTypeReplace.required.push("form_factor");
  },
  /RackTypeReplace required fields must match field_contracts\.replace/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType request nullability",
  (spec) => {
    spec.components.schemas.RackTypeUpdate.properties.manufacturer.type = [
      "integer",
      "null",
    ];
  },
  /RackTypeUpdate\.manufacturer nullability must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType request blankability",
  (spec) => {
    spec.components.schemas.RackTypeUpdate.properties.description.minLength = 1;
  },
  /RackTypeUpdate\.description blankability must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType Manufacturer identifier types",
  (spec) => {
    spec.components.schemas.RackTypeCreate.properties.manufacturer.type =
      "string";
    delete spec.components.schemas.RackTypeCreate.properties.manufacturer
      .format;
  },
  /RackTypeCreate\.manufacturer base type must be integer/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType text scalar types",
  (spec) => {
    spec.components.schemas.RackTypeReplace.properties.model.type = "integer";
  },
  /RackTypeReplace\.model base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType numeric scalar types",
  (spec) => {
    spec.components.schemas.RackTypeUpdate.properties.starting_unit.type =
      "string";
  },
  /RackTypeUpdate\.starting_unit base type must be integer/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType integer choice types",
  (spec) => {
    spec.components.schemas.RackTypeCreate.properties.width.type = "string";
    delete spec.components.schemas.RackTypeCreate.properties.width.format;
  },
  /RackTypeCreate\.width base type must be integer/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects nullable RackType Manufacturer responses",
  (spec) => {
    spec.components.schemas.RackType.properties.manufacturer.oneOf.push({
      type: "null",
    });
  },
  /RackType\.manufacturer response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType Manufacturer response references",
  (spec) => {
    spec.components.schemas.RackType.properties.manufacturer.oneOf[0].$ref =
      "#/components/schemas/RackType";
  },
  /RackType\.manufacturer must reference #\/components\/schemas\/ObjectReference/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType response timestamp nullability",
  (spec) => {
    spec.components.schemas.RackType.properties.created.type = "string";
  },
  /RackType\.created response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale RackType response scalar types",
  (spec) => {
    spec.components.schemas.RackType.properties.desc_units.type = "string";
  },
  /RackType\.desc_units base type must be boolean/u,
);

for (const [name, pathName, method, expectedSchema] of [
  ["POST", "/api/dcim/sites/", "post", "SiteCreate"],
  ["PUT", "/api/dcim/sites/{id}/", "put", "SiteReplace"],
  ["PATCH", "/api/dcim/sites/{id}/", "patch", "SiteUpdate"],
]) {
  assertOpenAPIRejected(
    `OpenAPI validation rejects a stale Site ${name} request reference`,
    (spec) => {
      spec.paths[pathName][method].requestBody.content[
        "application/json"
      ].schema = { $ref: "#/components/schemas/SiteWrite" };
    },
    new RegExp(
      escapeRegExp(
        `${name} ${pathName} must reference #/components/schemas/${expectedSchema}`,
      ),
      "u",
    ),
  );
}

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site create requiredness",
  (spec) => {
    const schema =
      spec.components.schemas.SiteCreate ??
      structuredClone(spec.components.schemas.SiteWrite);
    schema.required = [];
    spec.components.schemas.SiteCreate = schema;
  },
  /SiteCreate required fields must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site request nullability",
  (spec) => {
    const schema =
      spec.components.schemas.SiteReplace ??
      structuredClone(spec.components.schemas.SiteWrite);
    schema.properties.facility.type = ["string", "null"];
    spec.components.schemas.SiteReplace = schema;
  },
  /SiteReplace\.facility nullability must match field_contracts\.replace/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site request blankability",
  (spec) => {
    const schema =
      spec.components.schemas.SiteUpdate ??
      structuredClone(spec.components.schemas.SiteWrite);
    schema.properties.description.minLength = 1;
    spec.components.schemas.SiteUpdate = schema;
  },
  /SiteUpdate\.description blankability must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site request scalar types",
  (spec) => {
    const schema =
      spec.components.schemas.SiteCreate ??
      structuredClone(spec.components.schemas.SiteWrite);
    schema.properties.name.type = "integer";
    spec.components.schemas.SiteCreate = schema;
  },
  /SiteCreate\.name base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site response timestamp nullability",
  (spec) => {
    spec.components.schemas.Site.properties.created.type = "string";
  },
  /Site\.created response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale Site response scalar types",
  (spec) => {
    spec.components.schemas.Site.properties.rack_count.type = "string";
  },
  /Site\.rack_count base type must be integer/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale create requiredness",
  (spec) => {
    spec.components.schemas.IPAddressCreate.required = [];
  },
  /IPAddressCreate required fields must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects required PATCH fields",
  (spec) => {
    spec.components.schemas.IPAddressUpdate.required = ["address"];
  },
  /IPAddressUpdate required fields must match field_contracts\.update/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale request nullability",
  (spec) => {
    spec.components.schemas.IPAddressCreate.properties.role.type = "string";
  },
  /IPAddressCreate\.role nullability must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale request scalar types",
  (spec) => {
    spec.components.schemas.IPAddressCreate.properties.address.type = "integer";
  },
  /IPAddressCreate\.address base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale request choice types",
  (spec) => {
    spec.components.schemas.IPAddressCreate.properties.status.type = "integer";
  },
  /IPAddressCreate\.status base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale integer formats",
  (spec) => {
    spec.components.schemas.IPAddressCreate.properties.assigned_object_id.format =
      "int32";
  },
  /IPAddressCreate\.assigned_object_id format must be int64/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale request blankability",
  (spec) => {
    spec.components.schemas.IPAddressCreate.properties.role.minLength = 2;
  },
  /IPAddressCreate\.role blankability must match field_contracts\.create/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects missing nonblank constraints",
  (spec) => {
    delete spec.components.schemas.IPAddressReplace.properties.address
      .minLength;
  },
  /IPAddressReplace\.address blankability must match field_contracts\.replace/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale response nullability",
  (spec) => {
    spec.components.schemas.IPAddress.properties.display.type = [
      "string",
      "null",
    ];
  },
  /IPAddress\.display response nullability must match field_contracts\.response/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale response scalar types",
  (spec) => {
    spec.components.schemas.IPAddress.properties.display.type = "integer";
  },
  /IPAddress\.display base type must be string/u,
);

assertOpenAPIRejected(
  "OpenAPI validation rejects stale response choice types",
  (spec) => {
    spec.components.schemas.IPAddress.properties.family.properties.value.type =
      "string";
  },
  /IPAddress\.family\.value base type must be integer/u,
);
