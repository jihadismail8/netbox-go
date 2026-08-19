# Compatibility roadmap

- Status: canonical goal and dependency ledger
- Updated: 2026-08-03
- Compatibility Baseline: `fbb948d30e79ce657fac62994a22aca72c1770a9`
  (`v4.4.6-7-gfbb948d30`)

This roadmap defines the outcomes, ordering, checklists, and evidence needed to
reach a [Complete In-Scope Replacement](../CONTEXT.md) and then a production
release. It is deliberately goal-based rather than estimate-based. A goal
closes only when its exit evidence is retained and reviewed; code, routes,
tables, generated files, or unchecked assumptions never close a goal.

[Status](STATUS.md) is authoritative for current claims,
[Compatibility](COMPATIBILITY.md) defines T0–T4 and completion, and the
[execution playbook](IMPLEMENTATION_EXECUTION_PLAYBOOK.md) defines the detailed
implementation procedure. This roadmap is the queue and dependency map; it
does not broaden any Capability Profile.

## How to use this roadmap

### Goal states

| State         | Meaning                                                                   |
| ------------- | ------------------------------------------------------------------------- |
| `done`        | Exit evidence is retained and linked for the exact claimed source.        |
| `continuous`  | Currently green but must be re-established after relevant changes.        |
| `ready`       | Every entry condition is satisfied; implementation may start.             |
| `in-progress` | One named owner is executing a bounded increment.                         |
| `evidence`    | Implementation is complete; required external evidence is being retained. |
| `blocked`     | A named dependency or human decision is incomplete.                       |
| `candidate`   | Planning proposal only; it is not an accepted public contract.            |

Checkboxes record work inside a goal. They do not promote a compatibility tier
by themselves. When a checkbox represents evidence, it may be checked only
after the artifact is durable and linked. Every handoff must record the goal
ID, owner, status, dependencies, tested digest, commands, evidence links, and
residual risks.

### Delivery rules

- Keep `CW1-G00` green. A red or stale V0 blocks every merge.
- Complete one bounded goal increment at a time. Do not combine unrelated
  cleanup, feature scope, or production hardening.
- Deliver coherent Capability Profiles, not module percentages or generated
  CRUD breadth.
- REST must reach T2 before the corresponding gRPC behavior can reach T3;
  applicable browser workflows reach T4 only after their supported backend
  capabilities are verified.
- REST, gRPC, CLI, and Vue do not own separate business logic. Public adapters
  invoke the same typed application use cases; Vue uses REST.
- Authorization, validation, persistence, derived state, object changes, and
  required durable event intent share one application-owned transaction.
- Fail closed for credentials, permissions, visibility, fields, filters,
  choices, identifiers, origins, and undeclared routes.
- The development database remains disposable. `AutoMigrate` creates only a
  confirmed-absent table and never inspects, repairs, alters, or backfills an
  existing table.
- Python/Django is allowed only in the pinned differential oracle. Extension
  replacements run out of process.
- Retire displaced frozen artifacts only after capability completion under
  ADR 0004. ADR 0005 remains a one-time wrapper-cleanup exception.
- Do not start or merge a later Capability Profile before `CW1-V6-03`.
- Never weaken tests, coverage, security, normalizers, comparators, inventories,
  or evidence requirements to close a goal.

### Working philosophy

1. **Replace behavior, not Python structure.** A NetBox class, file, table, or
   generated route is evidence to investigate, never the unit of delivery.
2. **Prefer a narrow truth over broad scaffolding.** Accept one coherent
   Capability Profile, implement it vertically, and prove it before adding
   breadth.
3. **Keep one semantic core.** REST, gRPC, CLI, and Vue are adapters around
   shared typed application behavior, not independent products.
4. **Make invalid transitions explicit.** Domain types and use cases own
   validation, presence semantics, authorization inputs, and state changes;
   transport and persistence representations do not leak inward.
5. **Treat atomicity as behavior.** A mutation is correct only when its data,
   derived state, object changes, and required event intent commit or roll back
   together.
6. **Name secure divergence.** A deliberately hardened identity or extension
   behavior is documented and tested separately; it is never normalized into a
   false baseline-compatibility claim.
7. **Make evidence part of the deliverable.** Tests and harnesses show intent;
   only current, digest-bound, reviewed artifacts earn a tier or close a goal.
8. **Keep increments reversible and reviewable.** One owner, one goal ID, one
   declared file boundary, and one lowest-layer regression precede broader
   gates.
9. **Measure against accepted denominators.** File counts, route counts, tables,
   and compile success do not measure compatibility or completion.
10. **Separate compatibility from production.** Complete replacement proves
    the accepted behavior; production release additionally proves migrations,
    cutover, security, operability, resilience, and supply chain.

## Current program snapshot

| Dimension                   | Current                                  | Remaining                                                                |
| --------------------------- | ---------------------------------------- | ------------------------------------------------------------------------ |
| Baseline catalogue          | 155/155 resource/action entries recorded | 0 high-level catalogue entries; detailed business-rule discovery remains |
| Accepted in-scope breadth   | 13/153 entries at T1 (8.50%)             | 140 T0 entries (91.50%)                                                  |
| Resource breadth            | 13/131 resources at T1                   | 118 resources                                                            |
| Custom actions              | 0/22 promoted                            | 22 actions                                                               |
| Verified compatibility      | 0 retained T2, T3, or T4 capabilities    | Every accepted capability                                                |
| First-profile repository V0 | Retained source-v2 entry result          | Refresh continuously for every relevant merged source                    |
| Later profile candidates    | 19 catalogued candidates                 | All blocked on first-profile V6                                          |
| Module closeouts            | 0/10                                     | 10                                                                       |
| Extension Service closeout  | Architectural boundary accepted          | Contract/classification evidence                                         |
| Production readiness        | 0/7 programs signed off                  | PROD-1 through PROD-7                                                    |

Explicit inventory exclusions remain `/api/extras/scripts/` and anonymous
`/api/users/tokens/provision/`. GraphQL, Python runtime/plugin/script/report
compatibility, and pixel parity remain outside the accepted runtime boundary.
Production import/upgrade is not REST compatibility, but it is mandatory for a
production release through PROD-1 and PROD-2.

## Gate-to-goal map

| Gate | Outcome                             | Goals                                   | Current state                                        |
| ---- | ----------------------------------- | --------------------------------------- | ---------------------------------------------------- |
| 0    | Stable language and decisions       | CONTEXT plus ADR 0001–0005              | `done`                                               |
| 1    | Repository/shared-core baseline     | `CW1-G00`                               | `continuous`                                         |
| 2    | Unified identity and authorization  | `CW1-V1-01` through `CW1-V1-04`         | V1-01/I1/I2/I3 `done`; parent active                 |
| 3    | Traceable differential/parity proof | `CW1-V2-*`, `CW1-V4-*`                  | V2-01 `done`; remainder `blocked`                    |
| 4    | First Core Workflow Profile         | `CW1-V1-*` through `CW1-V6-*`           | T1; active program                                   |
| 5    | In-scope feature expansion          | `CP-P01`–`CP-P19`, then `MC-01`–`MC-10` | `blocked`                                            |
| 6    | Extension Service contracts         | `EXT-01`                                | `blocked` on a named inventory                       |
| 7    | Production release                  | `PROD-1`–`PROD-7`                       | `blocked` on replacement and applicable EXT closeout |

## Dependency map

```text
EXT-01 (profile-required contract) --> affected CP-Pxx goal

CW1-G00 (continuous V0)
   |
   +--> CW1-V1 identity/security --------+
   +--> CW1-V2 traceability/behavior ----+--> CW1-V4 REST T2 -> matching gRPC T3
   +--> CW1-V3 PostgreSQL/readiness -----+                         |
                                                                  v
                                                        CW1-V5 browser T4
                                                                  |
                                                                  v
                                                        CW1-V6 sign-off
                                                                  |
                                                                  v
                CP-P01..CP-P19 -> MC-01..MC-10
                                                                  |
                                                                  v
                                                Complete In-Scope Replacement
                                                                  |
                EXT-01 (deployment-only offered contracts) -------+
                                                                  |
                    PROD-1..PROD-6 -------------------------------+
                                                                  |
                                                                  v
                                                        PROD-7 release
```

