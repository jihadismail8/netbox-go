import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import crypto from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  calculateRetainedClaimDigest,
  validateTraceability,
} from "./validate_traceability.mjs";
import {
  calculateSourceDigestFromEntries,
  calculateSourceManifest,
  calculateSourceManifestAtGitRevision,
} from "./source_digest.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const PROFILE_PATH = path.join(
  ROOT,
  "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
);
const MATRIX_PATH = path.join(
  ROOT,
  "contracts/netbox/v4.4.6-post7/traceability/core-workflow-v1.yaml",
);
const IDENTITY_PATH = path.join(
  ROOT,
  "contracts/netbox/v4.4.6-post7/resources/identity.yaml",
);
const BASE_PROFILE = JSON.parse(fs.readFileSync(PROFILE_PATH, "utf8"));
const BASE_DOCUMENT = JSON.parse(fs.readFileSync(MATRIX_PATH, "utf8"));
const BASE_IDENTITY = JSON.parse(fs.readFileSync(IDENTITY_PATH, "utf8"));
const CURRENT_SOURCE_MANIFEST = calculateSourceManifest(ROOT);
const CURRENT_SOURCE_DIGEST = CURRENT_SOURCE_MANIFEST.digest;
const CURRENT_GIT_REVISION = execFileSync(
  "git",
  [
    "--no-replace-objects",
    "-C",
    ROOT,
    "rev-parse",
    "--verify",
    "HEAD^{commit}",
  ],
  { encoding: "utf8" },
).trim();
const DIFFERENT_FILE_DIGEST =
  "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff";
const EMPTY_PAYLOAD_DIGEST =
  "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

function validateMutation(mutator) {
  const document = structuredClone(BASE_DOCUMENT);
  mutator(document);
  return validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
  });
}

function validateProfileMutation({
  profileMutator = () => {},
  identityMutator = () => {},
  documentMutator = () => {},
} = {}) {
  const profile = structuredClone(BASE_PROFILE);
  const identity = structuredClone(BASE_IDENTITY);
  const document = structuredClone(BASE_DOCUMENT);
  profileMutator(profile);
  identityMutator(identity);
  documentMutator(document);
  return validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    identityMetadataDocument: identity,
    traceabilityDocument: document,
  });
}

function failureCodes(result) {
  return new Set(result.failures.map((failure) => failure.code));
}

function assertRejected(result, expectedCode) {
  assert.ok(
    failureCodes(result).has(expectedCode),
    `expected ${expectedCode}; received ${JSON.stringify(result.failures, null, 2)}`,
  );
}

function removeRow(document, predicate) {
  const index = document.rows.findIndex(predicate);
  assert.notEqual(index, -1, "mutation target row must exist");
  document.rows.splice(index, 1);
}

function evidenceReference() {
  return {
    path: "docs/evidence/README.md",
    anchor: "# Evidence ledger",
  };
}

const RESULT_EVIDENCE_PATH = "docs/evidence/2026-08-01-core-workflow-v1-v0.md";
const RESULT_EVIDENCE_ANCHOR =
  "# Core Workflow V0 recovery evidence — 2026-08-01";
const IDENTITY_SECURITY_EVIDENCE_PATH =
  "docs/evidence/2026-08-02-identity-security.md";
const IDENTITY_SECURITY_EVIDENCE_ANCHOR =
  "# Identity security evidence — 2026-08-02";
const DIFFERENT_SOURCE_DIGEST =
  "source-v2:sha256:a55fab792cea1100e5fd2cc641fad02345189dd27d0b28c3b7ed2b1e1dcc22e1";

function resultEvidenceReference() {
  return {
    path: RESULT_EVIDENCE_PATH,
    anchor: RESULT_EVIDENCE_ANCHOR,
    payload_sha256: EMPTY_PAYLOAD_DIGEST,
  };
}

function identitySecurityEvidenceReference() {
  return {
    path: IDENTITY_SECURITY_EVIDENCE_PATH,
    anchor: IDENTITY_SECURITY_EVIDENCE_ANCHOR,
    payload_sha256: EMPTY_PAYLOAD_DIGEST,
  };
}

const TESTED_MANIFEST_PATH = "docs/evidence/2026-08-03-post-cleanup-v0.md";
const MATRIX_RELATIVE_PATH =
  "contracts/netbox/v4.4.6-post7/traceability/core-workflow-v1.yaml";

function sha256(source) {
  return `sha256:${crypto.createHash("sha256").update(source).digest("hex")}`;
}

function twoDigestFixture(changedPath = MATRIX_RELATIVE_PATH) {
  const entries = structuredClone(CURRENT_SOURCE_MANIFEST.entries);
  const tested = entries.find((entry) => entry.path === changedPath);
  assert.ok(tested, `missing current source-manifest entry for ${changedPath}`);
  assert.equal(tested.kind, "file", `${changedPath} must be a file`);
  const current = CURRENT_SOURCE_MANIFEST.entries.find(
    (entry) => entry.path === changedPath,
  );
  tested.size += 1;
  tested.sha256 = DIFFERENT_FILE_DIGEST;
  const digest = calculateSourceDigestFromEntries(entries);
  const manifest = {
    schema_version: 2,
    digest,
    files: entries.length,
    entries,
  };
  const source = `${JSON.stringify(manifest, null, 2)}\n`;
  return {
    digest,
    manifestReference: {
      path: TESTED_MANIFEST_PATH,
      sha256: sha256(source),
    },
    manifestSource: source,
    claim: {
      path: changedPath,
      tested_mode: tested.mode,
      tested_size: tested.size,
      tested_sha256: tested.sha256,
      attestation_mode: current.mode,
      attestation_size: current.size,
      attestation_sha256: current.sha256,
    },
  };
}

function evidenceAttestation(document, verification, overrides = {}) {
  verification.tested_revision ??= CURRENT_GIT_REVISION;
  const attestation = {
    schema_version: 2,
    profile: BASE_DOCUMENT.profile,
    compatibility_baseline: structuredClone(
      BASE_DOCUMENT.compatibility_baseline,
    ),
    verification_set: verification.id,
    classification: verification.classification,
    tier: verification.tier,
    proof_dimensions: ["domain"],
    tested_digest: verification.tested_digest,
    tested_revision: verification.tested_revision,
    attestation_digest: CURRENT_SOURCE_DIGEST,
    claim_digest: calculateRetainedClaimDigest(document, verification.id),
    claim_manifest: [],
    result: "pass",
    ...(verification.tested_source_manifest !== undefined
      ? {
          tested_source_manifest: structuredClone(
            verification.tested_source_manifest,
          ),
        }
      : {}),
    ...(verification.classification === "extension"
      ? {
          extension_verification: structuredClone(
            verification.extension_verification,
          ),
        }
      : {}),
    ...overrides,
  };
  return `<!-- netbox-go-evidence-v2: ${JSON.stringify(attestation)} -->`;
}

function evidenceDocumentsFor(document, verification, overrides = {}) {
  return new Map([
    [
      RESULT_EVIDENCE_PATH,
      evidenceAttestation(document, verification, {
        proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        ...overrides,
      }),
    ],
  ]);
}

function configureCoveredProjectVerification(document) {
  const row = document.rows.find(
    (item) => item.id === "scenario.contract.oracle-pin",
  );
  const verification = {
    id: "project-contract-retained-test",
    classification: "project",
    tier: "not_applicable",
    state: "retained",
    tested_digest: CURRENT_SOURCE_DIGEST,
    evidence: [resultEvidenceReference()],
  };
  document.verification_sets.push(verification);
  row.verification_set = verification.id;

  const assessment = {
    id: "project-contract-confirmed-test",
    status: "confirmed",
  };
  document.assessment_sets.push(assessment);
  row.assessment_set = assessment.id;

  const proofSet = structuredClone(
    document.proof_sets.find(
      (item) => item.id === "project-contract-row-pending",
    ),
  );
  proofSet.id = "project-contract-covered-test";
  const inventory = document.reference_sets.find(
    (item) => item.id === "project",
  ).test_inventory;
  for (const dimension of ["rest_differential", "grpc_parity", "browser"]) {
    proofSet.proofs[dimension] = {
      status: "covered",
      references: [structuredClone(inventory[dimension].references[0])],
    };
  }
  document.proof_sets.push(proofSet);
  document.row_proofs[row.id] = proofSet.id;
  return verification;
}

function materializeCommittedValidationFixture() {
  const fixtureRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), "netbox-go-traceability-"),
  );
  for (const entry of CURRENT_SOURCE_MANIFEST.entries) {
    const destination = path.join(fixtureRoot, entry.path);
    fs.mkdirSync(path.dirname(destination), { recursive: true });
    if (entry.kind === "symlink") {
      fs.symlinkSync(entry.target, destination);
      continue;
    }
    fs.copyFileSync(path.join(ROOT, entry.path), destination);
    fs.chmodSync(destination, entry.mode === "100755" ? 0o755 : 0o644);
  }
  execFileSync("git", ["-C", fixtureRoot, "init", "-q", "-b", "main"]);
  execFileSync("git", ["-C", fixtureRoot, "config", "user.name", "Test"]);
  execFileSync("git", [
    "-C",
    fixtureRoot,
    "config",
    "user.email",
    "test@example.invalid",
  ]);
  execFileSync("git", ["-C", fixtureRoot, "add", "-f", "-A"]);
  execFileSync("git", [
    "-C",
    fixtureRoot,
    "commit",
    "-q",
    "-m",
    "traceability fixture",
  ]);
  const revision = execFileSync(
    "git",
    ["-C", fixtureRoot, "rev-parse", "HEAD"],
    { encoding: "utf8" },
  ).trim();

  const netboxPaths = new Set();
  const collectNetBoxPaths = (value) => {
    if (Array.isArray(value)) {
      value.forEach(collectNetBoxPaths);
      return;
    }
    if (value === null || typeof value !== "object") return;
    if (typeof value.path === "string" && value.path.startsWith("netbox/")) {
      netboxPaths.add(value.path);
    }
    Object.values(value).forEach(collectNetBoxPaths);
  };
  collectNetBoxPaths(BASE_DOCUMENT);
  for (const relativePath of netboxPaths) {
    const destination = path.join(fixtureRoot, relativePath);
    fs.mkdirSync(path.dirname(destination), { recursive: true });
    fs.copyFileSync(path.join(ROOT, relativePath), destination);
  }
  fs.mkdirSync(path.join(fixtureRoot, "netbox/.git"), { recursive: true });
  fs.writeFileSync(
    path.join(fixtureRoot, "netbox/.git/HEAD"),
    `${BASE_DOCUMENT.compatibility_baseline.git_sha}\n`,
    "utf8",
  );
  fs.mkdirSync(path.join(fixtureRoot, "docs/evidence"), { recursive: true });
  fs.copyFileSync(
    path.join(ROOT, "docs/evidence/README.md"),
    path.join(fixtureRoot, "docs/evidence/README.md"),
  );
  fs.writeFileSync(
    path.join(fixtureRoot, RESULT_EVIDENCE_PATH),
    `${RESULT_EVIDENCE_ANCHOR}\n`,
    "utf8",
  );
  fs.writeFileSync(
    path.join(fixtureRoot, IDENTITY_SECURITY_EVIDENCE_PATH),
    `${IDENTITY_SECURITY_EVIDENCE_ANCHOR}\n`,
    "utf8",
  );
  return { fixtureRoot, revision };
}

