# Core Workflow token-transport and unary safety — 2026-08-17

## Claim boundary

- Goal: `CW1-V1-02`.
- Increment: `CW1-V1-02-I2`.
- Executor: Codex GPT-5.6 Sol.
- State: `done`; the project owner accepted this bounded result at
  `2026-08-17T17:19:53Z`.
- Result: **accepted bounded source passed for the exact revision and digest
  below**.
- Capability Profile: `core-workflow-v1`, still T1 and pre-publication.

This artifact proves only the bounded token-transport and unary-method safety
slice specified by the
[`CW1-V1-02-I2` packet](../increments/CW1-V1-02-I2.md). It does not complete
the parent credential/session matrix, close V1, earn T2/T3/T4, change a
traceability row, publish the profile, complete a module or the rewrite, prove
deployment, or claim production readiness.

The artifact intentionally carries no `netbox-go-evidence-v2` marker. I2
promotes no compatibility or extension verification set, and fabricating a
tier marker or traceability change would exceed the reviewed increment. The
project owner accepted only this bounded result; the claim-only `done`
attestation does not promote a tier or complete the parent.

## Entry and tested source

- Packet-claim revision:
  `fffd7612753e3293d1865801767a3beaae7a289c` on `main`.
- Packet-claim source digest:
  `source-v2:sha256:3d6bfbe054a1f821968a661fa8b44c0fe2a8db4bb21e89bb926201ecc466b1c3`.
- Packet-claim owned entries: 2,999.
- Tested revision:
  `3049cc65e354e60015d16670dc53a42505b119fa` on `main`.
- Tested source digest:
  `source-v2:sha256:2c1579e228024d1573489a1cf05978d6cd85a8c0f41fdc1f2db61f34e7515e4c`.
- Tested owned entries: 3,002.
- Commit state: pushed to `origin/main`; the implementation worktree was clean
  before and after the exact-commit gates.
- Pinned upstream oracle revision:
  `fbb948d30e79ce657fac62994a22aca72c1770a9`.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.
- Increment execution began after the packet-claim CI completed at
  `2026-08-17T14:15:36Z`.
- Candidate CI completed when the tested revision's independent CI job
  succeeded at `2026-08-17T16:05:17Z`.

The packet-claim revision independently passed:

