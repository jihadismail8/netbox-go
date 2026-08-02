# Compatibility roadmap

This roadmap is gate-based. A gate closes when its current evidence is retained
and reviewed, not when a number of files, routes, or tests exists.

The target is a standalone Go backend and Vue frontend for the pinned NetBox
snapshot `fbb948d30e79ce657fac62994a22aca72c1770a9`
(`v4.4.6-7-gfbb948d30`). REST is exact in each declared profile; gRPC has
semantic parity through the same core. See [Compatibility](COMPATIBILITY.md)
for tiers and [Status](STATUS.md) for the dated audit.

## Delivery rules

- Deliver coherent [Capability Profiles](../CONTEXT.md), not generated-file
  counts or vague module percentages.
- REST and gRPC adapters share validation, authorization, transactions,
  persistence orchestration, state, and side effects.
- Vue uses REST and preserves declared workflows; it is not a second business
  implementation.
- The development database is disposable. `AutoMigrate` creates absent tables
  only; it never repairs, backfills, or alters an existing table shape.
- Python/Django is allowed only inside the pinned differential oracle job.
- A capability remains at its lower tier until the stronger boundary has a
  durable artifact.

## Gate status

| Gate | Outcome                            | Implementation state                                                                          | Evidence state                                                  |
| ---- | ---------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| 0    | Stable language and boundary       | Accepted in CONTEXT and ADRs                                                                  | Accepted                                                        |
| 1    | Inventory and shared-core boundary | Implemented for first profile                                                                 | Current full-gate artifact pending                              |
| 2    | Unified identity/RBAC              | Persisted groups, memberships, global/object grants, sessions/tokens, shared Principal        | Current security/parity artifact pending                        |
| 3    | Strict compatibility harness       | Fail-closed oracle/comparator and gRPC parity suites implemented                              | Post-hardening execution pending                                |
| 4    | First Core Workflow Profile        | 13 resources implemented with typed persistence/Vue adapters; displaced legacy stacks retired | T1, pre-publication; PostgreSQL/oracle/browser evidence pending |
| 5    | Module-by-module expansion         | Not started as a compatibility claim                                                          | Blocked on Gate 4 sign-off                                      |
| 6    | Extension-service contracts        | Architectural boundary accepted                                                               | Future profile                                                  |
| 7    | Production hardening/release       | Development conveniences only                                                                 | Not started                                                     |

## Gate 0 — decision baseline

Accepted decisions:

- the exact checked-in post-release commit is the Compatibility Baseline;
- the deployed runtime has no Python, Django, or upstream-source dependency;
- HTTPS REST and gRPC are both first-class public interfaces;
- REST preserves the baseline contract inside a reviewed profile;
- gRPC preserves the same domain meaning and outcomes;
- authentication fails closed and authorization is shared;
- Vue preserves supported workflows, not upstream pixels;
- GraphQL and Python plugins, scripts, and reports are out of scope; and
- future integrations are out-of-process Extension Services.

Evidence: [CONTEXT.md](../CONTEXT.md), [Architecture](ARCHITECTURE.md), and the
accepted [ADRs](adr/README.md).

## Gate 1 — inventory and shared-core boundary

Implemented:

- authoritative generated inventories: baseline REST 155; current REST 123
  (102 frozen + 13 canonical + 8 identity); current gRPC 179 (176 frozen + 3
  canonical); Vue 13;
- one shared application path for the 13 profile resources;
- exact-in-profile REST metadata/OpenAPI and canonical versioned protobufs;
- transport-neutral error and Principal flow; and
- runtime containment of deferred legacy REST/gRPC source.

Remaining exit work:

- retain the current repository/contract gate result; and
- replace the remaining dynamic application/domain maps with typed capability
  contracts. Typed per-table PostgreSQL rows and typed Vue adapters are already
  in place.

## Gate 2 — unified identity and authorization

Implemented:

- Go-owned users, groups, memberships, direct/group grants, object-scoped
  grants, sessions, and tokens;
- session/CSRF browser flow, NetBox-style REST token credentials, and gRPC
  bearer metadata resolving to one Principal;
- application-owned view/add/change/delete authorization and object visibility;
- one-time administrator bootstrap, password reset, non-superuser creation,
  and global model-permission grants through the protected CLI;
  and
- focused RBAC, token, session, REST, gRPC, and CLI tests.

Remaining exit work:

- retain a complete current security matrix covering missing/invalid
  credentials, global/group/object grants, visibility-before-pagination,
  token restrictions/order, CSRF/CORS/throttling, session rotation/invalidation,
  and CLI behavior through both public interfaces.

## Gate 3 — differential and parity proof

Implemented:

- isolated pinned-oracle and standalone-Go databases;
- SHA and effective-configuration refusal checks;
- deterministic first-profile scenarios;
- explicit normalizers with fail-closed treatment of statuses, validation,
  authorization, fields, paths, state, and side effects;
- comparator sensitivity test; and
- gRPC lifecycle, errors, permission, rollback, and assignment parity suites.

Remaining exit work:

- run the post-hardening strict oracle job and retain its report;
- retain the equivalent gRPC report after REST earns T2; and
- link per-capability results without promoting unexercised behavior.

## Gate 4 — Core Workflow Profile

The first profile contains:

```text
Manufacturer -> DeviceType -> InterfaceTemplate
                                      |
Site -> Rack -> Device ---------------+-> Interface

VRF -> Prefix -> IPAddress -> assign to Interface -> unassign
```

Supporting resources RackRole and DeviceRole are included. The profile exposes
list/get/create/replace/update/delete for all 13 resources plus assignment
semantics. Bulk operations, rack elevation, automatic allocation, cabling,
GraphQL, and Python extension execution remain deferred.

Implemented structure:

- canonical REST and gRPC adapters over one service;
- 13 private typed PostgreSQL tables and a typed object-change table;
- a 198-table missing-only startup registry (176 legacy + 8 identity + 13
  profile + 1 change);
- typed Vue DTO/form/filter adapters and profile-only routes;
- persisted RBAC and assignment/rollback behavior;
- strict oracle, PostgreSQL/concurrency, deployment, and real-browser harnesses;
  and
- physical removal of the 13 displaced legacy stacks.

Exit gate:

- repository and contract checks retained green;
- every declared REST capability reaches T2 against the pinned oracle;
- equivalent gRPC capabilities reach T3 through the shared core;
- identity extensions pass their separate contract/parity/security matrix;
- real PostgreSQL and deployment evidence is retained;
- both browser workflows and required negative cases reach T4;
- the dynamic application-boundary exception is closed; and
- the evidence-linked sign-off keeps omitted behavior visibly deferred.

Current status: **T1, pre-publication**. Harnesses exist, but the required latest
external executions are pending. Neither DCIM nor IPAM is complete.

## Gate 5 — module-by-module expansion

After Gate 4 closes, choose the next coherent workflow from the deferred
inventory. For each new profile:

1. catalogue exact operations, fields, filters, actions, permissions, and
   side effects;
2. implement typed domain/application/persistence boundaries;
3. reach T2 through strict REST differential tests;
4. reach T3 through equivalent gRPC scenarios;
5. reach T4 for declared browser workflows;
6. physically retire only the displaced legacy artifacts; and
7. keep all earlier profiles green.

Likely dependency order is remaining DCIM, remaining IPAM, Tenancy, remaining
Users, Virtualization, Circuits, Wireless, VPN, Core, and Extras. Cross-cutting
capabilities may be pulled forward without declaring their whole module
complete.

The
[execution playbook](IMPLEMENTATION_EXECUTION_PLAYBOOK.md#11-recommended-dependency-ordered-profiles)
decomposes that coarse order into 19 review-sized profile candidates derived
from the 155-entry baseline inventory, followed by explicit
[module-closeout passes](IMPLEMENTATION_EXECUTION_PLAYBOOK.md#12-module-closeout-passes).
Those candidates become public scope only after the profile-ready discovery
and review steps; frozen generated files never define the queue.

## Gate 6 — Extension Service contracts

Version authenticated REST/gRPC/event contracts for out-of-process
integrations. Define delivery, retry, idempotency, ordering, and failure
behavior. Extensions may observe and request authorized operations but may not
load Python into the core or write PostgreSQL directly.

## Gate 7 — production hardening and release

Before production, define and prove:

- schema versions, install/upgrade, backup/restore, and disaster recovery;
- TLS, secret rotation, session/token lifecycle, rate limits, and security
  review;
- metrics, tracing, audit output, alerts, and operational ownership;
- isolation, performance, soak, and failure-injection behavior; and
- deployment, rollback, API versioning, and support procedures.

Until this gate closes, database resets and missing-table-only `AutoMigrate`
remain development conveniences, not a production migration strategy.