test("the active profile has a complete authority-derived traceability denominator", () => {
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
  });

  assert.deepEqual(result.failures, []);
  assert.deepEqual(
    {
      scenarios: result.counts.scenarios,
      resource_operations: result.counts.resource_operations,
      resource_contracts: result.counts.resource_contracts,
      plan_rules: result.counts.plan_rules,
      rows: result.counts.rows,
      reference_sets: result.counts.reference_sets,
      applicability_sets: result.counts.applicability_sets,
      proof_sets: result.counts.proof_sets,
    },
    {
      scenarios: 17,
      resource_operations: 80,
      resource_contracts: 13,
      plan_rules: 183,
      rows: 293,
      reference_sets: 19,
      applicability_sets: 15,
      proof_sets: 15,
    },
  );
});

test("a reachable Git revision reconstructs committed owned bytes and executable modes", () => {
  const manifest = calculateSourceManifestAtGitRevision(
    ROOT,
    CURRENT_GIT_REVISION,
  );
  assert.match(manifest.digest, /^source-v2:sha256:[0-9a-f]{64}$/u);
  assert.equal(manifest.files, manifest.entries.length);
  const deploymentScript = manifest.entries.find(
    (entry) => entry.path === "tests/deployment/compose_smoke.sh",
  );
  assert.equal(deploymentScript?.kind, "file");
  assert.equal(deploymentScript?.mode, "100755");
});

