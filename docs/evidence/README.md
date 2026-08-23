# Evidence ledger

This directory records durable verification results for compatibility tiers and
release gates. Source files and test harnesses are linked from
[Testing](../TESTING.md); they are not pass evidence by themselves.

## Current ledger

The [2026-08-17 source-v2 V0](2026-08-17-core-workflow-v1-source-v2-v0.md),
[trusted-origin CORS evidence](2026-08-03-core-workflow-v1-cors-v0.md), and
[bounded token-authentication foundation](2026-08-17-core-workflow-v1-token-auth-i1-v0.md)
passed their exact committed local boundaries and linked independent GitHub
runs. The
[bounded token-transport and unary-safety result](2026-08-17-core-workflow-v1-token-transport-i2-v0.md)
also passed its exact committed local boundary, independent candidate CI, and
evidence-receipt CI. The
[bounded browser-session and CSRF result](2026-08-17-core-workflow-v1-browser-session-i3-v0.md)
passed its exact committed local, real-PostgreSQL, independent candidate CI,
and evidence-receipt CI boundaries. The exact
[password-change and session-rotation result](2026-08-19-core-workflow-v1-password-session-i4-v0.md)
passed its accepted-I3 red-first, focused, race, real-PostgreSQL, complete L4,
pinned repository, feature-candidate CI, and main exact-SHA CI boundaries. Its
claim-only revision
`d2a19d8e61838785be939c77e0e1a35ac95a075e` also passed its pinned local gate
and exact-SHA repository CI, making bounded I4 `evidence`. Digest-excluded
receipt revision `105bf69c90163a784e3e3359781ccb4e207e2d6c` then passed exact-SHA
repository CI; project-owner review remains.

The project owner reviewed and retained the entry V0, CORS, bounded token I1,
bounded token I2, and bounded browser-session I3 results. The
[2026-08-03 post-cleanup V0](2026-08-03-post-cleanup-v0.md) used the superseded
mode-blind v1 source digest and is now historical. The 2026-08-01
repository V0 and recovery-scoped PostgreSQL replay also remain historical in
the
[Core Workflow recovery artifact](2026-08-01-core-workflow-v1-v0.md). The
Profile remains T1 and pre-publication; V1-V6 remain open. `CW1-G00` passed for
the exact I4 tested source; `CW1-V1-01`, `CW1-V1-02-I1`, `CW1-V1-02-I2`, and
`CW1-V1-02-I3` are done. I4 is `evidence`; project-owner review remains. The
parent credential/session matrix remains open.
Feature work must re-establish V0 after its next owned-source change.

The source-included governance documents intentionally preserve the claim-time
wording in attestation revision
`d2a19d8e61838785be939c77e0e1a35ac95a075e`, which maps to the tested I4
revision and digest. This digest-excluded ledger is the later receipt-CI
addendum and does not advance I4 beyond `evidence`.

| Evidence                 | Required command or boundary                                  | Current state                                           |
| ------------------------ | ------------------------------------------------------------- | ------------------------------------------------------- |
| Repository quality       | Revised `make check` including non-mutating backend coverage  | Exact I4 claim and receipt CI passed; review remains    |
| Strict REST differential | `make compatibility-test`                                     | Harness present; current result pending                 |
| gRPC semantic parity     | `go test ./test/parity -count=1` plus corresponding T2 report | Tests present; current result pending                   |
| Real PostgreSQL          | DSN-enabled bootstrap/schema/concurrency/identity suites      | I1, I3, and I4 bounded slices passed; V1/V3 incomplete  |
| Standalone deployment    | `make deployment-smoke`                                       | CORS-scoped exact-commit result retained; V3 still open |
| Browser workflow         | `make browser-e2e`                                            | Harness present; current result pending                 |

### Entry V0/CORS claim-only reconciliation