The starting `CW1-V1-01` and `CW1-V2-01` outcomes are accepted. Bounded
`CW1-V1-02-I1` is accepted and done for the token credential foundation while
the `CW1-V1-02` parent remains active. `CW1-V1-02-I2` is accepted and done for
REST token grammar/outcomes and unary gRPC bearer/method safety.
`CW1-V1-02-I3` is accepted and done for the browser-session, CSRF,
login/logout, and REST credential-arbitration slice. The next separately
reviewed `CW1-V1-02` child and ready `CW1-V3-01` may run in parallel only when
their packets declare non-overlapping files and owners; otherwise serialize
them. All merges serialize through `CW1-G00`. Production foundations may
proceed in parallel only when they do not destabilize profile work or claim
production readiness early.

## Current execution board

| Goal                | State         | Immediate outcome                                | Hard dependency                      |
| ------------------- | ------------- | ------------------------------------------------ | ------------------------------------ |
| `CW1-G00`           | `continuous`  | Keep source-v2 V0 current                        | Every relevant merged source         |
| `CW1-V1-01`         | `done`        | Trusted-origin CORS retained                     | Retained entry-source `CW1-G00`      |
| `CW1-V1-02`         | `in-progress` | Continue the later bounded identity children     | `CW1-V1-01` and current `CW1-G00`    |
| `CW1-V1-03`–`V1-04` | `blocked`     | Complete and retain remaining identity/security  | Preceding V1 goal                    |
| `CW1-V2-01`         | `done`        | Accepted structural rule/scenario traceability   | Fresh source-v2 V0 follows this gate |
| `CW1-V2-02`–`V2-08` | `blocked`     | Close typed business behavior                    | V2-01 and named lane dependencies    |
| `CW1-V3-01`         | `ready`       | Make HTTP/gRPC readiness dependency-aware        | `CW1-G00`                            |
| `CW1-V3-02`–`V3-05` | `blocked`     | Close PostgreSQL/deployment evidence             | Named V2/V3 dependencies             |
| `CW1-V4-01`–`V4-03` | `blocked`     | Earn complete first-profile REST T2              | V1–V3 as declared                    |
| `CW1-V4-04`         | `blocked`     | Retain the identity extension report             | `CW1-V1-04`                          |
| `CW1-V4-05`         | `blocked`     | Earn corresponding gRPC T3 per capability        | Retained T2 for that capability      |
| `CW1-V4-06`         | `blocked`     | Retain the complete first-profile T2/T3 boundary | All V4 lanes                         |
| `CW1-V5-01`–`V5-04` | `blocked`     | Author complete browser scenarios                | Stable named V1/V2/REST contracts    |
| `CW1-V5-05`         | `blocked`     | Retain T4 for exercised workflows                | Corresponding T2/T3 and V5 authoring |
| `CW1-V6-*`          | `blocked`     | Converge evidence and sign off the first profile | V1–V5                                |
| `CP-P01`–`CP-P19`   | `blocked`     | Expand accepted feature breadth                  | `CW1-V6-03`                          |

No owner is assigned merely because a goal is `ready`. An executor claims one
goal by recording its increment specification before editing code.

# Phase A — finish `core-workflow-v1`

## `CW1-G00` — keep V0 current

Outcome: every merged source digest retains the deterministic repository gate
without weakening its policies.

Historical evidence:

- [x] The 2026-08-03 post-cleanup/audit source had a retained direct
      final-digest `make check` pass under the superseded mode-blind v1 source
      digest.
- [x] Pinned Go 1.26.0, Node 24.18.0, and npm 11.16.0 were used.
- [x] Coverage, inventory, contract, OpenAPI, architecture, frontend, and
      documentation checks are part of the gate.

Current source-v2 entry state: the exact revision and digest linked by the
[evidence ledger](evidence/README.md) have a human-reviewed retained V0. The
versioned digest includes executable mode and closed owned symlinks, so v1
artifacts cannot be relabeled as source-v2 evidence. Any source edit, including
this claimed increment, requires a fresh exact-source result before merge.

Repeat for every relevant change:

- [ ] Record the starting source digest and pinned tool versions.
- [ ] Name permitted files, forbidden scope, focused tests, and exit condition.
- [ ] Run the smallest L0/L1 checks before broader gates.
- [ ] Run affected backend/frontend L2 gates.
- [ ] Run `make check` at L3.
- [ ] Confirm test commands did not mutate owned source.
- [ ] Retain a credential-free exact-digest artifact when the previous V0 is
      invalidated.
- [ ] Keep coverage/package/exclusion baselines intact.
- [ ] Keep inventories, generated contracts, and architecture prohibitions
      drift-free.
- [ ] Confirm no deferred route, generated service, or retired wrapper became
      runtime-enabled or regenerated.

Exit: the merged digest is V0-green and conservatively recorded. Any red or
stale V0 blocks all other goals.

## V1 goals — identity, RBAC, and security

### `CW1-V1-01` — trusted-origin CORS

Entry: `CW1-G00` is current. SEC-009 already decides the policy; no new product
decision is required.

Completed packet: [CW1-V1-01 trusted-origin CORS](increments/CW1-V1-01.md).

- [x] Add an owned explicit origin allowlist to configuration.
- [x] Reject invalid origin configuration at startup.
- [x] Never emit a credentialed wildcard origin.
- [x] Give missing, malformed, or untrusted origins no CORS grant.
- [x] Give a trusted origin only declared methods and headers.
- [x] Test preflight and actual credentialed requests.
- [x] Keep CSRF mandatory for cookie-authenticated mutations.
- [x] Add focused configuration and canonical HTTP regressions.
- [x] Pass L0–L3 without broadening public endpoints.

Exit: trusted-origin behavior is explicit, fail-closed, and tested without
weakening CSRF.

### `CW1-V1-02` — credential, session, and token matrix

Entry: `CW1-V1-01` is merged and `CW1-G00` is green.

Completed bounded packets: [CW1-V1-02-I1 token credential
foundation](increments/CW1-V1-02.md), [CW1-V1-02-I2 token transport and unary
RPC safety](increments/CW1-V1-02-I2.md), and [CW1-V1-02-I3 browser session and
CSRF lifecycle](increments/CW1-V1-02-I3.md). The parent checklist remains open
until every later transport/session/throttle row is proved. I3 proves typed
session outcomes, valid-session-first arbitration, transactional login/logout,
exact CSRF pairing/recovery, and cookie shape. I2 proves only the direct-peer
token transport slice; throttles, trusted-proxy source resolution,
password-session policy, gRPC streams, and aggregate all-transport coverage
remain open.

- [ ] Test missing, malformed, unknown, expired, revoked, write-disabled, and
      IP-restricted API tokens.
- [x] Test a token belonging to an inactive user.
- [x] Prove baseline `last_used` ordering and one-minute throttling.
- [x] Prove an unknown token key performs no durable write.
- [x] Test valid, invalid, and expired browser sessions.
- [x] Test logout invalidation.
- [ ] Test password-change invalidation.
- [x] Test session fixation resistance and rotation.
- [x] Test missing and invalid CSRF for cookie-authenticated mutations.
- [ ] Test login and token-creation throttling.
- [ ] Cover REST and applicable gRPC credential paths with stable outcomes.

Exit: every credential row has a passing focused test and a stable expected
reason/state result.

### `CW1-V1-03` — authorization, administration, and secrets

Entry: `CW1-V1-02` is complete and its credential semantics are stable.

- [ ] Cover superuser, direct global, group global, user-object, group-object,
      no-grant, and wrong-action principals.
- [ ] Cover view/add/change/delete distinctions.
- [ ] Prove visibility is applied before count, ordering, and pagination.
- [ ] Prove identical REST and gRPC Principal decisions.
- [ ] Test empty-database administrator bootstrap and refusal afterward.
- [ ] Test protected password reset, superuser-authorized user creation, and
      global model-permission grant.
