#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const BASELINE = "v4.4.6-post7";
const OUT_DIR = path.join(ROOT, "contracts", "netbox", BASELINE, "inventory");
const CHECK = process.argv.includes("--check");
const PROFILE = JSON.parse(
  fs.readFileSync(
    path.join(
      ROOT,
      "contracts",
      "netbox",
      BASELINE,
      "profiles",
      "core-workflow-v1.yaml",
    ),
    "utf8",
  ),
);
const IDENTITY = JSON.parse(
  fs.readFileSync(
    path.join(
      ROOT,
      "contracts",
      "netbox",
      BASELINE,
      "resources",
      "identity.yaml",
    ),
    "utf8",
  ),
);
const MODULES = [
  "circuits",
  "core",
  "dcim",
  "extras",
  "ipam",
  "tenancy",
  "users",
  "virtualization",
  "vpn",
  "wireless",
];

const profilePaths = new Set(
  PROFILE.resources.map((resource) => canonicalPath(resource.rest_path)),
);

function canonicalPath(value) {
  return `/${value.split("/").filter(Boolean).join("/")}`;
}

function classification(value) {
  const apiPath = canonicalPath(value);
  if (profilePaths.has(apiPath)) return "in_profile";
  if (
    apiPath === "/api/extras/scripts" ||
    apiPath === "/api/users/tokens/provision"
  ) {
    return "out_of_scope";
  }
  return "deferred";
}

function owner(module, entryClassification) {
  if (entryClassification === "in_profile") return `${module}-core-workflow`;
  if (entryClassification === "out_of_scope") return "architecture";
  return `${module}-future-profile`;
}

function document(source, entries) {
  return {
    schema_version: 1,
    compatibility_baseline: BASELINE,
    source,
    generated: true,
    entries: entries.sort((a, b) => a.id.localeCompare(b.id)),
  };
}