test("an exact committed revision and exact claim marker form a valid retained attestation", () => {
  const { fixtureRoot, revision } = materializeCommittedValidationFixture();
  try {
    const manifest = calculateSourceManifest(fixtureRoot);
    assert.deepEqual(
      calculateSourceManifestAtGitRevision(fixtureRoot, revision),
      manifest,
    );
    const document = structuredClone(BASE_DOCUMENT);
    const verification = configureCoveredProjectVerification(document);
    verification.tested_digest = manifest.digest;
    verification.tested_revision = revision;
    const resultPayload = `${RESULT_EVIDENCE_ANCHOR}\n`;
    verification.evidence[0].payload_sha256 = sha256(
      Buffer.from(resultPayload, "utf8"),
    );
    const marker = evidenceAttestation(document, verification, {
      attestation_digest: manifest.digest,
      proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
    });
    const result = validateTraceability({
      root: fixtureRoot,
      profilePath: path.join(
        fixtureRoot,
        "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
      ),
      traceabilityDocument: document,
      evidenceDocuments: new Map([
        [RESULT_EVIDENCE_PATH, `${resultPayload}${marker}`],
      ]),
    });
    assert.deepEqual(result.failures, []);

    const extensionDocument = structuredClone(BASE_DOCUMENT);
    const extensionVerification = {
      id: "identity-cli-partial-committed-test",
      classification: "extension",
      tier: "not_applicable",
      state: "retained",
      tested_digest: manifest.digest,
      tested_revision: revision,
      evidence: [resultEvidenceReference()],
      extension_verification: structuredClone(
        BASE_PROFILE.identity_extension.verification,
      ),
    };
    extensionDocument.verification_sets.push(extensionVerification);
    const extensionRow = extensionDocument.rows.find(
      (row) => row.id === "rule.plan.identity.admin.bootstrap-empty-store-once",
    );
    extensionRow.verification_set = extensionVerification.id;
    extensionDocument.assessment_sets.push({
      id: "identity-cli-partial-committed-assessment",
      status: "confirmed",
    });
    extensionRow.assessment_set = "identity-cli-partial-committed-assessment";
    const extensionProof = structuredClone(
      extensionDocument.proof_sets.find(
        (proof) => proof.id === "identity-cli-row-pending",
      ),
    );
    extensionProof.id = "identity-cli-partial-committed-proof";
    extensionProof.proofs.cli_security = {
      status: "covered",
      references: [
        structuredClone(
          extensionDocument.reference_sets.find(
            (referenceSet) => referenceSet.id === "identity",
          ).test_inventory.cli_security.references[0],
        ),
      ],
    };
    extensionDocument.proof_sets.push(extensionProof);
    extensionDocument.row_proofs[extensionRow.id] = extensionProof.id;
    extensionVerification.evidence[0].payload_sha256 = sha256(
      Buffer.from(resultPayload, "utf8"),
    );
    const extensionMarker = evidenceAttestation(
      extensionDocument,
      extensionVerification,
      {
        attestation_digest: manifest.digest,
        proof_dimensions: ["cli_security"],
      },
    );
    const extensionResult = validateTraceability({
      root: fixtureRoot,
      profilePath: path.join(
        fixtureRoot,
        "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
      ),
      traceabilityDocument: extensionDocument,
      evidenceDocuments: new Map([
        [RESULT_EVIDENCE_PATH, `${resultPayload}${extensionMarker}`],
      ]),
    });
    assert.deepEqual(extensionResult.failures, []);
    assert.deepEqual(extensionVerification.extension_verification, {
      contract: "partial",
      parity: "partial",
      security: "partial",
    });
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("a committed CLI-only extension split requires only its applicable completed axis", () => {
  const { fixtureRoot, revision } = materializeCommittedValidationFixture();
  try {
    const manifest = calculateSourceManifest(fixtureRoot);
    const profile = structuredClone(BASE_PROFILE);
    const identity = structuredClone(BASE_IDENTITY);
    const document = structuredClone(BASE_DOCUMENT);
    profile.identity_extension.verification.security = "complete";
    identity.verification.security = "complete";

    const originalExtensionVerification = document.verification_sets.find(
      (verification) => verification.id === "identity-extension-pending",
    );
    originalExtensionVerification.extension_verification.security = "complete";
    const targetVerification = {
      id: "identity-cli-security-retained-test",
      classification: "extension",
      tier: "not_applicable",
      state: "retained",
      tested_digest: manifest.digest,
      tested_revision: revision,
      evidence: [resultEvidenceReference()],
      extension_verification: structuredClone(
        profile.identity_extension.verification,
      ),
    };
    const supportVerification = {
      id: "identity-security-support-retained-test",
      classification: "extension",
      tier: "not_applicable",
      state: "retained",
      tested_digest: manifest.digest,
      tested_revision: revision,
      evidence: [identitySecurityEvidenceReference()],
      extension_verification: structuredClone(
        profile.identity_extension.verification,
      ),
    };
    document.verification_sets.push(targetVerification, supportVerification);

    const confirmedAssessment = {
      id: "identity-security-confirmed-test",
      status: "confirmed",
    };
    document.assessment_sets.push(confirmedAssessment);
    const verificationSets = new Map(
      document.verification_sets.map((verification) => [
        verification.id,
        verification,
      ]),
    );
    const applicabilitySets = new Map(
      document.applicability_sets.map((applicability) => [
        applicability.id,
        applicability,
      ]),
    );
    const proofSets = new Map(
      document.proof_sets.map((proof) => [proof.id, proof]),
    );
    const identityInventory = document.reference_sets.find(
      (referenceSet) => referenceSet.id === "identity",
    ).test_inventory;
    const targetRowID = "rule.plan.identity.admin.bootstrap-empty-store-once";
    const securityDimensions = [
      "browser",
      "rest_extension_contract",
      "cli_security",
    ];
    const extensionRows = document.rows.filter(
      (row) =>
        verificationSets.get(row.verification_set)?.classification ===
        "extension",
    );
    for (const [index, row] of extensionRows.entries()) {
      row.verification_set =
        row.id === targetRowID ? targetVerification.id : supportVerification.id;
      row.assessment_set = confirmedAssessment.id;
      const proof = structuredClone(proofSets.get(document.row_proofs[row.id]));
      proof.id = `identity-security-covered-${index}`;
      const applicability = applicabilitySets.get(
        document.row_applicability[row.id],
      );
      for (const dimension of securityDimensions) {
        if (applicability.dimensions[dimension].status !== "applicable") {
          continue;
        }
        proof.proofs[dimension] = {
          status: "covered",
          references: [
            structuredClone(identityInventory[dimension].references[0]),
          ],
        };
      }
      document.proof_sets.push(proof);
      document.row_proofs[row.id] = proof.id;
    }
    const usedProofSets = new Set(Object.values(document.row_proofs));
    document.proof_sets = document.proof_sets.filter((proof) =>
      usedProofSets.has(proof.id),
    );
    const usedVerificationSets = new Set(
      document.rows.map((row) => row.verification_set),
    );
    document.verification_sets = document.verification_sets.filter(
      (verification) => usedVerificationSets.has(verification.id),
    );

    const targetPayload = `${RESULT_EVIDENCE_ANCHOR}\n`;
    const supportPayload = `${IDENTITY_SECURITY_EVIDENCE_ANCHOR}\n`;
    targetVerification.evidence[0].payload_sha256 = sha256(
      Buffer.from(targetPayload, "utf8"),
    );
    supportVerification.evidence[0].payload_sha256 = sha256(
      Buffer.from(supportPayload, "utf8"),
    );
    const targetMarker = evidenceAttestation(document, targetVerification, {
      attestation_digest: manifest.digest,
      proof_dimensions: ["cli_security"],
    });
    const supportMarker = evidenceAttestation(document, supportVerification, {
      attestation_digest: manifest.digest,
      proof_dimensions: securityDimensions,
    });
    const result = validateTraceability({
      root: fixtureRoot,
      profilePath: path.join(
        fixtureRoot,
        "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
      ),
      profileDocument: profile,
      identityMetadataDocument: identity,
      traceabilityDocument: document,
      evidenceDocuments: new Map([
        [RESULT_EVIDENCE_PATH, `${targetPayload}${targetMarker}`],
        [IDENTITY_SECURITY_EVIDENCE_PATH, `${supportPayload}${supportMarker}`],
      ]),
    });
    assert.deepEqual(result.failures, []);
    assert.match(targetMarker, /"proof_dimensions":\["cli_security"\]/u);
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("source digests distinguish executable mode changes", () => {
  const entries = structuredClone(CURRENT_SOURCE_MANIFEST.entries);
  const executable = entries.find(
    (entry) => entry.kind === "file" && entry.mode === "100755",
  );
  assert.ok(executable, "fixture requires an executable owned file");
  const before = calculateSourceDigestFromEntries(entries);
  executable.mode = "100644";
  assert.notEqual(calculateSourceDigestFromEntries(entries), before);
});

test("source manifests reject an omitted executable mode", () => {
  const entries = structuredClone(CURRENT_SOURCE_MANIFEST.entries);
  const executable = entries.find(
    (entry) => entry.kind === "file" && entry.mode === "100755",
  );
  assert.ok(executable, "fixture requires an executable owned file");
  delete executable.mode;
  assert.throws(
    () => calculateSourceDigestFromEntries(entries),
    /source-manifest file entry is invalid/u,
  );
});

test("historical source reconstruction ignores local Git replacement refs", () => {
  const { fixtureRoot, revision } = materializeCommittedValidationFixture();
  try {
    const trustedManifest = calculateSourceManifestAtGitRevision(
      fixtureRoot,
      revision,
    );
    fs.appendFileSync(
      path.join(fixtureRoot, "AGENTS.md"),
      "\nforged replacement content\n",
      "utf8",
    );
    execFileSync("git", ["-C", fixtureRoot, "add", "AGENTS.md"]);
    execFileSync("git", [
      "-C",
      fixtureRoot,
      "commit",
      "-q",
      "-m",
      "forged replacement",
    ]);
    const replacement = execFileSync(
      "git",
      ["-C", fixtureRoot, "rev-parse", "HEAD"],
      { encoding: "utf8" },
    ).trim();
    execFileSync("git", ["-C", fixtureRoot, "replace", revision, replacement]);
    const replacedAgents = execFileSync(
      "git",
      ["-C", fixtureRoot, "show", `${revision}:AGENTS.md`],
      { encoding: "utf8" },
    );
    assert.match(replacedAgents, /forged replacement content/u);
    assert.deepEqual(
      calculateSourceManifestAtGitRevision(fixtureRoot, revision),
      trustedManifest,
    );
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("historical source reachability ignores local Git grafts", () => {
  const { fixtureRoot, revision } = materializeCommittedValidationFixture();
  try {
    const tree = execFileSync(
      "git",
      ["--no-replace-objects", "-C", fixtureRoot, "rev-parse", "HEAD^{tree}"],
      { encoding: "utf8" },
    ).trim();
    const unreachable = execFileSync(
      "git",
      ["--no-replace-objects", "-C", fixtureRoot, "commit-tree", tree],
      {
        encoding: "utf8",
        env: {
          ...process.env,
          GIT_AUTHOR_DATE: "2000-01-04T00:00:00Z",
          GIT_COMMITTER_DATE: "2000-01-04T00:00:00Z",
        },
        input: "unreachable source fixture\n",
      },
    ).trim();
    fs.writeFileSync(
      path.join(fixtureRoot, ".git/info/grafts"),
      `${revision} ${unreachable}\n`,
      "ascii",
    );
    execFileSync("git", [
      "-c",
      "advice.graftFileDeprecated=false",
      "-C",
      fixtureRoot,
      "merge-base",
      "--is-ancestor",
      unreachable,
      revision,
    ]);
    assert.throws(
      () => calculateSourceManifestAtGitRevision(fixtureRoot, unreachable),
      /is not reachable from the current HEAD/u,
    );
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("retained claim digests bind exact consumers, assessments, applicability, and proofs", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.tested_revision = CURRENT_GIT_REVISION;
  const original = calculateRetainedClaimDigest(document, verification.id);

  const consumerMutation = structuredClone(document);
  consumerMutation.rows.find(
    (row) => row.id === "scenario.contract.surface-manifest",
  ).verification_set = verification.id;
  assert.notEqual(
    calculateRetainedClaimDigest(consumerMutation, verification.id),
    original,
  );

  const assessmentMutation = structuredClone(document);
  assessmentMutation.assessment_sets.find(
    (assessment) => assessment.id === "project-contract-confirmed-test",
  ).status = "not_applicable";
  assert.notEqual(
    calculateRetainedClaimDigest(assessmentMutation, verification.id),
    original,
  );

  const proofMutation = structuredClone(document);
  proofMutation.proof_sets.find(
    (proof) => proof.id === "project-contract-covered-test",
  ).proofs.browser.status = "partial";
  assert.notEqual(
    calculateRetainedClaimDigest(proofMutation, verification.id),
    original,
  );
});

test("reviewed semantic bindings remain attached to the owning capability and operation", () => {
  const expected = new Map([
    [
      "dcim.rack-placement-validation",
      ["dcim.Device", "workflow", "dcim.Device"],
    ],
    [
      "plan.dcim.rack.manual-position",
      ["dcim.Device", "validate", "dcim.Device"],
    ],
    [
      "plan.dcim.device-create.defer-other-template-families",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    [
      "plan.dcim.interface-template.bridge-deferred",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    [
      "plan.dcim.device-role.delete-descendant-cascade",
      ["dcim.DeviceRole", "delete", "dcim.DeviceRole"],
    ],
    [
      "plan.dcim.device-role.delete-device-protection",
      ["dcim.DeviceRole", "delete", "dcim.DeviceRole"],
    ],
    [
      "plan.dcim.device.airflow-create-inheritance",
      ["dcim.Device", "create", "dcim.Device"],
    ],
    [
      "plan.ipam.assignment.unassigned-requires-pair",
      ["ipam.IPAddress", "assign", "ipam.IPAddress"],
    ],
    ["plan.identity.token.one-time-secret", ["identity", "create", "identity"]],
    [
      "plan.identity.token.owner-only-management",
      ["identity", "authorize", "identity"],
    ],
    [
      "plan.identity.token.defer-baseline-crud",
      ["identity", "reject", "identity"],
    ],
    ["plan.identity.token.maintenance-off", ["profile", "validate", "project"]],
    [
      "plan.identity.interfaces.explicit-contract",
      ["identity", "contract", "identity-interface-support"],
    ],
    [
      "plan.identity.token.authenticated-extension",
      ["identity", "contract", "identity"],
    ],
    [
      "plan.identity.token.extension-not-t2",
      ["identity", "validate", "identity"],
    ],
    [
      "plan.ipam.assignment.grpc-actions",
      ["ipam.IPAddress", "contract", "ipam.IPAddress"],
    ],
    [
      "plan.ipam.assignment.shared-transport-use-case",
      ["ipam.IPAddress", "contract", "ipam.IPAddress"],
    ],
    [
      "plan.ipam.interface-vrf.deferred",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    [
      "plan.ipam.interface-vrf.future-set-null",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    [
      "plan.ipam.prefix.flags-no-auto-allocation",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    [
      "plan.ipam.uniqueness.global-default",
      ["profile", "validate", "baseline-common-support"],
    ],
    [
      "plan.ipam.uniqueness.vrf-enforce-unique",
      ["ipam.VRF", "validate", "ipam.VRF"],
    ],
    [
      "plan.profile.transport.central-error-mapping",
      ["profile", "contract", "project"],
    ],
    [
      "plan.common.response.standard-envelope",
      ["profile", "get", "baseline-common-support"],
    ],
    [
      "plan.common.collection.reject-bulk",
      ["profile", "reject", "deferred-baseline-support"],
    ],
    ["plan.common.fields.reject-undeclared", ["profile", "reject", "project"]],
    [
      "plan.common.resource.six-single-object-operations",
      ["profile", "contract", "project"],
    ],
    [
      "plan.profile.mutation.single-transaction",
      ["profile", "persist", "project"],
    ],
    [
      "plan.dcim.rack.direct-save-full-validation",
      ["dcim.Rack", "update", "dcim.Rack"],
    ],
    [
      "plan.dcim.rack.direct-save-mounted-device-protection",
      ["dcim.Rack", "update", "dcim.Rack"],
    ],
  ]);

  for (const [sourceID, binding] of expected) {
    const row = BASE_DOCUMENT.rows.find((item) => item.source_id === sourceID);
    assert.ok(row, `missing reviewed semantic binding for ${sourceID}`);
    assert.deepEqual(
      [row.capability, row.operation, row.reference_set],
      binding,
      `semantic binding drift for ${sourceID}`,
    );
  }
});

test("reviewed multi-operation and multi-capability bindings remain explicit", () => {
  const coveredOperations = new Map([
    [
      "contract.transport-errors",
      [
        "authenticate",
        "authorize",
        "list",
        "get",
        "create",
        "replace",
        "update",
        "delete",
        "assign",
        "unassign",
        "validate",
      ],
    ],
    ["dcim.device-instantiates-interfaces", ["create", "get", "update"]],
    ["identity.missing-credential-denied", ["list", "get"]],
    [
      "identity.permission-matrix",
      [
        "authorize",
        "list",
        "get",
        "create",
        "replace",
        "update",
        "delete",
        "assign",
        "unassign",
      ],
    ],
    [
      "identity.token-lifecycle",
      ["create", "authenticate", "authorize", "list", "delete"],
    ],
    ["ipam.prefix-and-address-semantics", ["list", "create"]],
    [
      "plan.identity.token.authenticated-extension",
      ["list", "create", "delete"],
    ],
    [
      "plan.identity.interfaces.explicit-contract",
      [
        "authenticate",
        "authorize",
        "list",
        "get",
        "create",
        "delete",
        "login",
        "logout",
        "password_change",
        "administer",
      ],
    ],
    ["plan.identity.token.one-time-secret", ["create", "list"]],
    [
      "plan.identity.token.owner-only-management",
      ["authorize", "list", "delete"],
    ],
    [
      "plan.identity.token.defer-baseline-crud",
      ["list", "get", "create", "replace", "update", "delete"],
    ],
    ["plan.identity.token.revoke-immediate", ["delete", "authenticate"]],
    ["plan.ipam.assignment.grpc-actions", ["assign", "unassign"]],
    [
      "plan.ipam.assignment.shared-transport-use-case",
      ["update", "assign", "unassign"],
    ],
    [
      "plan.profile.transport.shared-application-use-case",
      [
        "list",
        "get",
        "create",
        "replace",
        "update",
        "delete",
        "assign",
        "unassign",
      ],
    ],
    [
      "plan.deferred.physical-cabling",
      ["list", "get", "create", "replace", "update", "delete", "query"],
    ],
    [
      "plan.common.response.standard-envelope",
      ["list", "get", "create", "replace", "update"],
    ],
  ]);
  for (const [sourceID, expectedOperations] of coveredOperations) {
    const row = BASE_DOCUMENT.rows.find((item) => item.source_id === sourceID);
    assert.ok(row, `missing reviewed operation coverage for ${sourceID}`);
    assert.deepEqual(
      row.covered_operations,
      expectedOperations,
      `covered-operation drift for ${sourceID}`,
    );
  }

  const affectedCapabilities = new Map([
    [
      "dcim.device-instantiates-interfaces",
      ["dcim.Device", "dcim.InterfaceTemplate", "dcim.Interface"],
    ],
    ["dcim.device-instantiation-rollback", ["dcim.Device", "dcim.Interface"]],
    [
      "dcim.relationship-delete-semantics",
      [
        "dcim.Site",
        "dcim.Manufacturer",
        "dcim.RackRole",
        "dcim.RackType",
        "dcim.Rack",
        "dcim.DeviceRole",
        "dcim.DeviceType",
        "dcim.InterfaceTemplate",
        "dcim.Device",
        "dcim.Interface",
        "ipam.IPAddress",
      ],
    ],
    ["ipam.address-edge-rules", ["ipam.IPAddress", "dcim.Interface"]],
    ["ipam.assign-interface", ["ipam.IPAddress", "dcim.Interface"]],
    ["ipam.assignment-presence-matrix", ["ipam.IPAddress", "dcim.Interface"]],
    ["identity.missing-credential-denied", ["identity", "profile"]],
    ["identity.permission-matrix", ["identity", "profile"]],
    ["ipam.prefix-and-address-semantics", ["ipam.Prefix", "ipam.IPAddress"]],
    [
      "plan.ipam.uniqueness.global-default",
      ["profile", "ipam.Prefix", "ipam.IPAddress"],
    ],
    [
      "plan.ipam.uniqueness.vrf-enforce-unique",
      ["ipam.VRF", "ipam.Prefix", "ipam.IPAddress"],
    ],
  ]);
  const addAffectedGroup = (sourceIDs, capabilities) => {
    for (const sourceID of sourceIDs) {
      affectedCapabilities.set(sourceID, capabilities);
    }
  };
  addAffectedGroup(
    [
      "plan.dcim.device-create.copy-template-fields",
      "plan.dcim.device-create.omit-template-description",
    ],
    ["dcim.Device", "dcim.InterfaceTemplate", "dcim.Interface"],
  );
  addAffectedGroup(
    [
      "plan.dcim.device-create.record-changes",
      "plan.dcim.device-create.atomic-commit",
      "plan.dcim.device-create.interfaces-and-changes-atomic",
      "plan.dcim.delete.device-cascades-interfaces",
    ],
    ["dcim.Device", "dcim.Interface"],
  );
  addAffectedGroup(
    [
      "plan.dcim.device-create.snapshot-templates",
      "plan.dcim.device-create.bridge-free-scope",
    ],
    ["dcim.Device", "dcim.InterfaceTemplate"],
  );
  addAffectedGroup(
    [
      "plan.dcim.interface-template.non-retroactive",
      "plan.dcim.interface-template.edits-do-not-modify-interfaces",
    ],
    ["dcim.InterfaceTemplate", "dcim.Interface"],
  );
  addAffectedGroup(
    [
      "plan.dcim.rack-type.propagates-rack-physical-fields",
      "plan.dcim.rack-type.propagation-skips-rack-full-validation",
      "plan.dcim.delete.rack-type-protected",
    ],
    ["dcim.RackType", "dcim.Rack"],
  );
  addAffectedGroup(
    [
      "plan.dcim.device-type.height-increase-fit",
      "plan.dcim.device-type.height-zero-position-protection",
      "plan.dcim.delete.device-type-protected",
    ],
    ["dcim.DeviceType", "dcim.Device"],
  );
  addAffectedGroup(
    [
      "plan.dcim.rack.direct-save-mounted-device-protection",
      "plan.dcim.delete.rack-protected",
    ],
    ["dcim.Rack", "dcim.Device"],
  );
  addAffectedGroup(
    [
      "plan.dcim.device-role.delete-device-protection",
      "plan.dcim.delete.device-role-protected",
    ],
    ["dcim.DeviceRole", "dcim.Device"],
  );
  addAffectedGroup(
    [
      "plan.dcim.delete.interface-cascades-ip-addresses",
      "plan.dcim.delete.interface-cascade-change-effects",
    ],
    ["dcim.Interface", "ipam.IPAddress"],
  );
  addAffectedGroup(
    [
      "plan.ipam.assignment.atomic-update",
      "plan.ipam.assignment.both-null-unassign",
      "plan.ipam.assignment.both-omitted-preserve",
      "plan.ipam.assignment.edge-prefix-exceptions",
      "plan.ipam.assignment.grpc-actions",
      "plan.ipam.assignment.grpc-update-presence",
      "plan.ipam.assignment.interface-only-target",
      "plan.ipam.assignment.invalid-pairs-reject",
      "plan.ipam.assignment.no-interface-vrf-equality",
      "plan.ipam.assignment.partial-reassignment",
      "plan.ipam.assignment.reject-network-broadcast",
      "plan.ipam.assignment.shared-transport-use-case",
      "plan.ipam.assignment.unassigned-requires-pair",
      "ipam.IPAddress.AssignIPAddress",
      "ipam.IPAddress.UnassignIPAddress",
    ],
    ["ipam.IPAddress", "dcim.Interface"],
  );
  addAffectedGroup(
    [
      "plan.identity.authorization.shared-rbac",
      "plan.identity.interfaces.explicit-contract",
      "plan.identity.rbac.direct-group-global-object",
      "plan.identity.rbac.visibility-before-pagination",
      "plan.identity.rbac.authorization-before-write",
      "plan.identity.token.credential-use-baseline",
      "plan.identity.token.fail-closed-conditions",
      "plan.identity.token.last-used-before-restrictions",
      "plan.identity.token.last-used-rate",
      "plan.identity.token.rest-token-scheme",
      "plan.identity.token.shared-authentication-path",
      "plan.identity.token.unknown-key-no-write",
    ],
    ["identity", "profile"],
  );
  addAffectedGroup(
    [
      "plan.profile.authorization.authenticate-default",
      "plan.profile.authorization.permission-parity",
    ],
    ["profile", "identity"],
  );
  affectedCapabilities.set("plan.dcim.device.airflow-create-inheritance", [
    "dcim.Device",
    "dcim.DeviceType",
  ]);
  affectedCapabilities.set("plan.dcim.delete.site-protected", [
    "dcim.Site",
    "dcim.Rack",
    "dcim.Device",
  ]);
  affectedCapabilities.set("plan.dcim.delete.rack-role-protected", [
    "dcim.RackRole",
    "dcim.Rack",
  ]);
  affectedCapabilities.set("plan.dcim.delete.manufacturer-protected", [
    "dcim.Manufacturer",
    "dcim.RackType",
    "dcim.DeviceType",
  ]);
  affectedCapabilities.set("plan.dcim.delete.device-type-cascades-templates", [
    "dcim.DeviceType",
    "dcim.InterfaceTemplate",
  ]);
  affectedCapabilities.set("plan.ipam.delete.vrf-protected", [
    "ipam.VRF",
    "ipam.Prefix",
    "ipam.IPAddress",
  ]);
  for (const sourceID of [
    "plan.ipam.interface-vrf.deferred",
    "plan.ipam.interface-vrf.future-set-null",
  ]) {
    affectedCapabilities.set(sourceID, [
      "profile",
      "dcim.Interface",
      "ipam.VRF",
    ]);
  }
  affectedCapabilities.set("plan.ipam.prefix.flags-no-auto-allocation", [
    "profile",
    "ipam.Prefix",
  ]);
  for (const [sourceID, expectedCapabilities] of affectedCapabilities) {
    const row = BASE_DOCUMENT.rows.find((item) => item.source_id === sourceID);
    assert.ok(
      row,
      `missing reviewed affected-capability binding for ${sourceID}`,
    );
    assert.deepEqual(
      row.affected_capabilities,
      expectedCapabilities,
      `affected-capability drift for ${sourceID}`,
    );
  }
});

test("transport-specific rows select proof sets with exact applicable boundaries", () => {
  const expected = new Map([
    [
      "scenario.identity.local-administration",
      "identity-cli-network-row-pending",
    ],
    ["scenario.identity.session-login-logout", "identity-browser-row-pending"],
    [
      "rule.plan.identity.grpc.no-cookie-equivalent",
      "grpc-contract-row-pending",
    ],
    ["rule.plan.identity.token.grpc-bearer", "grpc-contract-row-pending"],
    ["rule.plan.common.list.grpc-typed-fields", "grpc-contract-row-pending"],
    ["operation.ipam.ip-address.assign-ipaddress", "grpc-runtime-row-pending"],
    [
      "operation.ipam.ip-address.unassign-ipaddress",
      "grpc-runtime-row-pending",
    ],
    [
      "rule.plan.identity.interfaces.explicit-contract",
      "project-identity-interface-row-pending",
    ],
    [
      "rule.plan.common.fields.promote-vertically",
      "project-runtime-row-pending",
    ],
    ["rule.plan.ipam.uniqueness.global-default", "project-runtime-row-pending"],
    [
      "rule.plan.identity.token.credential-use-baseline",
      "rest-runtime-row-pending",
    ],
    [
      "rule.plan.identity.authorization.shared-rbac",
      "rest-grpc-runtime-row-pending",
    ],
    [
      "rule.plan.identity.token.authenticated-extension",
      "identity-rest-grpc-contract-row-pending",
    ],
    [
      "rule.plan.identity.token.defer-baseline-crud",
      "identity-rest-contract-row-pending",
    ],
    [
      "rule.plan.identity.session.current-principal",
      "identity-browser-grpc-row-pending",
    ],
  ]);
  for (const [rowID, proofSetID] of expected) {
    assert.equal(
      BASE_DOCUMENT.row_proofs[rowID],
      proofSetID,
      `proof boundary drift for ${rowID}`,
    );
  }
});

test("proof-set templates preserve exact transport applicability", () => {
  const dimensions = BASE_DOCUMENT.proof_dimensions;
  const expected = new Map([
    [
      "baseline-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "identity-extension-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "pending",
        "not_applicable",
        "pending",
        "not_applicable",
      ],
    ],
    [
      "rest-grpc-runtime-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "project-runtime-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "project-contract-row-pending",
      [
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "identity-browser-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
        "pending",
        "pending",
        "not_applicable",
      ],
    ],
    [
      "identity-cli-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "pending",
      ],
    ],
    [
      "grpc-contract-row-pending",
      [
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "pending",
        "not_applicable",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "rest-runtime-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
      ],
    ],
    [
      "identity-browser-grpc-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "pending",
        "pending",
        "pending",
        "not_applicable",
      ],
    ],
    [
      "identity-rest-grpc-contract-row-pending",
      [
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "pending",
        "not_applicable",
        "pending",
        "not_applicable",
      ],
    ],
    [
      "identity-rest-contract-row-pending",
      [
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "not_applicable",
        "pending",
        "not_applicable",
      ],
    ],
    [
      "identity-cli-network-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "pending",
        "not_applicable",
        "pending",
        "pending",
      ],
    ],
    [
      "project-identity-interface-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
        "pending",
      ],
    ],
    [
      "grpc-runtime-row-pending",
      [
        "pending",
        "pending",
        "pending",
        "not_applicable",
        "pending",
        "not_applicable",
        "not_applicable",
        "not_applicable",
      ],
    ],
  ]);
  assert.equal(BASE_DOCUMENT.proof_sets.length, expected.size);
  for (const proofSet of BASE_DOCUMENT.proof_sets) {
    assert.deepEqual(
      dimensions.map((dimension) => proofSet.proofs[dimension].status),
      expected.get(proofSet.id),
      `proof applicability drift for ${proofSet.id}`,
    );
  }

  const applicabilityIDByProof = new Map([
    ["baseline-row-pending", "baseline-row"],
    ["identity-extension-row-pending", "identity-extension-row"],
    ["rest-grpc-runtime-row-pending", "rest-grpc-runtime-row"],
    ["project-runtime-row-pending", "project-runtime-row"],
    ["project-contract-row-pending", "project-contract-row"],
    ["identity-browser-row-pending", "identity-browser-row"],
    ["identity-cli-row-pending", "identity-cli-row"],
    ["grpc-contract-row-pending", "grpc-contract-row"],
    ["rest-runtime-row-pending", "rest-runtime-row"],
    ["identity-browser-grpc-row-pending", "identity-browser-grpc-row"],
    [
      "identity-rest-grpc-contract-row-pending",
      "identity-rest-grpc-contract-row",
    ],
    ["identity-rest-contract-row-pending", "identity-rest-contract-row"],
    ["identity-cli-network-row-pending", "identity-cli-network-row"],
    [
      "project-identity-interface-row-pending",
      "project-identity-interface-row",
    ],
    ["grpc-runtime-row-pending", "grpc-runtime-row"],
  ]);
  assert.equal(BASE_DOCUMENT.applicability_sets.length, expected.size);
  for (const proofSet of BASE_DOCUMENT.proof_sets) {
    const applicabilitySet = BASE_DOCUMENT.applicability_sets.find(
      (item) => item.id === applicabilityIDByProof.get(proofSet.id),
    );
    assert.ok(applicabilitySet, `missing applicability for ${proofSet.id}`);
    assert.deepEqual(
      dimensions.map(
        (dimension) => applicabilitySet.dimensions[dimension].status,
      ),
      expected
        .get(proofSet.id)
        .map((status) =>
          status === "not_applicable" ? "not_applicable" : "applicable",
        ),
      `reviewed applicability drift for ${proofSet.id}`,
    );
  }
});

test("a missing scenario row is rejected", () => {
  const result = validateMutation((document) => {
    removeRow(
      document,
      (row) =>
        row.kind === "scenario" && row.source_id === "contract.oracle-pin",
    );
  });
  assertRejected(result, "missing-scenario-row");
});

test("a missing resource-operation row is rejected", () => {
  const result = validateMutation((document) => {
    removeRow(
      document,
      (row) =>
        row.kind === "resource_operation" && row.source_id === "dcim.Site.list",
    );
  });
  assertRejected(result, "missing-operation-row");
});

test("a missing resource-contract row is rejected", () => {
  const result = validateMutation((document) => {
    removeRow(
      document,
      (row) =>
        row.kind === "resource_contract" && row.source_id === "dcim.Site",
    );
  });
  assertRejected(result, "missing-resource-contract-row");
});

test("a missing stable plan-rule row is rejected", () => {
  const result = validateMutation((document) => {
    removeRow(
      document,
      (row) =>
        row.kind === "plan_rule" &&
        row.source_id === "plan.profile.surface.manifest-exact",
    );
  });
  assertRejected(result, "missing-plan-rule-row");
});

test("a stale reference path is rejected", () => {
  const result = validateMutation((document) => {
    document.reference_sets[0].test_inventory.rest_differential.references[0].path =
      "tests/compatibility/does-not-exist.mjs";
  });
  assertRejected(result, "stale-reference-path");
});

test("a stale exact anchor is rejected", () => {
  const result = validateMutation((document) => {
    document.reference_sets[0].test_inventory.rest_differential.references[0].anchor =
      "an anchor which does not exist";
  });
  assertRejected(result, "stale-reference-anchor");
});

test("a resource reference set cannot select another resource metadata object", () => {
  const result = validateMutation((document) => {
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    site.metadata_ref.key = "dcim.Manufacturer";
  });
  assertRejected(result, "invalid-metadata-selector");
});

test("a proof set with a missing proof dimension is rejected", () => {
  const result = validateMutation((document) => {
    delete document.proof_sets[0].proofs.cli_security;
  });
  assertRejected(result, "missing-proof-dimension");
});

test("every row must select one reviewed applicability set", () => {
  const result = validateMutation((document) => {
    delete document.row_applicability[document.rows[0].id];
  });
  assertRejected(result, "missing-row-applicability");
});

test("an applicability mapping cannot name an unknown row", () => {
  const result = validateMutation((document) => {
    document.row_applicability["unknown.row"] = "baseline-row";
  });
  assertRejected(result, "extra-row-applicability");
});

test("applicability sets must declare every proof dimension", () => {
  const result = validateMutation((document) => {
    delete document.applicability_sets[0].dimensions.cli_security;
  });
  assertRejected(result, "missing-applicability-dimension");
});

test("not-applicable authority requires an exact narrow reason", () => {
  const result = validateMutation((document) => {
    const declaration =
      document.applicability_sets[0].dimensions.rest_extension_contract;
    delete declaration.reason;
  });
  assertRejected(result, "invalid-applicability");
});

test("an applicable dimension cannot be changed to a not-applicable proof", () => {
  const result = validateMutation((document) => {
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.domain;
    proof.status = "not_applicable";
    proof.references = [];
    proof.reason = "Adversarial proof-only applicability change.";
  });
  assertRejected(result, "proof-applicability-mismatch");
});

test("a reviewed not-applicable dimension cannot be changed to pending proof", () => {
  const result = validateMutation((document) => {
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.rest_extension_contract;
    proof.status = "pending";
    proof.reason = "Adversarial proof-only applicability change.";
  });
  assertRejected(result, "proof-applicability-mismatch");
});

test("one proof set cannot serve different reviewed applicability sets", () => {
  const result = validateMutation((document) => {
    document.row_proofs["rule.plan.common.list.grpc-typed-fields"] =
      "project-contract-row-pending";
  });
  assertRejected(result, "shared-proof-applicability");
});

test("the reviewed 293-row applicability authority cannot be remapped", () => {
  const result = validateMutation((document) => {
    document.row_applicability["rule.plan.identity.token.one-time-secret"] =
      "identity-rest-contract-row";
    document.row_proofs["rule.plan.identity.token.one-time-secret"] =
      "identity-rest-contract-row-pending";
  });
  assertRejected(result, "reviewed-applicability-drift");
});

test("a coordinated all-N/A applicability and proof pair is rejected", () => {
  const result = validateMutation((document) => {
    const reason = "Adversarial row-local N/A bypass.";
    document.applicability_sets.push({
      id: "adversarial-row-na",
      dimensions: Object.fromEntries(
        document.proof_dimensions.map((dimension) => [
          dimension,
          { status: "not_applicable", reason },
        ]),
      ),
    });
    document.proof_sets.push({
      id: "adversarial-row-na-pending",
      proofs: Object.fromEntries(
        document.proof_dimensions.map((dimension) => [
          dimension,
          { status: "not_applicable", references: [], reason },
        ]),
      ),
    });
    const rowID = "rule.plan.identity.token.one-time-secret";
    document.row_applicability[rowID] = "adversarial-row-na";
    document.row_proofs[rowID] = "adversarial-row-na-pending";
  });
  assertRejected(result, "reviewed-applicability-drift");
});

test("a multi-capability row must link a reference set for every affected capability", () => {
  const result = validateMutation((document) => {
    const row = document.rows.find(
      (item) => item.source_id === "ipam.prefix-and-address-semantics",
    );
    row.related_reference_sets = ["ipam.Prefix"];
  });
  assertRejected(result, "missing-affected-reference-set");
});

test("affected capabilities must include the primary row capability", () => {
  const result = validateMutation((document) => {
    const row = document.rows.find(
      (item) => item.source_id === "ipam.prefix-and-address-semantics",
    );
    row.affected_capabilities = ["ipam.IPAddress"];
  });
  assertRejected(result, "invalid-affected-capability");
});

test("a duplicate row ID is rejected", () => {
  const result = validateMutation((document) => {
    document.rows.push(structuredClone(document.rows[0]));
  });
  assertRejected(result, "duplicate-row");
});

test("a baseline tier above the active resource tier is rejected", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "baseline-t1-pending",
    );
    verification.tier = "T2";
    verification.state = "retained";
    verification.evidence = [evidenceReference()];
    delete verification.pending_reason;
  });
  assertRejected(result, "tier-inflation");
});

test("a multi-capability row tier is capped only by its primary owner", () => {
  const profile = structuredClone(BASE_PROFILE);
  profile.resources.find(
    (resource) => resource.module === "dcim" && resource.name === "Site",
  ).tier = "T2";
  const document = structuredClone(BASE_DOCUMENT);
  const verification = {
    id: "adversarial-site-t2",
    classification: "baseline",
    tier: "T2",
    state: "retained",
    tested_digest: CURRENT_SOURCE_DIGEST,
    evidence: [resultEvidenceReference()],
  };
  document.verification_sets.push(verification);
  const row = document.rows.find(
    (item) => item.source_id === "plan.dcim.delete.site-protected",
  );
  row.verification_set = verification.id;
  document.assessment_sets.push({
    id: "adversarial-site-confirmed",
    status: "confirmed",
  });
  row.assessment_set = "adversarial-site-confirmed";
  const proofSet = structuredClone(
    document.proof_sets.find((item) => item.id === "baseline-row-pending"),
  );
  proofSet.id = "adversarial-site-t2-covered";
  const inventory = document.reference_sets.find(
    (item) => item.id === "dcim.Site",
  ).test_inventory;
  proofSet.proofs.rest_differential = {
    status: "covered",
    references: [structuredClone(inventory.rest_differential.references[0])],
  };
  document.proof_sets.push(proofSet);
  document.row_proofs[row.id] = proofSet.id;

  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential"],
        }),
      ],
    ]),
  });
  assert.ok(
    !result.failures.some(
      (failure) =>
        failure.code === "tier-inflation" &&
        (failure.message.includes("dcim.Rack") ||
          failure.message.includes("dcim.Device")),
    ),
    JSON.stringify(result.failures, null, 2),
  );
  const siteClosure = result.failures.find(
    (failure) =>
      failure.code === "profile-tier-unproven" &&
      failure.message.startsWith("dcim.Site "),
  );
  assert.ok(siteClosure, JSON.stringify(result.failures, null, 2));
  assert.ok(siteClosure.message.includes("operation.dcim.site.list"));
  assert.ok(
    !siteClosure.message.includes(
      "scenario.dcim.relationship-delete-semantics",
    ),
  );
});