- [ ] Prove no password appears in argv, output, logs, SQL logs, or evidence.
- [ ] Prove no anonymous network bootstrap, reset, user creation, or grant.
- [ ] Review session-cookie and CSRF bootstrap flags/shape.
- [ ] Prove API-token secret material appears once on creation only.
- [ ] Prove list/get never return reusable secret material.
- [ ] Scan all retained artifacts for credentials and hashes.

Exit: focused application, REST, gRPC, and CLI tests cover every row; identity
operations remain classified as extensions where applicable.

### `CW1-V1-04` — retain V1

Entry: `CW1-V1-01` through `CW1-V1-03` pass on one unchanged tested digest.

- [ ] Run the complete credential, authorization, administration, and secret
      matrix.
- [ ] Record commands, versions, times, totals, and non-secret configuration.
- [ ] Link every matrix row to both public interfaces where applicable.
- [ ] Scan the evidence bundle for secrets.
- [ ] Record extension classification separately from baseline compatibility.
- [ ] Apply the two-digest attestation procedure if claim metadata changes.

Exit: current V1 evidence is durable and reviewed. V1 does not imply T2 or T3.

## V2 goals — traceability and business behavior

### `CW1-V2-01` — machine-readable traceability

Entry: `CW1-G00` is current. Existing profile/resource/scenario metadata is
input, not evidence.

- [x] Create one row for every manifest scenario and every stable
      Implementation Plan rule.
- [x] Give each row a stable ID, capability, and operation.
- [x] Link exact pinned source and upstream-test references.
- [x] Link profile/resource metadata.
- [x] Link domain, application, and PostgreSQL test paths.
- [x] Link REST differential, corresponding gRPC, and applicable browser cases.
- [x] Record retained evidence/tier or an explicit pending/not-applicable reason.
- [x] Reject missing rows, stale paths, unclassified rules, and tier inflation.
- [x] Wire the validator into `contracts-check`.

Exit: structural traceability is complete and unresolved behavior remains
visibly pending rather than falsely green.

### `CW1-V2-02` — common CRUD and presence semantics

Entry: `CW1-V2-01` is green.

- [ ] Represent all 13 resources × list/get/create/PUT/PATCH/delete.
- [ ] Cover defaults, read-only fields, and nullability.
- [ ] Cover absent, explicit null, zero, empty, and concrete values.
- [ ] Cover validation, unauthenticated, forbidden, not-found, conflict, and
      delete-protection reasons.
- [ ] Add the lowest-layer positive and negative regressions.
- [ ] Prove failed mutations leave no forbidden durable state.
- [ ] Update every traceability row.

Exit: every common rule has positive/negative typed tests and no unresolved
implementation row.

### `CW1-V2-03` — DCIM hierarchy, uniqueness, and rack placement

Entry: `CW1-V2-01` and relevant common semantics are complete.

- [ ] Prove global, sibling, manufacturer, site, and location uniqueness where
      the baseline applies it.
- [ ] Prove DeviceRole hierarchy cycles, cascade, and protection behavior.
- [ ] Prove rack bounds, half-unit placement, occupancy, and protected changes.
- [ ] Prove RackType propagation behavior.
- [ ] Prove DeviceType height transitions.
- [ ] Add domain, application, and real-PostgreSQL regressions where required.

Exit: every listed hierarchy/placement rule is implemented, atomic where
required, and traceable.

### `CW1-V2-04` — Device, InterfaceTemplate, and Interface transactions

Entry: common semantics and relevant DCIM rules are stable.

- [ ] Prove nullable Device name and case-insensitive uniqueness behavior.
- [ ] Prove InterfaceTemplate snapshot timing and non-retroactivity.
- [ ] Prove atomic Device plus instantiated-Interface creation.
- [ ] Prove Interface immovability and protected deletion.
- [ ] Force a late template failure and roll back Device, Interfaces, and object
      changes.
- [ ] Prove authorization occurs before persistence.

Exit: every multi-object Device mutation has positive and forced-failure
rollback evidence.

### `CW1-V2-05` — IPAM values, containment, and uniqueness

Entry: traceability and common presence semantics are stable.

- [ ] Prove Prefix canonicalization, containment, and `/0` rejection.
- [ ] Prove host-bit input rejection and canonical suggestion.
- [ ] Prove utilization flags and VRF/global uniqueness.
- [ ] Preserve IP host address and mask.
- [ ] Prove role exceptions and IPv4/IPv6 edge-prefix behavior.
- [ ] Prove SLAAC is IPv6-only and DNS names normalize as declared.
- [ ] Preserve the accepted rule that Interface–VRF equality is not invented.
- [ ] Add domain, application, and PostgreSQL regressions.

Exit: every accepted/rejected IPAM transition is implemented and traceable.

### `CW1-V2-06` — assignment transaction atomicity

Entry: Interface and IPAddress behavior from `CW1-V2-04` and `CW1-V2-05` is
stable.

- [ ] Cover the full assignment presence matrix.
- [ ] Cover assignment, reassignment, and unassignment.
- [ ] Cover invalid object pairs and duplicate/concurrent assignment.
- [ ] Make REST PATCH and gRPC actions invoke one typed use case.
- [ ] Prove failure rolls back address, relation, derived state, and object
      changes.
- [ ] Prove one application-owned PostgreSQL transaction.

Exit: assignment behavior is atomic and has exact stable outcomes.

### `CW1-V2-07` — queries, visibility, projections, and object changes

Entry: common semantics are stable; final authorization rows depend on
`CW1-V1-03`.

- [ ] Cover every declared filter, search key, and ordering key.
- [ ] Cover bounded pagination and invalid values.
- [ ] Cover nested projections and read-only counters.
- [ ] Apply visibility before count, order, and page.
- [ ] Prove authorization before persistence.
- [ ] Prove exact object-change rows and required side effects.
- [ ] Prove rollback removes object changes and derived state.

Exit: every query and object-change rule has positive/negative typed evidence.

### `CW1-V2-08` — close V2

Entry: `CW1-V2-01` through `CW1-V2-07` are complete.

- [ ] Confirm no unclassified scenario or plan rule remains.
- [ ] Confirm every invariant has positive and negative tests.
- [ ] Confirm every multi-object mutation proves atomic rollback.
- [ ] Confirm all referenced test paths exist.
- [ ] Leave pending reasons only for still-open V1 evidence and V3–V5 external
      boundaries.
- [ ] Pass `contracts-check`, affected L1/L2 gates, and `CW1-G00`.
- [ ] Retain the V2 traceability/evidence report.

Exit: V2 is reviewed and retained; this still does not promote T2.

`CW1-V2-03`, `CW1-V2-04`, `CW1-V2-05`, and parts of `CW1-V2-07` may run in
parallel only with disjoint resource-family ownership. `CW1-V2-06` depends on
both DCIM and IPAM lanes.

## V3 goals — real PostgreSQL and deployment

### `CW1-V3-01` — dependency-aware readiness

Entry: `CW1-G00` is current. This may overlap early V1/V2 work only when the
increment specifications declare disjoint permitted files and ownership.

- [ ] Define liveness as process-alive only.
- [ ] Make HTTP `/ready` check required PostgreSQL connectivity.
- [ ] Make gRPC health reflect required dependency state.
- [ ] Make dependency loss fail readiness.
- [ ] Make dependency recovery restore readiness.
- [ ] Add deterministic tests without arbitrary sleeps.
- [ ] Keep business logic out of probe adapters.

Exit: HTTP and gRPC readiness truthfully track PostgreSQL loss/recovery.

### `CW1-V3-02` — bootstrap and schema contract

Entry: `CW1-G00` is current and `CW1-V2-01` traceability exists. This may
overlap `CW1-V3-01` only with disjoint permitted files and ownership.

- [ ] Test empty-database bootstrap.
- [ ] Prove a correct existing table remains untouched.
- [ ] Create a missing table in a non-empty database.
- [ ] Prove a malformed existing table is not inspected or repaired.
- [ ] Validate deterministic 198-entry startup ordering without treating the
      count as capability evidence.
- [ ] Validate private tables, columns, constraints, indexes, FK actions, and
      content-type uniqueness.
