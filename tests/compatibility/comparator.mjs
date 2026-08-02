const numericFields = new Set([
  "width",
  "u_height",
  "starting_unit",
  "position",
  "mtu",
  "speed",
  "device_count",
  "prefix_count",
  "rack_count",
  "devicetype_count",
  "interface_template_count",
  "interface_count",
  "count_ipaddresses",
  "ipaddress_count",
  "family",
  "children",
  "_depth",
]);

const timestampFields = new Set(["created", "last_updated"]);

export class SemanticMismatch extends Error {
  constructor(label, path, expected, actual) {
    super(
      `${label}: semantic mismatch at ${path}\n` +
        `  oracle: ${format(expected)}\n` +
        `  go:     ${format(actual)}`,
    );
    this.name = "SemanticMismatch";
    this.label = label;
    this.path = path;
    this.expected = expected;
    this.actual = actual;
  }
}

function format(value) {
  return JSON.stringify(value, null, 2);
}

export class IdentifierBindings {
  constructor() {
    this.paths = { oracle: new Map(), go: new Map() };
    this.symbols = new Map();
  }

  bind(symbol, resourcePath, oracleID, goID) {
    if (typeof symbol !== "string" || symbol === "" || !resourcePath.startsWith("/") || !resourcePath.endsWith("/")) {
      throw new Error(`invalid identifier binding ${format({ symbol, resourcePath })}`);
    }
    for (const [system, id] of [["oracle", oracleID], ["go", goID]]) {
      if (!Number.isInteger(id) || id <= 0) {
        throw new Error(`${symbol}: ${system} identifier must be a positive integer`);
      }
      const detailPath = `${resourcePath}${id}/`;
      const existing = this.paths[system].get(detailPath);
      if (existing && existing !== symbol) {
        throw new Error(`${system} ${detailPath} is already bound to ${existing}`);
      }
      this.paths[system].set(detailPath, symbol);
    }
    const existingSymbol = this.symbols.get(symbol);
    const binding = { resourcePath, oracleID, goID };
    if (existingSymbol && format(existingSymbol) !== format(binding)) {
      throw new Error(`${symbol} is already bound to a different identifier pair`);
    }
    this.symbols.set(symbol, binding);
  }

  symbolForURL(system, value, label) {
    if (!Object.hasOwn(this.paths, system)) throw new Error(`${label}: unknown identifier system ${system}`);
    const parsed = new URL(value, "http://compatibility.invalid");
    const symbol = this.paths[system].get(parsed.pathname);
    if (!symbol) throw new Error(`${label}: unbound ${system} identifier URL ${parsed.pathname}`);
    return symbol;
  }

  symbolForID(system, resourcePath, id, label) {
    if (!Number.isInteger(id) || id <= 0) {
      throw new Error(`${label}: expected a positive integer identifier, got ${format(id)}`);
    }
    return this.symbolForURL(system, `${resourcePath}${id}/`, label);
  }
}

function normalizeURL(value, label, field, context = {}) {
  if (typeof value !== "string") {
    throw new Error(`${label}.${field}: expected a URL string, got ${format(value)}`);
  }
  let parsed;
  try {
    parsed = new URL(value, "http://compatibility.invalid");
  } catch (error) {
    throw new Error(`${label}.${field}: invalid URL ${format(value)}: ${error.message}`);
  }
  // The pinned normalizer ignores only the configured origin. Preserve the
  // path, trailing slash, and query byte-for-byte. Resource identifiers are
  // bound to a scenario symbol because the disposable databases allocate
  // them independently; importantly, this does not manufacture a slash.
  const detailMatch = parsed.pathname.match(/\/\d+\/$/);
  let path = parsed.pathname;
  if (detailMatch) {
    const symbol = context.bindings
      ? context.bindings.symbolForURL(context.system, value, `${label}.${field}`)
      : "id";
    path = parsed.pathname.replace(/\/\d+\/$/, `/:${symbol}/`);
  }
  return `${path}${parsed.search}`;
}