The accepted source-v2 V0 and CORS candidate at
`07aaca9264e5af9830e39fc95dd3289c0b5c9476` was retained by claim-only revision
`5d6f9cf9f5a5dc5e5e09de58f4f0e76e0f7a6381`. Candidate run
[31994516467](https://github.com/jihadismail8/netbox-go/actions/runs/31994516467)
and claim run
[31997963053](https://github.com/jihadismail8/netbox-go/actions/runs/31997963053)
both passed at their exact SHAs. Reconstruction from both reachable Git trees
on 2026-08-23 produced the same 2,993-entry source manifest, SHA-256
`8f82bbc5b6ba5fe0c6acefe17cd552fe5bd9523414a5f224fec0b7040e45ea49`,
and source digest
`source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82`.

The canonical JSON below records the complete historical claim-only path
mapping. Its exact UTF-8 bytes, including the final newline, have SHA-256
`cda816c8a9fdbf1f0a859395bc6f6399cbad3123121b602080695cd157e57e69`.

```json
{
  "schema_version": 1,
  "tested_revision": "07aaca9264e5af9830e39fc95dd3289c0b5c9476",
  "attestation_revision": "5d6f9cf9f5a5dc5e5e09de58f4f0e76e0f7a6381",
  "tested_digest": "source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82",
  "attestation_digest": "source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82",
  "tested_files": 2993,
  "attestation_files": 2993,
  "tested_manifest_sha256": "sha256:8f82bbc5b6ba5fe0c6acefe17cd552fe5bd9523414a5f224fec0b7040e45ea49",
  "attestation_manifest_sha256": "sha256:8f82bbc5b6ba5fe0c6acefe17cd552fe5bd9523414a5f224fec0b7040e45ea49",
  "paths": [
    {
      "path": "docs/evidence/2026-08-03-core-workflow-v1-cors-v0.md",
      "old": {
        "present": false
      },
      "new": {
        "present": true,
        "mode": "100644",
        "size": 6769,
        "sha256": "sha256:197a4178374252a8527c61f9e9448f22f4f859b9447270137e3560947c82b3f9"
      }
    },
    {
      "path": "docs/evidence/2026-08-17-core-workflow-v1-source-v2-v0.md",
      "old": {
        "present": false
      },
      "new": {
        "present": true,
        "mode": "100644",
        "size": 5119,
        "sha256": "sha256:a13bb032ad66ecca5c88e0043d1d78786c33874f42f417f86aad3b43eaa5b00b"
      }
    },
    {
      "path": "docs/evidence/README.md",
      "old": {
        "present": true,
        "mode": "100644",
        "size": 5706,
        "sha256": "sha256:e5ad226a59319b36174c79a2627269d788585a7dc61216bed1d02563e7fe4360"
      },
      "new": {
        "present": true,
        "mode": "100644",
        "size": 6195,
        "sha256": "sha256:3b1829c3112e1f3b3560fc87f928148b314555613c81e187ffc947c64d172a6d"
      }
    }
  ]
}
```

This reconciliation is digest-excluded historical evidence bookkeeping only.
It makes no new retained claim and promotes no consumer row, extension axis,
profile, or compatibility tier. It does not make the old V0 current for later
source changes. The new bookkeeping becomes retained only after this
reconciliation revision passes its own exact-SHA repository gate.

GitHub CI currently runs V0 only. Real PostgreSQL, deployment, differential
oracle, gRPC promotion, and browser gates remain operator-run until durable,
connected CI jobs retain their credential-free artifacts.

## Required artifact shape

A result may be cited from [Project status](../STATUS.md) only when it records:

- the full reachable tested Git revision and its exact versioned
  `make source-digest` value; retained source-v2 claims are unavailable when
  repository Git metadata is unavailable, and the oracle remains identified
  separately by its pinned Git SHA; all authoritative historical-object reads
  disable local Git replacement refs and grafts;
- command, toolchain, and relevant non-secret configuration;
- start/end time and pass/fail status;
- the pinned oracle SHA/configuration for differential runs;
- per-capability scenario totals and failures where a tier is affected;
- durable-state and side-effect checks required by the scenario; and
- a credential-free artifact location retained by CI or committed as a concise
  summary.

Raw logs must not contain passwords, cookies, CSRF values, bearer/API tokens,
hashes, DSNs, or complete configuration objects. Temporary paths under `/tmp`
are useful diagnostics but are not durable evidence after the environment is
gone.

## Promotion rule

1. Retain the successful current artifact.
2. Link it from this ledger and the affected Capability Profile/report.
3. Review that the comparator used only committed normalizers.
4. Bind evidence to only the exact rows exercised. Promote a tier-owning
   resource only after every applicable row it owns closes that tier boundary.
5. Keep the contract pre-publication until every first-profile exit condition
   is satisfied together.

An earlier successful run cannot be reused after comparator, behavioral
contract, implementation, fixture, or security changes without a fresh
execution. Claim-only tier/link/status metadata uses the audited exception
below.

## Claim-only attestation

Updating tier/evidence links and Status after a successful run changes
`make source-digest`, even though the tested behavior did not change. Use the
two-digest protocol in
[the execution playbook](../IMPLEMENTATION_EXECUTION_PLAYBOOK.md#two-digest-claim-attestation):

1. retain gates against an unchanged `tested_digest`;
2. allow only reviewed evidence summaries/links, exact tier/contract-state
   metadata, metadata-only generated docs, and Status claim changes;
3. record a file/content-hash diff proving no behavior, schema, scenario,
   fixture, comparator, normalizer, security rule, or toolchain changed;
4. compute the final `attestation_digest` and run `make check`; and
5. retain a mapping from the final digest to the tested digest and artifacts.

Anything outside that claim-only boundary invalidates the affected evidence and
requires a fresh external run.

Retained claims use `source-v2:sha256:<hex>` and
`netbox-go-evidence-v2`. The tested revision must be a commit reachable from
the claimed `HEAD`; its complete owned-source manifest is reconstructed from
the Git tree. File mode (`100644`/`100755`), byte size/hash, and closed relative
symlinks are part of that manifest. Two-digest claims retain the tested
manifest and must match it to the Git tree entry for entry.

Result payload SHA-256 commitments cover exact raw non-marker bytes before any
decoding; the whole artifact must be valid UTF-8 and marker lines must be ASCII.
The marker's claim digest covers the complete verification/evidence references
and each exact consumer row's full assessment, applicability, and proof
assignment/object. These checks prove structural integrity and prevent claim
rebinding. They cannot prove a command ran or its output/environment was
truthful, so durable runner provenance and designated human evidence acceptance
remain mandatory.