test("a foreign retained row cannot promote a secondary affected resource", () => {
  const profile = structuredClone(BASE_PROFILE);
  profile.resources.find(
    (resource) => resource.module === "dcim" && resource.name === "Rack",
  ).tier = "T2";
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    traceabilityDocument: structuredClone(BASE_DOCUMENT),
  });
  const closure = result.failures.find(
    (failure) =>
      failure.code === "profile-tier-unproven" &&
      failure.message.startsWith("dcim.Rack "),
  );
  assert.ok(closure, JSON.stringify(result.failures, null, 2));
  assert.ok(closure.message.includes("operation.dcim.rack.list"));
  assert.ok(!closure.message.includes("rule.plan.dcim.delete.site-protected"));
});

test("a profile resource tier cannot lead unresolved row and support boundaries", () => {
  const profile = structuredClone(BASE_PROFILE);
  profile.resources.find(
    (resource) => resource.module === "dcim" && resource.name === "Site",
  ).tier = "T2";
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    traceabilityDocument: structuredClone(BASE_DOCUMENT),
  });
  assertRejected(result, "profile-tier-unproven");
  assertRejected(result, "tier-support-pending");
});

test("a gRPC-only retained baseline set requires only applicable transport evidence", () => {
  const profile = structuredClone(BASE_PROFILE);
  profile.resources.find(
    (resource) => resource.module === "ipam" && resource.name === "IPAddress",
  ).tier = "T3";
  const document = structuredClone(BASE_DOCUMENT);
  const verification = {
    id: "adversarial-ipaddress-t3",
    classification: "baseline",
    tier: "T3",
    state: "retained",
    tested_digest: CURRENT_SOURCE_DIGEST,
    evidence: [resultEvidenceReference()],
  };
  document.verification_sets.push(verification);
  const row = document.rows.find(
    (item) => item.id === "operation.ipam.ip-address.assign-ipaddress",
  );
  row.verification_set = verification.id;
  document.assessment_sets.push({
    id: "adversarial-ipaddress-t3-confirmed",
    status: "confirmed",
  });
  row.assessment_set = "adversarial-ipaddress-t3-confirmed";
  const proofSet = structuredClone(
    document.proof_sets.find((item) => item.id === "grpc-contract-row-pending"),
  );
  proofSet.id = "adversarial-ipaddress-t3-covered";
  const inventory = document.reference_sets.find(
    (item) => item.id === "ipam.IPAddress",
  ).test_inventory;
  proofSet.proofs.grpc_parity = {
    status: "covered",
    references: [structuredClone(inventory.grpc_parity.references[0])],
  };
  document.proof_sets.push(proofSet);
  document.row_proofs[row.id] = proofSet.id;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["grpc_parity"],
        }),
      ],
    ]),
  });
  assert.ok(
    !result.failures.some(
      (failure) =>
        ["invalid-evidence-attestation", "evidence-boundary-mismatch"].includes(
          failure.code,
        ) &&
        (failure.message.includes("rest_differential") ||
          failure.message.includes("grpc_parity")),
    ),
    JSON.stringify(result.failures, null, 2),
  );
});