function normalizeTimestamp(value, label, field) {
  if (typeof value !== "string" || Number.isNaN(Date.parse(value))) {
    throw new Error(`${label}.${field}: expected a timestamp, got ${format(value)}`);
  }
  return "<timestamp>";
}

function normalizeNumber(value, label, field) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`${label}.${field}: expected a finite number, got ${format(value)}`);
  }
  return value;
}

function normalizeValue(value, field, label, context = {}) {
  if (field === "id") {
    const id = normalizeNumber(value, label, field);
    if (!Number.isInteger(id) || id <= 0) {
      throw new Error(`${label}.${field}: expected a positive integer, got ${format(value)}`);
    }
    if (!context.resourceURL || !context.bindings) return "<id>";
    return `<${context.bindings.symbolForURL(context.system, context.resourceURL, `${label}.${field}`)}>`;
  }
  if (field === "assigned_object_id") {
    if (value === null) return null;
    const id = normalizeNumber(value, label, field);
    if (!Number.isInteger(id) || id <= 0) {
      throw new Error(`${label}.${field}: expected a positive relationship ID, got ${format(value)}`);
    }
    const resourcePath = context.referencePaths?.[field];
    if (!context.bindings || !resourcePath) return "<relationship-id>";
    return `<${context.bindings.symbolForID(context.system, resourcePath, id, `${label}.${field}`)}>`;
  }
  if (field === "url") return normalizeURL(value, label, field, context);
  if (timestampFields.has(field)) return normalizeTimestamp(value, label, field);

  // Relationship IDs differ between the disposable databases. Bind each ID
  // to its relationship URL and compare the stable resource path/display.
  if (value && typeof value === "object" && !Array.isArray(value) && Object.hasOwn(value, "display")) {
    if (typeof value.display !== "string") {
      throw new Error(`${label}.${field}.display: expected a string, got ${format(value.display)}`);
    }
    const relationshipID = normalizeNumber(value.id, `${label}.${field}`, "id");
    if (!Number.isInteger(relationshipID) || relationshipID <= 0) {
      throw new Error(`${label}.${field}.id: expected a positive integer, got ${format(value.id)}`);
    }
    if (typeof value.url !== "string") {
      throw new Error(`${label}.${field}.url: expected a URL string, got ${format(value.url)}`);
    }
    const parsed = new URL(value.url, "http://compatibility.invalid");
    const match = parsed.pathname.match(/\/(\d+)\/$/);
    if (!match || Number(match[1]) !== relationshipID || parsed.search !== "") {
      throw new Error(`${label}.${field}: relationship URL does not bind ID ${relationshipID}`);
    }
    const symbol = context.bindings
      ? context.bindings.symbolForURL(context.system, value.url, `${label}.${field}`)
      : "relationship-id";
    return {
      id: `<${symbol}>`,
      url: normalizeURL(value.url, `${label}.${field}`, "url", context),
      display: value.display,
    };
  }

  if (numericFields.has(field) && value !== null) return normalizeNumber(value, label, field);
  if (Array.isArray(value)) return value.map((entry, index) => normalizeValue(entry, `${field}[${index}]`, label, context));
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.keys(value)
        .sort()
        .map((key) => [key, normalizeValue(value[key], key, `${label}.${field}`, context)]),
    );
  }
  return value;
}