- [ ] Preserve missing-table-only `AutoMigrate`.

Exit: real PostgreSQL tests prove the bootstrap/schema policy.

### `CW1-V3-03` — locking, rollback, reconnect, and restart

Entry: `CW1-V3-02` is complete; `CW1-V1-03` is complete for identity rows; and
the corresponding `CW1-V2-03` through `CW1-V2-07` behavior is complete.

- [ ] Prove concurrent duplicate, placement, allocation, and assignment
      strategies.
- [ ] Prove rollback removes derived objects and object changes.
- [ ] Prove identity/session/token/CLI state survives reconnect and restart.
- [ ] Use deterministic locking assertions.
- [ ] Do not substitute SQLite for PostgreSQL concurrency evidence.

Exit: every declared concurrency, rollback, and restart row has real
PostgreSQL proof.

### `CW1-V3-04` — deployment loss/recovery smoke

Entry: `CW1-V3-01` through `CW1-V3-03` are complete.

- [ ] Use a unique clean Compose project and volume.
- [ ] Use no Django migration or initialization-SQL mount.
- [ ] Bootstrap schema and content types.
- [ ] Assert real dependency health/readiness.
- [ ] Exercise PostgreSQL loss and recovery.
- [ ] Restart the application and prove idempotence.
- [ ] Tear down only owned resources.
- [ ] Retain no credential artifacts.

Exit: deployment smoke proves fresh start, dependency truth, recovery, restart,
and safe teardown.

### `CW1-V3-05` — retain V3

Entry: V1/V2 are closed and all V3 implementation goals pass on one digest.

- [ ] Run the explicit real-PostgreSQL DSN suite from Testing.
- [ ] Run `make deployment-smoke`.
- [ ] Use the same source digest for both.
- [ ] Record versions, times, totals, and redacted configuration.
- [ ] Scan artifacts for secrets.
- [ ] Link durable artifacts and apply claim attestation.

Exit: reviewed PostgreSQL and deployment artifacts share one tested digest.

## V4 goals — REST T2, then corresponding gRPC T3

### `CW1-V4-01` — close differential scenarios and fixtures

Entry: V2 is closed and V3 implementation is stable.

- [ ] Isolate deterministic state for every profile scenario.
- [ ] Enforce oracle SHA and effective-configuration refusal.
- [ ] Keep the comparator sensitivity self-test.
- [ ] Bind every generated ID to a named scenario object.
- [ ] Use only committed normalizers.
- [ ] Add every missing permission, presence, conflict, invariant, rollback,
      and durable-effect case.
- [ ] Keep identity divergences outside baseline T2 and report them separately.

Exit: the harness can exercise every declared T2 row without an unclassified
normalization or fixture gap.

### `CW1-V4-02` — DCIM REST T2

Entry: V1–V3 are retained/current and `CW1-V4-01` is complete.

- [ ] Exercise six operations for all 10 DCIM resources.
- [ ] Match status, path/slash, authentication, errors, JSON types, fields,
      choices, pagination, state, and side effects.
- [ ] Cover DeviceRole hierarchy.
- [ ] Cover RackType/Rack propagation and placement.
- [ ] Cover DeviceType transitions and Device naming.
- [ ] Cover Interface immovability and template non-retroactivity.
- [ ] Cover limited users, presence matrices, protected deletes, and exact
      object-change outcomes.
- [ ] Retain results per capability; leave failures at T1.

Exit: passing row subsets may be retained per capability, but rows are proof
units rather than tier owners. Each DCIM resource remains T1 until every
applicable row it owns closes T2; this goal remains open until all 10 resources
are T2.

### `CW1-V4-03` — IPAM REST T2

Entry: V1–V3 are retained/current and `CW1-V4-01` is complete. It may use an
independent isolated fixture lane from DCIM.

- [ ] Exercise six operations for VRF, Prefix, and IPAddress.
- [ ] Cover canonicalization, containment, uniqueness, and edge prefixes.
- [ ] Cover assignment invalid pairs, reassignment, unassignment, and rollback.
- [ ] Cover limited users, presence, protected deletes, state, and side effects.
- [ ] Match every strict comparator dimension.
- [ ] Retain results per capability; leave failures at T1.

Exit: passing row subsets may be retained per capability, but each IPAM
resource remains T1 until every applicable row it owns closes T2. This goal
remains open until VRF, Prefix, IPAddress, and declared assignment REST
behavior are all T2.

### `CW1-V4-04` — identity extension report

Entry: V1 is retained and extension classification is stable.

- [ ] Test extension REST/gRPC semantics separately.
- [ ] Never normalize a secure divergence into T2.
- [ ] Link credential, RBAC, administration, and secret outcomes.
- [ ] Keep the extension label in inventory and documentation.

Useful partial artifacts may be retained against exact rows and proof
dimensions without completing an extension axis. An axis becomes `complete`
only when every applicable extension row/dimension and required project-
support row closes. Extension evidence never earns T2/T3/T4.

Exit: identity evidence is complete without changing baseline tier counts.

### `CW1-V4-05` — corresponding gRPC T3

Entry: each individual REST capability must already have retained T2.

- [ ] Use typed transport-native messages.
- [ ] Invoke the same application use case as REST.
- [ ] Prove lifecycle, error, RBAC/list visibility, rollback, assignment, state,
      and object-change equivalence.
- [ ] Do not chain through REST.
- [ ] Do not leak generic workflow or storage types.
- [ ] Run and retain parity per capability.

Exit: passing row subsets may retain exact evidence, but they do not promote a
tier independently. A resource capability reaches T3 only when every
applicable row it owns closes the T3 boundary; this goal remains open until
every corresponding first-profile resource capability is T3.

### `CW1-V4-06` — close V4

Entry: all 13 resources and assignment behavior have T2/T3; the identity
extension report is complete.

- [ ] Run `make compatibility-comparator-test`.
- [ ] Run `make compatibility-test`.
- [ ] Run `go test ./test/parity -count=1` from `netbox-backend`.
- [ ] Use one source digest.
- [ ] Retain per-capability reports.
- [ ] Keep failed/deferred rows visible.
- [ ] Scan artifacts and complete attestation.

Exit: the first-profile REST/gRPC boundary is fully retained, never inferred
from a partial aggregate run.

## V5 goals — Vue Workflow Parity

Browser test authoring may overlap late V4, but T4 execution and promotion wait
for the corresponding T2 and T3 capabilities.

### `CW1-V5-01` — browser authentication and sessions

Entry: `CW1-G00` is current and V1 behavior is stable.

- [ ] Login through the real browser.
- [ ] Refresh the session.
- [ ] Perform a CSRF-protected mutation.
- [ ] Logout and prove invalidation.
- [ ] Prove limited-user denial.
- [ ] Keep secrets/session/CSRF out of localStorage, sessionStorage, and
      IndexedDB.
- [ ] Use only documented cookies.
- [ ] Keep failure diagnostics credential-free.

Exit: every declared browser authentication/session outcome is authored and
passes its focused lower-boundary test. This closes authoring, not T4.

### `CW1-V5-02` — DCIM workflow and rollback

Entry: the relevant DCIM V2 behavior and REST contract are stable.

- [ ] Create Manufacturer, roles, RackType, Site, and Rack prerequisites.
- [ ] Create DeviceType and InterfaceTemplate.
- [ ] Create Device and display instantiated Interface.
- [ ] Force a late template failure and prove Device, Interfaces, and object
      changes roll back.
- [ ] Exercise Interface delete warning, cancel, and confirm.
- [ ] Assert resulting IP and object-change state.

Exit: the complete DCIM operator chain and forced-rollback scenarios are
authored and pass their focused lower-boundary tests. This closes authoring,
not T4.

### `CW1-V5-03` — IPAM edit/filter/assignment workflow

Entry: the relevant IPAM/assignment V2 behavior and REST contract are stable.

- [ ] Create, edit, list, and filter VRF, Prefix, and IPAddress.
- [ ] Assign, reassign, and unassign IPAddress.
- [ ] Force assignment rollback.
- [ ] Assert exact post-operation state.