test("an arbitrary baseline row cannot use gRPC-only N/A to bypass T2", () => {
  const profile = structuredClone(BASE_PROFILE);
  profile.resources.find(
    (resource) => resource.module === "ipam" && resource.name === "IPAddress",
  ).tier = "T2";
  const document = structuredClone(BASE_DOCUMENT);
  document.verification_sets.push({
    id: "adversarial-ipaddress-t2",
    classification: "baseline",
    tier: "T2",
    state: "retained",
    evidence: [evidenceReference()],
  });
  const row = document.rows.find(
    (item) => item.source_id === "plan.ipam.assignment.atomic-update",
  );
  row.verification_set = "adversarial-ipaddress-t2";
  document.row_proofs[row.id] = "grpc-contract-row-pending";
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    traceabilityDocument: document,
  });
  assert.ok(
    result.failures.some(
      (failure) =>
        failure.code === "tier-boundary-missing" &&
        failure.message.includes(row.id) &&
        failure.message.includes("rest_differential"),
    ),
    JSON.stringify(result.failures, null, 2),
  );
});

test("a resource cannot be promoted through a vacuous reviewed boundary", () => {
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.resources.find(
        (resource) => resource.module === "dcim" && resource.name === "Site",
      ).tier = "T2";
    },
    documentMutator(document) {
      const verification = document.verification_sets.find(
        (item) => item.id === "baseline-t1-pending",
      );
      verification.tier = "T2";
      verification.state = "retained";
      verification.evidence = [evidenceReference()];
      delete verification.pending_reason;
      const reason = "Adversarially removed REST applicability.";
      const applicability = document.applicability_sets.find(
        (item) => item.id === "baseline-row",
      ).dimensions.rest_differential;
      applicability.status = "not_applicable";
      applicability.reason = reason;
      const proof = document.proof_sets.find(
        (item) => item.id === "baseline-row-pending",
      ).proofs.rest_differential;
      proof.status = "not_applicable";
      proof.references = [];
      proof.reason = reason;
    },
  });
  assert.ok(
    result.failures.some(
      (failure) =>
        failure.code === "profile-tier-unproven" &&
        failure.message.includes("vacuous boundaries: rest_differential"),
    ),
    JSON.stringify(result.failures, null, 2),
  );
});

