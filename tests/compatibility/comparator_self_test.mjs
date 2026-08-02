import {
  assertSemanticEqual,
  normalizeDenial,
  normalizeValidation,
  projectPage,
  projectResource,
  IdentifierBindings,
  SemanticMismatch,
} from "./comparator.mjs";

const oracle = {
  count: 1,
  results: [{ display: "edge-01", status: "active", relationship: { display: "dc-a" } }],
};
const deliberatelyDivergentGo = structuredClone(oracle);
deliberatelyDivergentGo.results[0].status = "planned";

let rejected = false;
try {
  assertSemanticEqual("deliberate-divergence", oracle, deliberatelyDivergentGo);
} catch (error) {
  if (!(error instanceof SemanticMismatch) || error.path !== "$.results[0].status") throw error;
  rejected = true;
}

if (!rejected) {
  throw new Error("comparator accepted a deliberate semantic divergence");
}

assertSemanticEqual("identical-control", oracle, structuredClone(oracle));

function mustReject(label, oracleValue, goValue, expectedPath) {
  try {
    assertSemanticEqual(label, oracleValue, goValue);
  } catch (error) {
    if (!(error instanceof SemanticMismatch) || error.path !== expectedPath) throw error;
    return;
  }
  throw new Error(`${label}: comparator accepted a forbidden normalization`);
}

mustReject(
  "authorization status",
  normalizeDenial({ status: 401, body: { detail: "denied" } }),
  normalizeDenial({ status: 403, body: { detail: "denied" } }),
  "$.status",
);
mustReject(
  "validation reason",
  normalizeValidation({ status: 400, body: { name: ["oracle reason"] } }),
  normalizeValidation({ status: 400, body: { name: ["different reason"] } }),
  "$.body.name[0]",
);

const metadata = { name: "Example", writable_fields: [], response_only_fields: ["url"] };
mustReject(
  "resource URL trailing slash",
  projectResource({ url: "http://oracle/api/examples/17/" }, metadata, "oracle"),
  projectResource({ url: "/api/examples/91" }, metadata, "Go"),
  "$.url",
);
mustReject(
  "page query defaults",
  projectPage(
    { count: 3, next: null, previous: "http://oracle/api/examples/?limit=2", results: [] },
    { ...metadata, writable_fields: [], response_only_fields: [] },
    "oracle page",
  ),
  projectPage(
    { count: 3, next: null, previous: "/api/examples/?limit=2&offset=0", results: [] },
    { ...metadata, writable_fields: [], response_only_fields: [] },
    "Go page",
  ),
  "$.previous",
);
mustReject(
  "page query ordering",
  projectPage(
    { count: 0, next: "http://oracle/api/examples/?limit=2&offset=4", previous: null, results: [] },
    { ...metadata, writable_fields: [], response_only_fields: [] },
    "oracle page",
  ),
  projectPage(
    { count: 0, next: "/api/examples/?offset=4&limit=2", previous: null, results: [] },
    { ...metadata, writable_fields: [], response_only_fields: [] },
    "Go page",
  ),
  "$.next",
);
mustReject(
  "identifier type",
  { id: "<id>" },
  { id: "17" },
  "$.id",
);
mustReject(
  "choice envelope",
  { status: { value: "active", label: "Active" } },
  { status: "active" },
  "$.status",
);

const identifiers = new IdentifierBindings();
identifiers.bind("Site.primary", "/api/dcim/sites/", 17, 91);
const identifiedMetadata = {
  name: "Site",
  writable_fields: [],
  response_only_fields: ["id", "url"],
};
assertSemanticEqual(
  "scenario-bound identifiers",
  projectResource(
    { id: 17, url: "http://oracle/api/dcim/sites/17/" },
    identifiedMetadata,
    "oracle Site",
    null,
    { bindings: identifiers, system: "oracle" },
  ),
  projectResource(
    { id: 91, url: "/api/dcim/sites/91/" },
    identifiedMetadata,
    "Go Site",
    null,
    { bindings: identifiers, system: "go" },
  ),
);

let unboundRejected = false;
try {
  projectResource(
    { id: 92, url: "/api/dcim/sites/92/" },
    identifiedMetadata,
    "unbound Go Site",
    null,
    { bindings: identifiers, system: "go" },
  );
} catch (error) {
  if (!String(error.message).includes("unbound go identifier URL")) throw error;
  unboundRejected = true;
}
if (!unboundRejected) throw new Error("comparator accepted an unbound generated identifier");

console.log("compatibility comparator self-test: forbidden normalizations rejected");