- Workflow: `repository-gate`.
- Run:
  [32037645156](https://github.com/jihadismail8/netbox-go/actions/runs/32037645156).
- Job:
  [95411143348](https://github.com/jihadismail8/netbox-go/actions/runs/32037645156/job/95411143348).
- Start: `2026-08-17T14:03:00Z`.
- Completion: `2026-08-17T14:15:36Z`.
- Conclusion: **success**.

## Implemented boundary

- The shared application service exposes transport-neutral typed token
  failures for missing, unknown, soft-revoked, expired, inactive-owner,
  unavailable-source, and denied-source states. Infrastructure failures and
  write-disabled denials remain separately typed and are not mislabeled as
  credential failures.
- Baseline REST accepts the pinned DRF single-header-value `Token` grammar,
  including ASCII case folding and Python byte-whitespace separation. It rejects
  unsupported, malformed, invalid-UTF-8, or duplicate field values before
  lookup and renders the exact declared NetBox details as HTTP 403 without
  `WWW-Authenticate`.
- The Go-owned identity-extension middleware remains isolated with its
  pre-existing parser and HTTP 401 boundary.
- Unary gRPC accepts exactly one strict `Bearer` metadata value with ASCII
  case folding, one or more ASCII spaces, and one opaque visible-ASCII
  credential. Malformed or duplicate values stop before lookup.
- Only the standard unary Health Check is public. Descriptor-driven tests
  classify all 85 canonical unary methods as protected: 28 reads and 57
  writes. Unknown methods default to write semantics and never inherit safety
  from `Get`- or `List`-looking names.
- REST and gRPC use the same application operation and prove matching
  credential kind, effective principal, lookup/touch call ordering, and
  handler containment where applicable. REST retains baseline-specific detail;
  gRPC retains generic safe messages and stable semantic codes.
- Restricted-token evaluation is deliberately bounded to the canonical direct
  peer. Forwarded-header and trusted-proxy resolution remain deferred.

I1 lookup classification, strict touch timing, revocation containment, and
PostgreSQL touch behavior remain unchanged. I2 changed no store, row, schema,
migration, transaction, route registration, protobuf, generated output,
OpenAPI, Vue source, profile, scenario, traceability, dependency, toolchain,
coverage policy, inventory, comparator, normalizer, fixture, or legacy stack.

## Changed source files

Production and tests:

1. `netbox-backend/internal/application/identity/service.go`
2. `netbox-backend/internal/application/identity/credential_errors.go`
3. `netbox-backend/internal/application/identity/credential_matrix_test.go`
4. `netbox-backend/internal/adapters/rest/netbox/identity/http.go`
5. `netbox-backend/internal/adapters/rest/netbox/identity/token_transport_matrix_test.go`
6. `netbox-backend/internal/adapters/rest/netbox/identity/credential_matrix_test.go`
7. `netbox-backend/internal/adapters/rest/netbox/router/router_test.go`
8. `netbox-backend/internal/adapters/grpc/identity/interceptor.go`
9. `netbox-backend/internal/adapters/grpc/identity/token_transport_matrix_test.go`
10. `netbox-backend/test/parity/identity_grpc_test.go`

Source-included governance:

11. `docs/increments/CW1-V1-02-I2.md`
12. `docs/increments/README.md`
13. `docs/ROADMAP.md`
14. `docs/STATUS.md`
15. `docs/TESTING.md`
16. `docs/IMPLEMENTATION_PLAN.md`
17. `AGENTS.md`

The tested commit contains 2,167 insertions and 92 deletions across exactly
these 17 paths. The present artifact and evidence-ledger link are
digest-excluded post-candidate receipt paths.

## Red-first and independent review evidence

All eight packet-named matrices were discovered before production correction:

```text
TestTokenCredentialOutcomeClassification
TestBaselineTokenAuthorizationGrammar
TestBaselineTokenOutcomeRendering
TestBaselineTokenHTTPMethodSafety
TestUnaryAuthenticatorBearerMetadataGrammar
TestUnaryAuthenticatorRPCSafetyClassification
TestUnaryAuthenticatorCredentialOutcomeMappings
TestRESTAndGRPCTokenCredentialParity
```

The pre-correction runs failed for the intended absent typed causes, REST
case/byte-whitespace/duplicate/state rendering, gRPC case/shape validation,
and unknown method-safety inference. Diagnostics used semantic case labels and
did not print credential material.

Independent read-only reviews then checked the full 17-path candidate. Review
caused additional evidence strengthening before commit:

1. accepted REST and gRPC credentials are proved to reach the application
   lookup unchanged without displaying them;
2. the extension REST parser is distinguished from the baseline parser;
3. parity rows assert typed causes and exact lookup-before-touch call order;
4. a write-enabled gRPC mutation-success row is present; and
5. the packet distinguishes effect-call parity from durable persistence.

The final reviewers found no code, security, scope, secret-containment, or
governance blocker. Deferred session, proxy, and stream behavior remained
explicit rather than being silently classified as covered.

## Focused and race results

Every focused matrix, every affected package, and the exact affected-package
race run passed. The exact race interval was `2026-08-17T14:42:02Z` through
`2026-08-17T14:43:02Z`, exit 0:

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test -race \
  ./internal/application/identity \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/rest/netbox/router \
  ./internal/adapters/grpc/identity \
  ./test/parity -count=1
```

Race results:

- application identity: 52.934s;
- REST identity: 28.303s;
- composed REST router: 10.785s;
- gRPC identity: 1.012s; and
- REST/gRPC parity: 22.326s.

The corresponding non-race affected-package command also passed:

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test \
  ./internal/application/identity \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/rest/netbox/router \
  ./internal/adapters/grpc/identity \
  ./test/parity -count=1
```

The focused commands were the exact application, baseline REST, composed
router, unary gRPC, and cross-transport expressions recorded in the
[increment packet](../increments/CW1-V1-02-I2.md#focused-runs). Each exited 0
after implementation, and a final combined discovery/review run confirmed all
eight top-level matrices were still present and green.

## Repository L0-L3 and independent CI

The final working-tree full gate passed from `2026-08-17T15:48:20Z` through
`2026-08-17T15:50:16Z`. After commit, the exact tested revision passed the same
gate again from `2026-08-17T15:50:53Z` through
`2026-08-17T15:52:42Z`, both at exit 0:

```bash
env \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  GOCACHE=/tmp/go-cache \
  GOFLAGS=-buildvcs=false \
  make check
```

Results included:

- backend lint, tests/race, builds, architecture, generated/frozen checks, and
  coverage policy passed;
- backend coverage 12.4365% (36,134/290,548) met the retained 12.4001% floor;
- frontend formatting, lint, application/test typechecks, coverage tests, and
  production build passed; 28 files and 125 tests passed at 61.72% statement
  coverage;
- inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries;
- traceability validator regressions passed 96/96;
- the active profile remained 13 resources, 3 interfaces, 17 scenarios, and
  293 traceability rows;
- OpenAPI remained 33 paths and 86 operations; and
- all local links were valid across 153 Markdown files.

Independent candidate result:

- Workflow: `repository-gate`.
- Run:
  [32043671270](https://github.com/jihadismail8/netbox-go/actions/runs/32043671270).
- Job:
  [95427263720](https://github.com/jihadismail8/netbox-go/actions/runs/32043671270/job/95427263720).
- Head SHA: `3049cc65e354e60015d16670dc53a42505b119fa`.
- Run created: `2026-08-17T15:53:04Z`.
- Job start: `2026-08-17T15:53:06Z`.
- Completion: `2026-08-17T16:05:17Z`.
- Conclusion: **success**.
- CI frontend statement coverage: 61.79% across the same 28 files and 125
  tests; the local exact-commit result above was 61.72%.

Every setup, repository-gate, post, and completion step reported success.

Evidence-receipt result before project-owner acceptance:

- Workflow: `repository-gate`.
- Evidence-receipt revision:
  `37f5fb2d78b431cf308812b2100a47064709e7a6` on `main`.
- Source digest:
  `source-v2:sha256:2c1579e228024d1573489a1cf05978d6cd85a8c0f41fdc1f2db61f34e7515e4c`.
- Owned entries: 3,002.
- Run:
  [32044758176](https://github.com/jihadismail8/netbox-go/actions/runs/32044758176).
- Job:
  [95430159051](https://github.com/jihadismail8/netbox-go/actions/runs/32044758176/job/95430159051).
- Head SHA: `37f5fb2d78b431cf308812b2100a47064709e7a6`.
- Job start: `2026-08-17T16:14:26Z`.
- Job completion: `2026-08-17T16:27:06Z`.
- Run completion: `2026-08-17T16:27:07Z`.
- Conclusion: **success**; every job step succeeded.

## Gate disposition

- L0-L3: passed locally on the exact tested digest; repository CI passed on
  the exact tested commit.
- L4 real PostgreSQL: not applicable because I2 changes no persistence shape,
  predicate, transaction, locking, or durable-effect implementation.
- L5 differential REST: skipped because I2 does not close a complete T2 owner
  row.
- L6 gRPC: local semantic parity passed; promotion evidence was skipped because
  the corresponding REST T2 boundary is absent.
- L7 deployment: not applicable because deployment wiring did not change.
- L8 browser: not applicable because Vue/browser behavior did not change.

No skipped boundary is implied by the local or CI results.

## Digest, secret, and containment checks

- The tested source digest remained
  `source-v2:sha256:2c1579e228024d1573489a1cf05978d6cd85a8c0f41fdc1f2db61f34e7515e4c`
  with 3,002 entries across the final focused, race, repository, commit, and CI
  sequence.
- Go formatting, Prettier, trailing-whitespace, scoped-diff, local-link, and
  generated/frozen checks passed.
- A scoped high-risk signature scan found no private-key marker,
  credential-bearing database URI, long literal authorization credential,
  JWT pattern, or provider token signature.
- No raw password, cookie, CSRF value, API/bearer credential, secret hash, DSN,
  complete configuration object, or credential-bearing header value is
  retained in this artifact.
- Test fixtures use synthetic opaque values; their values do not appear in
  test names, failure messages, logs retained here, or evidence links.
- The post-candidate artifact and ledger are excluded from `source-v2`. Their
  evidence-only commit left the tested digest and entry count unchanged and
  passed its exact-SHA repository CI before project-owner acceptance.

## Human closeout attestation

The project owner accepted this bounded result in the project thread at
`2026-08-17T17:19:53Z`. The acceptance changes only the state of
`CW1-V1-02-I2` from `evidence` to `done`; it does not close `CW1-V1-02`, change
a traceability row, complete an identity verification axis, or promote V1,
T2, T3, T4, the Capability Profile, publication, deployment, or production
readiness.

- Tested implementation revision:
  `3049cc65e354e60015d16670dc53a42505b119fa`.
- Tested implementation digest:
  `source-v2:sha256:2c1579e228024d1573489a1cf05978d6cd85a8c0f41fdc1f2db61f34e7515e4c`
  with 3,002 entries.
- Pre-closeout evidence revision:
  `37f5fb2d78b431cf308812b2100a47064709e7a6`; its source digest remained the
  tested implementation digest.
- Accepted closeout revision:
  `105bcf426d2ae52be38f88afede774a2f54abec6`.
- Accepted closeout digest:
  `source-v2:sha256:61e2821bb85036df3f62219c807086086cf42a0b8b62b149a87cf0c0262e8b03`
  with 3,002 entries.
- Exact-commit local result: the pinned-environment `make check` ran from
  `2026-08-17T17:33:55Z` through `2026-08-17T17:35:30Z` at exit 0. Backend
  coverage remained 12.4365% (36,134/290,548) against the 12.4001%
  (36,013/290,425) floor; 28 frontend files and 125 tests passed at 61.72%
  statement coverage; 96 traceability-validator tests passed; the active
  profile remained 13 resources, 3 interfaces, 17 scenarios, and 293 rows;
  inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries; OpenAPI remained 33 paths and 86 operations; and
  all local links were valid across 154 Markdown files.
- Exact-commit independent result: repository-gate
  [run 32051029055](https://github.com/jihadismail8/netbox-go/actions/runs/32051029055),
  [job 95450168136](https://github.com/jihadismail8/netbox-go/actions/runs/32051029055/job/95450168136),
  started `2026-08-17T17:35:49Z`, completed
  `2026-08-17T17:45:49Z`, and reported success for every step at the accepted
  closeout SHA.

The exact local command was:

```bash
env \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  GOCACHE=/tmp/go-cache \
  GOFLAGS=-buildvcs=false \
  make check
```

Replacement- and graft-disabled Git reads reconstructed the reachable before
and after trees. Every changed entry retained mode `100644`. The exact
`37f5fb2d` to `105bcf42` file mapping is below; hashes are SHA-256 over the raw
blob bytes.

| Path                                                                 | Before bytes | Before SHA-256                                                     | After bytes | After SHA-256                                                      |
| -------------------------------------------------------------------- | -----------: | ------------------------------------------------------------------ | ----------: | ------------------------------------------------------------------ |
| `AGENTS.md`                                                          |        5,235 | `c6f0773d9a4c2f812537bed0bd484953416081bef3d4d38ea6f37ace37b41de4` |       5,191 | `6c01cd762ca9f3f708ebf25d9c756657a951a649031fc234dfb306875cef6510` |
| `docs/IMPLEMENTATION_PLAN.md`                                        |       84,988 | `a11aaa4c1a615e4a03776cb0291e4a2f5b2ac186cb9170c2c7f92cbf7f0a3795` |      85,105 | `f9e74798d50aa745522775e5e7eb362b5dfe384da9cd5e658dbcd446a9a521d3` |
| `docs/ROADMAP.md`                                                    |       75,377 | `cc2833630d9558252a1aec3f7bc49938eaed3be4f60cc35284e58c2b3ef47248` |      75,368 | `6454fbeee1644234f7b7e7ee03b54e3fe1adb33a1064ba72556be5a4bf91c3a8` |
| `docs/STATUS.md`                                                     |       22,285 | `90c604e170c28f2ea0105cd9f3de749817065bef63cb48f3f780e8c3373b5d71` |      22,235 | `1bc5b2ecc4f5f1946aae46d7aaf82e2382300254a60abf5ed26e447bbebc59b8` |
| `docs/TESTING.md`                                                    |       13,010 | `0f36978049cc60349463ce0880a242b2de39bc3fe6626b2fb3e10b46291ed51b` |      13,019 | `c92ce7fcc12e65e99a1ccac5091a7182941d78fc5eed1913916d54deb54e0fcc` |
| `docs/increments/CW1-V1-02-I2.md`                                    |       33,152 | `0e416329941805a88b7d77d2be11718fedec802b10c319ea296304ea36989cf3` |      33,331 | `b4745e81599822e2559e8c240832824d217f76dc2f19d1a5c21ab7a290f2d634` |
| `docs/increments/README.md`                                          |        3,098 | `f7f9901cfaf3444b5688f5b9f4ffc51361237b3aa64315405d814aedab1c8650` |       3,074 | `5ee0bf0310ff55665fad63a6d2f70fb9d3d5bd0d996e2f8882fe67f7e7b2be86` |
| `docs/evidence/2026-08-17-core-workflow-v1-token-transport-i2-v0.md` |       13,292 | `b5352a3856888023ff5779352e8c334fd2fd104c52f7719cf3ea4b8eeabf7387` |      13,995 | `358858186caf0510e3dfcbe36f529dd23d9fe08c4d0dcb20c6996751c34d38fb` |
| `docs/evidence/README.md`                                            |        6,673 | `5f364a2cf4b604118b8656437556b99d8d013b9f83e75a81aaf2b0762539954b` |       6,555 | `8233c2eaa4291c2c4ff284496217ae7626895aae3991a82e6f318f0fb68c5f4c` |

The first seven rows are source-v2 inputs and constitute the complete
claim-only source diff. The final two rows are excluded evidence paths. No
entry was added, removed, renamed, retargeted, or made executable, and no
other byte changed. The complete manifests retained 3,002 entries before and
after the claim.

The claim-only transition changed no application behavior, route, schema,
migration, security rule, scenario, fixture, comparator, normalizer,
dependency, toolchain, coverage policy, verification set, or parent-goal
state. This artifact still intentionally carries no `netbox-go-evidence-v2`
marker because I2 promotes no traceability consumer or extension verification
set. The evidence-only receipt containing this attestation is excluded from
`source-v2`; it must preserve the accepted closeout digest and pass its own
exact-SHA repository CI.

Human reviewer: project owner.

Human decision: accepted only `CW1-V1-02-I2` in the project thread at
`2026-08-17T17:19:53Z`; the parent goal remains open.

## Residual work and risk

The following `CW1-V1-02` work remains open and unclaimed:

- browser-session validity, expiry, rotation, logout, and password-change
  invalidation;
- complete cookie and CSRF lifecycle behavior;
- session-plus-token credential precedence and provenance;
- trusted-proxy and forwarded-client-source resolution;
- login and token-creation throttling, with exact scope and storage semantics
  still requiring reviewed child-increment documentation;
- gRPC Health Watch, reflection, and unknown-stream authentication; and
- aggregate credential/session closeout.

The current frontend dependency audit still reports three high-severity
advisories. They predate I2 and were not changed or waived; dependency changes
were outside this increment.

The parent `CW1-V1-02`, V1, T2/T3/T4, Capability Profile publication, full
rewrite, deployment, and production-readiness claims all remain open. The
project owner accepted the bounded I2 result, so I2 is `done`; no broader claim
changes.