test("project support joins only at its reviewed baseline tier boundary", () => {
  const t2 = validateProfileMutation({
    profileMutator(profile) {
      profile.resources.find(
        (resource) => resource.module === "dcim" && resource.name === "Site",
      ).tier = "T2";
    },
    documentMutator(document) {
      const verification = document.verification_sets.find(
        (item) => item.id === "baseline-t1-pending",
      );
      verification.tier = "T2";
      verification.state = "retained";
      verification.evidence = [evidenceReference()];
      delete verification.pending_reason;
    },
  });
  const t2Support = t2.failures.find(
    (failure) => failure.code === "tier-support-pending",
  );
  assert.ok(t2Support, JSON.stringify(t2.failures, null, 2));
  assert.ok(t2Support.message.includes("scenario.contract.surface-manifest"));
  assert.ok(
    !t2Support.message.includes("rule.plan.common.list.grpc-typed-fields"),
  );
  assert.ok(
    !t2Support.message.includes("rule.plan.identity.token.one-time-secret"),
  );

  const t3 = validateProfileMutation({
    profileMutator(profile) {
      profile.resources.find(
        (resource) => resource.module === "dcim" && resource.name === "Site",
      ).tier = "T3";
    },
    documentMutator(document) {
      const verification = document.verification_sets.find(
        (item) => item.id === "baseline-t1-pending",
      );
      verification.tier = "T3";
      verification.state = "retained";
      verification.evidence = [evidenceReference()];
      delete verification.pending_reason;
    },
  });
  const t3Support = t3.failures.find(
    (failure) => failure.code === "tier-support-pending",
  );
  assert.ok(t3Support, JSON.stringify(t3.failures, null, 2));
  assert.ok(
    t3Support.message.includes("rule.plan.common.list.grpc-typed-fields"),
  );
});

test("extra IPAddress RPC operations are REST N/A at T2 and require gRPC at T3", () => {
  const validateTier = (tier) =>
    validateProfileMutation({
      profileMutator(profile) {
        profile.resources.find(
          (resource) =>
            resource.module === "ipam" && resource.name === "IPAddress",
        ).tier = tier;
      },
      documentMutator(document) {
        const verification = document.verification_sets.find(
          (item) => item.id === "baseline-t1-pending",
        );
        verification.tier = tier;
        verification.state = "retained";
        verification.evidence = [evidenceReference()];
        delete verification.pending_reason;
      },
    });

  const t2 = validateTier("T2");
  for (const rowID of [
    "operation.ipam.ip-address.assign-ipaddress",
    "operation.ipam.ip-address.unassign-ipaddress",
  ]) {
    assert.ok(
      !t2.failures.some(
        (failure) =>
          failure.code === "tier-boundary-missing" &&
          failure.message.includes(rowID) &&
          failure.message.includes("rest_differential"),
      ),
      JSON.stringify(t2.failures, null, 2),
    );
  }

  const t3 = validateTier("T3");
  for (const rowID of [
    "operation.ipam.ip-address.assign-ipaddress",
    "operation.ipam.ip-address.unassign-ipaddress",
  ]) {
    assert.ok(
      t3.failures.some(
        (failure) =>
          failure.code === "tier-boundary-missing" &&
          failure.message.includes(rowID) &&
          failure.message.includes("grpc_parity"),
      ),
      JSON.stringify(t3.failures, null, 2),
    );
  }
});

test("identity verification states reject lookalike completion values", () => {
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.identity_extension.verification.contract = "completed";
    },
    identityMutator(identity) {
      identity.verification.contract = "completed";
    },
  });
  assertRejected(result, "invalid-extension-verification");
});

test("the identity extension rejects unsupported profile properties", () => {
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.identity_extension.unreviewed_escape = true;
    },
  });
  assertRejected(result, "invalid-extension-verification");
});

test("identity metadata verification must exactly match the active profile", () => {
  const result = validateProfileMutation({
    identityMutator(identity) {
      identity.verification.contract = "complete";
    },
  });
  assertRejected(result, "extension-verification-drift");
});

test("an identity extension cannot receive a T tier", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "identity-extension-pending",
    );
    verification.tier = "T1";
  });
  assertRejected(result, "extension-tier-claim");
});

test("retained complete identity extension requires every applicable extension boundary", () => {
  const complete = {
    contract: "complete",
    parity: "complete",
    security: "complete",
  };
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.identity_extension.verification = structuredClone(complete);
    },
    identityMutator(identity) {
      identity.verification = structuredClone(complete);
    },
    documentMutator(document) {
      const verification = document.verification_sets.find(
        (item) => item.id === "identity-extension-pending",
      );
      verification.state = "retained";
      verification.evidence = [evidenceReference()];
      verification.extension_verification = structuredClone(complete);
      delete verification.pending_reason;
    },
  });
  assertRejected(result, "extension-boundary-missing");
  assertRejected(result, "extension-verification-pending");
  assertRejected(result, "extension-support-pending");
});

test("completed identity axes sweep every row with an applicable boundary", () => {
  for (const [axis, expectedDimension] of [
    ["contract", "rest_extension_contract"],
    ["parity", "grpc_parity"],
    ["security", "browser"],
  ]) {
    const result = validateProfileMutation({
      profileMutator(profile) {
        profile.identity_extension.verification[axis] = "complete";
      },
      identityMutator(identity) {
        identity.verification[axis] = "complete";
      },
      documentMutator(document) {
        const verification = document.verification_sets.find(
          (item) => item.id === "identity-extension-pending",
        );
        verification.extension_verification[axis] = "complete";
      },
    });
    assert.ok(
      result.failures.some(
        (failure) =>
          failure.code === "extension-boundary-missing" &&
          failure.message.includes(expectedDimension),
      ),
      `${axis}: ${JSON.stringify(result.failures, null, 2)}`,
    );
  }
});

test("complete identity verification cannot redefine every proof as N/A", () => {
  const complete = {
    contract: "complete",
    parity: "complete",
    security: "complete",
  };
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.identity_extension.verification = structuredClone(complete);
    },
    identityMutator(identity) {
      identity.verification = structuredClone(complete);
    },
    documentMutator(document) {
      const reason = "Adversarial all-boundaries-not-applicable declaration.";
      document.applicability_sets.push({
        id: "adversarial-all-na",
        dimensions: Object.fromEntries(
          document.proof_dimensions.map((dimension) => [
            dimension,
            { status: "not_applicable", reason },
          ]),
        ),
      });
      document.proof_sets.push({
        id: "adversarial-all-na-pending",
        proofs: Object.fromEntries(
          document.proof_dimensions.map((dimension) => [
            dimension,
            { status: "not_applicable", references: [], reason },
          ]),
        ),
      });
      const extensionVerification = document.verification_sets.find(
        (item) => item.id === "identity-extension-pending",
      );
      extensionVerification.state = "retained";
      extensionVerification.evidence = [evidenceReference()];
      extensionVerification.extension_verification = structuredClone(complete);
      delete extensionVerification.pending_reason;
      for (const assessment of document.assessment_sets) {
        assessment.status = "confirmed";
        delete assessment.reason;
        delete assessment.resolution_goal;
        delete assessment.conflict_references;
      }
      for (const row of document.rows) {
        const verification = document.verification_sets.find(
          (item) => item.id === row.verification_set,
        );
        if (verification?.classification !== "extension") continue;
        document.row_applicability[row.id] = "adversarial-all-na";
        document.row_proofs[row.id] = "adversarial-all-na-pending";
      }
    },
  });
  assertRejected(result, "extension-applicability-missing");
});