function baselineInventory() {
  const entries = [];
  const oracleRoot = path.join(ROOT, "netbox", "netbox");

  for (const module of MODULES) {
    const urlsPath = path.join(oracleRoot, module, "api", "urls.py");
    const viewsPath = path.join(oracleRoot, module, "api", "views.py");
    const urls = fs.readFileSync(urlsPath, "utf8");
    const views = fs.readFileSync(viewsPath, "utf8");
    const registrations = new Map();

    for (const match of urls.matchAll(
      /router\.register\(\s*['"]([^'"]+)['"]\s*,\s*views\.(\w+)/g,
    )) {
      const [, resource, viewSet] = match;
      const apiPath = `/api/${module}/${resource}`;
      const entryClassification = classification(apiPath);
      registrations.set(viewSet, resource);
      entries.push({
        id: `baseline:rest:${module}:${resource}`,
        kind: "resource",
        module,
        path: `${apiPath}/`,
        methods: ["GET", "POST", "PUT", "PATCH", "DELETE"],
        classification: entryClassification,
        tier: "T0",
        owner: owner(module, entryClassification),
        oracle_view: viewSet,
      });
    }

    const classMatches = [...views.matchAll(/^class\s+(\w+)\(([^\n]+)\):/gm)];
    const classDefinitions = new Map();
    for (let index = 0; index < classMatches.length; index += 1) {
      const current = classMatches[index];
      const start = current.index ?? 0;
      const end = classMatches[index + 1]?.index ?? views.length;
      const block = views.slice(start, end);
      const actions = [];
      for (const action of block.matchAll(
        /@action\(([\s\S]*?)\)\s*\n\s*def\s+(\w+)/g,
      )) {
        actions.push({
          options: action[1],
          functionName: action[2],
          declaredBy: current[1],
        });
      }
      classDefinitions.set(current[1], {
        bases: current[2].split(",").map((base) => base.trim()),
        actions,
      });
    }

    function inheritedActions(className, visited = new Set()) {
      if (visited.has(className)) return [];
      visited.add(className);
      const definition = classDefinitions.get(className);
      if (!definition) return [];
      return [
        ...definition.actions,
        ...definition.bases.flatMap((base) => inheritedActions(base, visited)),
      ];
    }

    for (const [viewSet, resource] of registrations) {
      const seenActionPaths = new Set();
      for (const action of inheritedActions(viewSet)) {
        const options = action.options;
        const functionName = action.functionName;
        const detail = /detail\s*=\s*True/.test(options);
        const urlPath =
          options.match(/url_path\s*=\s*['"]([^'"]+)['"]/)?.[1] ??
          functionName.replaceAll("_", "-");
        const configuredMethods = options.match(
          /methods\s*=\s*\[([^\]]+)\]/,
        )?.[1];
        const methods = configuredMethods
          ? [...configuredMethods.matchAll(/['"]([^'"]+)['"]/g)].map((item) =>
              item[1].toUpperCase(),
            )
          : ["GET"];
        const apiPath = `/api/${module}/${resource}`;
        const actionPath = detail
          ? `${apiPath}/{id}/${urlPath}/`
          : `${apiPath}/${urlPath}/`;
        if (seenActionPaths.has(actionPath)) continue;
        seenActionPaths.add(actionPath);
        const entryClassification = "deferred";
        entries.push({
          id: `baseline:rest-action:${module}:${resource}:${urlPath}`,
          kind: "action",
          module,
          path: actionPath,
          methods,
          classification: entryClassification,
          tier: "T0",
          owner: `${module}-future-profile`,
          oracle_view: `${action.declaredBy}.${functionName}`,
        });
      }
    }

    for (const explicit of urls.matchAll(
      /path\(\s*['"]([^'"]+)['"]\s*,\s*views\.(\w+)/g,
    )) {
      const route = explicit[1];
      if (route === "") continue;
      const view = explicit[2];
      const apiPath = `/api/${module}/${route}`
        .replaceAll("<int:pk>", "{id}")
        .replaceAll("//", "/");
      const methods =
        view === "DashboardView"
          ? ["GET", "PUT", "PATCH", "DELETE"]
          : view === "TokenProvisionView"
            ? ["POST"]
            : ["GET", "POST"];
      const entryClassification =
        canonicalPath(apiPath) === "/api/users/tokens/provision"
          ? "out_of_scope"
          : "deferred";
      entries.push({
        id: `baseline:rest-explicit:${module}:${route.replaceAll(/[^a-z0-9]+/gi, "-")}`,
        kind: "action",
        module,
        path: apiPath,
        methods,
        classification: entryClassification,
        tier: "T0",
        owner: owner(module, entryClassification),
        oracle_view: view,
      });
    }
  }

  return document(
    "checked-in oracle netbox/netbox/*/api/{urls,views}.py at fbb948d30e79ce657fac62994a22aca72c1770a9",
    entries,
  );
}

function currentRESTInventory() {
  const registryPath = path.join(
    ROOT,
    "netbox-backend",
    "internal",
    "routers",
    "netbox_drf_autogen.go",
  );
  const source = fs.readFileSync(registryPath, "utf8");
  const entries = [];

  for (const match of source.matchAll(/path:\s*"(\/api\/([^/]+)\/[^\"]+)"/g)) {
    const apiPath = canonicalPath(match[1]);
    const module = match[2];
    const resource = apiPath.split("/").at(-1);
    const entryClassification = classification(apiPath);
    entries.push({
      id: `current:rest:legacy:${module}:${resource}`,
      kind: "legacy-resource",
      module,
      path: `${apiPath}/`,
      methods: ["GET", "POST", "PUT", "PATCH", "DELETE"],
      classification: entryClassification,
      tier: "T0",
      owner: owner(module, entryClassification),
      runtime_enabled: false,
      implementation: "frozen-direct-gorm-registry",
      lifecycle: "frozen-unpublished",
    });
  }

  const openAPIPath = path.join(
    ROOT,
    "netbox-backend",
    "api",
    "openapi",
    "netbox-go-v1.yaml",
  );
  const openAPI = JSON.parse(fs.readFileSync(openAPIPath, "utf8"));
  const operationExists = (method, apiPath) =>
    Boolean(openAPI.paths?.[apiPath]?.[method.toLowerCase()]);

  for (const resource of PROFILE.resources) {
    const collectionPath = resource.rest_path;
    const detailPath = `${resource.rest_path}{id}/`;
    for (const [method, apiPath] of [
      ["GET", collectionPath],
      ["POST", collectionPath],
      ["GET", detailPath],
      ["PUT", detailPath],
      ["PATCH", detailPath],
      ["DELETE", detailPath],
    ]) {
      if (!operationExists(method, apiPath)) {
        throw new Error(
          `canonical OpenAPI is missing ${method} ${apiPath} for ${resource.module}.${resource.name}`,
        );
      }
    }
    entries.push({
      id: `current:rest:v1:${resource.module}:${resource.name}`,
      kind: "canonical-resource",
      module: resource.module,
      path: resource.rest_path,
      methods: ["GET", "POST", "PUT", "PATCH", "DELETE"],
      classification: "in_profile",
      tier: resource.tier,
      owner: resource.owner,
      runtime_enabled: true,
      implementation: "shared-core-workflow-adapter",
      lifecycle: PROFILE.contract_state,
    });
  }

  for (const operation of IDENTITY.rest_operations) {
    if (!operationExists(operation.method, operation.path)) {
      throw new Error(
        `canonical OpenAPI is missing identity operation ${operation.method} ${operation.path}`,
      );
    }
    const slug = `${operation.method}-${operation.path}`
      .toLowerCase()
      .replaceAll(/[^a-z0-9]+/g, "-")
      .replaceAll(/(^-|-$)/g, "");
    entries.push({
      id: `current:rest:extension:identity:${slug}`,
      kind: "canonical-operation",
      module: "identity",
      path: operation.path,
      methods: [operation.method],
      classification: "extension",
      tier: "not_applicable",
      owner: PROFILE.identity_extension.owner,
      runtime_enabled: true,
      implementation: "go-owned-identity-adapter",
      lifecycle: PROFILE.contract_state,
      verification: IDENTITY.verification,
    });
  }

  return document(
    "frozen direct-GORM registry plus canonical OpenAPI/runtime adapters",
    entries,
  );
}

function currentVueInventory() {
  const profileModelsPath = path.join(
    ROOT,
    "netbox-frontend",
    "src",
    "router",
    "models",
    "core-profile.ts",
  );
  const entries = [];
  const source = fs.readFileSync(profileModelsPath, "utf8");
  const models = [
    ...source.matchAll(
      /coreModel\(\s*'(dcim|ipam)'\s*,\s*'([^']+)'\s*,\s*'[^']*'\s*,\s*'[^']*'\s*,\s*'([^']+)'/g,
    ),
  ].map((match) => ({ module: match[1], model: match[2], path: match[3] }));

  for (const model of models) {
    const apiPath = `${canonicalPath(`/api/${model.module}/${model.path}`)}/`;
    const resource = PROFILE.resources.find(
      (candidate) => candidate.rest_path === apiPath,
    );
    if (!resource) {
      throw new Error(
        `Vue profile model ${model.module}.${model.model} has undeclared API path ${apiPath}`,
      );
    }
    entries.push({
      id: `current:vue:${model.module}:${model.model}`,
      kind: "resource-page",
      module: model.module,
      model: model.model,
      api_path: apiPath,
      classification: "in_profile",
      tier: resource.tier,
      owner: "frontend-core-workflow",
      runtime_enabled: true,
      backend_capability_enabled: true,
      lifecycle: PROFILE.contract_state,
    });
  }
  return document(
    "netbox-frontend/src/router/models/core-profile.ts; published by model-registry.ts",
    entries,
  );
}

function serviceBlocks(content) {
  const blocks = [];
  for (const match of content.matchAll(/^service\s+(\w+)\s*\{/gm)) {
    let depth = 0;
    let end = match.index ?? 0;
    for (; end < content.length; end += 1) {
      if (content[end] === "{") depth += 1;
      if (content[end] === "}") {
        depth -= 1;
        if (depth === 0) {
          end += 1;
          break;
        }
      }
    }
    blocks.push({ name: match[1], body: content.slice(match.index, end) });
  }
  return blocks;
}

function currentGRPCInventory() {
  const protoDir = path.join(ROOT, "netbox-backend", "api", "netbox_go", "v1");
  const entries = [];
  for (const file of fs
    .readdirSync(protoDir)
    .filter((name) => name.endsWith(".proto"))
    .sort()) {
    const content = fs.readFileSync(path.join(protoDir, file), "utf8");
    for (const service of serviceBlocks(content)) {
      const annotatedPath = service.body.match(
        /[a-z]+:\s*"(\/api\/[^"{]+)/,
      )?.[1];
      const apiPath = annotatedPath ? canonicalPath(annotatedPath) : null;
      const module = apiPath?.split("/")[2] ?? "internal";
      const entryClassification = "deferred";
      entries.push({
        id: `current:grpc:legacy:${service.name}`,
        kind: "legacy-service",
        module,
        service: `api.netbox_go.v1.${service.name}`,
        rpcs: [...service.body.matchAll(/^\s*rpc\s+(\w+)/gm)].map(
          (rpc) => rpc[1],
        ),
        annotated_rest_path: apiPath ? `${apiPath}/` : null,
        classification: entryClassification,
        tier: "T0",
        owner: "grpc-legacy-retirement",
        runtime_enabled: false,
        lifecycle: "frozen-unpublished",
      });
    }
  }

  const canonicalRoot = path.join(ROOT, "netbox-backend", "api", "proto");
  function protoFiles(directory) {
    return fs
      .readdirSync(directory, { withFileTypes: true })
      .flatMap((entry) => {
        const candidate = path.join(directory, entry.name);
        if (entry.isDirectory()) return protoFiles(candidate);
        return entry.isFile() && entry.name.endsWith(".proto")
          ? [candidate]
          : [];
      });
  }
  for (const filePath of protoFiles(canonicalRoot).sort()) {
    const content = fs.readFileSync(filePath, "utf8");
    const packageName = content.match(/^package\s+([^;]+);/m)?.[1];
    if (!packageName) continue;
    for (const service of serviceBlocks(content)) {
      const module = packageName.split(".")[1];
      const isExtension = module === "identity";
      const moduleResources = PROFILE.resources.filter(
        (resource) => resource.module === module,
      );
      const tier = moduleResources[0]?.tier ?? "T0";
      entries.push({
        id: `current:grpc:v1:${packageName}.${service.name}`,
        kind: "canonical-service",
        module,
        service: `${packageName}.${service.name}`,
        rpcs: [...service.body.matchAll(/^\s*rpc\s+(\w+)/gm)].map(
          (rpc) => rpc[1],
        ),
        classification: isExtension ? "extension" : "in_profile",
        tier: isExtension ? "not_applicable" : tier,
        owner: `${module}-core-workflow`,
        runtime_enabled: true,
        lifecycle: PROFILE.contract_state,
        ...(isExtension
          ? {
              verification: IDENTITY.verification,
            }
          : {}),
      });
    }
  }
  return document(
    "frozen netbox-backend/api/netbox_go/v1 plus canonical netbox-backend/api/proto",
    entries,
  );
}

const outputs = new Map([
  ["baseline-rest.yaml", baselineInventory()],
  ["current-rest.yaml", currentRESTInventory()],
  ["current-grpc.yaml", currentGRPCInventory()],
  ["current-vue.yaml", currentVueInventory()],
]);

const drift = [];
for (const [name, value] of outputs) {
  const filePath = path.join(OUT_DIR, name);
  const serialized = `${JSON.stringify(value, null, 2)}\n`;
  if (CHECK) {
    if (
      !fs.existsSync(filePath) ||
      fs.readFileSync(filePath, "utf8") !== serialized
    ) {
      drift.push(path.relative(ROOT, filePath));
    }
  } else {
    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    fs.writeFileSync(filePath, serialized);
  }
  console.log(`${name}: ${value.entries.length} entries`);
}

if (drift.length > 0) {
  console.error("Contract inventory is out of date:");
  for (const file of drift) console.error(`  ${file}`);
  process.exitCode = 1;
} else if (CHECK) {
  console.log("Contract inventory is up to date");
}
