import fs from "node:fs/promises";
import path from "node:path";
import process from "node:process";

import {
  assertSemanticEqual,
  IdentifierBindings,
  normalizeDenial,
  normalizeValidation,
  projectPage,
  projectResource,
} from "./comparator.mjs";

const repositoryRoot = path.resolve(import.meta.dirname, "../..");
const oracleURL = requiredEnvironment("NETBOX_ORACLE_URL").replace(/\/$/, "");
const goURL = requiredEnvironment("NETBOX_GO_URL").replace(/\/$/, "");
const oracleToken = requiredEnvironment("NETBOX_ORACLE_TOKEN");
const goUsername = requiredEnvironment("NETBOX_GO_USERNAME");
const goPassword = requiredEnvironment("NETBOX_GO_PASSWORD");
const artifactDirectory = requiredEnvironment("NETBOX_COMPAT_ARTIFACT_DIR");

const exchanges = [];
const semanticFailures = [];
const timestampTimeline = new Map();
const identifiers = new IdentifierBindings();
let assertionCount = 0;

function requiredEnvironment(name) {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
}

function redact(value) {
  if (Array.isArray(value)) return value.map(redact);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).map(([key, entry]) => [
        key,
        ["secret", "password", "token", "authorization", "cookie"].includes(key.toLowerCase())
          ? "<redacted>"
          : redact(entry),
      ]),
    );
  }
  return value;
}

class API {
  constructor(name, baseURL, token = "") {
    this.name = name;
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(
    method,
    resourcePath,
    { body, authenticated = true, headers = {}, redirect = "follow" } = {},
  ) {
    const requestHeaders = { Accept: "application/json", ...headers };
    if (body !== undefined) requestHeaders["Content-Type"] = "application/json";
    if (authenticated && this.token) requestHeaders.Authorization = `Token ${this.token}`;
    const response = await fetch(`${this.baseURL}${resourcePath}`, {
      method,
      headers: requestHeaders,
      body: body === undefined ? undefined : JSON.stringify(body),
      redirect,
      signal: AbortSignal.timeout(30_000),
    });
    const raw = await response.text();
    let parsed = null;
    if (raw !== "") {
      try {
        parsed = JSON.parse(raw);
      } catch {
        parsed = raw;
      }
    }
    const result = { status: response.status, body: parsed, headers: response.headers };
    exchanges.push({
      system: this.name,
      method,
      path: resourcePath,
      request: redact(body),
      status: response.status,
      location: response.headers.get("location"),
      response: redact(parsed),
    });
    return result;
  }
}

function compareValidationResponses(label, oracleResponse, goResponse, statuses = [400]) {
  expectStatus(`oracle ${label}`, oracleResponse, statuses);
  expectStatus(`Go ${label}`, goResponse, statuses);
  compare(label, normalizeValidation(oracleResponse), normalizeValidation(goResponse));
}

function redirectOutcome(response) {
  let location = response.headers.get("location");
  if (location !== null) {
    const parsed = new URL(location, "http://compatibility.invalid");
    location = `${parsed.pathname}${parsed.search}`;
  }
  return { status: response.status, location };
}

function expectGoProfileFieldRejection(label, response, field) {
  expectStatus(`Go ${label}`, response, [400]);
  compare(
    `Go ${label} reason`,
    {
      status: 400,
      body: { [field]: ["This field is not supported by the active capability profile."] },
    },
    normalizeValidation(response),
  );
}

class CookieSession {
  constructor(baseURL) {
    this.baseURL = baseURL;
    this.cookies = new Map();
  }

  absorb(headers) {
    const setCookies = typeof headers.getSetCookie === "function" ? headers.getSetCookie() : [headers.get("set-cookie")].filter(Boolean);
    for (const line of setCookies) {
      const pair = line.split(";", 1)[0];
      const separator = pair.indexOf("=");
      if (separator < 1) continue;
      const name = pair.slice(0, separator);
      const value = pair.slice(separator + 1);
      if (value === "") this.cookies.delete(name);
      else this.cookies.set(name, value);
    }
  }

  cookieHeader() {
    return [...this.cookies.entries()].map(([name, value]) => `${name}=${value}`).join("; ");
  }