test("one retained extension verification set cannot hide a pending extension row", () => {
  const complete = {
    contract: "complete",
    parity: "complete",
    security: "complete",
  };
  const targetRowID = "rule.plan.identity.token.one-time-secret";
  const result = validateProfileMutation({
    profileMutator(profile) {
      profile.identity_extension.verification = structuredClone(complete);
    },
    identityMutator(identity) {
      identity.verification = structuredClone(complete);
    },
    documentMutator(document) {
      const retained = document.verification_sets.find(
        (item) => item.id === "identity-extension-pending",
      );
      retained.state = "retained";
      retained.evidence = [evidenceReference()];
      retained.extension_verification = structuredClone(complete);
      delete retained.pending_reason;
      document.verification_sets.push({
        id: "adversarial-extension-pending",
        classification: "extension",
        tier: "not_applicable",
        state: "pending",
        evidence: [],
        pending_reason: "One row remains deliberately pending.",
        extension_verification: structuredClone(complete),
      });
      document.rows.find((row) => row.id === targetRowID).verification_set =
        "adversarial-extension-pending";
      for (const assessment of document.assessment_sets) {
        assessment.status = "confirmed";
        delete assessment.reason;
        delete assessment.resolution_goal;
        delete assessment.conflict_references;
      }
    },
  });
  assert.ok(
    result.failures.some(
      (failure) =>
        failure.code === "extension-verification-pending" &&
        failure.message.includes(targetRowID),
    ),
    JSON.stringify(result.failures, null, 2),
  );
});

test("contradicted behavior cannot inherit retained verification", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "baseline-t1-pending",
    );
    verification.state = "retained";
    verification.evidence = [evidenceReference()];
    delete verification.pending_reason;
  });
  assertRejected(result, "contradicted-tier-claim");
});

test("unresolved behavior cannot inherit covered proof", () => {
  const result = validateMutation((document) => {
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.domain;
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    proof.status = "covered";
    proof.references = [
      structuredClone(site.test_inventory.domain.references[0]),
    ];
    delete proof.reason;
  });
  assertRejected(result, "unresolved-covered-proof");
});

test("covered proof must belong to an affected capability inventory", () => {
  const result = validateMutation((document) => {
    const assessment = document.assessment_sets.find(
      (item) => item.id === "unresolved-v2",
    );
    assessment.status = "confirmed";
    delete assessment.reason;
    delete assessment.resolution_goal;
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.domain;
    proof.status = "covered";
    proof.references = [evidenceReference()];
    delete proof.reason;
  });
  assertRejected(result, "unowned-proof-reference");
});

test("a capability inventory and its consuming proof cannot be changed together", () => {
  const result = validateMutation((document) => {
    const arbitraryReference = evidenceReference();
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    site.test_inventory.domain.references.push(arbitraryReference);

    const proofSet = structuredClone(
      document.proof_sets.find((item) => item.id === "baseline-row-pending"),
    );
    proofSet.id = "adversarial-site-list-proof";
    proofSet.proofs.domain.status = "covered";
    proofSet.proofs.domain.references = [arbitraryReference];
    delete proofSet.proofs.domain.reason;
    document.proof_sets.push(proofSet);

    const row = document.rows.find(
      (item) => item.id === "operation.dcim.site.list",
    );
    document.row_proofs[row.id] = proofSet.id;
    document.assessment_sets.push({
      id: "adversarial-confirmed",
      status: "confirmed",
    });
    row.assessment_set = "adversarial-confirmed";
  });
  assertRejected(result, "reviewed-reference-authority-drift");
});

test("a reviewed test inventory cannot be reassigned to another capability", () => {
  const result = validateMutation((document) => {
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    site.capabilities.push("ipam.IPAddress");
  });
  assertRejected(result, "reviewed-reference-authority-drift");
});

test("a reviewed plan rule cannot be rebound to another capability", () => {
  const result = validateMutation((document) => {
    const row = document.rows.find(
      (item) => item.id === "rule.plan.dcim.site.name-global",
    );
    row.capability = "dcim.Manufacturer";
    row.operation = "get";
    row.reference_set = "dcim.Manufacturer";
    delete row.affected_capabilities;
    delete row.related_reference_sets;
  });
  assertRejected(result, "reviewed-row-semantics-drift");
});

test("pinned source and upstream tests cannot be substituted across resources", () => {
  const result = validateMutation((document) => {
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    const manufacturer = document.reference_sets.find(
      (item) => item.id === "dcim.Manufacturer",
    );
    site.pinned_source = structuredClone(manufacturer.pinned_source);
    site.upstream_tests = structuredClone(manufacturer.upstream_tests);
  });
  assertRejected(result, "reviewed-reference-authority-drift");
});

test("retained project and extension evidence must cite a result artifact", () => {
  for (const verificationSetID of [
    "project-pending",
    "identity-extension-pending",
  ]) {
    const result = validateMutation((document) => {
      const verification = document.verification_sets.find(
        (item) => item.id === verificationSetID,
      );
      verification.state = "retained";
      verification.evidence = [
        {
          path: "docs/IMPLEMENTATION_PLAN.md",
          anchor: "# NetBox Go rewrite implementation plan",
        },
      ];
      delete verification.pending_reason;
    });
    assertRejected(result, "invalid-retained-evidence");
  }
});

test("the evidence ledger index is not itself a retained result artifact", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "project-pending",
    );
    verification.state = "retained";
    verification.evidence = [evidenceReference()];
    delete verification.pending_reason;
  });
  assertRejected(result, "invalid-retained-evidence");
});

test("a normalized alias cannot turn the evidence ledger into a result artifact", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = document.verification_sets.find(
    (item) => item.id === "project-pending",
  );
  verification.state = "retained";
  verification.tested_digest = CURRENT_SOURCE_DIGEST;
  verification.evidence = [
    {
      path: "docs/evidence/./README.md",
      anchor: "# Evidence ledger",
    },
  ];
  delete verification.pending_reason;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        "docs/evidence/./README.md",
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-retained-evidence");
  assertRejected(result, "stale-reference-path");
});

test("an unrelated historical result cannot attest a retained verification", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "identity-extension-pending",
    );
    verification.state = "retained";
    verification.tested_digest = CURRENT_SOURCE_DIGEST;
    verification.evidence = [resultEvidenceReference()];
    delete verification.pending_reason;
  });
  assertRejected(result, "missing-evidence-attestation");
});

test("retained verification requires an immutable source digest", () => {
  const result = validateMutation((document) => {
    const verification = document.verification_sets.find(
      (item) => item.id === "project-pending",
    );
    verification.state = "retained";
    verification.evidence = [resultEvidenceReference()];
    delete verification.pending_reason;
  });
  assertRejected(result, "invalid-evidence-source-digest");
});

test("retained verification requires an explicit tested revision", () => {
  const document = structuredClone(BASE_DOCUMENT);
  configureCoveredProjectVerification(document);
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
  });
  assertRejected(result, "invalid-tested-revision");
});

test("a structurally matching attestation rejects an unreachable tested revision", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.tested_revision = "0".repeat(40);

  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          compatibility_baseline: {
            git_sha: BASE_DOCUMENT.compatibility_baseline.git_sha,
            id: BASE_DOCUMENT.compatibility_baseline.id,
          },
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-tested-revision");
  assert.deepEqual(
    result.failures.filter((failure) =>
      [
        "invalid-retained-evidence",
        "invalid-evidence-source-digest",
        "missing-evidence-attestation",
        "invalid-evidence-attestation",
      ].includes(failure.code),
    ),
    [],
  );
});

test("a same-digest claim rejects a dirty worktree that differs from its tested commit", () => {
  const { fixtureRoot, revision } = materializeCommittedValidationFixture();
  try {
    fs.appendFileSync(
      path.join(fixtureRoot, "AGENTS.md"),
      "\nuncommitted evidence-source mutation\n",
      "utf8",
    );
    const manifest = calculateSourceManifest(fixtureRoot);
    const document = structuredClone(BASE_DOCUMENT);
    const verification = configureCoveredProjectVerification(document);
    verification.tested_digest = manifest.digest;
    verification.tested_revision = revision;
    const resultPayload = `${RESULT_EVIDENCE_ANCHOR}\n`;
    verification.evidence[0].payload_sha256 = sha256(
      Buffer.from(resultPayload, "utf8"),
    );
    const result = validateTraceability({
      root: fixtureRoot,
      profilePath: path.join(
        fixtureRoot,
        "contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml",
      ),
      traceabilityDocument: document,
      evidenceDocuments: new Map([
        [
          RESULT_EVIDENCE_PATH,
          `${resultPayload}${evidenceAttestation(document, verification, {
            attestation_digest: manifest.digest,
            proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          })}`,
        ],
      ]),
    });
    assertRejected(result, "invalid-tested-revision");
    assert.deepEqual(
      result.failures.filter((failure) =>
        [
          "invalid-retained-evidence",
          "invalid-evidence-source-digest",
          "missing-evidence-attestation",
          "invalid-evidence-attestation",
        ].includes(failure.code),
      ),
      [],
    );
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});

test("a project attestation must name and cover its consumer boundaries", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [RESULT_EVIDENCE_PATH, evidenceAttestation(document, verification)],
    ]),
  });
  assertRejected(result, "invalid-evidence-attestation");
  assertRejected(result, "evidence-boundary-mismatch");
});

test("an attested boundary cannot outrun its row proof", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const proofSet = document.proof_sets.find(
    (item) => item.id === "project-contract-covered-test",
  );
  proofSet.proofs.grpc_parity = {
    status: "pending",
    references: [],
    reason: "Adversarially left pending behind a retained artifact.",
  };
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "evidence-boundary-mismatch");
});

test("the attestation digest must equal the current owned-source digest", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.tested_digest = DIFFERENT_SOURCE_DIGEST;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          tested_digest: DIFFERENT_SOURCE_DIGEST,
          attestation_digest:
            "source-v2:sha256:0000000000000000000000000000000000000000000000000000000000000000",
          claim_manifest: [
            {
              path: MATRIX_RELATIVE_PATH,
              tested_mode: "100644",
              tested_size: 1,
              tested_sha256: DIFFERENT_FILE_DIGEST,
              attestation_mode: "100644",
              attestation_size: 1,
              attestation_sha256: DIFFERENT_FILE_DIGEST,
            },
          ],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("an invented two-digest manifest cannot replace an exact committed tested revision", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const fixture = twoDigestFixture();
  verification.tested_digest = fixture.digest;
  verification.tested_source_manifest = fixture.manifestReference;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          tested_digest: fixture.digest,
          tested_source_manifest: fixture.manifestReference,
          claim_manifest: [fixture.claim],
        }),
      ],
      [TESTED_MANIFEST_PATH, fixture.manifestSource],
    ]),
  });
  assertRejected(result, "invalid-tested-revision");
});

test("a fabricated historical digest without a committed full manifest is rejected", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.tested_digest = DIFFERENT_SOURCE_DIGEST;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          tested_digest: DIFFERENT_SOURCE_DIGEST,
          claim_manifest: [
            {
              path: MATRIX_RELATIVE_PATH,
              tested_mode: "100644",
              tested_size: 1,
              tested_sha256: DIFFERENT_FILE_DIGEST,
              attestation_mode: "100644",
              attestation_size: 1,
              attestation_sha256:
                "sha256:0000000000000000000000000000000000000000000000000000000000000000",
            },
          ],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-tested-source-manifest");
});