export function projectResource(resource, metadata, label = metadata.name, selectedFields = null, context = {}) {
  if (!resource || typeof resource !== "object" || Array.isArray(resource)) {
    throw new Error(`${label}: expected a resource object, got ${format(resource)}`);
  }
  const fields = selectedFields ?? [...metadata.writable_fields, ...metadata.response_only_fields];
  if (fields.includes("id") && fields.includes("url")) {
    const id = normalizeNumber(resource.id, label, "id");
    if (!Number.isInteger(id) || id <= 0 || typeof resource.url !== "string") {
      throw new Error(`${label}: id/url identity binding is invalid`);
    }
    const parsed = new URL(resource.url, "http://compatibility.invalid");
    const match = parsed.pathname.match(/\/(\d+)\/$/);
    if (!match || Number(match[1]) !== id || parsed.search !== "") {
      throw new Error(`${label}: resource URL does not bind ID ${id} with a trailing slash`);
    }
  }
  const projected = {};
  for (const field of fields) {
    if (!Object.hasOwn(resource, field)) {
      throw new Error(`${label}: declared field ${field} is missing from the response`);
    }
    projected[field] = normalizeValue(resource[field], field, label, { ...context, resourceURL: resource.url });
  }
  return projected;
}

export function projectPage(page, metadata, label = `${metadata.name} page`, context = {}) {
  if (!page || typeof page !== "object" || !Array.isArray(page.results)) {
    throw new Error(`${label}: expected a paginated response`);
  }
  const count = normalizeNumber(page.count, label, "count");
  if (!Number.isInteger(count) || count < 0) {
    throw new Error(`${label}.count: expected a non-negative integer`);
  }
  const normalizePageLink = (value, field) => {
    if (value === null) return null;
    if (typeof value !== "string") throw new Error(`${label}.${field}: expected URL or null`);
    const parsed = new URL(value, "http://compatibility.invalid");
    // The committed normalizer ignores only the origin. Query ordering,
    // encoding, defaults, and trailing slash remain byte-observable.
    return `${parsed.pathname}${parsed.search}`;
  };
  return {
    count,
    next: normalizePageLink(page.next, "next"),
    previous: normalizePageLink(page.previous, "previous"),
    results: page.results.map((entry, index) => projectResource(entry, metadata, `${label}.results[${index}]`, null, context)),
  };
}

export function normalizeValidation(response) {
  // normalizers.yaml explicitly forbids discarding validation reasons. Keep
  // the status, field names, and every reason; only sort object keys to make
  // JSON object ordering irrelevant.
  return { status: response.status, body: canonicalJSON(response.body) };
}

export function normalizeDenial(response) {
  // 401 and 403 are deliberately distinct compatibility outcomes.
  return { status: response.status, body: canonicalJSON(response.body) };
}

function canonicalJSON(value) {
  if (Array.isArray(value)) return value.map(canonicalJSON);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.keys(value)
        .sort()
        .map((key) => [key, canonicalJSON(value[key])]),
    );
  }
  return value;
}

export function assertSemanticEqual(label, oracle, go) {
  const difference = firstDifference(oracle, go, "$");
  if (difference) {
    throw new SemanticMismatch(label, difference.path, difference.oracle, difference.go);
  }
}

function firstDifference(oracle, go, path) {
  if (Object.is(oracle, go)) return null;
  if (typeof oracle !== typeof go || oracle === null || go === null) {
    return { path, oracle, go };
  }
  if (Array.isArray(oracle) || Array.isArray(go)) {
    if (!Array.isArray(oracle) || !Array.isArray(go) || oracle.length !== go.length) {
      return { path, oracle, go };
    }
    for (let index = 0; index < oracle.length; index += 1) {
      const difference = firstDifference(oracle[index], go[index], `${path}[${index}]`);
      if (difference) return difference;
    }
    return null;
  }
  if (typeof oracle === "object") {
    const oracleKeys = Object.keys(oracle).sort();
    const goKeys = Object.keys(go).sort();
    if (oracleKeys.join("\0") !== goKeys.join("\0")) return { path, oracle: oracleKeys, go: goKeys };
    for (const key of oracleKeys) {
      const difference = firstDifference(oracle[key], go[key], `${path}.${key}`);
      if (difference) return difference;
    }
    return null;
  }
  return { path, oracle, go };
}