Exit: the complete IPAM operator workflow and rollback scenarios are authored
and pass their focused lower-boundary tests. This closes authoring, not T4.

### `CW1-V5-04` — browser negative and presence outcomes

Entry: `CW1-V5-01` through `CW1-V5-03` are authored.

- [ ] Exercise explicit-null PATCH.
- [ ] Exercise validation, conflict, and not-found UX.
- [ ] Exercise protected-deletion UX.
- [ ] Verify read-side choice envelopes and scalar mutation values.
- [ ] Assert exact durable/object-change state.
- [ ] Prove the UI does not substitute for backend authorization.

Exit: every applicable negative, presence, and state outcome is authored and
passes its focused lower-boundary test. This closes authoring, not T4.

### `CW1-V5-05` — retain T4

Entry: corresponding capabilities are T2/T3 and V5 authoring is complete.

- [ ] Run `make browser-e2e` against a fresh built deployment.
- [ ] Pass both workflows and every applicable negative case.
- [ ] Record digest, toolchains, times, and totals.
- [ ] Retain credential-free artifacts.
- [ ] Link only workflows actually exercised.
- [ ] Apply the two-digest attestation procedure.

Exit: partial browser artifacts may be retained and only exercised workflows
may be promoted, but this goal remains open until both declared workflows and
every applicable required outcome pass and are retained.

## V6 goals — first-profile sign-off

### `CW1-V6-01` — evidence convergence audit

Entry: V1–V5 are retained and no untested behavior change occurred.

- [ ] Confirm every V0–V5 artifact attests the same behavioral
      `tested_digest`; rerun every gate not on that digest.
- [ ] Rerun `make check` unchanged.
- [ ] Verify every artifact's commands, toolchains, configuration, timestamps,
      totals, state/effect checks, and retention.
- [ ] Confirm the dynamic application exception remains closed.
- [ ] Confirm runtime REST routes, gRPC reflection, and Vue contain only declared
      profile/extension surfaces.
- [ ] Keep every deferred field/action visible.
- [ ] Remove any prose claiming all DCIM, IPAM, or NetBox is complete.

Exit: compatibility, security, data, UI, and operations evidence is internally
consistent.

### `CW1-V6-02` — protobuf freeze and claim attestation

Entry: `CW1-V6-01` is complete and affected capabilities have T3.

- [ ] Freeze the affected v1 protobuf capabilities.
- [ ] Establish a checked `buf breaking` baseline.
- [ ] Require additive or explicitly versioned future changes.
- [ ] Link V0–V5 from the ledger, profile, and Status.
- [ ] Review the claim-only diff manifest.
- [ ] Compute the attestation digest and run final `make check`.
- [ ] Retain a credential-free tested-to-attested digest mapping.

Exit: the final claimed revision is green and mapped to reviewed behavior.

### `CW1-V6-03` — joint human sign-off

Entry: `CW1-V6-02` is complete.

- [ ] Compatibility review signed.
- [ ] Security review signed.
- [ ] Data/transaction review signed.
- [ ] Operations/evidence review signed.
- [ ] Residual limitations, exclusions, and deferred scope accepted.
- [ ] Sign-off recorded in the evidence ledger.
- [ ] Confirm no later profile was opened early.

Exit: `core-workflow-v1` is signed off and Gate 5 becomes unblocked.

# Phase B — expand the accepted baseline

The following 19 entries are planning candidates, not accepted public
contracts. Materialize exactly one through CP-00 to CP-03 before implementation
unless a human review explicitly authorizes more than one profile in flight.

## Reusable Capability Profile checklist

- [ ] Keep `CW1-G00` green.
- [ ] Confirm every predecessor/dependency is signed off.
- [ ] CP-00: select one coherent workflow, owners, prerequisites, non-goals, and
      retirement boundary.
- [ ] Satisfy every Definition-of-Ready item in playbook section 9.
- [ ] CP-01: interrogate the pinned baseline source and upstream tests.
- [ ] CP-02: write and review machine-readable resource/action/profile
      contracts.
- [ ] CP-03: create deterministic fixtures and deliberately failing acceptance
      scenarios.
- [ ] Obtain any required human-reviewed ADR before CP-02.
- [ ] CP-04–CP-09: implement typed domain, application, PostgreSQL, exact REST,
      semantic gRPC, and applicable Vue behavior.
- [ ] CP-10: compose only declared capabilities and prove containment.
- [ ] CP-11: climb L0–L8; earn T2 before T3 and applicable T4.
- [ ] CP-12: retain credential-free evidence only for exact exercised rows,
      complete L9 tested-digest confirmation and claim attestation, and promote
      only a tier-owning resource whose full applicable owner-row boundary is
      closed.
- [ ] CP-13: only after capability completion under ADR 0004, retire only the
      exact frozen artifacts proved displaced and unreachable.
- [ ] CP-14: publish only after contract stability, support status, evidence,
      and upgrade implications are reviewed; when cross-module fields/actions
      remain deferred, keep them explicit and schedule the applicable
      module-closeout goal.
- [ ] Keep the profile pre-publication until all exit conditions pass together.

## Candidate profile queue