test("a tested-source manifest must match its owned matrix commitment", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const fixture = twoDigestFixture();
  verification.tested_digest = fixture.digest;
  verification.tested_source_manifest = fixture.manifestReference;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          claim_manifest: [fixture.claim],
        }),
      ],
      [TESTED_MANIFEST_PATH, `${fixture.manifestSource} `],
    ]),
  });
  assertRejected(result, "invalid-tested-source-manifest");
});

test("an incomplete tested-source manifest cannot replace the committed Git tree", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const fixture = twoDigestFixture();
  const manifest = JSON.parse(fixture.manifestSource);
  manifest.entries.pop();
  manifest.files = manifest.entries.length;
  manifest.digest = calculateSourceDigestFromEntries(manifest.entries);
  const source = `${JSON.stringify(manifest, null, 2)}\n`;
  verification.tested_digest = manifest.digest;
  verification.tested_source_manifest = {
    path: TESTED_MANIFEST_PATH,
    sha256: sha256(source),
  };
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          claim_manifest: [fixture.claim],
        }),
      ],
      [TESTED_MANIFEST_PATH, source],
    ]),
  });
  assertRejected(result, "invalid-tested-revision");
});

test("an invented source-code diff cannot replace the committed Git tree", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const fixture = twoDigestFixture("scripts/validate_traceability.mjs");
  verification.tested_digest = fixture.digest;
  verification.tested_source_manifest = fixture.manifestReference;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          claim_manifest: [fixture.claim],
        }),
      ],
      [TESTED_MANIFEST_PATH, fixture.manifestSource],
    ]),
  });
  assertRejected(result, "invalid-tested-revision");
});

test("an invented symlink retarget cannot replace the committed Git tree", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const entries = structuredClone(CURRENT_SOURCE_MANIFEST.entries);
  const symlink = entries.find((entry) => entry.kind === "symlink");
  assert.ok(symlink, "the repository source manifest must contain a symlink");
  symlink.target = `${symlink.target}-changed`;
  const digest = calculateSourceDigestFromEntries(entries);
  const manifest = {
    schema_version: 2,
    digest,
    files: entries.length,
    entries,
  };
  const source = `${JSON.stringify(manifest, null, 2)}\n`;
  const manifestReference = {
    path: TESTED_MANIFEST_PATH,
    sha256: sha256(source),
  };
  verification.tested_digest = digest;
  verification.tested_source_manifest = manifestReference;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
          claim_manifest: [
            {
              path: symlink.path,
              tested_mode: "100644",
              tested_size: 1,
              tested_sha256: DIFFERENT_FILE_DIGEST,
              attestation_mode: "100644",
              attestation_size: 1,
              attestation_sha256: DIFFERENT_FILE_DIGEST,
            },
          ],
        }),
      ],
      [TESTED_MANIFEST_PATH, source],
    ]),
  });
  assertRejected(result, "invalid-tested-revision");
});

test("a result attestation cannot substitute a different source digest", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);

  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          tested_digest: DIFFERENT_SOURCE_DIGEST,
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("a result artifact cannot carry duplicate attestations for one claim", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const marker = evidenceAttestation(document, verification, {
    proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
  });

  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [RESULT_EVIDENCE_PATH, `${marker}\n${marker}`],
    ]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("a result artifact cannot carry an unowned attestation marker", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const owned = evidenceAttestation(document, verification, {
    proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
  });
  const extra = owned.replace(verification.id, "unowned-retained-verification");
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([[RESULT_EVIDENCE_PATH, `${owned}\n${extra}`]]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("an attestation rejects duplicate JSON member names", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const marker = evidenceAttestation(document, verification, {
    proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
  }).replace(
    `"tested_digest":"${CURRENT_SOURCE_DIGEST}"`,
    `"tested_digest":"${DIFFERENT_SOURCE_DIGEST}","tested_digest":"${CURRENT_SOURCE_DIGEST}"`,
  );
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([[RESULT_EVIDENCE_PATH, marker]]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("an attestation rejects JSON whitespace outside RFC 8259", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const marker = evidenceAttestation(document, verification, {
    proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
  }).replace(": {", ": {\f");
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([[RESULT_EVIDENCE_PATH, marker]]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("a retained evidence path cannot use a symlinked lexical alias", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const alias =
    "netbox-backend/docs/evidence/2026-08-01-core-workflow-v1-v0.md";
  verification.evidence[0].path = alias;
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        alias,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-retained-evidence");
});

test("retained evidence commits the non-attestation result payload", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const marker = evidenceAttestation(document, verification, {
    proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
  });
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [RESULT_EVIDENCE_PATH, `tampered result\n${marker}`],
    ]),
  });
  assertRejected(result, "invalid-evidence-payload");
});

test("retained evidence rejects invalid UTF-8 before parsing markers", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  const invalidPayload = Buffer.from([0x80, 0x0a]);
  verification.evidence[0].payload_sha256 = sha256(invalidPayload);
  const marker = Buffer.from(
    evidenceAttestation(document, verification, {
      proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
    }),
    "utf8",
  );
  const artifact = Buffer.concat([invalidPayload, marker]);
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([[RESULT_EVIDENCE_PATH, artifact]]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("retained payload commitments hash exact raw bytes", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.evidence[0].payload_sha256 = sha256(Buffer.from([0x80, 0x0a]));
  const marker = Buffer.from(
    evidenceAttestation(document, verification, {
      proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
    }),
    "utf8",
  );
  const artifact = Buffer.concat([Buffer.from([0x81, 0x0a]), marker]);
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([[RESULT_EVIDENCE_PATH, artifact]]),
  });
  assertRejected(result, "invalid-evidence-payload");
});

test("verification-set classification fields are closed", () => {
  const project = validateMutation((document) => {
    document.verification_sets.find(
      (item) => item.id === "project-pending",
    ).extension_verification = {
      contract: "complete",
      parity: "complete",
      security: "complete",
    };
  });
  assertRejected(project, "invalid-verification");

  const baseline = validateMutation((document) => {
    document.verification_sets.find(
      (item) => item.id === "baseline-t1-pending",
    ).state = "not_applicable";
  });
  assertRejected(baseline, "invalid-verification");
});

test("a retained verification cannot keep a stale pending reason", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = configureCoveredProjectVerification(document);
  verification.pending_reason = "This stale text must not survive retention.";
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: evidenceDocumentsFor(document, verification),
  });
  assertRejected(result, "invalid-verification");
});

test("an unused retained verification set cannot claim T4", () => {
  const document = structuredClone(BASE_DOCUMENT);
  const verification = {
    id: "unused-baseline-t4-retained",
    classification: "baseline",
    tier: "T4",
    state: "retained",
    tested_digest: CURRENT_SOURCE_DIGEST,
    evidence: [resultEvidenceReference()],
  };
  document.verification_sets.push(verification);
  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_differential", "grpc_parity", "browser"],
        }),
      ],
    ]),
  });
  assertRejected(result, "unused-verification-set");
});

test("an unused assessment set is rejected", () => {
  const result = validateMutation((document) => {
    document.assessment_sets.push({
      id: "unused-confirmed-assessment",
      status: "confirmed",
    });
  });
  assertRejected(result, "unused-assessment-set");
});

test("a completed extension attestation names every claimed proof boundary", () => {
  const complete = {
    contract: "complete",
    parity: "complete",
    security: "complete",
  };
  const profile = structuredClone(BASE_PROFILE);
  const identity = structuredClone(BASE_IDENTITY);
  const document = structuredClone(BASE_DOCUMENT);
  profile.identity_extension.verification = structuredClone(complete);
  identity.verification = structuredClone(complete);
  const verification = document.verification_sets.find(
    (item) => item.id === "identity-extension-pending",
  );
  verification.state = "retained";
  verification.tested_digest = CURRENT_SOURCE_DIGEST;
  verification.evidence = [resultEvidenceReference()];
  verification.extension_verification = structuredClone(complete);
  delete verification.pending_reason;

  const result = validateTraceability({
    root: ROOT,
    profilePath: PROFILE_PATH,
    profileDocument: profile,
    identityMetadataDocument: identity,
    traceabilityDocument: document,
    evidenceDocuments: new Map([
      [
        RESULT_EVIDENCE_PATH,
        evidenceAttestation(document, verification, {
          proof_dimensions: ["rest_extension_contract"],
        }),
      ],
    ]),
  });
  assertRejected(result, "invalid-evidence-attestation");
});

test("every row must select an explicit proof set", () => {
  const result = validateMutation((document) => {
    delete document.row_proofs[document.rows[0].id];
  });
  assertRejected(result, "missing-row-proof");
});

test("partial or covered proof sets must be row-scoped", () => {
  const result = validateMutation((document) => {
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.domain;
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    proof.status = "partial";
    proof.references = [
      structuredClone(site.test_inventory.domain.references[0]),
    ];
    proof.reason = "A deliberately shared live proof must fail closed.";
  });
  assertRejected(result, "shared-live-proof-set");
});

test("pending row proof cannot carry a test reference", () => {
  const result = validateMutation((document) => {
    const proof = document.proof_sets.find(
      (item) => item.id === "baseline-row-pending",
    ).proofs.domain;
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    proof.references = [
      structuredClone(site.test_inventory.domain.references[0]),
    ];
  });
  assertRejected(result, "invalid-proof");
});

test("schema-forbidden matrix properties are rejected", () => {
  const result = validateMutation((document) => {
    document.unvalidated_escape = true;
  });
  assertRejected(result, "schema-validation");
});

test("scenario covered operations must belong to the operation catalogue", () => {
  const result = validateMutation((document) => {
    document.rows
      .find((row) => row.kind === "scenario")
      .covered_operations.push("undeclared_operation");
  });
  assertRejected(result, "unknown-covered-operation");
});

test("baseline resources cannot omit pinned source or upstream test links", () => {
  const result = validateMutation((document) => {
    const site = document.reference_sets.find(
      (item) => item.id === "dcim.Site",
    );
    site.pinned_source = {
      status: "not_applicable",
      references: [],
      reason: "adversarial omission",
    };
  });
  assertRejected(result, "missing-upstream-link");
});

test("upstream-backed support sets cannot cite source or tests outside pinned NetBox", () => {
  for (const referenceSetID of [
    "identity",
    "identity-baseline-support",
    "identity-interface-support",
    "baseline-common-support",
    "deferred-baseline-support",
  ]) {
    const result = validateMutation((document) => {
      const referenceSet = document.reference_sets.find(
        (item) => item.id === referenceSetID,
      );
      referenceSet.pinned_source.references = [evidenceReference()];
      referenceSet.upstream_tests.references = [evidenceReference()];
    });
    assertRejected(result, "invalid-upstream-reference");
  }
});

test("special reference sets retain exact classification and capability ownership", () => {
  const result = validateMutation((document) => {
    const support = document.reference_sets.find(
      (item) => item.id === "baseline-common-support",
    );
    support.capabilities.push("contract");
  });
  assertRejected(result, "invalid-reference-capabilities");
});