  async request(method, resourcePath, body) {
    const headers = { Accept: "application/json" };
    if (this.cookies.size > 0) headers.Cookie = this.cookieHeader();
    if (body !== undefined) headers["Content-Type"] = "application/json";
    if (!["GET", "HEAD", "OPTIONS"].includes(method) && this.cookies.has("csrftoken")) {
      headers["X-CSRFToken"] = this.cookies.get("csrftoken");
    }
    const response = await fetch(`${this.baseURL}${resourcePath}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: AbortSignal.timeout(30_000),
    });
    this.absorb(response.headers);
    const raw = await response.text();
    let parsed = null;
    if (raw !== "") {
      try {
        parsed = JSON.parse(raw);
      } catch {
        parsed = raw;
      }
    }
    exchanges.push({
      system: "go-session",
      method,
      path: resourcePath,
      request: redact(body),
      status: response.status,
      response: redact(parsed),
    });
    return { status: response.status, body: parsed, headers: response.headers };
  }
}

function expectStatus(label, response, expected) {
  if (!expected.includes(response.status)) {
    throw new Error(`${label}: expected HTTP ${expected.join(" or ")}, received ${response.status}: ${JSON.stringify(response.body)}`);
  }
  assertionCount += 1;
}

function compare(label, oracle, go) {
  try {
    assertSemanticEqual(label, oracle, go);
    assertionCount += 1;
  } catch (error) {
    semanticFailures.push({ label, message: error.message, stack: error.stack });
    console.error(`compatibility divergence: ${error.message}`);
  }
}

function recordFailure(label, error) {
  const message = error instanceof Error ? error.message : String(error);
  semanticFailures.push({ label, message, stack: error instanceof Error ? error.stack : undefined });
  console.error(`compatibility divergence: ${label}: ${message}`);
}

function expectedResponseFields(metadata, omitted = new Set()) {
  return [...metadata.writable_fields, ...metadata.response_only_fields].filter((field) => !omitted.has(field));
}

function assertExactGoEnvelope(label, body, fields, metadata = null) {
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    throw new Error(`Go response must be a resource object`);
  }
  const expected = [...fields].sort();
  const actual = Object.keys(body).sort();
  if (expected.join("\0") !== actual.join("\0")) {
    const missing = expected.filter((field) => !actual.includes(field));
    const undeclared = actual.filter((field) => !expected.includes(field));
    throw new Error(
      `Go field envelope drift; missing=[${missing.join(", ")}], undeclared=[${undeclared.join(", ")}]`,
    );
  }
  for (const [field, value] of Object.entries(body)) {
    if (!value || typeof value !== "object" || Array.isArray(value)) continue;
    const nestedFields = Object.keys(value).sort();
    const expectedNestedFields = metadata?.choice_fields?.[field]
      ? ["label", "value"]
      : ["display", "id", "url"];
    if (nestedFields.join("\0") !== expectedNestedFields.join("\0")) {
      throw new Error(
        `Go nested field envelope drift at ${field}; expected=[${expectedNestedFields.join(", ")}], actual=[${nestedFields.join(", ")}]`,
      );
    }
  }
  assertionCount += 1;
}

function parseRFC3339Nanoseconds(value, label) {
  if (typeof value !== "string") throw new Error(`${label}: expected an RFC3339 string`);
  const match = value.match(/^(.*:\d{2})(?:\.(\d{1,9}))?(Z|[+-]\d{2}:\d{2})$/);
  if (!match) throw new Error(`${label}: expected an RFC3339 timestamp, received ${JSON.stringify(value)}`);
  const epochMilliseconds = Date.parse(`${match[1]}${match[3]}`);
  if (!Number.isFinite(epochMilliseconds)) throw new Error(`${label}: invalid RFC3339 timestamp`);
  const fraction = BigInt((match[2] ?? "").padEnd(9, "0"));
  return BigInt(epochMilliseconds) * 1_000_000n + fraction;
}

function observeTimestamps(system, resourceName, phase, body, mutation = false) {
  const label = `${system} ${phase} ${resourceName}`;
  const created = parseRFC3339Nanoseconds(body?.created, `${label}.created`);
  const lastUpdated = parseRFC3339Nanoseconds(body?.last_updated, `${label}.last_updated`);
  if (created > lastUpdated) {
    throw new Error(`${label}: created must not be later than last_updated`);
  }
  const key = `${system}:${resourceName}`;
  const previous = timestampTimeline.get(key);
  if (previous) {
    if (created !== previous.created) {
      throw new Error(`${label}: created timestamp changed after creation`);
    }
    if (mutation && lastUpdated <= previous.lastUpdated) {
      throw new Error(`${label}: mutation did not advance last_updated`);
    }
    if (!mutation && lastUpdated !== previous.lastUpdated) {
      throw new Error(`${label}: read changed last_updated`);
    }
  }
  timestampTimeline.set(key, { created, lastUpdated });
  assertionCount += 2;
}

const createOmissions = {
  Site: ["device_count", "prefix_count", "rack_count"],
  Manufacturer: ["devicetype_count"],
  RackRole: ["rack_count"],
  Rack: ["device_count"],
  DeviceType: ["device_count"],
  VRF: ["ipaddress_count", "prefix_count"],
};

function compareCreateResource(label, oracleBody, goBody, metadata) {
  // RelatedObjectCountField annotations are absent only from NetBox's POST
  // serializer instance. This table is an asserted oracle contract, not a
  // dynamic intersection which could conceal response-shape drift.
  const omitted = new Set(createOmissions[metadata.name] ?? []);
  for (const field of omitted) {
    for (const [system, body] of [
      ["oracle", oracleBody],
      ["Go", goBody],
    ]) {
      if (Object.hasOwn(body, field)) {
        semanticFailures.push({
          label,
          message: `${system} ${metadata.name} create unexpectedly emitted ${field}`,
        });
      }
    }
  }
  const fields = expectedResponseFields(metadata, omitted);
  compareResourceProjection(label, oracleBody, goBody, metadata, fields);
}

function compareResourceProjection(label, oracleBody, goBody, metadata, fields = null) {
  try {
    const selectedFields = fields ?? expectedResponseFields(metadata);
    assertExactGoEnvelope(label, goBody, selectedFields, metadata);
    compare(
      label,
      projectResource(oracleBody, metadata, `oracle ${metadata.name}`, selectedFields, projectionContext("oracle", metadata)),
      projectResource(goBody, metadata, `Go ${metadata.name}`, selectedFields, projectionContext("go", metadata)),
    );
  } catch (error) {
    recordFailure(label, error);
  }
}

function comparePageProjection(label, oracleBody, goBody, metadata) {
  try {
    const expectedPageFields = ["count", "next", "previous", "results"];
    assertExactGoEnvelope(`${label} page`, goBody, expectedPageFields);
    for (const [index, entry] of goBody.results.entries()) {
      assertExactGoEnvelope(`${label} result ${index}`, entry, expectedResponseFields(metadata));
    }
    compare(
      label,
      projectPage(oracleBody, metadata, `oracle ${metadata.name} page`, projectionContext("oracle", metadata)),
      projectPage(goBody, metadata, `Go ${metadata.name} page`, projectionContext("go", metadata)),
    );
  } catch (error) {
    recordFailure(label, error);
  }
}

function projectionContext(system, metadata) {
  return {
    bindings: identifiers,
    system,
    referencePaths:
      metadata.name === "IPAddress"
        ? { assigned_object_id: "/api/dcim/interfaces/" }
        : {},
  };
}

function bindCreatedPair(symbol, metadata, oracleBody, goBody) {
  identifiers.bind(symbol, metadata.rest_path, oracleBody.id, goBody.id);
}

async function loadMetadata() {
  const files = [
    "contracts/netbox/v4.4.6-post7/resources/dcim.yaml",
    "contracts/netbox/v4.4.6-post7/resources/ipam.yaml",
  ];
  const resources = new Map();
  for (const file of files) {
    const document = JSON.parse(await fs.readFile(path.join(repositoryRoot, file), "utf8"));
    for (const resource of document.resources) resources.set(resource.name, resource);
  }
  if (resources.size !== 13) throw new Error(`expected 13 first-profile resources, found ${resources.size}`);
  return resources;
}

const definitions = [
  {
    name: "Site",
    payload: () => ({
      name: "Compatibility Site",
      slug: "compatibility-site",
      status: "active",
      facility: "TEST-DC",
      description: "Pinned oracle compatibility site",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "Manufacturer",
    payload: () => ({
      name: "Compatibility Networks",
      slug: "compatibility-networks",
      description: "Pinned oracle manufacturer",
    }),
  },
  {
    name: "RackRole",
    payload: () => ({
      name: "Compatibility Rack Role",
      slug: "compatibility-rack-role",
      color: "9e9e9e",
      description: "Pinned oracle rack role",
    }),
  },
  {
    name: "RackType",
    payload: (ids) => ({
      manufacturer: ids.Manufacturer,
      model: "Compatibility Rack Type",
      slug: "compatibility-rack-type",
      form_factor: "4-post-cabinet",
      width: 19,
      u_height: 42,
      starting_unit: 1,
      desc_units: false,
      description: "Pinned oracle rack type",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "Rack",
    payload: (ids) => ({
      site: ids.Site,
      name: "Compatibility Rack",
      facility_id: null,
      rack_type: ids.RackType,
      status: "active",
      role: ids.RackRole,
      serial: "COMPAT-RACK-SERIAL",
      asset_tag: "COMPAT-RACK-ASSET",
      airflow: "front-to-rear",
      description: "Pinned oracle rack",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "DeviceRole",
    payload: () => ({
      parent: null,
      name: "Compatibility Device Role",
      slug: "compatibility-device-role",
      color: "607d8b",
      vm_role: true,
      description: "Pinned oracle device role",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "DeviceType",
    payload: (ids) => ({
      manufacturer: ids.Manufacturer,
      model: "Compatibility Device Type",
      slug: "compatibility-device-type",
      part_number: "COMPAT-PN-1",
      u_height: 1,
      exclude_from_utilization: false,
      is_full_depth: true,
      airflow: "front-to-rear",
      description: "Pinned oracle device type",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "InterfaceTemplate",
    payload: (ids) => ({
      device_type: ids.DeviceType,
      name: "eth0",
      label: "",
      type: "1000base-t",
      enabled: true,
      mgmt_only: false,
      description: "Pinned oracle interface template",
    }),
  },
  {
    name: "Device",
    payload: (ids) => ({
      device_type: ids.DeviceType,
      role: ids.DeviceRole,
      name: "compatibility-device-01",
      site: ids.Site,
      rack: ids.Rack,
      position: 10,
      face: "front",
      status: "active",
      serial: "COMPAT-DEVICE-SERIAL",
      asset_tag: "COMPAT-DEVICE-ASSET",
      airflow: "front-to-rear",
      description: "Pinned oracle device",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "Interface",
    payload: (ids) => ({
      device: ids.Device,
      name: "eth1",
      label: "",
      type: "1000base-t",
      enabled: true,
      mgmt_only: false,
      mtu: 1500,
      speed: 1_000_000,
      duplex: "full",
      description: "Pinned oracle interface",
    }),
  },
  {
    name: "VRF",
    payload: () => ({
      name: "Compatibility VRF",
      rd: "65000:100",
      enforce_unique: true,
      description: "Pinned oracle VRF",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "Prefix",
    payload: (ids) => ({
      prefix: "198.51.100.0/24",
      vrf: ids.VRF,
      status: "active",
      is_pool: false,
      mark_utilized: false,
      description: "Pinned oracle prefix",
      comments: "core-workflow-v1",
    }),
  },
  {
    name: "IPAddress",
    payload: (ids) => ({
      address: "198.51.100.10/24",
      vrf: ids.VRF,
      status: "active",
      dns_name: "compatibility-device-01.example.test",
      description: "Pinned oracle IP address",
      comments: "core-workflow-v1",
      assigned_object_type: "dcim.interface",
      assigned_object_id: ids.Interface,
    }),
  },
];

function declaredFilterValues(name, ids) {
  const values = {
    Site: { name: "Compatibility Site", slug: "compatibility-site", status: "active" },
    Manufacturer: { name: "Compatibility Networks", slug: "compatibility-networks" },
    RackRole: { name: "Compatibility Rack Role", slug: "compatibility-rack-role" },
    RackType: {
      manufacturer_id: ids.Manufacturer,
      manufacturer_slug: "compatibility-networks",
      model: "Compatibility Rack Type",
      slug: "compatibility-rack-type",
    },
    Rack: {
      site_id: ids.Site,
      site_slug: "compatibility-site",
      name: "Compatibility Rack",
      status: "active",
      role_id: ids.RackRole,
      role_slug: "compatibility-rack-role",
      rack_type_id: ids.RackType,
      rack_type_slug: "compatibility-rack-type",
    },
    DeviceRole: { name: "Compatibility Device Role", slug: "compatibility-device-role" },
    DeviceType: {
      manufacturer_id: ids.Manufacturer,
      manufacturer_slug: "compatibility-networks",
      model: "Compatibility Device Type",
      slug: "compatibility-device-type",
    },
    InterfaceTemplate: {
      device_type_id: ids.DeviceType,
      name: "eth0",
      type: "1000base-t",
      enabled: true,
      mgmt_only: false,
    },
    Device: {
      site_id: ids.Site,
      site_slug: "compatibility-site",
      rack_id: ids.Rack,
      device_type_id: ids.DeviceType,
      device_type_slug: "compatibility-device-type",
      role_id: ids.DeviceRole,
      role_slug: "compatibility-device-role",
      name: "compatibility-device-01",
      status: "active",
    },
    Interface: {
      device_id: ids.Device,
      device_name: "compatibility-device-01",
      name: "eth1",
      type: "1000base-t",
      enabled: true,
      mgmt_only: false,
    },
    VRF: { name: "Compatibility VRF", rd: "65000:100", enforce_unique: true },
    Prefix: {
      vrf_id: ids.VRF,
      vrf_rd: "65000:100",
      prefix: "198.51.100.0/24",
      family: 4,
      status: "active",
      within: "198.51.0.0/16",
      within_include: "198.51.100.0/24",
      contains: "198.51.100.128/25",
    },
    IPAddress: {
      vrf_id: ids.VRF,
      vrf_rd: "65000:100",
      address: "198.51.100.10/24",
      family: 4,
      parent: "198.51.100.0/24",
      status: "active",
      assigned: true,
      interface_id: ids.Interface,
      device_id: ids.Device,
    },
  }[name];
  if (!values) throw new Error(`no declared filter values for ${name}`);
  return values;
}

async function createGoToken() {
  const session = new CookieSession(goURL);
  const csrf = await session.request("GET", "/api/auth/csrf/");
  expectStatus("Go CSRF bootstrap", csrf, [200]);
  if (!session.cookies.has("csrftoken")) throw new Error("Go CSRF endpoint did not set csrftoken");
  const login = await session.request("POST", "/api/auth/login/", { username: goUsername, password: goPassword });
  expectStatus("Go session login", login, [200]);
  if (!session.cookies.has("netbox_session") || !session.cookies.has("csrftoken")) {
    throw new Error("Go login did not establish the required session and CSRF cookies");
  }
  const current = await session.request("GET", "/api/auth/session/");
  expectStatus("Go current session", current, [200]);
  if (current.body?.user?.username !== goUsername || current.body?.user?.is_superuser !== true) {
    throw new Error(`Go session principal mismatch: ${JSON.stringify(current.body)}`);
  }
  const token = await session.request("POST", "/api/auth/tokens/", {
    description: "compatibility oracle write token",
    write_enabled: true,
    allowed_ips: ["127.0.0.0/8", "::1/128"],
  });
  expectStatus("Go token creation", token, [201]);
  if (typeof token.body?.secret !== "string" || token.body.secret.length < 20) {
    throw new Error("Go token creation did not return its one-time secret");
  }
  return { session, tokenID: token.body.id, secret: token.body.secret };
}

async function run() {
  await fs.mkdir(artifactDirectory, { recursive: true });
  const metadata = await loadMetadata();
  for (const definition of definitions) {
    if (!metadata.has(definition.name)) throw new Error(`profile metadata missing ${definition.name}`);
  }

  const oracle = new API("oracle", oracleURL, oracleToken);
  const go = new API("go", goURL);

  const [oracleDenied, goDenied] = await Promise.all([
    oracle.request("GET", "/api/dcim/sites/?limit=1", { authenticated: false }),
    go.request("GET", "/api/dcim/sites/?limit=1", { authenticated: false }),
  ]);
  compare("missing credentials are denied", normalizeDenial(oracleDenied), normalizeDenial(goDenied));

  const oracleAuthenticated = await oracle.request("GET", "/api/dcim/sites/?limit=1");
  expectStatus("pinned oracle token authentication", oracleAuthenticated, [200]);

  const goIdentity = await createGoToken();
  go.token = goIdentity.secret;
  const goAuthenticated = await go.request("GET", "/api/dcim/sites/?limit=1");
  expectStatus("Go token authentication", goAuthenticated, [200]);

  // APPEND_SLASH/redirect behavior is observable API behavior. Use manual
  // redirects so fetch cannot silently turn a mismatched route into a 200.
  for (const resourcePath of ["/api/dcim/sites?limit=1", "/api/dcim/sites/999999"]) {
    const [oracleRedirect, goRedirect] = await Promise.all([
      oracle.request("GET", resourcePath, { redirect: "manual" }),
      go.request("GET", resourcePath, { redirect: "manual" }),
    ]);
    compare(`trailing-slash redirect ${resourcePath}`, redirectOutcome(oracleRedirect), redirectOutcome(goRedirect));
  }

  // The missing-slug case is selected so both APIs identify exactly one
  // declared field; message wording remains implementation-local.
  const invalidSite = { name: "Intentionally Invalid Site" };
  const [oracleInvalidSite, goInvalidSite] = await Promise.all([
    oracle.request("POST", "/api/dcim/sites/", { body: invalidSite }),
    go.request("POST", "/api/dcim/sites/", { body: invalidSite }),
  ]);
  compareValidationResponses("required-field validation semantics", oracleInvalidSite, goInvalidSite);

  const unknownSite = {
    name: "Unknown Field Site",
    slug: "unknown-field-site",
    compatibility_unknown_field: "must-fail-closed",
  };
  const [oracleUnknownSite, goUnknownSite] = await Promise.all([
    oracle.request("POST", "/api/dcim/sites/", { body: unknownSite }),
    go.request("POST", "/api/dcim/sites/", { body: unknownSite }),
  ]);
  expectStatus("oracle ignores unknown request field outside its serializer", oracleUnknownSite, [201]);
  expectGoProfileFieldRejection(
    "unknown request field rejection",
    goUnknownSite,
    "compatibility_unknown_field",
  );
  const oracleUnknownDeleted = await oracle.request(
    "DELETE",
    `/api/dcim/sites/${oracleUnknownSite.body.id}/`,
  );
  expectStatus("oracle unknown-field fixture cleanup", oracleUnknownDeleted, [204]);

  // A real NetBox field outside core-workflow-v1 is explicitly deferred. The
  // oracle may accept it; the standalone profile boundary must reject it and
  // must not accidentally publish that field. Clean up an oracle object if it
  // was accepted so later comparisons retain identical profile state.
  const deferredSite = { name: "Deferred Field Site", slug: "deferred-field-site", tenant: null };
  const [oracleDeferredSite, goDeferredSite] = await Promise.all([
    oracle.request("POST", "/api/dcim/sites/", { body: deferredSite }),
    go.request("POST", "/api/dcim/sites/", { body: deferredSite }),
  ]);
  expectStatus("oracle accepts explicitly deferred request field", oracleDeferredSite, [201]);
  expectGoProfileFieldRejection("deferred request field rejection", goDeferredSite, "tenant");
  const oracleDeferredDeleted = await oracle.request(
    "DELETE",
    `/api/dcim/sites/${oracleDeferredSite.body.id}/`,
  );
  expectStatus("oracle deferred-field fixture cleanup", oracleDeferredDeleted, [204]);

  const nullSite = { name: "Null Field Site", slug: "null-field-site", description: null };
  const [oracleNullSite, goNullSite] = await Promise.all([
    oracle.request("POST", "/api/dcim/sites/", { body: nullSite }),
    go.request("POST", "/api/dcim/sites/", { body: nullSite }),
  ]);
  compareValidationResponses("explicit-null rejection", oracleNullSite, goNullSite);

  const siteMetadata = metadata.get("Site");
  const minimalSite = { name: "Minimal Default Site", slug: "minimal-default-site" };
  const [oracleMinimalSite, goMinimalSite] = await Promise.all([
    oracle.request("POST", siteMetadata.rest_path, { body: minimalSite }),
    go.request("POST", siteMetadata.rest_path, { body: minimalSite }),
  ]);
  expectStatus("oracle create minimal default Site", oracleMinimalSite, [201]);
  expectStatus("Go create minimal default Site", goMinimalSite, [201]);
  bindCreatedPair("Site.defaults", siteMetadata, oracleMinimalSite.body, goMinimalSite.body);
  compareCreateResource(
    "omitted fields receive oracle defaults",
    oracleMinimalSite.body,
    goMinimalSite.body,
    siteMetadata,
  );
  for (const [api, created] of [
    [oracle, oracleMinimalSite],
    [go, goMinimalSite],
  ]) {
    const deleted = await api.request("DELETE", `${siteMetadata.rest_path}${created.body.id}/`);
    expectStatus(`${api.name} minimal default Site cleanup`, deleted, [204]);
  }

  const ids = { oracle: {}, go: {} };
  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    if (definition.name === "Prefix") {
      const hostBitsPayload = (system) => ({
        prefix: "198.51.100.7/24",
        vrf: ids[system].VRF,
        status: "active",
      });
      const [oracleHostBits, goHostBits] = await Promise.all([
        oracle.request("POST", resourceMetadata.rest_path, { body: hostBitsPayload("oracle") }),
        go.request("POST", resourceMetadata.rest_path, { body: hostBitsPayload("go") }),
      ]);
      expectStatus("oracle rejects prefix host bits", oracleHostBits, [400]);
      expectStatus("Go rejects prefix host bits", goHostBits, [400]);
      compare(
        "prefix host-bit validation semantics",
        normalizeValidation(oracleHostBits),
        normalizeValidation(goHostBits),
      );
    }
    const oraclePayload = definition.payload(ids.oracle);
    const goPayload = definition.payload(ids.go);
    const [oracleCreated, goCreated] = await Promise.all([
      oracle.request("POST", resourceMetadata.rest_path, { body: oraclePayload }),
      go.request("POST", resourceMetadata.rest_path, { body: goPayload }),
    ]);
    expectStatus(`oracle create ${definition.name}`, oracleCreated, [201]);
    expectStatus(`Go create ${definition.name}`, goCreated, [201]);
    ids.oracle[definition.name] = oracleCreated.body.id;
    ids.go[definition.name] = goCreated.body.id;
    bindCreatedPair(`${definition.name}.primary`, resourceMetadata, oracleCreated.body, goCreated.body);
    compareCreateResource(`create ${definition.name}`, oracleCreated.body, goCreated.body, resourceMetadata);
    observeTimestamps("oracle", definition.name, "create", oracleCreated.body);
    observeTimestamps("Go", definition.name, "create", goCreated.body);
  }

  // Device creation instantiates InterfaceTemplates in both systems. Bind the
  // generated Interface explicitly by its declared scenario identity before
  // any page comparison; an unpaired generated identifier is a hard failure.
  const interfaceMetadata = metadata.get("Interface");
  const templateInterfaceQuery = (system) =>
    new URLSearchParams({
      device_id: String(ids[system].Device),
      name: "eth0",
      ordering: "id",
      limit: "2",
    }).toString();
  const [oracleTemplateInterfaces, goTemplateInterfaces] = await Promise.all([
    oracle.request("GET", `${interfaceMetadata.rest_path}?${templateInterfaceQuery("oracle")}`),
    go.request("GET", `${interfaceMetadata.rest_path}?${templateInterfaceQuery("go")}`),
  ]);
  expectStatus("oracle template-created Interface discovery", oracleTemplateInterfaces, [200]);
  expectStatus("Go template-created Interface discovery", goTemplateInterfaces, [200]);
  for (const [system, page] of [["oracle", oracleTemplateInterfaces.body], ["Go", goTemplateInterfaces.body]]) {
    if (page?.count !== 1 || page.results?.length !== 1 || page.results[0]?.name !== "eth0") {
      throw new Error(`${system} template-created Interface discovery was not unique`);
    }
    assertionCount += 1;
  }
  bindCreatedPair(
    "Interface.template.eth0",
    interfaceMetadata,
    oracleTemplateInterfaces.body.results[0],
    goTemplateInterfaces.body.results[0],
  );
  comparePageProjection(
    "template-created Interface",
    oracleTemplateInterfaces.body,
    goTemplateInterfaces.body,
    interfaceMetadata,
  );

  // Re-read every resource after the whole relationship graph exists. This
  // verifies computed counters, nested relationships, assignment projection,
  // canonical network values, and interface-template instantiation effects.
  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    const [oracleLoaded, goLoaded] = await Promise.all([
      oracle.request("GET", `${resourceMetadata.rest_path}${ids.oracle[definition.name]}/`),
      go.request("GET", `${resourceMetadata.rest_path}${ids.go[definition.name]}/`),
    ]);
    expectStatus(`oracle get ${definition.name}`, oracleLoaded, [200]);
    expectStatus(`Go get ${definition.name}`, goLoaded, [200]);
    compareResourceProjection(`get ${definition.name}`, oracleLoaded.body, goLoaded.body, resourceMetadata);
    observeTimestamps("oracle", definition.name, "get", oracleLoaded.body);
    observeTimestamps("Go", definition.name, "get", goLoaded.body);
  }

  for (const query of ["compatibility", "definitely-no-profile-match"]) {
    const encoded = new URLSearchParams({ q: query, ordering: "id", limit: "100", offset: "0" });
    const [oracleSearch, goSearch] = await Promise.all([
      oracle.request("GET", `${siteMetadata.rest_path}?${encoded}`),
      go.request("GET", `${siteMetadata.rest_path}?${encoded}`),
    ]);
    expectStatus(`oracle Site search ${query}`, oracleSearch, [200]);
    expectStatus(`Go Site search ${query}`, goSearch, [200]);
    comparePageProjection(`Site search ${query}`, oracleSearch.body, goSearch.body, siteMetadata);
  }

  for (const [field, value] of [
    ["compatibility_unknown_filter", "ignored-by-oracle"],
    ["tenant_id", "999999"],
  ]) {
    const query = new URLSearchParams({ [field]: value, limit: "1" });
    const [oracleDeferredFilter, goDeferredFilter] = await Promise.all([
      oracle.request("GET", `${siteMetadata.rest_path}?${query}`),
      go.request("GET", `${siteMetadata.rest_path}?${query}`),
    ]);
    expectStatus(`oracle accepts out-of-profile filter ${field}`, oracleDeferredFilter, [200]);
    expectStatus(`Go rejects out-of-profile filter ${field}`, goDeferredFilter, [400]);
    compare(
      `Go out-of-profile filter ${field} reason`,
      { status: 400, body: { [field]: ["Unsupported filter."] } },
      normalizeValidation(goDeferredFilter),
    );
  }

  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    const patch = { description: `Updated compatibility description for ${definition.name}` };
    const [oraclePatched, goPatched] = await Promise.all([
      oracle.request("PATCH", `${resourceMetadata.rest_path}${ids.oracle[definition.name]}/`, { body: patch }),
      go.request("PATCH", `${resourceMetadata.rest_path}${ids.go[definition.name]}/`, { body: patch }),
    ]);
    expectStatus(`oracle patch ${definition.name}`, oraclePatched, [200]);
    expectStatus(`Go patch ${definition.name}`, goPatched, [200]);
    compareResourceProjection(`patch ${definition.name}`, oraclePatched.body, goPatched.body, resourceMetadata);
    observeTimestamps("oracle", definition.name, "patch", oraclePatched.body, true);
    observeTimestamps("Go", definition.name, "patch", goPatched.body, true);
  }

  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    const oracleReplacement = {
      ...definition.payload(ids.oracle),
      description: `Replaced compatibility description for ${definition.name}`,
    };
    const goReplacement = {
      ...definition.payload(ids.go),
      description: `Replaced compatibility description for ${definition.name}`,
    };
    const [oracleReplaced, goReplaced] = await Promise.all([
      oracle.request("PUT", `${resourceMetadata.rest_path}${ids.oracle[definition.name]}/`, {
        body: oracleReplacement,
      }),
      go.request("PUT", `${resourceMetadata.rest_path}${ids.go[definition.name]}/`, { body: goReplacement }),
    ]);
    expectStatus(`oracle replace ${definition.name}`, oracleReplaced, [200]);
    expectStatus(`Go replace ${definition.name}`, goReplaced, [200]);
    compareResourceProjection(`replace ${definition.name}`, oracleReplaced.body, goReplaced.body, resourceMetadata);
    observeTimestamps("oracle", definition.name, "replace", oracleReplaced.body, true);
    observeTimestamps("Go", definition.name, "replace", goReplaced.body, true);
  }

  const ipAddressMetadata = metadata.get("IPAddress");
  const patchIPAddress = async (label, oraclePatch, goPatch = oraclePatch) => {
    const [oraclePatched, goPatched] = await Promise.all([
      oracle.request("PATCH", `${ipAddressMetadata.rest_path}${ids.oracle.IPAddress}/`, {
        body: oraclePatch,
      }),
      go.request("PATCH", `${ipAddressMetadata.rest_path}${ids.go.IPAddress}/`, { body: goPatch }),
    ]);
    expectStatus(`oracle ${label}`, oraclePatched, [200]);
    expectStatus(`Go ${label}`, goPatched, [200]);
    compareResourceProjection(label, oraclePatched.body, goPatched.body, ipAddressMetadata);
    return { oracle: oraclePatched.body, go: goPatched.body };
  };

  const assignmentPreserved = await patchIPAddress("assignment omitted preserves state", {
    description: "Assignment-presence omission check",
  });
  for (const [system, body] of Object.entries(assignmentPreserved)) {
    if (body.assigned_object_type !== "dcim.interface" || body.assigned_object_id === null) {
      throw new Error(`${system} did not preserve assignment when assignment fields were omitted`);
    }
  }

  const assignmentCleared = await patchIPAddress("both assignment nulls unassign", {
    assigned_object_type: null,
    assigned_object_id: null,
  });
  for (const [system, body] of Object.entries(assignmentCleared)) {
    if (body.assigned_object_type !== null || body.assigned_object_id !== null || body.assigned_object !== null) {
      throw new Error(`${system} did not clear the complete assignment projection`);
    }
  }

  await patchIPAddress(
    "concrete assignment pair assigns",
    { assigned_object_type: "dcim.interface", assigned_object_id: ids.oracle.Interface },
    { assigned_object_type: "dcim.interface", assigned_object_id: ids.go.Interface },
  );
  const assignmentIDPreserved = await patchIPAddress("partial assignment type preserves ID", {
    assigned_object_type: "dcim.interface",
  });
  for (const [system, body] of Object.entries(assignmentIDPreserved)) {
    if (body.assigned_object_id === null) {
      throw new Error(`${system} did not preserve assigned_object_id during a type-only patch`);
    }
  }
  await patchIPAddress(
    "partial assignment ID preserves type",
    { assigned_object_id: ids.oracle.Interface },
    { assigned_object_id: ids.go.Interface },
  );

  const deviceMetadata = metadata.get("Device");
  for (const [label, overrides] of [
    ["occupied rack unit", { name: "compatibility-device-occupied", asset_tag: "COMPAT-OCCUPIED" }],
    ["out-of-bounds rack unit", { name: "compatibility-device-bounds", asset_tag: "COMPAT-BOUNDS", position: 43 }],
  ]) {
    const oraclePayload = { ...definitions.find(({ name }) => name === "Device").payload(ids.oracle), ...overrides };
    const goPayload = { ...definitions.find(({ name }) => name === "Device").payload(ids.go), ...overrides };
    const [oracleInvalidDevice, goInvalidDevice] = await Promise.all([
      oracle.request("POST", deviceMetadata.rest_path, { body: oraclePayload }),
      go.request("POST", deviceMetadata.rest_path, { body: goPayload }),
    ]);
    compareValidationResponses(`rack placement rejects ${label}`, oracleInvalidDevice, goInvalidDevice);
    for (const [api, expectedName] of [
      [oracle, overrides.name],
      [go, overrides.name],
    ]) {
      const query = new URLSearchParams({ name: expectedName, limit: "1" });
      const page = await api.request("GET", `${deviceMetadata.rest_path}?${query}`);
      expectStatus(`${api.name} rack placement rollback`, page, [200]);
      if (Number(page.body?.count) !== 0) {
        throw new Error(`${api.name} persisted invalid rack placement ${expectedName}`);
      }
      assertionCount += 1;
    }
  }

  const prefixMetadata = metadata.get("Prefix");
  for (const [label, prefix] of [
    ["Prefix rejects /0", "0.0.0.0/0"],
    ["IPv6 Prefix rejects host bits", "2001:db8:100::1/64"],
  ]) {
    const payload = (system) => ({ prefix, vrf: ids[system].VRF, status: "active" });
    const [oracleInvalidPrefix, goInvalidPrefix] = await Promise.all([
      oracle.request("POST", prefixMetadata.rest_path, { body: payload("oracle") }),
      go.request("POST", prefixMetadata.rest_path, { body: payload("go") }),
    ]);
    compareValidationResponses(label, oracleInvalidPrefix, goInvalidPrefix);
  }

  const ipv6PrefixPayload = (system) => ({
    prefix: "2001:db8:100::/64",
    vrf: ids[system].VRF,
    status: "active",
  });
  const [oracleIPv6Prefix, goIPv6Prefix] = await Promise.all([
    oracle.request("POST", prefixMetadata.rest_path, { body: ipv6PrefixPayload("oracle") }),
    go.request("POST", prefixMetadata.rest_path, { body: ipv6PrefixPayload("go") }),
  ]);
  expectStatus("oracle create IPv6 Prefix", oracleIPv6Prefix, [201]);
  expectStatus("Go create IPv6 Prefix", goIPv6Prefix, [201]);
  bindCreatedPair("Prefix.ipv6", prefixMetadata, oracleIPv6Prefix.body, goIPv6Prefix.body);
  compareCreateResource("create canonical IPv6 Prefix", oracleIPv6Prefix.body, goIPv6Prefix.body, prefixMetadata);

  const invalidAddressCases = [
    ["IPAddress rejects /0", "0.0.0.1/0", "active", ""],
    ["assigned IPv4 network ID rejects", "198.51.100.0/24", "active", ""],
    ["assigned IPv4 broadcast rejects", "198.51.100.255/24", "active", ""],
    ["assigned IPv6 network ID rejects", "2001:db8:100::/64", "active", ""],
    ["IPv4 SLAAC rejects", "198.51.100.30/32", "slaac", ""],
  ];
  for (const [label, address, status, dnsName] of invalidAddressCases) {
    const payload = (system) => ({
      address,
      vrf: ids[system].VRF,
      status,
      dns_name: dnsName,
      assigned_object_type: "dcim.interface",
      assigned_object_id: ids[system].Interface,
    });
    const [oracleInvalidAddress, goInvalidAddress] = await Promise.all([
      oracle.request("POST", ipAddressMetadata.rest_path, { body: payload("oracle") }),
      go.request("POST", ipAddressMetadata.rest_path, { body: payload("go") }),
    ]);
    compareValidationResponses(label, oracleInvalidAddress, goInvalidAddress);
  }

  const validAddressCases = [
    ["assigned IPv4 /31 network exception", "198.51.100.20/31", "active", ""],
    ["assigned IPv4 /32 exception", "198.51.100.22/32", "active", ""],
    ["assigned IPv6 /127 network exception", "2001:db8:100::/127", "active", ""],
    ["assigned IPv6 /128 exception and SLAAC DNS normalization", "2001:db8:100::2/128", "slaac", "MiXeD.Example.TEST"],
  ];
  for (const [caseIndex, [label, address, status, dnsName]] of validAddressCases.entries()) {
    const payload = (system) => ({
      address,
      vrf: ids[system].VRF,
      status,
      dns_name: dnsName,
      assigned_object_type: "dcim.interface",
      assigned_object_id: ids[system].Interface,
    });
    const [oracleCreatedAddress, goCreatedAddress] = await Promise.all([
      oracle.request("POST", ipAddressMetadata.rest_path, { body: payload("oracle") }),
      go.request("POST", ipAddressMetadata.rest_path, { body: payload("go") }),
    ]);
    expectStatus(`oracle ${label}`, oracleCreatedAddress, [201]);
    expectStatus(`Go ${label}`, goCreatedAddress, [201]);
    bindCreatedPair(
      `IPAddress.edge.${caseIndex + 1}`,
      ipAddressMetadata,
      oracleCreatedAddress.body,
      goCreatedAddress.body,
    );
    compareCreateResource(label, oracleCreatedAddress.body, goCreatedAddress.body, ipAddressMetadata);
    if (dnsName !== "") {
      for (const [system, body] of [
        ["oracle", oracleCreatedAddress.body],
        ["Go", goCreatedAddress.body],
      ]) {
        if (body.dns_name !== dnsName.toLowerCase()) {
          throw new Error(`${system} did not lowercase DNS name ${dnsName}`);
        }
        assertionCount += 1;
      }
    }
    const [oracleDeletedAddress, goDeletedAddress] = await Promise.all([
      oracle.request("DELETE", `${ipAddressMetadata.rest_path}${oracleCreatedAddress.body.id}/`),
      go.request("DELETE", `${ipAddressMetadata.rest_path}${goCreatedAddress.body.id}/`),
    ]);
    expectStatus(`oracle cleanup ${label}`, oracleDeletedAddress, [204]);
    expectStatus(`Go cleanup ${label}`, goDeletedAddress, [204]);
  }

  const [oracleDeletedIPv6Prefix, goDeletedIPv6Prefix] = await Promise.all([
    oracle.request("DELETE", `${prefixMetadata.rest_path}${oracleIPv6Prefix.body.id}/`),
    go.request("DELETE", `${prefixMetadata.rest_path}${goIPv6Prefix.body.id}/`),
  ]);
  expectStatus("oracle cleanup IPv6 Prefix", oracleDeletedIPv6Prefix, [204]);
  expectStatus("Go cleanup IPv6 Prefix", goDeletedIPv6Prefix, [204]);

  const extraSites = [];
  for (const suffix of ["alpha", "zulu"]) {
    const payload = {
      name: `Compatibility Site ${suffix}`,
      slug: `compatibility-site-${suffix}`,
      status: "active",
      facility: "",
      description: `Pagination ${suffix}`,
      comments: "",
    };
    const [oracleCreated, goCreated] = await Promise.all([
      oracle.request("POST", siteMetadata.rest_path, { body: payload }),
      go.request("POST", siteMetadata.rest_path, { body: payload }),
    ]);
    expectStatus(`oracle create pagination site ${suffix}`, oracleCreated, [201]);
    expectStatus(`Go create pagination site ${suffix}`, goCreated, [201]);
    bindCreatedPair(`Site.pagination.${suffix}`, siteMetadata, oracleCreated.body, goCreated.body);
    compareCreateResource(`create pagination site ${suffix}`, oracleCreated.body, goCreated.body, siteMetadata);
    extraSites.push({ oracle: oracleCreated.body.id, go: goCreated.body.id });
  }

  // Exercise every filter declared by the selected capability metadata. The
  // values intentionally cover scalar, boolean, relation-ID, relation-slug,
  // network-containment, and assignment filters.
  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    const values = {
      oracle: declaredFilterValues(definition.name, ids.oracle),
      go: declaredFilterValues(definition.name, ids.go),
    };
    const generatedFields = Object.keys(values.oracle).sort();
    const declaredFields = [...resourceMetadata.filters].sort();
    if (generatedFields.join("\0") !== declaredFields.join("\0")) {
      throw new Error(
        `${definition.name} filter fixture drift: declared ${declaredFields.join(", ")}; generated ${generatedFields.join(", ")}`,
      );
    }
    for (const field of resourceMetadata.filters) {
      const query = (system) =>
        new URLSearchParams({
          [field]: String(values[system][field]),
          ordering: "id",
          limit: "100",
          offset: "0",
        }).toString();
      const [oraclePage, goPage] = await Promise.all([
        oracle.request("GET", `${resourceMetadata.rest_path}?${query("oracle")}`),
        go.request("GET", `${resourceMetadata.rest_path}?${query("go")}`),
      ]);
      expectStatus(`oracle ${definition.name}.${field} filter`, oraclePage, [200]);
      expectStatus(`Go ${definition.name}.${field} filter`, goPage, [200]);
      if (Number(oraclePage.body?.count) < 1 || Number(goPage.body?.count) < 1) {
        throw new Error(
          `${definition.name}.${field} fixture did not select a resource: oracle=${oraclePage.body?.count}, Go=${goPage.body?.count}`,
        );
      }
      comparePageProjection(
        `filter ${definition.name}.${field}`,
        oraclePage.body,
        goPage.body,
        resourceMetadata,
      );
    }
  }

  // Exercise each declared ordering key. Site and Interface have multiple
  // rows, while the remaining resources still prove field admission and wire
  // parity against the oracle.
  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    for (const ordering of resourceMetadata.ordering) {
      const query = new URLSearchParams({ ordering, limit: "100", offset: "0" }).toString();
      const [oraclePage, goPage] = await Promise.all([
        oracle.request("GET", `${resourceMetadata.rest_path}?${query}`),
        go.request("GET", `${resourceMetadata.rest_path}?${query}`),
      ]);
      expectStatus(`oracle ${definition.name}.${ordering} ordering`, oraclePage, [200]);
      expectStatus(`Go ${definition.name}.${ordering} ordering`, goPage, [200]);
      comparePageProjection(
        `ordering ${definition.name}.${ordering}`,
        oraclePage.body,
        goPage.body,
        resourceMetadata,
      );
    }
  }

  for (const query of [
    "?ordering=name&limit=2&offset=0",
    "?ordering=name&limit=2&offset=2",
    "?ordering=-name&limit=2&offset=0",
    "?status=active&ordering=name&limit=2&offset=0",
  ]) {
    const [oraclePage, goPage] = await Promise.all([
      oracle.request("GET", `${siteMetadata.rest_path}${query}`),
      go.request("GET", `${siteMetadata.rest_path}${query}`),
    ]);
    expectStatus(`oracle site page ${query}`, oraclePage, [200]);
    expectStatus(`Go site page ${query}`, goPage, [200]);
    comparePageProjection(`site ordering and pagination ${query}`, oraclePage.body, goPage.body, siteMetadata);
  }

  // List every profile resource, including the template-created Interface.
  for (const definition of definitions) {
    const resourceMetadata = metadata.get(definition.name);
    const [oraclePage, goPage] = await Promise.all([
      oracle.request("GET", `${resourceMetadata.rest_path}?ordering=id&limit=100&offset=0`),
      go.request("GET", `${resourceMetadata.rest_path}?ordering=id&limit=100&offset=0`),
    ]);
    expectStatus(`oracle list ${definition.name}`, oraclePage, [200]);
    expectStatus(`Go list ${definition.name}`, goPage, [200]);
    comparePageProjection(`list ${definition.name}`, oraclePage.body, goPage.body, resourceMetadata);
  }

  for (const resourceName of ["Rack", "DeviceType"]) {
    const resourceMetadata = metadata.get(resourceName);
    const [oracleProtected, goProtected] = await Promise.all([
      oracle.request("DELETE", `${resourceMetadata.rest_path}${ids.oracle[resourceName]}/`),
      go.request("DELETE", `${resourceMetadata.rest_path}${ids.go[resourceName]}/`),
    ]);
    expectStatus(`oracle protects referenced ${resourceName}`, oracleProtected, [409]);
    expectStatus(`Go protects referenced ${resourceName}`, goProtected, [409]);
    compare(
      `protected-delete reason for ${resourceName}`,
      { status: oracleProtected.status, body: oracleProtected.body },
      { status: goProtected.status, body: goProtected.body },
    );
    const [oracleStillPresent, goStillPresent] = await Promise.all([
      oracle.request("GET", `${resourceMetadata.rest_path}${ids.oracle[resourceName]}/`),
      go.request("GET", `${resourceMetadata.rest_path}${ids.go[resourceName]}/`),
    ]);
    expectStatus(`oracle preserves protected ${resourceName}`, oracleStillPresent, [200]);
    expectStatus(`Go preserves protected ${resourceName}`, goStillPresent, [200]);
    compareResourceProjection(
      `protected delete leaves ${resourceName} unchanged`,
      oracleStillPresent.body,
      goStillPresent.body,
      resourceMetadata,
    );
  }

  for (const pair of extraSites.reverse()) {
    const [oracleDeleted, goDeleted] = await Promise.all([
      oracle.request("DELETE", `${siteMetadata.rest_path}${pair.oracle}/`),
      go.request("DELETE", `${siteMetadata.rest_path}${pair.go}/`),
    ]);
    expectStatus("oracle delete pagination site", oracleDeleted, [204]);
    expectStatus("Go delete pagination site", goDeleted, [204]);
  }

  // Device owns Interfaces, and an Interface owns its assigned IP addresses.
  // Deleting the Device must commit the entire cascade atomically.
  const [oracleDeviceDeleted, goDeviceDeleted] = await Promise.all([
    oracle.request("DELETE", `${deviceMetadata.rest_path}${ids.oracle.Device}/`),
    go.request("DELETE", `${deviceMetadata.rest_path}${ids.go.Device}/`),
  ]);
  expectStatus("oracle Device cascade delete", oracleDeviceDeleted, [204]);
  expectStatus("Go Device cascade delete", goDeviceDeleted, [204]);
  for (const resourceName of ["Device", "Interface", "IPAddress"]) {
    const resourceMetadata = metadata.get(resourceName);
    const [oracleMissing, goMissing] = await Promise.all([
      oracle.request("GET", `${resourceMetadata.rest_path}${ids.oracle[resourceName]}/`),
      go.request("GET", `${resourceMetadata.rest_path}${ids.go[resourceName]}/`),
    ]);
    expectStatus(`oracle cascade removed ${resourceName}`, oracleMissing, [404]);
    expectStatus(`Go cascade removed ${resourceName}`, goMissing, [404]);
  }
  const [oracleCascadedInterfaces, goCascadedInterfaces] = await Promise.all([
    oracle.request("GET", `${metadata.get("Interface").rest_path}?device_id=${ids.oracle.Device}&limit=100`),
    go.request("GET", `${metadata.get("Interface").rest_path}?device_id=${ids.go.Device}&limit=100`),
  ]);
  expectStatus("oracle cascade Interface list", oracleCascadedInterfaces, [200]);
  expectStatus("Go cascade Interface list", goCascadedInterfaces, [200]);
  comparePageProjection(
    "Device cascade leaves no Interfaces",
    oracleCascadedInterfaces.body,
    goCascadedInterfaces.body,
    metadata.get("Interface"),
  );
  const cascadeDeleted = new Set(["Device", "Interface", "IPAddress"]);

  for (const definition of [...definitions].reverse()) {
    if (cascadeDeleted.has(definition.name)) continue;
    const resourceMetadata = metadata.get(definition.name);
    const [oracleDeleted, goDeleted] = await Promise.all([
      oracle.request("DELETE", `${resourceMetadata.rest_path}${ids.oracle[definition.name]}/`),
      go.request("DELETE", `${resourceMetadata.rest_path}${ids.go[definition.name]}/`),
    ]);
    expectStatus(`oracle delete ${definition.name}`, oracleDeleted, [204]);
    expectStatus(`Go delete ${definition.name}`, goDeleted, [204]);
    const [oracleMissing, goMissing] = await Promise.all([
      oracle.request("GET", `${resourceMetadata.rest_path}${ids.oracle[definition.name]}/`),
      go.request("GET", `${resourceMetadata.rest_path}${ids.go[definition.name]}/`),
    ]);
    expectStatus(`oracle deleted ${definition.name} is absent`, oracleMissing, [404]);
    expectStatus(`Go deleted ${definition.name} is absent`, goMissing, [404]);
  }

  const revoke = await goIdentity.session.request("DELETE", `/api/auth/tokens/${goIdentity.tokenID}/`);
  expectStatus("Go token revocation", revoke, [204]);
  const revoked = await go.request("GET", "/api/dcim/sites/?limit=1");
  expectStatus("revoked Go token rejected", revoked, [403]);
  const logout = await goIdentity.session.request("POST", "/api/auth/logout/");
  expectStatus("Go session logout", logout, [204]);

  await fs.writeFile(path.join(artifactDirectory, "exchanges.json"), `${JSON.stringify(exchanges, null, 2)}\n`);

  if (semanticFailures.length > 0) {
    await fs.writeFile(
      path.join(artifactDirectory, "divergences.json"),
      `${JSON.stringify(semanticFailures, null, 2)}\n`,
    );
    throw new Error(`${semanticFailures.length} semantic compatibility divergence(s); see divergences.json`);
  }

  const report = {
    status: "passed",
    profile: "core-workflow-v1",
    resources: definitions.map(({ name }) => name),
    normalizers_asserted: [
      "origin-only URL normalization with exact path/query/trailing slash",
      "RFC3339 timestamp presence and scenario-relative ordering",
      "scenario-bound generated identifiers",
      "exact authorization status and validation reasons",
      "no undeclared Go response fields",
    ],
    differential_scenarios: [
      "profile input and filter fail-closed boundary",
      "null omission and defaults",
      "search",
      "rack placement and rollback",
      "assignment presence matrix",
      "IPv4 IPv6 and SLAAC edge rules",
      "protected delete and atomic ownership cascade",
    ],
    intentionally_deferred_oracle_inputs: ["Site.tenant", "Site.compatibility_unknown_field"],
    assertions: assertionCount,
    exchanges: exchanges.length,
  };
  await fs.writeFile(path.join(artifactDirectory, "report.json"), `${JSON.stringify(report, null, 2)}\n`);
  console.log(
    `compatibility oracle: ${definitions.length} resources, ${assertionCount} assertions, ${exchanges.length} HTTP exchanges passed`,
  );
}

run().catch(async (error) => {
  const failure = {
    status: "failed",
    error: { name: error.name, message: error.message, stack: error.stack },
    assertions_completed: assertionCount,
    semantic_divergences: semanticFailures,
    recent_exchanges: exchanges.slice(-20),
  };
  try {
    await fs.mkdir(artifactDirectory, { recursive: true });
    await fs.writeFile(path.join(artifactDirectory, "exchanges.json"), `${JSON.stringify(exchanges, null, 2)}\n`);
    await fs.writeFile(path.join(artifactDirectory, "failure.json"), `${JSON.stringify(failure, null, 2)}\n`);
  } catch (artifactError) {
    console.error(`could not write failure artifact: ${artifactError.message}`);
  }
  console.error(error.stack || error.message);
  process.exitCode = 1;
});