| Goal     | Candidate scope                                                                                                                                                                                                                                        | Workflow                                                                       | Dependencies/special decision                                                                                                                                                                                                                                                                   |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CP-P01` | `dcim-hierarchy-v1`: Region, SiteGroup, Location, Platform                                                                                                                                                                                             | Geographic/organizational hierarchy, Site/Device placement, platform selection | Signed-off first profile                                                                                                                                                                                                                                                                        |
| `CP-P02` | `dcim-modular-hardware-v1`: ModuleTypeProfile, ModuleType, ModuleBayTemplate, ModuleBay, Module, DeviceBayTemplate, DeviceBay, InventoryItemRole, InventoryItemTemplate, InventoryItem, MACAddress                                                     | Modular templates, bay occupancy, inventory hierarchy, MAC assignment          | Core DeviceType/Device/Interface                                                                                                                                                                                                                                                                |
| `CP-P03` | `dcim-ports-power-v1`: ConsolePortTemplate, ConsolePort, ConsoleServerPortTemplate, ConsoleServerPort, FrontPortTemplate, FrontPort, RearPortTemplate, RearPort, PowerPortTemplate, PowerPort, PowerOutletTemplate, PowerOutlet, PowerPanel, PowerFeed | Physical terminations and power topology without cabling                       | `CP-P01`, `CP-P02`, core Device                                                                                                                                                                                                                                                                 |
| `CP-P04` | `ipam-registry-routing-v1`: RIR, Aggregate, ASNRange, ASN, RouteTarget; available-ASN action                                                                                                                                                           | Registry/range management and safe ASN allocation                              | Core IPAM                                                                                                                                                                                                                                                                                       |
| `CP-P05` | `ipam-vlan-v1`: Role, VLANGroup, VLAN, VLANTranslationPolicy, VLANTranslationRule; available-VLAN action                                                                                                                                               | Scoped VLANs, VID allocation, translation                                      | `CP-P01` for scope promotion; tenancy may remain explicitly deferred                                                                                                                                                                                                                            |
| `CP-P06` | `ipam-address-services-v1`: IPRange, FHRPGroup, FHRPGroupAssignment, ServiceTemplate, Service; three available-IP/prefix actions                                                                                                                       | Address/prefix allocation, redundant gateways, services                        | Core Prefix/IPAddress; `CP-P04` and `CP-P05` only for registry, role, and VLAN relationships this profile promotes                                                                                                                                                                              |
| `CP-P07` | `tenancy-v1`: TenantGroup, Tenant, ContactGroup, ContactRole, Contact, ContactAssignment                                                                                                                                                               | Tenant/contact hierarchy and typed assignments                                 | Signed-off target modules for each promoted assignment; unclosed consumers and their tenant fields remain deferred                                                                                                                                                                              |
| `CP-P08` | `users-admin-v1`: Users Config, User, Group, Permission, Token                                                                                                                                                                                         | Reviewed baseline administration over the Go identity store                    | Reuse one identity/Principal/RBAC store; distinguish baseline `/api/users/tokens/` from `/api/auth/tokens/`; implement only reviewed baseline `/api/users/config/` behavior; never recreate removed unsafe public `/config` or disguise it under another route; anonymous provisioning excluded |
| `CP-P09` | `dcim-operational-v1`: RackReservation, VirtualChassis, VirtualDeviceContext; rack-elevation action                                                                                                                                                    | Rack reservations, virtual chassis/VDCs, elevation                             | Core Rack/Device; `CP-P01`, `CP-P07`, `CP-P08`                                                                                                                                                                                                                                                  |
| `CP-P10` | `virtualization-v1`: ClusterGroup, ClusterType, Cluster, VirtualMachine, VirtualDisk, virtualization Interface                                                                                                                                         | Clusters, VMs, disks/interfaces, shared IP/VLAN semantics                      | DCIM hierarchy/platform, Users/Tenancy, IPAM/VLAN, existing DeviceRole semantics                                                                                                                                                                                                                |
| `CP-P11` | `circuits-v1`: 11 provider, circuit, group, and virtual-circuit resources                                                                                                                                                                              | Physical/virtual circuits with A/Z terminations                                | Tenancy and DCIM termination owners                                                                                                                                                                                                                                                             |
| `CP-P12` | `dcim-cabling-v1`: Cable, CableTermination, ConnectedDevice; eight trace/path actions                                                                                                                                                                  | Typed physical connections and path traversal                                  | `CP-P03`, `CP-P11`; raw endpoints forbidden                                                                                                                                                                                                                                                     |
| `CP-P13` | `wireless-v1`: WirelessLANGroup, WirelessLAN, WirelessLink                                                                                                                                                                                             | WLAN organization and wireless Interface links                                 | VLAN, Tenancy, physical/virtual Interfaces                                                                                                                                                                                                                                                      |
| `CP-P14` | `vpn-v1`: 10 IKE/IPSec/Tunnel/L2VPN resources                                                                                                                                                                                                          | Ordered security policies and typed terminations                               | IPAM, DCIM, Virtualization, Circuits, Tenancy                                                                                                                                                                                                                                                   |
| `CP-P15` | `core-audit-v1`: ObjectType, ObjectChange                                                                                                                                                                                                              | Supported content-type discovery and authorized immutable change history       | All existing object-change producers                                                                                                                                                                                                                                                            |
| `CP-P16` | `core-automation-v1`: BackgroundQueue, BackgroundTask, BackgroundWorker, DataFile, DataSource, Job; five task/sync actions                                                                                                                             | Durable jobs and synchronization without Python                                | Human ADR before CP-02: state machine, leases, cancellation, retry/idempotency, extension split                                                                                                                                                                                                 |
| `CP-P17` | `extras-metadata-v1`: seven tag/custom-field/config resources; choices/render actions                                                                                                                                                                  | Typed metadata/config context and safe rendering                               | Human ADR before CP-02: closed cross-module typed-object registry plus template language, sandbox, and resource-limit contract                                                                                                                                                                  |
| `CP-P18` | `extras-productivity-v1`: seven user/content resources; dashboard action                                                                                                                                                                               | Shortcuts, annotations, images, safe export, dashboard                         | Human ADR before CP-02: attachment blob storage, integrity, retention/deletion, backup, serving policy, and safe export-rendering boundary                                                                                                                                                      |
| `CP-P19` | `extras-automation-v1`: EventRule, Webhook, NotificationGroup, Notification, Subscription                                                                                                                                                              | Authenticated delivery from committed domain events                            | Human ADR before CP-02: stable event versioning, transactional-outbox ownership, delivery guarantees, retry/replay/dead-letter behavior, and Extension Service contracts                                                                                                                        |

Exact resources, actions, discovery topics, and route shapes remain in
[playbook section 11](IMPLEMENTATION_EXECUTION_PLAYBOOK.md#11-recommended-dependency-ordered-profiles).

## Inventory reconciliation

- [x] 118 future resources are represented by the candidate queue.
- [x] 22 future custom-action entries are represented.
- [x] DCIM accounts for 35 resources and 9 actions.
- [x] IPAM accounts for 15 resources and 5 actions.
- [x] Tenancy accounts for 6 resources.
- [x] Users accounts for 5 resources; anonymous provision remains excluded.
- [x] Virtualization accounts for 6 resources.
- [x] Circuits accounts for 11 resources.
- [x] Wireless accounts for 3 resources.
- [x] VPN accounts for 10 resources.
- [x] Core accounts for 8 resources and 5 actions.
- [x] Extras accounts for 19 resources and 3 actions; Scripts remains excluded.
- [ ] Recalculate all figures from the machine inventory whenever a candidate
      becomes an accepted profile.

# Phase C — module-closeout goals

Completing candidate profiles does not automatically complete a module. Run
the closeouts below in order unless a machine-readable dependency graph proves
a safer order.

| Goal    | Module         | State     | Entry condition                                                |
| ------- | -------------- | --------- | -------------------------------------------------------------- |
| `MC-01` | DCIM           | `blocked` | All accepted DCIM and cross-module dependencies signed off     |
| `MC-02` | IPAM           | `blocked` | `MC-01` plus all IPAM dependencies signed off                  |
| `MC-03` | Tenancy        | `blocked` | `MC-02`; DCIM/IPAM target relationships stable                 |
| `MC-04` | Users/Identity | `blocked` | `MC-03`; User/RBAC/token profiles and consumers stable         |
| `MC-05` | Virtualization | `blocked` | `MC-04`; relevant DCIM/IPAM/Tenancy/Users semantics stable     |
| `MC-06` | Circuits       | `blocked` | `MC-05`; Tenancy and DCIM termination semantics stable         |
| `MC-07` | Wireless       | `blocked` | `MC-06`; VLAN/Tenancy/Interface semantics stable               |
| `MC-08` | VPN            | `blocked` | `MC-07`; IPAM/DCIM/Virtualization/Circuits/Tenancy stable      |
| `MC-09` | Core           | `blocked` | `MC-08`; all audit/job producers and consumers stable          |
| `MC-10` | Extras         | `blocked` | `MC-09`; typed object/event/template/storage boundaries stable |

For each closeout:

- [ ] Confirm every dependency profile is signed off.
- [ ] Diff every baseline route/action against every accepted profile.
- [ ] Diff every serializer field, writable flag, default, nullability state,
      brief/nested projection, filter, search/order field, and computed value.
- [ ] Promote or explicitly exclude every deferred tenant, tag, custom-field,
      journal, image, content-type, assignment, termination, platform, VRF,
      VLAN, primary-IP, config-context, and template relationship.
- [ ] Cover bulk create/update/delete/import/rename or explicitly exclude it
      from the replacement claim.
- [ ] Revalidate generic references against the typed object registry.
- [ ] Revalidate permissions and list visibility after cross-module joins.
- [ ] Revalidate delete protection, cascade, and set-null behavior across
      module boundaries.
- [ ] Revalidate object-change snapshots and durable events.
- [ ] Pass the full-module strict REST differential matrix.
- [ ] Pass corresponding gRPC semantic-parity matrices.
- [ ] Pass every declared operator workflow.
- [ ] Prove zero runtime-enabled legacy route/service overlap.
- [ ] Retire only now-displaced frozen artifacts under ADR 0004.
- [ ] Update the baseline ledger so no accidental T0 entry is hidden.
- [ ] Confirm every in-scope entry meets the canonical definition of complete.

“All routes exist,” “all tables exist,” and “all generated services compile”
do not close a module.

# Phase D — Extension Service goal

## `EXT-01` — close the Extension Service boundary

State: `blocked` for closeout until a named accepted-profile or deployment
extension inventory, or a reviewed zero-extension declaration, exists. The
architectural boundary is accepted. Individual replacement contracts remain
`candidate` until reviewed.

Entry: accepted named profile/deployment manifests, pinned source
versions/effective settings where relevant, and a complete
plugin/script/report inventory—or a reviewed declaration that the named scope
has none.

Outcome: remove Python-runtime extensibility from the deployed core; classify
every deployment-specific workflow and, where a replacement is offered, expose
an explicit, versioned, authenticated out-of-process contract without RBAC,
invariant, transaction, or persistence bypass.

`EXT-01` does not promote a baseline capability or bring Python runtime
compatibility into scope.

Only an Extension Service contract required by an accepted Capability Profile
blocks that profile and therefore the Complete In-Scope Replacement. Exhaustive
migration of excluded Python plugins, scripts, and reports does not block the
153-entry replacement claim. An offered deployment replacement, or a production
source that depends on one, must close its applicable `EXT-01` work before
PROD-7.

- [ ] Inventory each deployment-specific Python plugin, script, and report and
      assign a workflow owner.
- [ ] Classify each as unnecessary, core behavior, external automation, event
      consumer, or intentionally unsupported.
- [ ] Route anything classified as core product behavior through the Capability
      Profile factory; never close it through an Extension Service contract.
- [ ] Define a versioned authenticated REST, gRPC, and/or event contract for
      each offered replacement.
- [ ] Define Principal resolution, authorization, tenant/object visibility,
      and data minimization.
- [ ] Define idempotency, replay, delivery guarantee, ordering, and concurrency.
- [ ] Define retries, timeout, cancellation, dead letter, and manual replay.
- [ ] Define payload versions and consumer compatibility window.
- [ ] Define signing/key rotation and secret storage.
- [ ] Enforce rate, size, destination allowlist, and SSRF controls.
- [ ] Assign observability, alert, support, and failure ownership.
- [ ] Supply a consumer contract suite and sandbox fixture.
- [ ] Write required asynchronous delivery intent to a transactional outbox in
      the core transaction; perform network I/O only after commit.
- [ ] Prove extensions cannot load Python, redefine core invariants, bypass
      RBAC, or write PostgreSQL directly.
- [ ] Migrate/retest offered replacements outside the Go process.
- [ ] Document every intentionally unsupported workflow.

Zero-extension exit:

- [ ] Retain the reviewed named inventory or zero-extension declaration.
- [ ] Confirm the named scope offers no Extension Service contract and every
      discovered workflow has an explicit non-extension classification.

When one or more replacements are offered, exit evidence also requires:

- [ ] Versioned schemas and breaking-change report.
- [ ] Complete classification/migration ledger with owners.
- [ ] Real-PostgreSQL transaction/outbox atomicity, rollback, and delivery
      failure tests where asynchronous delivery is required.
- [ ] Applicable delivery-guarantee, idempotency, retry, replay, manual-replay,
      dead-letter, timeout, and cancellation tests.
- [ ] Applicable ordering and concurrency tests at the declared boundary.
- [ ] Payload-version and consumer-compatibility-window tests.
- [ ] Signing-key rotation and secret-storage evidence.
- [ ] Authentication, RBAC, visibility, redaction, signature, SSRF, rate, and
      size tests.
- [ ] Consumer contract/sandbox results.
- [ ] Alert/runbook/support-owner review and retained evidence.

# Phase E — production-readiness goals

Production engineering may proceed in parallel only when it does not
destabilize profile completion. PROD-1 through PROD-6 may become
implementation/evidence-ready earlier, but foundations alone do not close
them. PROD-7 cannot start until the Complete In-Scope Replacement and applicable
`EXT-01` work are signed off; it cannot close until PROD-1 through PROD-6 are
rerun and retained green on one frozen release candidate.

## `PROD-1` — versioned schema lifecycle

Outcome: replace disposable bootstrap with a deterministic,
operator-controlled production schema lifecycle.

- [ ] Introduce ordered immutable schema versions with checksums.
- [ ] Run migration as a deployment step separate from ordinary startup.
- [ ] Acquire a migration lock and fail on unknown, modified, partial, or
      out-of-order versions.
- [ ] Document install, upgrade, rollback/forward recovery, and observability.
- [ ] Use expand → migrate/backfill → contract for rolling changes.
- [ ] Test clean install, N-1 upgrade, failed-step recovery, concurrent
      migrator refusal, and mixed-version deployment.
- [ ] Disable `AutoMigrate` in production.
- [ ] Retain missing-table-only `AutoMigrate` solely for disposable dev/test.
- [ ] Never use `AutoMigrate` to inspect, repair, backfill, or upgrade existing
      tables.

Exit: immutable migration manifest; PostgreSQL evidence for clean install, N-1
upgrade, failed-step recovery, concurrent-migrator refusal, and mixed-version
deployment; proof that ordinary production startup performs no schema mutation,
`AutoMigrate` is disabled, checksum-invalid or partial state and schema versions
outside the application's declared compatibility range fail safely, and
explicitly supported mixed-version deployments remain operational; and a
reviewed migration/rollback runbook.

## `PROD-2` — data migration and cutover

Dependencies: PROD-1 target schema/version semantics and a stable supported
source/capability boundary. A plugin/script/report-bearing deployment also
depends on its applicable `EXT-01` classification ledger.

Outcome: move supported source data into the production schema through a
deterministic, restartable, reconcilable, and rehearsed cutover path without
silent loss or secret disclosure.

- [ ] Define a repeatable offline-capable export/import contract.
- [ ] Pin supported source versions and effective settings.
- [ ] Preflight plugins, custom fields, and content types; reject unsupported
      input before writes.
- [ ] Import in dependency order, preserving IDs or emitting an auditable map.
- [ ] Define secure password/token/session migration; emit no secrets to logs.
- [ ] Reconcile counts, keys, relationships, hierarchy, IP/network values,
      permissions, object changes, and representative REST responses.
- [ ] Make import restartable/idempotent or define exact failure cleanup.
- [ ] Rehearse abort, rollback, and final reconciliation.
- [ ] Define write freeze, delta strategy, RPO, and go/no-go authority.
- [ ] Never infer in-place upgrade safety from similar table names.

Exit: retained preflight, import, reconciliation, restart/failure, rollback,
and timed cutover-rehearsal evidence; supported-source/effective-settings
identity; stable-ID or ID-map reconciliation; credential-leak verification;
and a signed runbook.

## `PROD-3` — security hardening

Dependencies: the final surface and deployment topology, including jobs,
webhooks, files/templates, import, and Extension Services.

Outcome: establish reviewed controls for every exposed trust boundary and close
all security findings classified as release blockers.

- [ ] Threat-model browser, REST, gRPC, CLI, migration/import, jobs, webhooks,
      files/templates, and Extension Services.
- [ ] Enforce TLS on public boundaries and mTLS where required.
- [ ] Use external secret management, rotation, and least-privilege identities.
- [ ] Define password, session, API-token, signing-key, certificate, and
      recovery lifecycles.
- [ ] Enforce request/header/body/upload limits, timeouts, secure headers,
      origin policy, and rate limits.
- [ ] Fuzz decoders, filters, authentication, `net/netip`, archives, URLs, and
      templates.
- [ ] Scan dependencies, images, source, and infrastructure; assign owners and
      deadlines to findings.
- [ ] Complete an independent security review and close every release blocker.

Exit: approved threat model, control/fuzz/scan evidence, TLS/secret-rotation
proof, a triage ledger with owner/deadline for every nonblocking finding, and
independent-review proof that every release blocker is closed.

## `PROD-4` — operability

Outcome: make the production service observable, supportable, safely probed,
and recoverable by named responders under declared SLOs.

- [ ] Define SLIs/SLOs for availability, latency, errors, DB saturation, queue
      age, delivery failures, and data freshness.
- [ ] Emit structured credential-free logs with request/trace correlation.
- [ ] Emit metrics/traces at application, PostgreSQL, queue, and extension
      boundaries without secret/high-cardinality labels.
- [ ] Export append-only audit data with retention/access policy.
- [ ] Implement truthful liveness, readiness, and startup probes.
- [ ] Implement graceful shutdown, draining, bounded work, and resource limits.
- [ ] Create alerts with named responders and actionable runbooks.
- [ ] Write runbooks for startup, migration, rollback, secret compromise,
      revocation, DB degradation, queue recovery, failed webhook, restore, and
      capacity exhaustion.

Exit: reviewed SLOs/dashboards, telemetry/secret checks, append-only audit-export
retention/access proof, probe and shutdown tests, named alert/respondership
ownership, and runbook drill records.

## `PROD-5` — reliability and performance

Dependencies: representative data, PROD-4 instrumentation, a production-like
deployment, and applicable queue/extension contracts.

Outcome: prove declared performance, concurrency, resilience, and recovery
budgets on representative and pathological workloads.

- [ ] Build representative production-sized and pathological datasets.
- [ ] Set per-interface latency, throughput, concurrency, and payload budgets.
- [ ] Add query-count/no-N+1 assertions for critical list/nested views.
- [ ] Load-test reads, mutations, filters, allocation, authentication, and
      browser workflows.
- [ ] Run long soak and burst tests with leak/resource monitoring.
- [ ] Inject PostgreSQL disconnect/failover, slow queries, worker death,
      extension timeout, disk pressure, and network partition.
- [ ] Prove retries/idempotency do not duplicate mutations or events.
- [ ] Define and test the backup schedule, restore verification, RPO, RTO, and
      DR failover.

Exit: exact representative/pathological dataset descriptors; query-count and
no-N+1 results; explicit pass/fail against declared per-interface budgets; and
retained load/soak/burst, fault-injection, duplicate-prevention,
backup/restore, and DR evidence measured against every declared RPO/RTO.

## `PROD-6` — supply chain and deployment

Dependencies: PROD-1 migration order, PROD-3 controls, PROD-4 probes/runbooks,
and PROD-5 capacity findings.

Outcome: ship reproducible, verifiable, least-privilege artifacts through a
rehearsed deployment and rollback process with tested capacity controls.

- [ ] Produce reproducible builds from pinned toolchains and lockfiles.
- [ ] Use minimal non-root images, read-only filesystems, and dropped
      capabilities where practical.
- [ ] Produce SBOM, provenance, signatures, vulnerability policy, and license
      policy.
- [ ] Keep compiler/package-manager credentials out of final images.
- [ ] Define configuration schema, safe defaults, and external secret
      references.
- [ ] Rehearse staging and canary/rolling deployment, compatibility-aware
      migration order, and automated rollback.
- [ ] Run production-image smoke in connected CI.
- [ ] Document/test resource requests, autoscaling constraints, topology, and
      disruption budgets.

Exit: matching independently reproduced artifact/image content digests; separate
SBOM, provenance, and signature verification; hardened final-image credential
inspection; policy reports; configuration validation; connected-CI
production-image smoke; rollout/rollback evidence; and tested resource-request,
autoscaling, topology, and disruption-budget evidence.

## `PROD-7` — release and cutover

Outcome: execute a controlled, reversible, monitored production cutover from
one frozen and fully attested release candidate.

Entry: the Complete In-Scope Replacement and applicable `EXT-01` work are
signed off; PROD-1 through PROD-6 exit suites and controls are ready; and no
unresolved release blocker remains.

- [ ] Freeze one release candidate. Any blocker fix creates a new candidate
      digest, invalidates prior PROD-1–PROD-6 and aggregate evidence, and
      restarts the freeze-and-rerun sequence.
- [ ] Rerun and retain PROD-1 through PROD-6 on that exact candidate before
      cutover.
- [ ] Run repository, security, domain, PostgreSQL, REST compatibility, gRPC,
      deployment, browser, performance, supply-chain, and operations suites on
      that candidate.
- [ ] Check every published API version for accidental breakage.
- [ ] Complete backup and rehearse the exact migration/cutover.
- [ ] Perform shadow/differential validation where feasible.
- [ ] Hold a signed pre-cutover go/no-go and rollback-authority checkpoint.
- [ ] Execute write freeze, export/import, reconciliation, smoke, and
      limited-user workflow.
- [ ] Hold a signed post-reconciliation continue-or-rollback checkpoint.
- [ ] Monitor the post-cutover window with named support ownership.
- [ ] Publish versioned release notes with supported profiles and exclusions.

Exit: tested source digest; claim attestation digest; exact signed image/artifact
digest; verified provenance mapping source, toolchain, and SBOM to that
artifact; proof that deployment and cutover smoke used the same artifact
digest; aggregate gate manifest; breaking-change report; signed go/no-go and
rollback-decision record; signed cutover/reconciliation record; monitored
post-cutover report; and published support/version policy.

# Program completion checklists

## Complete In-Scope Replacement

- [ ] All 131 in-scope resources are covered by accepted profiles and required
      evidence.
- [ ] All 22 in-scope custom actions are covered.
- [ ] Every in-scope baseline REST capability is T2.
- [ ] Every corresponding gRPC capability is T3.
- [ ] Every applicable declared browser workflow is T4.
- [ ] `MC-01` through `MC-10` are complete.
- [ ] Extensions/divergences retain separate contract/parity/security status.
- [ ] Every Extension Service contract required by an accepted profile is
      complete; every intentionally unsupported workflow in each named
      `EXT-01` inventory remains explicit.
- [ ] Frozen runtime/source ownership is gone wherever displaced.
- [ ] Runtime REST equals accepted profiles plus named
      health/readiness/schema/identity surfaces and has no dedicated
      `GET /ping`.
- [ ] gRPC reflection lists only canonical published services and health.
- [ ] Vue exposes only declared workflows.
- [ ] Explicit exclusions and unsupported workflows in named inventories remain
      visible.
- [ ] No documentation claims compatibility for Python
      runtime/plugins/scripts/reports, GraphQL, or pixel parity.
- [ ] Joint compatibility, security, and capability-owner sign-off is retained.

## Production release

- [ ] Complete In-Scope Replacement is signed off.
- [ ] PROD-1 through PROD-7 are retained green on the release candidate.
- [ ] Upgrade/import, backup/restore, rollback, DR, security, performance, and
      operations are rehearsed.
- [ ] Support ownership, SLOs, runbooks, API/version policy, and exclusions are
      published.
- [ ] Data-migration, operations, security, and release owners sign off.
- [ ] Gate 7 has a reviewed go/no-go decision.

Until the production checklist closes, the artifact is not approved as
production-ready; after cutover and before closeout it remains inside the
controlled acceptance/rollback window. Missing-table-only `AutoMigrate` remains
a disposable-development convenience and is never a production migration
strategy.

## Implementation-agent handoff template

Use the [playbook section 17 template and required executor preamble](IMPLEMENTATION_EXECUTION_PLAYBOOK.md#17-increment-specification-for-gpt-56-luna)
verbatim. Add the roadmap goal identity and tested/attestation fields shown
below; do not give an implementation agent an unbounded “implement the next
module” prompt.

Completed bounded packets are indexed under
[implementation handoffs](increments/README.md).

```text
Goal ID:
Owner:
State:
Increment ID:
Profile and capability:
Outcome:

Authoritative inputs:
- Capability Profile/resource metadata:
- Baseline source locations:
- Canonical ADR/standards sections:
- Existing tests/evidence:

Hard dependencies:
Entry evidence:
- Required previous gate:
- Current failing test/diagnostic:

Permitted files/packages:
-

Forbidden scope:
- No undeclared route/field/action.
- No generated-file edit.
- No new dependency/toolchain/exclusion/normalizer unless explicitly listed.
- No legacy deletion before capability completion.

Required implementation:
1.
2.

Required scenarios:
- Lowest-layer positive regression:
- Validation:
- Authorization/visibility:
- Not found/conflict:
- PATCH presence:
- Rollback/concurrency and durable state:
- State/object changes:
- REST/gRPC equivalence:
- Vue/browser, if applicable:

Commands, in order:
1. Focused L0/L1:
2. Affected L2:
3. Broader required gates:

Exit conditions:
- All listed commands green.
- No broader behavior change.
- Documentation/evidence updated only to the proved tier.

Completion report:
- Changed files and why:
- Exact command outcomes:
- Tested source digest:
- Attestation digest, if claims changed:
- Claim-only diff file/content-hash manifest, if attested:
- Human reviewer/approval, if attested:
- Durable evidence links:
- Skipped external gates and why:
- Residual risk / next goal:
```

An increment may close one subgoal or a coherent portion of it. It may not mark
a parent goal, tier, module, Complete In-Scope Replacement, or production
release complete without that boundary's retained exit evidence.
