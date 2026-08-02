# Whole-project implementation execution playbook

- Status: canonical operational companion
- Audience: implementation agents (primary handoff: GPT-5.6 Luna), reviewers,
  and release operators
- Compatibility Baseline:
  `fbb948d30e79ce657fac62994a22aca72c1770a9`
  (`v4.4.6-7-gfbb948d30`)
- Updated: 2026-08-01

This playbook turns the accepted architecture, compatibility contract, coding
standards, and roadmap into an executable sequence. It does not replace those
documents or make new product decisions. Its immediate purpose is to recover
the partially completed first Capability Profile without guessing; its
long-term purpose is to give each later profile the same repeatable path from
discovery to retained evidence.

The executor must treat every unchecked item as work unless a dated execution
status below explicitly records that the procedure has been performed. A
completed structural step is still not retained compatibility-tier evidence.

## Navigation

1. [Authority and conflict resolution](#1-authority-and-conflict-resolution)
2. [Mission and end states](#2-mission-and-end-states)
3. [Execution philosophy](#3-execution-philosophy)
4. [Executor operating contract](#4-executor-operating-contract)
5. [Stable project rule catalogue](#5-stable-project-rule-catalogue)
6. [Historical handoff snapshot](#6-historical-handoff-snapshot)
7. [Immediate recovery sequence](#7-immediate-recovery-sequence)
8. [First-profile V1-V6 completion](#8-first-profile-completion-plan-v1-through-v6)
9. [Later-profile definition of ready](#9-definition-of-ready-for-a-later-capability-profile)
10. [Capability Profile factory](#10-repeatable-capability-profile-factory)
11. [Dependency-ordered profiles](#11-recommended-dependency-ordered-profiles)
12. [Module-closeout passes](#12-module-closeout-passes)
13. [Extension Service gate](#13-extension-service-gate)
14. [Production readiness](#14-production-readiness-program)
15. [Verification command ladder](#15-verification-command-ladder)
16. [Evidence procedure](#16-evidence-procedure)
17. [GPT-5.6 Luna increment specification](#17-increment-specification-for-gpt-56-luna)
18. [Definition of done](#18-definition-of-done)
19. [Forbidden shortcuts](#19-forbidden-shortcuts-and-hard-stop-review-findings)
20. [Implementation blueprints](#20-implementation-blueprints)

## 1. Authority and conflict resolution

Read these sources before changing code:

1. [Project language](../CONTEXT.md) defines the project's vocabulary and
   boundary.
2. [Accepted ADRs](adr/README.md) govern hard-to-reverse decisions.
3. [Compatibility](COMPATIBILITY.md) governs behavioral proof and promotion.
4. The machine-readable
   [Capability Profile](../contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml)
   governs the exact declared first-profile surface.
5. [Architecture](ARCHITECTURE.md) and
   [Coding standards](CODING_STANDARDS.md) govern dependency direction and
   implementation rules.
6. [Testing](TESTING.md) and [Evidence](evidence/README.md) govern gates and
   durable results.
7. [Project status](STATUS.md) governs what may currently be claimed.
8. [Implementation plan](IMPLEMENTATION_PLAN.md) governs the first profile's
   accepted behavior and sequence.
9. [Roadmap](ROADMAP.md) governs delivery gates after the first profile.
10. This playbook governs day-to-day execution order and handoff format.

If two sources conflict:

- an accepted ADR wins over implementation precedent;
- the Capability Profile wins for declared surface;
- Compatibility wins for evidence and tier claims;
- the checked-in upstream source and strict oracle decide baseline behavior;
- generated code, tests written only against this implementation, comments,
  artifact counts, and prior agent summaries never override those authorities;
- stop and record a real unresolved hard-to-reverse decision rather than
  silently choosing a new architecture.

Two existing contradictions have settled operational interpretations:

1. ADR 0004 allows displaced legacy artifacts to be retired only after their
   replacement satisfies capability completion. The first-profile legacy
   stacks were historically removed while the profile remained T1. Do not
   restore them merely to recreate the old ordering, but do not delete another
   displaced stack until the corresponding replacement has passed its
   completion gate or the ADR is explicitly amended.
2. Secret material is never logged or retained in evidence. A hardened session
   handle is emitted only through the documented HttpOnly `Set-Cookie` flow;
   the CSRF bootstrap may emit its documented non-HttpOnly SameSite cookie and
   response-body value; and a generated API-token secret may be returned only
   once in its documented creation response. Password hashes, reusable token
   material after creation, DSNs, and full configuration are never serialized.

Baseline compatibility tiers apply to baseline capabilities. Secure identity
extensions cannot earn REST T2 against an upstream operation they intentionally
replace; they use a separate contract, shared-core parity, and security gate.

## 2. Mission and end states

The project has three nested outcomes:

### Outcome A — close the first Core Workflow Profile

- [ ] Restore and retain a green V0 repository gate.
- [x] Close the dynamic/raw-map application-boundary exception.
- [ ] Retain the V1 identity and authorization matrix.
- [ ] Prove every declared first-profile invariant and rollback path at V2.
- [ ] Retain real-PostgreSQL and standalone deployment evidence at V3.
- [ ] Earn REST T2 and then gRPC T3 for all 13 baseline resources and declared
      assignment behavior at V4.
- [ ] Earn Vue workflow T4 at V5.
- [ ] Review all V0-V5 artifacts together and record V6 sign-off.

### Outcome B — cover the reviewed baseline profiles

- [x] Retain and revalidate the existing 155-entry
      [baseline inventory](../contracts/netbox/v4.4.6-post7/inventory/baseline-rest.yaml);
      recatalogue only when the pinned baseline changes by accepted decision.
- [ ] Group them into dependency-correct, reviewable Capability Profiles.
- [ ] Give every promoted baseline capability exact REST behavior, same-core
      gRPC semantics, PostgreSQL correctness, shared RBAC, and applicable Vue
      Workflow Parity.
- [ ] Keep every omitted operation visibly T0/deferred.
- [ ] Retire displaced frozen scaffolding only after the governing capability
      completion gate.
- [ ] Never call an entire module complete from route or file counts.

### Outcome C — production readiness

- [ ] Replace disposable schema handling with versioned install/upgrade and
      rollback procedures.
- [ ] Prove backup/restore, security, TLS, observability, performance, failure
      handling, deployment, rollback, and support operations.
- [ ] Publish only reviewed interface versions and evidence-backed support
      claims.
- [ ] Complete Gate 7 before representing the system as production-ready.

The rewrite is not a file-for-file Python-to-Go translation. It is a
standalone Go/Vue replacement of observable behavior. GraphQL and in-process
Python plugins, scripts, and reports remain outside the accepted runtime
boundary. Future extensibility uses authenticated out-of-process Extension
Services.

## 3. Execution philosophy

1. **Behavior before structure.** A type, route, table, component, or generated
   descriptor is not the capability. Accepted/rejected transitions and durable
   effects are.
2. **Profile before implementation.** Scope must be declared before code can
   earn a compatibility claim.
3. **Vertical workflow slices.** Implement the smallest useful workflow across
   domain, application, persistence, REST, gRPC, Vue, and evidence. Avoid
   horizontal layers that leave dozens of unproved CRUD stubs.
4. **One semantic center.** REST and gRPC translate into the same typed
   application use case. Vue consumes REST. No adapter becomes a second
   business implementation.
5. **Typed center, translation edges.** REST DTOs, protobuf messages, and
   PostgreSQL rows stop at adapters. Domain and application contracts describe
   domain meaning and field presence explicitly.
6. **Fail closed.** Unknown credentials, permissions, filters, fields, enum
   values, content types, and undeclared routes fail rather than broaden scope.
7. **Transactionally honest.** Authorization, validation, state transition,
   persistence, derived state, and required object changes succeed or roll back
   together.
8. **Evidence before promotion.** A harness proves nothing until the required
   current external boundary passes and its credential-free result is retained.
9. **Preserve surprising baseline behavior.** Do not “clean up” upstream
   semantics inside the exact REST surface. Lock surprising behavior with an
   oracle scenario.
10. **Secure divergences are explicit.** A security extension or deliberate
    incompatibility is named, classified, and tested separately; it is never
    normalized into a T2 pass.
11. **Small, reversible increments.** Each increment has one declared outcome,
    bounded files, a focused test ladder, and a clean rollback boundary.
12. **Conservative language.** Say “implemented,” “exercised,” and “retained”
    precisely. Say “complete” only at the Capability Profile boundary with all
    required evidence.

## 4. Executor operating contract

GPT-5.6 Luna, or any replacement executor, must follow this contract for every
increment.

### Before editing

- [ ] Read this playbook and the exact linked canonical sections.
- [ ] Read the entire target files and their nearest tests; do not patch from a
      search snippet alone.
- [ ] Inspect repository-local instructions and generators.
- [ ] Capture the current failing command and exact diagnostics.
- [ ] Identify user-owned or unrelated changes and preserve them.
- [ ] Identify generated files before editing; change their owned source only.
- [ ] State the increment's capability, permitted files, non-goals, tests, and
      exit condition.
- [ ] Confirm V0 status. If V0 is red, do recovery work only.

### While editing

- [ ] Keep one coherent increment in progress.
- [ ] Make the smallest complete change that fixes the stated failure or
      advances the stated capability.
- [ ] Add the lowest-layer regression before or with the behavior.
- [ ] Do not weaken assertions, comparators, normalizers, lint, coverage,
      authorization, schema checks, or generated-drift checks to make a gate
      green.
- [ ] Do not add a skip, exclusion, broad `nolint`, type escape, or
      compatibility normalization without the exact rule, path/symbol, reason,
      owner, and removal milestone.
- [ ] Do not change pinned toolchains or dependencies merely because the local
      machine has a different version.
- [ ] Do not mix unrelated formatting, renaming, cleanup, dependency updates,
      or new capabilities into the increment.
- [ ] Do not delete legacy code before ADR 0004's completion condition.
- [ ] Re-run a focused test after each logical edit, then climb the gate ladder.

### Before reporting completion

- [ ] Review the diff or, when Git metadata is unavailable, review the complete
      changed files and `make source-digest`.
- [ ] Run every gate required by the change's layer and risk.
- [ ] Confirm generated outputs are deterministic and documentation links pass.
- [ ] Scan logs/artifacts for credentials and environment-specific data.
- [ ] Update profile/status/evidence only to the highest retained boundary.
- [ ] List commands exactly as run, outcomes, skipped external gates, residual
      risks, and the next unblocked increment.
- [ ] Never report success while a required command is failing.

### Ask or stop only when

- the baseline source and oracle still leave two materially different public
  behaviors possible;
- a hard-to-reverse contract, security, data-loss, deployment, or ownership
  decision is absent from ADRs;
- completion requires credentials, network access, a protected external
  service, or destructive action outside granted scope;
- user-owned changes overlap the same lines and cannot be safely preserved;
- the requested change would require broadening a Capability Profile;
- a new exception would be necessary; or
- a test can pass only by weakening a governing check.

Do not ask design questions whose answers are already in the profile,
standards, upstream source, tests, or accepted ADRs.

## 5. Stable project rule catalogue

These IDs make review comments and executor handoffs unambiguous. The linked
canonical documents remain authoritative if this summary ever drifts.

### Governance rules

- **GOV-001 — Fixed oracle.** Baseline behavior means the checked-in source at
  `fbb948d30e79ce657fac62994a22aca72c1770a9`, never a tag, latest release, or
  remembered NetBox behavior.
- **GOV-002 — Standalone runtime.** Python/Django and the upstream tree may be
  used only for pinned-source discovery and the differential development
  oracle, never in the delivered build, migration, startup, deployment, or
  operations.
- **GOV-003 — Source-of-truth scope.** The product manages intended state; it
  does not silently acquire monitoring or device-control responsibilities.
- **GOV-004 — Profile-bounded work.** An operation, field, relationship,
  filter, action, or workflow is public only when declared in a reviewed
  Capability Profile.
- **GOV-005 — Accurate claims.** Artifact counts and implementation presence do
  not imply compatibility. Status may claim only retained evidence.
- **GOV-006 — Canonical language.** Use Managed Object, Compatibility Baseline,
  Standalone Runtime, Interface Parity, Workflow Parity, Capability Profile,
  and Extension Service. Qualify technical uses of “model.”
- **GOV-007 — Same-increment documentation.** Contract, profile, generated
  docs, tests, and status affected by a behavior change move together.
- **GOV-008 — Narrow increments.** One increment has one reviewable outcome and
  no unrelated refactoring.
- **GOV-009 — Preserve the workspace.** Never discard, overwrite, reformat, or
  commit unrelated user changes.
- **GOV-010 — Exceptions are debts.** Every exception names rule, exact
  path/symbol, reason, owner, and removal milestone. Unowned or project-wide
  exceptions are prohibited.
- **GOV-011 — ADR threshold.** Record durable cross-cutting or hard-to-reverse
  decisions in an ADR; keep reversible execution details in this playbook.
- **GOV-012 — No feature work on red V0.** Repair and retain `make check`
  before beginning a new capability.
- **GOV-013 — No fabricated certainty.** If behavior is unknown, inspect the
  pinned source and add an oracle scenario; do not infer from names or generic
  NetBox knowledge.

### Architecture rules

- **ARCH-001 — Dependency direction.** Dependencies point adapter →
  application → domain.
- **ARCH-002 — Pure domain.** Domain imports no Gin, protobuf, GORM, SQL
  driver, generated API type, global config, or global database package.
- **ARCH-003 — Pure application.** Application imports no transport or storage
  implementation.
- **ARCH-004 — Thin adapters.** REST/gRPC adapters authenticate, decode,
  perform transport-shape checks, invoke a use case, and encode. They do not
  own authorization policy, semantic validation, transactions, or queries.
- **ARCH-005 — Consumer-owned ports.** Narrow repository/integration
  interfaces live beside the use case that consumes them. Generic CRUD
  repositories are prohibited.
- **ARCH-006 — Constructor injection.** No new mutable package globals,
  service locators, hidden `init()` registration, or singleton DB/cache use.
- **ARCH-007 — Precise packages.** Do not create `util`, `common`,
  `helper`, or equivalent dumping grounds.
- **ARCH-008 — One use case.** One externally visible operation maps to one
  transport-neutral application use case and one transaction boundary.
- **ARCH-009 — No transport chaining.** REST never calls gRPC and gRPC never
  calls REST.
- **ARCH-010 — Modular monolith.** The default deployment is one Go process
  serving REST and gRPC; do not split services without an accepted decision.
- **ARCH-011 — Runtime containment.** Deferred direct-GORM REST and table-shaped
  gRPC scaffolding remain frozen, unpublished, and runtime-disabled.
- **ARCH-012 — No dynamic-core regression.** The retired
  `internal/domain/workflow`, `internal/application/workflow`, and
  `internal/adapters/postgres/workflow` packages must not be recreated or
  imported. Raw Managed Object application contracts, generic workflow ports,
  and transitional generic-service constructors are equally prohibited.

### Compatibility rules

- **COMP-001 — Exact REST.** Preserve declared paths, methods, trailing slashes,
  content types, auth, permission results, status codes, fields, defaults,
  nullability, choice envelopes, validation reasons, pagination, filters,
  ordering, actions, transactions, state, and side effects.
- **COMP-002 — Presence matters.** Create, PUT, and PATCH use distinct commands;
  PATCH preserves omitted, explicit-null, zero, empty, and concrete values when
  the public contract distinguishes them.
- **COMP-003 — Strict normalizers.** Only committed rules may normalize origins,
  generated IDs, and volatile timestamps. Never normalize status, validation
  reason, authorization, missing/extra fields, state, or side effects.
- **COMP-004 — Verified oracle configuration.** Differential jobs refuse the
  wrong SHA, database, timezone, authentication policy, plugin state, fixture,
  or effective configuration.
- **COMP-005 — Bound identifiers.** Scenario-generated IDs bind to named
  objects. An unbound or ambiguously rebound ID is a comparator failure.
- **COMP-006 — Tier order.** T0 catalogue → T1 implementation/scaffolding → T2
  strict REST differential → T3 same-core gRPC semantics → T4 applicable
  browser workflow.
- **COMP-007 — Extensions stay separate.** An extension is explicitly
  classified and tested against its contract/security matrix, never counted as
  upstream T2.
- **COMP-008 — Deferred means rejected or absent.** An omitted field or action
  is not silently accepted, ignored, or partially persisted.
- **COMP-009 — Earlier profiles stay green.** A later profile cannot advance by
  regressing an already retained profile.

### Security rules

- **SEC-001 — Authenticate by default.** Every Managed Object read and mutation
  requires authentication unless a future profile explicitly enables a narrow
  exception; anonymous mutation is forbidden.
- **SEC-002 — Narrow public endpoints.** Only health, readiness, login, and CSRF
  entry points may be public, and they expose no Managed Object data.
- **SEC-003 — One Principal.** Session cookies, REST API tokens, and gRPC bearer
  metadata resolve through the same user, group, restriction, Principal, and
  RBAC implementation.
- **SEC-004 — Authorization in application.** Adapters authenticate; typed use
  cases enforce view/add/change/delete and object constraints.
- **SEC-005 — Visibility before pagination.** Authorization filters the set
  before count, ordering, limit, and offset.
- **SEC-006 — Browser credentials.** Use hardened HttpOnly/SameSite session
  cookies and CSRF for state-changing cookie requests. Never put credentials,
  permissions, CSRF, or session state in `localStorage`.
- **SEC-007 — Secret disclosure.** Never log a credential or retain it in
  evidence. Emit only the documented hardened session cookie, CSRF bootstrap
  cookie/body value, and one-time token-creation secret; never return hashes or
  reusable secrets afterward.
- **SEC-008 — Safe administration.** Bootstrap, recovery, local user creation,
  and permission grants are protected CLI operations with passwords on
  protected stdin, never anonymous network endpoints.
- **SEC-009 — CORS and throttling.** Credentialed CORS has an explicit origin
  allowlist. Login and token creation use bounded throttling without logging
  inputs.
- **SEC-010 — Safe errors.** Public errors never expose SQL, GORM, filesystem,
  stack, secret, or internal configuration text.
- **SEC-011 — TLS production boundary.** Production REST uses HTTPS and gRPC
  uses a secured channel or trusted TLS-terminating ingress.
- **SEC-012 — Extension isolation.** Extension Services authenticate, invoke
  authorized public/application operations, and never write PostgreSQL
  directly.

### Data and transaction rules

- **DATA-001 — Typed identity.** Core IDs are typed signed `int64`; zero does
  not represent database null.
- **DATA-002 — Explicit nullability.** Relationships and optional values use
  explicit presence types, not magic zero, empty string, or empty slice.
- **DATA-003 — Domain values.** Prefix/IP use `net/netip`-backed values;
  slugs, choices, DNS names, timestamps, and rack half-units use domain types.
- **DATA-004 — Private rows.** PostgreSQL rows are adapter-private and never
  become domain objects or public DTOs.
- **DATA-005 — PostgreSQL authority.** Correctness does not depend on Redis,
  SQLite behavior, a cache, or a mock.
- **DATA-006 — Safe queries.** Use context, parameters, explicit projections,
  deterministic ordering, and bounded pagination.
- **DATA-007 — Real boundary proof.** Constraints, FK behavior, indexes, locks,
  CIDR, JSONB, arrays, and transactions require real PostgreSQL tests.
- **DATA-008 — Concurrency is designed.** Rack placement, uniqueness, and
  allocation name their locking/isolation strategy and include concurrent
  tests.
- **DATA-009 — Atomic mutation.** Authorization, load, validation, persistence,
  derived state, required object change, and durable event intent share one
  application-owned transaction.
- **DATA-010 — No nested commits.** Repositories join the active unit of work
  and do not independently commit.
- **DATA-011 — No network in transaction.** Durable external delivery uses a
  future transactional outbox; cache changes happen after commit.
- **DATA-012 — Bulk atomicity.** Baseline bulk behavior is not implemented as a
  loop of separately committed single-object calls.
- **DATA-013 — Missing-table-only AutoMigrate.** Call `AutoMigrate` only
  after `HasTable` confirms the table is absent. Never inspect, alter,
  repair, backfill, or drift-correct an existing table through AutoMigrate.
- **DATA-014 — Fatal bootstrap failure.** Schema/bootstrap/content-type seeding
  failures stop startup.
- **DATA-015 — Disposable development schema.** Until Gate 7, incompatible
  development/test shape changes require a fresh disposable database, not
  branches for arbitrary partial upgrades.

### Go rules

- **GO-001 — Formatting.** Handwritten Go passes `gofmt`; imports use the
  pinned formatter when available.
- **GO-002 — Naming.** Use short lowercase package/file names, standard
  initialisms, no package stutter, and export only stable contracts.
- **GO-003 — Explain intent.** Comments explain invariants or trade-offs.
  `TODO`/`FIXME` names an issue or milestone; `nolint` is local, names
  the linter, and explains why.
- **GO-004 — Context first.** I/O-capable methods take
  `context.Context` first, propagate cancellation, and release every
  cancellation function.
- **GO-005 — Typed errors.** Domain/application errors are transport-neutral
  with stable validation, unauthenticated, forbidden, not-found, conflict, and
  internal reasons; wrap causes with `%w`.
- **GO-006 — Central error mapping.** REST and gRPC map typed errors centrally;
  individual handlers do not invent statuses.
- **GO-007 — Structured boundary logs.** Emit one structured log at the
  process/transport boundary, with secrets and full requests removed.

### REST rules

- **REST-001 — Explicit DTOs.** Public DTOs preserve baseline
  `snake_case` and map explicitly to typed commands/results.
- **REST-002 — Explicit allowlists.** Filters, search, ordering, identifiers,
  and writable fields use allowlists; all query values are parameterized.
- **REST-003 — Safe projections.** No public `SELECT *`, raw row
  serialization, or raw-map write.
- **REST-004 — Exact errors and envelopes.** Preserve error shapes, nested/brief
  objects, choice `{value, label}` envelopes, nullability, and pagination
  envelopes.
- **REST-005 — Token path only.** Automation tokens are accepted only through
  the declared NetBox-compatible token authentication path.

### gRPC/protobuf rules

- **GRPC-001 — Handwritten versioned source.** Canonical contracts live under
  packages such as `netbox.dcim.v1`, not frozen table-generated APIs.
- **GRPC-002 — Capability services.** Design services around workflows/modules,
  never tables.
- **GRPC-003 — Presence-aware updates.** Use presence fields and
  `google.protobuf.FieldMask`; use standard timestamps and typed structured
  values.
- **GRPC-004 — Safe evolution.** Reserve removed field numbers and names. Freeze
  each affected v1 capability when its T3 scenarios pass; subsequent breaking
  changes require a new version even if broader publication comes later.
- **GRPC-005 — No storage leakage.** Messages expose no table/column/GORM/SQL
  operators, generic condition trees, hashes, audit internals, or server-owned
  writable fields.
- **GRPC-006 — Shared semantics.** RPC validation beyond malformed transport
  shapes, authorization, transactions, state, and effects belong to the same
  application use case as REST.
- **GRPC-007 — No fictional gateway.** Omit HTTP annotations unless an actual
  gateway is served and tested.

### Vue/TypeScript rules

- **VUE-001 — Vue 3 typed composition.** Use Composition API,
  `<script setup lang="ts">`, typed props/emits, PascalCase components,
  and `useX` composables.
- **VUE-002 — Strict TypeScript.** Production `any` is prohibited by
  default; receive unknown boundary data as `unknown`, validate/narrow, then
  return typed values.
- **VUE-003 — Owned API boundary.** Only API infrastructure imports Axios.
  Features use typed modules/composables and never string-build endpoints.
- **VUE-004 — Explicit wire adapters.** Wire DTOs remain
  `snake_case`; UI view models exist only through deliberate adapters.
- **VUE-005 — Central failures.** Normalize API errors into one discriminated
  type; centralize cancellation, query serialization, trailing slash, and
  401/403 behavior.
- **VUE-006 — Backend authority.** UI affordances are not authorization or
  semantic validation.
- **VUE-007 — Safe HTML.** `v-html` exists only in one audited sanitized
  Markdown component with adversarial tests.
- **VUE-008 — Owned tests.** Colocate component/composable tests and use a real
  browser over the built app for T4.

### Generated-code and dependency rules

- **GEN-001 — Immutable outputs.** Never hand-edit generated source, including
  files with missing markers. Change the owner input/generator and regenerate.
- **GEN-002 — No generated business logic.** Frozen models, DAOs, caches,
  handlers, services, routers, registries, and table protobufs receive no new
  behavior.
- **GEN-003 — Deterministic generation.** Pin tools, sort inputs, fail on
  missing/unmatched input, omit environment data, and require idempotence.
- **GEN-004 — Narrow generation.** A generator does not patch databases, merge
  handwritten files, change `go.mod`, or perform unrelated scaffolding.
- **GEN-005 — Repository hygiene.** Commit needed generated source and lockfiles;
  do not commit binaries, coverage output, frontend builds, editor state,
  secrets, or disposable artifacts.
- **GEN-006 — Immutable third party.** Change vendored code only through a
  reviewed dependency update.
- **GEN-007 — Pinned runtime.** Use Go `1.26.0`, Node `24.18.0`, npm
  `11.16.0`, and repository-pinned protobuf/lint/generation tools. A local
  mismatch is an environment failure, not permission to repin.

### Test, evidence, and operations rules

- **TEST-001 — Hermetic default gate.** Tests own processes, servers, DB,
  clocks, randomness, ports, and fixtures; no unmanaged live service.
- **TEST-002 — Race and warning free.** Backend unit tests run race-enabled;
  formatting, lint, typecheck, generation, and builds have no warnings/errors.
- **TEST-003 — Deterministic synchronization.** No arbitrary sleeps, execution
  order, or wall-clock assumptions.
- **TEST-004 — Regression at lowest layer.** Every defect fix adds the smallest
  test that would have failed before it.
- **TEST-005 — No new exclusions.** Legacy exclusions need owner and removal
  milestone. New exclusions are prohibited by default.
- **TEST-006 — Coverage does not regress.** Coverage is mandatory now; never
  exclude security or compatibility-critical code to improve a number.
- **TEST-007 — External boundaries are distinct.** PostgreSQL, REST oracle,
  gRPC parity, deployment, and browser suites are separate required jobs.
- **TEST-008 — Supporting checks are not tiers.** Build, route enumeration,
  generation, OpenAPI equality, SQLite rollback, and happy-path CRUD do not
  earn T2/T3/T4.
- **TEST-009 — Evidence is current.** Behavioral contract, comparator, fixture,
  implementation, or security changes invalidate affected prior artifacts.
  Claim-only tier/link/status metadata follows the two-digest attestation.
- **TEST-010 — Evidence is durable and clean.** Record source digest/revision,
  command, toolchain, non-secret config, timestamps, result, scenario totals,
  state/effect checks, and a durable credential-free location. `/tmp` is
  diagnostic only.
- **OPS-001 — Production claims wait.** Missing-table-only AutoMigrate and DB
  resets are development conveniences, not an upgrade strategy.
- **OPS-002 — Observable ownership.** Every production service, alert,
  migration, backup, and recovery path has a named operator and runbook.
- **OPS-003 — Rehearse rollback.** A release is incomplete without tested
  rollback or forward-recovery behavior and data-loss boundaries.

## 6. Historical handoff snapshot

> **Historical snapshot — superseded after R5:** The observations below record
> the interrupted-cutover handoff that justified R0-R8. Preserve them as the
> diagnostic history; do not read them as current implementation state. The
> recovery status in section 7 and [Project status](STATUS.md) govern current
> claims. This snapshot was never retained tier evidence.

### What was true at handoff

- The first profile remains **T1 and pre-publication**. No V0-V5 durable bundle
  is linked from the evidence ledger.
- The production REST and gRPC composition uses typed DCIM/IPAM services.
- `go build ./...` succeeds from `netbox-backend`.
- Frontend formatting, lint, application/test typechecks, 125 tests, and build
  have succeeded in focused runs, but the installed local Node/npm
  (`22.23.1`/`10.9.8`) do not match the required
  `24.18.0`/`11.16.0`. Those runs are diagnostics, not V0.
- Contract generation, inventory validation, OpenAPI validation, and local-link
  checks have succeeded in focused runs.
- Repository Git history/status is unavailable in this workspace because
  `.git` has no usable metadata. Use complete-file review and
  `make source-digest`; do not infer a clean revision.

### Immediate red conditions at handoff

1. `make check` stops at Go formatting because
   `netbox-backend/internal/adapters/rest/netbox/workflow/device_handlers.go`
   is not `gofmt` clean.
2. `netbox-backend/test/parity` does not compile after a partially completed
   typed-fixture migration:
   - `newParityIPAMServer` now accepts `composition.Core`, but old
     three-argument calls remain in
     `grpc_profile_errors_assignment_test.go`,
     `resource_parity_test.go`, and
     `grpc_profile_lifecycle_test.go`;
   - obsolete `restadapter.NewHandler(service, sites)` construction remains
     in `resource_parity_test.go`, `grpc_profile_lifecycle_test.go`,
     `site_parity_test.go`, and `site_visibility_parity_test.go`; and
   - five parity call sites still construct
     `NewTransitionalDCIMServer`.
3. `composition.Core` still builds and exposes the generic
   `application/workflow.Service`, even though the production runtime no
   longer consumes it.
4. The dynamic exception is larger than one obsolete field:
   typed transport mappers and repositories still import generic
   `domain/workflow` resources/kinds or row ownership from
   `adapters/postgres/workflow`. It must be unwound in typed slices, not
   deleted wholesale.
5. Backend coverage collection/non-regression is mandatory but is not included
   in the current `make check`; the existing `cover` target writes
   `netbox-backend/cover.out` into the source tree.
6. The router still registers a redundant unauthenticated `GET /ping`, which
   is outside the accepted public health/readiness/login/CSRF boundary.

The prior implementation thread was interrupted by an execution usage limit
and later context-window failures. It was not waiting on an unresolved product
or architecture decision.

## 7. Immediate recovery sequence

Do R0-R8 in order. Do not start a later profile, add first-profile scope, or
change public behavior during recovery.

### Recovery execution status

This table is a structural work log, not compatibility evidence. The detailed
checklists below remain the repeatable audit procedure.

| Steps   | Current execution state                                                                                                                                                                                                       |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R0-R5.6 | Structurally complete: typed composition/parity, canonical `/ping` containment, precise row ownership, generic-package retirement, disposable-PostgreSQL verification, and non-mutating coverage enforcement are implemented. |
| R6-R8   | Closure is governed by the retained V0 artifact and exact digest linked from Status/Evidence. The checklists remain the repeatable procedure and do not themselves make a current claim.                                      |

### R0 — re-establish the observed baseline

**Purpose:** ensure the handoff snapshot is still current and protect existing
work.

**Checklist**

- [ ] Read `CONTEXT.md`, the four ADRs, Compatibility, Standards, Testing,
      Status, and the first-profile manifest.
- [ ] Record `go version`, `node --version`, `npm --version`, and pinned
      versions without changing pins.
- [ ] Run `make source-digest` and keep the value in the work log.
- [ ] Run read-only formatting and focused compile checks.
- [ ] Search for every transitional constructor and generic-workflow import.
- [ ] Classify every changed file as handwritten, generated, frozen, or
      evidence output.
- [ ] Record unrelated user changes; do not reformat them.

**Commands**

```bash
make source-digest
cd netbox-backend
make fmt-check
go build ./...
go test ./test/parity -run '^$' -count=1
rg -n 'NewTransitionalDCIMServer|NewDCIMServer|application/workflow|domain/workflow|postgres/workflow|\.Workflow' internal test
```

**Exit:** the exact current failures and all generic references are captured.
No file has changed.

### R1 — repair formatting only

**Purpose:** remove the first mechanical V0 blocker without combining semantic
work.

**Permitted change**

- Format only
  `internal/adapters/rest/netbox/workflow/device_handlers.go`, unless
  `gofmt -l` reports another file already touched by the interrupted
  increment.

**Checklist**

- [ ] Run `gofmt -w` on the explicit file, not the entire repository.
- [ ] Review the resulting file to ensure the change is formatting-only.
- [ ] Run `make fmt-check`.
- [ ] Do not claim V0; parity is still red.

**Exit:** backend format check is green with no semantic diff.

### R2 — establish one typed parity composition fixture

**Purpose:** make parity tests exercise the same typed services as production
while retaining the ability to inject allow/deny/object-scope authorization.

**Approach**

1. Keep `composition.NewCore(db)` as the production-default entry point.
2. Add the narrowest explicit composition seam needed by tests, accepting the
   typed `authz.ResourceAuthorizer` contract (and preserving the optional
   `authz.ResourceListScopeAuthorizer` behavior when the implementation
   provides it) plus any deterministic clock already required by constructors.
   Do not introduce a service locator or a bag of `any`.
3. Have the production constructor call that seam with
   `authz.PermissionAuthorizer{}` and the system clock.
4. In `test/parity/typed_profile_stack_test.go`, build a complete
   `composition.Core` from the test DB and chosen authorizer.
5. Expose exactly three test helpers:
   `newParityRESTRouter(core, principal)`,
   `newParityDCIMServer(core)`, and `newParityIPAMServer(core)`.
6. The helpers must require all 13 typed services. A missing dependency must
   fail immediately rather than fall back to the generic workflow.

**Tests**

- [ ] Production `NewCore` still chooses the real permission authorizer.
- [ ] An allow-all parity fixture permits the expected operation.
- [ ] A permission-aware fixture denies the same operation for an ungranted
      Principal through both transports.
- [ ] Object list scope still applies before count/pagination.
- [ ] No production server constructor accepts the generic service.

**Exit:** one explicit typed fixture can reproduce every existing parity
authorization setup; no test needs `workflowapp.Service`.

### R3 — finish the parity call-site migration

**Purpose:** make the full parity package compile and run exclusively through
typed REST/gRPC adapters.

**Checklist**

- [ ] Replace all old `newParityIPAMServer(t, db, authorizer)` calls with a
      pre-composed typed core.
- [ ] Replace all two-argument `restadapter.NewHandler(service, sites)`
      constructions with `newParityRESTRouter(core, principal)` or the full
      typed constructor when the test needs direct registration.
- [ ] Replace all `NewTransitionalDCIMServer` calls with
      `newParityDCIMServer(core)`.
- [ ] Preserve each test's original Principal and permission setup.
- [ ] Preserve shared database state between REST and gRPC in same-state tests.
- [ ] Do not rewrite assertions or fixtures merely to match new output.
- [ ] Delete a transitional test helper only after `rg` proves no caller.

**Focused command**

```bash
cd netbox-backend
go test ./test/parity -count=1
```

**Required parity groups**

- [ ] Site lifecycle and visibility.
- [ ] All-resource lifecycle.
- [ ] Stable error mapping.
- [ ] RBAC allow/deny and visibility.
- [ ] Device InterfaceTemplate instantiation and rollback.
- [ ] IPAddress assign/unassign and rollback.
- [ ] Same committed state and object-change effects through REST and gRPC.

**Exit:** the package compiles, every parity test passes, and searches show no
parity use of the transitional generic REST/gRPC constructors.

### R4 — remove generic production composition and close public containment

**Purpose:** stop constructing an unused generic application service in every
runtime process.

**Checklist**

- [ ] Remove `Core.Workflow`.
- [ ] Remove construction of `workflowapp.Service` and
      `workflowpostgres.NewStore` from `composition.NewCore`.
- [ ] Remove now-unused imports.
- [ ] Confirm REST server, gRPC server, admin CLI, contract tests, and router
      tests compile using typed fields only.
- [ ] Confirm no generic object store is initialized for first-profile runtime
      requests.
- [ ] Remove `GET /ping` from canonical router registration and the SPA
      bypass; keep `/health` and `/ready` as the only public process probes.
- [ ] Add/update router tests proving the canonical route inventory contains no
      dedicated `GET /ping`, ordinary SPA history fallback is not a diagnostic
      route, and profile/schema routes still require authentication.

**Commands**

```bash
cd netbox-backend
go test ./internal/platform/composition ./internal/server ./internal/adapters/rest/netbox/router ./test/contract/rest -count=1
go build ./...
```

**Exit:** production composition has no generic workflow field or service.

### R5 — close the dynamic application-boundary exception

**Purpose:** replace the remaining raw-map/generic contracts without a risky
mass deletion.

Before R5.1, run the pinned `make check` once after R3/R4 to reveal failures
outside focused packages; it is diagnostic until R5.6 closes coverage. Perform
the substeps as separate reviewable increments, keep behavior/wire shapes
unchanged, and rerun the applicable full gate after each slice.

#### R5.1 — inventory the exception by ownership

- [ ] Produce an `rg` ledger of every import/use of
      `domain/workflow`, `application/workflow`, and
      `adapters/postgres/workflow`.
- [ ] Classify each use as domain command/result, authorization identifier,
      transport mapper, row/table definition, repository mapping, test fixture,
      or dead code.
- [ ] Include generic error/validation/choice/list/reference helpers and old
      REST/gRPC constructors in the ledger; current uses extend into authz,
      status mapping, typed adapters, migrations, and tests.
- [ ] Map each use to one of the 13 resource owners.
- [ ] Add an architecture/import test that fails on a new generic import
      outside the shrinking allowlist.

#### R5.2 — replace transport-facing dynamic resource values

- [ ] Give every REST mapper a typed input/result and explicit response DTO.
- [ ] Give every gRPC mapper a typed domain/application input/result.
- [ ] Converge the duplicated Site/VRF/generic `Field[T]` implementations on
      one precise typed application-presence abstraction after tests prove
      identical omitted/null/present semantics; do not leave it owned by the
      retiring generic workflow or place it in a vague `common` package.
- [ ] Preserve all returned read-only counts, nested brief objects, choice
      envelopes, nullability, and timestamp semantics.
- [ ] Keep raw JSON only for a profile-declared JSON-valued field, validated at
      its boundary.
- [ ] Add exact mapper tests for absent/null/zero/value and numeric JSON types.

#### R5.3 — move shared identifiers out of the generic domain

- [ ] Replace generic resource “kind” strings used by authorization/change
      logging with a precise typed identifier owned by authz/changelog or the
      resource module.
- [ ] Keep the exact content-type `app_label`/`model` values required by
      permissions and object changes.
- [ ] Do not turn the identifier into a generic repository API.

#### R5.4 — move PostgreSQL row ownership

- [ ] Move each first-profile row and mapper from
      `adapters/postgres/workflow` to its owning `dcim`, `ipam`, or
      `changelog` adapter package.
- [ ] Preserve exact table names, columns, constraints, indexes, FK actions,
      content types, and startup registry order.
- [ ] Update the missing-table registry through owned handwritten inputs.
- [ ] Run real-PostgreSQL shape/constraint tests before deleting the old owner.

#### R5.5 — delete the generic core only at zero references

- [ ] `rg` proves no first-profile runtime, adapter, composition, migration,
      or parity import of `domain/workflow`, `application/workflow`, or
      `adapters/postgres/workflow`, and no transitional REST/gRPC constructor
      remains.
- [ ] Remove dead generic packages and transitional constructors.
- [ ] Tighten the architecture test from an allowlist to a total prohibition.
- [ ] Update the exception ledger in Coding Standards as closed.
- [ ] Do not remove frozen deferred-resource scaffolding.

**Exit:** no first-profile use case accepts/returns unbounded maps; application
and domain are typed; no affected capability is blocked from T2 by the
exception.

### R5.6 — make backend coverage a real non-mutating V0 gate

**Purpose:** satisfy the existing mandatory coverage/non-regression standard
without leaving source-tree artifacts or gaming the measured package set.

- [ ] Change the backend coverage target to write an atomic profile under an
      owned temporary path and remove it after deriving the summary.
- [ ] Measure the same hermetic handwritten packages owned by the unit gate;
      keep the existing reviewed frozen/live-client exclusions explicit.
- [ ] Establish the first trustworthy machine-readable baseline after R5
      cleanup; record total and package set rather than inventing a target.
- [ ] Fail on coverage regression and prohibit new unowned exclusions.
- [ ] Include the non-mutating coverage check in backend/root `make check`.
- [ ] Keep `netbox-backend/cover.out` and
      `netbox-frontend/coverage/` ignored as disposable local output.
- [ ] Prove `make check` leaves the source digest unchanged.

**Exit:** backend and frontend coverage are both part of V0, produce no
source-tree artifact, and enforce a reviewed non-regression baseline.

### R6 — restore the complete V0 repository gate

**Purpose:** establish the mandatory feature-development baseline.

**Environment**

- Go: `1.26.0`
- Node: `24.18.0`
- npm: `11.16.0`
- repository-pinned golangci-lint, protoc, Buf, and protobuf plugins

Use the pinned runtime through the developer's version manager. Do not change
`.go-version`, `.node-version`, `.nvmrc`, `packageManager`, engines,
or lockfiles to fit the host.

**Command**

```bash
make check
```

**Checklist**

- [ ] Backend format, vet, pinned lint, race-unit, build, frozen live-client
      compile, frozen protobuf compile, and canonical Buf checks pass.
- [ ] Backend coverage uses the reviewed package set, meets its recorded
      non-regression baseline, and leaves no profile in the source tree.
- [ ] Frontend `npm ci`, toolchain check, Prettier, ESLint, application/test
      typechecks, coverage, and production build pass.
- [ ] Generated inventory drift, Capability Profile, protobuf descriptor,
      generated contract docs, OpenAPI, and Markdown links pass.
- [ ] There are no warnings, allow-fail jobs, new exclusions, skipped suites,
      or unexpected generated changes.
- [ ] Record source digest, toolchain, start/end times, and result.

**Exit:** a credential-free V0 artifact for the final post-recovery source is
durably retained and linked. Only then may profile completion work resume.

### R7 — reconcile documentation with the recovered implementation

- [ ] Update Status with the exact source digest and observed structure.
- [ ] Mark the dynamic exception closed only after R5 exit.
- [ ] Replace obsolete “once the manifest exists” wording; it exists now.
- [ ] Keep the first profile T1 until the external gates pass.
- [ ] Record the historical early legacy retirement as a deviation from ADR
      0004; do not present it as precedent.
- [ ] Regenerate owned inventories/contract docs if an owned input changed.
- [ ] Run `make docs-check`, `make generated-check`, and
      `make contracts-check`.

### R8 — freeze the recovery boundary

- [ ] Confirm no public route, field, protobuf contract, table shape, or Vue
      workflow changed unintentionally.
- [ ] Confirm every old parity assertion still exercises the same scenario.
- [ ] Confirm all three typed runtime services are the only canonical gRPC
      services registered, plus health.
- [ ] Confirm runtime REST exposes only profile, identity, schema, health,
      readiness, and Vue routes.
- [ ] Confirm the canonical route inventory contains no dedicated `GET /ping`;
      an ordinary SPA history fallback does not count as that endpoint.
- [ ] Record residual first-profile evidence work as V1-V6, not as bugs hidden
      under “done.”

**Exit:** the project has a trustworthy V0 and a clean typed base for sign-off.

## 8. First-profile completion plan: V1 through V6

V0 is a prerequisite, not a substitute for the following boundaries.

### V1 — identity, session, token, CLI, and RBAC evidence

**Credential matrix**

- [ ] Missing, malformed, unknown, expired, revoked, write-disabled, and
      disallowed-IP tokens.
- [ ] Token belonging to an inactive user.
- [ ] Baseline `last_used` update ordering and one-minute throttling,
      including no write for an unknown key.
- [ ] Valid browser session, invalid/expired session, logout, password-change
      invalidation, fixation resistance, and rotation.
- [ ] Missing/invalid CSRF on state-changing cookie requests.
- [ ] Explicit CORS allow/deny origins and no wildcard credentialed origin.
- [ ] Login and token-creation throttling.

**Authorization matrix**

- [ ] Superuser.
- [ ] Direct global grant.
- [ ] Group-derived global grant.
- [ ] User object grant.
- [ ] Group object grant.
- [ ] No grant and wrong action grant.
- [ ] View/add/change/delete distinctions.
- [ ] Object visibility before count/order/pagination.
- [ ] Identical REST and gRPC Principal decisions.

**Administration matrix**

- [ ] One-time empty-database administrator bootstrap.
- [ ] Bootstrap refusal after an administrator exists.
- [ ] Protected password reset.
- [ ] Authenticated active-superuser local user creation.
- [ ] Authenticated global model-permission grant.
- [ ] Passwords stay off argv and logs.
- [ ] No anonymous network bootstrap, reset, creation, or grant endpoint.

**Secret review**

- [ ] Session-cookie flags and CSRF bootstrap contract, including the expected
      readable SameSite CSRF cookie and matching response-body value.
- [ ] One-time API-token secret appears only in the creation response.
- [ ] List/get never return reusable secret material.
- [ ] No hash, token, cookie, CSRF value, password, DSN, or full config in
      logs/evidence.

**Exit:** the current security matrix is retained against both interfaces.
Identity operations remain classified as extensions where applicable.

### V2 — domain and application evidence

Create a traceability row for every manifest scenario and every rule in
`IMPLEMENTATION_PLAN.md`. At minimum cover:

- [ ] all 13 resources' create/get/list/PUT/PATCH/delete positive paths;
- [ ] defaults, read-only fields, nullability, absent/null/zero/value updates;
- [ ] global, sibling, manufacturer-, site-, location-, and VRF-scoped
      uniqueness exactly where the baseline applies it;
- [ ] DeviceRole hierarchy cycle/cascade/protection;
- [ ] Rack bounds, half-unit placement, occupancy, protected changes, and
      RackType propagation behavior;
- [ ] Device nullable name and case-insensitive uniqueness;
- [ ] InterfaceTemplate snapshot and atomic Device/Interface creation;
- [ ] Prefix canonicalization, containment, `/0` rejection, utilization
      flags, VRF/global uniqueness, and host-bit suggestion;
- [ ] IP host/mask preservation, role exceptions, IPv4 and IPv6 edge prefixes,
      SLAAC, DNS lowercase, assignment presence matrix, and no invented
      Interface-VRF equality rule;
- [ ] explicit filters, search, ordering, bounded pagination, nested
      projections, and read-only counters;
- [ ] validation, unauthenticated, forbidden, not-found, conflict, delete
      protection, and rollback stable reasons;
- [ ] authorization before persistence and required object-change records.

**Exit:** every declared invariant is linked to a positive/negative test; every
multi-object mutation proves atomic rollback.

### V3 — real PostgreSQL and standalone deployment

Run the DSN-enabled suites listed in Testing, then deployment smoke.

**PostgreSQL checklist**

- [ ] Empty-database bootstrap.
- [ ] Existing correct table remains untouched.
- [ ] A missing table is created in a non-empty database.
- [ ] A deliberately malformed existing table is not inspected or repaired.
- [ ] Startup registry contains the expected current entries in deterministic
      order; do not treat the count as capability evidence.
- [ ] Private typed table names, columns, constraints, indexes, FK actions, and
      content-type uniqueness.
- [ ] Concurrent duplicate/placement/allocation/assignment behavior uses the
      declared isolation or locking strategy.
- [ ] Transaction rollback also removes derived objects and object changes.
- [ ] Identity/session/token/CLI state survives reconnect/restart.

**Deployment checklist**

- [ ] Unique clean Compose project and volume.
- [ ] No Django migration or initialization-SQL mount.
- [ ] Health and readiness reflect real dependencies.
- [ ] Content types and schema bootstrap successfully.
- [ ] Application restart is idempotent.
- [ ] Teardown owns only its unique project and leaves no credential artifacts.

**Exit:** durable real-PostgreSQL and deployment artifacts refer to the same
source digest.

### V4 — REST T2, then gRPC T3

Run in this order:

```bash
make compatibility-comparator-test
make compatibility-test
cd netbox-backend
go test ./test/parity -count=1
```

**REST T2 review**

- [ ] Oracle SHA and effective configuration refusal are active.
- [ ] Comparator self-test proves rejection of intentional divergences.
- [ ] Every profile scenario has isolated, deterministic fixture state.
- [ ] Status, path/trailing slash, auth, validation reason, numeric JSON type,
      fields, choices, pagination, state, and side effects match.
- [ ] Every generated ID is bound to a named scenario object.
- [ ] Only committed normalizers were used.
- [ ] Secure identity divergence is excluded from baseline T2 and reported in
      its own matrix.

**gRPC T3 review**

- [ ] Corresponding REST capability already has retained T2.
- [ ] RPCs use typed transport-native messages and the same application use
      cases.
- [ ] Lifecycle, errors, RBAC/list visibility, rollback, assignment, state, and
      object changes are equivalent.
- [ ] No RPC goes through REST or transitional generic workflow code.

**Exit:** each capability is promoted individually. A partial green report
does not promote the whole profile.

### V5 — Vue Workflow Parity

Run `make browser-e2e` against a fresh built deployment.

**Required browser outcomes**

- [ ] Login, session refresh, CSRF mutation, logout, and limited-user denial.
- [ ] Manufacturer/roles/RackType/Site/Rack creation needed for a Device.
- [ ] DeviceType plus InterfaceTemplate, Device creation, and instantiated
      Interface display.
- [ ] Template-instantiation failure rolls back Device, Interfaces, and object
      changes.
- [ ] VRF, Prefix, and IPAddress create/edit/list/filter workflows.
- [ ] Interface assignment, reassignment, unassignment, and rollback.
- [ ] Explicit-null PATCH, validation, conflict, not-found, and protected
      deletion UX.
- [ ] Interface delete warning, cancel, confirm, and resulting IP/change state.
- [ ] Correct read-side choice envelopes and scalar mutation values.
- [ ] No credential/session/CSRF state in `localStorage`,
      `sessionStorage`, or IndexedDB, and no secret in retained diagnostics;
      only the documented cookies may hold browser credential/CSRF state.

Component tests remain required but cannot substitute for these outcomes.

### V6 — reviewed first-profile sign-off

- [ ] Re-run `make check` on the unchanged tested digest before promotion
      metadata changes.
- [ ] Link V0-V5 artifacts from the ledger, profile report, and Status using the
      two-digest attestation procedure in section 16.
- [ ] Run `make check` again on the final attestation digest and prove the
      only post-test changes are reviewed evidence summaries, links, and
      tier/status metadata.
- [ ] Verify exact source digest, toolchains, non-secret config, timestamps,
      scenario totals, durable-state/effect checks, and artifact retention.
- [ ] Verify the dynamic application exception is closed.
- [ ] Verify runtime inventories/reflection/routes contain only declared
      profile and named extension surfaces.
- [ ] Verify every deferred field/action remains documented.
- [ ] Verify no status prose calls DCIM, IPAM, or all NetBox complete.
- [ ] Freeze each affected v1 protobuf capability when its T3 scenarios pass;
      after that, evolution is additive and a breaking change requires a new
      package version even if broader profile publication occurs later.
- [ ] Obtain human review of compatibility, security, data, and operations
      artifacts together.

**Exit:** `core-workflow-v1` is signed off. Only then may Gate 5 feature work
merge.

## 9. Definition of ready for a later Capability Profile

A future profile is not ready for implementation until every item below is
reviewed.

### Scope

- [ ] Profile ID, owner, purpose, operator workflow, prerequisites, and
      non-goals are named.
- [ ] Scope comes from
      `contracts/netbox/v4.4.6-post7/inventory/baseline-rest.yaml`, not from
      frozen handlers or protobuf services.
- [ ] Every included route, method, action, field, relationship, filter, and
      computed value is listed.
- [ ] Every intentionally deferred field, relationship, action, bulk shape,
      and cross-module dependency is listed.
- [ ] Extension/divergence entries have a separate classification and
      acceptance matrix.
- [ ] The profile remains pre-publication.

### Baseline discovery

- [ ] Relevant upstream URL/router, viewset, serializer, filterset, model,
      validation, permission, signal/job, and test sources are cited at the
      exact pinned SHA.
- [ ] Default settings, timezone, authentication policy, plugins, database, and
      fixtures needed by the oracle are fixed.
- [ ] Exact paths, trailing slashes, methods, content types, statuses, and error
      envelopes are recorded.
- [ ] Create/PUT/PATCH writable fields, defaults, read-only values, nullability,
      and field-presence states are recorded.
- [ ] Nested/brief objects, choice envelopes, numeric JSON types, counters, and
      display values are recorded.
- [ ] Search, every allowed filter, ordering, count, limit, offset, and
      visibility order are recorded.
- [ ] Uniqueness, deletion, bulk atomicity, concurrency, transaction, object
      change, job/event, and durable side effects are recorded.
- [ ] Surprising or uncertain behavior has a differential scenario rather than
      a prose guess.

### Acceptance design

- [ ] Positive lifecycle scenario.
- [ ] Invalid input for every invariant.
- [ ] Absent/null/zero/empty/concrete PATCH matrix.
- [ ] PUT replacement/default behavior.
- [ ] Unauthenticated, forbidden, object-hidden, and allowed cases.
- [ ] Not-found and conflict cases.
- [ ] Delete protect/cascade/set-null behavior.
- [ ] Transaction rollback after a late failure.
- [ ] Concurrent scenario where a race can violate an invariant.
- [ ] Durable state and object-change assertions.
- [ ] Equivalent gRPC outcome for every baseline operation.
- [ ] Browser workflow and negative UX cases when operator-facing.
- [ ] Comparator sensitivity case for any new normalizer.

### Contract design

- [ ] Resource metadata and machine-readable profile validate.
- [ ] REST wire DTOs and OpenAPI preserve exact baseline spelling/shape.
- [ ] Versioned protobuf is transport-native, typed, presence-aware, and
      module/capability-oriented.
- [ ] New protobuf fields have stable numbers; removed fields reserve number
      and name.
- [ ] Domain/application types contain no REST, protobuf, GORM, or raw-map
      leakage.
- [ ] PostgreSQL table/constraint/index plan and locking strategy are reviewed.
- [ ] Vue feature boundary and typed DTO/form/filter mapping are identified.

If any box remains open, the next increment is discovery/profile work, not
application code.

## 10. Repeatable Capability Profile factory

A profile is an acceptance unit and may require several small vertical
increments. Do not implement “all domain objects,” then “all repositories,”
then “all handlers.” Each increment should leave one coherent subset usable and
testable across its required layers.

### CP-00 — select the workflow

**Inputs:** deferred baseline ledger, dependency graph, operator need, and all
earlier retained profiles.

**Steps**

1. Choose the smallest workflow with an observable outcome.
2. Include only supporting resources required for that outcome.
3. Identify dependencies that remain fixture-only versus publicly promoted.
4. Assign profile ID and owners for backend, frontend, compatibility, security,
   and evidence.
5. Record non-goals and the retirement boundary.

**Exit:** a one-page profile proposal can answer “what task becomes possible,
through which interfaces, and what stays deferred?”

### CP-01 — interrogate the pinned baseline

**Steps**

1. Trace router → view/action → serializer → validation/model → permissions →
   durable effects.
2. Trace filtersets, ordering, search, pagination, nested serializers, and
   choice labels.
3. Trace delete behavior, signals, object changes, jobs/events, and transaction
   boundaries.
4. Read upstream tests, but verify behavior against the oracle rather than
   treating test names as the contract.
5. Record exact source paths and line/function names in resource metadata or
   business-logic notes.
6. Add a discovery scenario for every surprising branch.

**Exit:** no public behavior in the proposed profile depends on memory or
generic framework expectations.

### CP-02 — write the machine-readable contract

**Steps**

1. Add versioned resource metadata.
2. Add profile resource/action entries with methods and operation states.
3. Declare included and deferred fields, relations, filters, ordering, actions,
   bulk forms, and workflow scenarios.
4. Classify each operation as baseline or extension.
5. Set initial tier to T0/T1 as appropriate; never pre-award evidence.
6. Generate and review OpenAPI/contract documentation from owned inputs.
7. Validate that the runtime remains unpublished.

**Exit:** profile validation and generated-drift checks pass; reviewers can
determine exact scope without reading implementation code.

### CP-03 — build fixtures and failing acceptance scenarios

**Steps**

1. Create deterministic named fixtures in isolated databases.
2. Bind generated IDs by semantic names.
3. Add positive, negative, authz, conflict, delete, rollback, state, and effect
   scenarios.
4. Add comparator rules only for unavoidable origins, bound IDs, or volatile
   timestamps.
5. Prove every new normalizer rejects a nearby semantic difference.
6. Author gRPC scenarios against the same semantic outcomes.
7. Author browser workflow outline and selectors owned by user-visible meaning,
   not brittle DOM position.

**Exit:** scenarios fail for the expected unimplemented reason and would catch
an intentionally wrong implementation.

### CP-04 — implement typed domain behavior

**Location:** `netbox-backend/internal/domain/<module>/`.

**Steps**

1. Define typed IDs, value objects, aggregates/entities, choices, and explicit
   optional/presence types.
2. Encode invariants and state transitions without transport/persistence
   imports.
3. Distinguish create, replace, patch, and action semantics.
4. Return typed stable errors/reasons.
5. Add table-driven positive, boundary, and negative tests.
6. Add property/fuzz tests for parsers, ranges, networks, hierarchies, or other
   combinatorial values where useful.

**Exit:** domain tests describe and prove invariants with no I/O.

### CP-05 — implement typed application use cases

**Location:** `netbox-backend/internal/application/<module>/`.

**Steps**

1. Define one typed command/query/result per operation or coherent action.
2. Define narrow consumer-owned repository/integration ports.
3. Accept `context.Context` and authenticated Principal.
4. Authorize before observable data access or persistence.
5. Apply object visibility before count/pagination.
6. Load current/related state, invoke domain behavior, and coordinate one unit
   of work.
7. Record required object changes and durable event intent in that unit.
8. Return transport-neutral typed errors.
9. Add orchestration, permission, visibility, late-failure, and rollback tests.

**Exit:** use cases require no Gin/protobuf/GORM types and neither transport
needs business logic.

### CP-06 — implement PostgreSQL adapters

**Location:** `netbox-backend/internal/adapters/postgres/<module>/`.

**Steps**

1. Define private typed rows and explicit domain mappings.
2. Choose the private schema deliberately. Preserve an upstream table
   name/shape only when an accepted migration contract requires it; REST
   compatibility does not imply database-schema compatibility.
3. Define FK actions, unique/partial indexes, checks, content types, and
   deterministic query order.
4. Implement explicit projections and parameterized filters.
5. Push visibility/count/order/limit/offset into SQL only when semantics are
   exact; use an authorization-before-pagination hybrid for non-portable
   predicates.
6. Join the application-owned transaction.
7. Implement an explicit lock/isolation strategy for contested invariants.
8. Register only missing tables for development AutoMigrate.
9. Test on real PostgreSQL, including constraint names/outcomes, concurrency,
   rollback, and malformed-existing-table non-repair.

**Exit:** real PostgreSQL proves storage semantics and no row type escapes the
adapter.

### CP-07 — implement exact REST

**Location:** `netbox-backend/internal/adapters/rest/netbox/<module>/`.

**Steps**

1. Define explicit request, patch-presence, brief/nested, list-envelope, and
   response DTOs.
2. Decode only declared fields and reject/handle unknown/deferred input exactly
   as the profile says.
3. Map DTOs explicitly to typed application commands.
4. Route through shared authentication and central error mapping.
5. Preserve paths, slashes, methods, status, headers, content types, errors,
   choices, pagination, filters, ordering, and durable effects.
6. Register only declared runtime routes.
7. Update owned OpenAPI source/generation and route-contract tests.
8. Make the pre-authored differential scenarios ready to exercise the composed
   runtime; run in-process route/mapper tests now.

**Exit:** REST adapter, exact DTO mapping, OpenAPI, and in-process contract tests
are green. T2 is earned only after CP-10 composition and CP-11 strict oracle
execution.

### CP-08 — implement semantic gRPC

**Locations:** `netbox-backend/api/proto/netbox/<module>/v1/` and
`internal/adapters/grpc/<module>/`.

**Steps**

1. Design typed messages around operations, not JSON or tables.
2. Use explicit presence and FieldMask for partial updates.
3. Map request to the same application command as REST.
4. Resolve bearer metadata to the same Principal.
5. Map typed errors centrally to stable gRPC codes/details.
6. Implement in-process parity scenarios for accepted/rejected transitions,
   permissions, committed state, rollback, and object changes corresponding to
   the planned REST scenarios.
7. Regenerate protobuf outputs with pinned tools; never edit them.

**Exit:** the typed adapter and in-process same-core parity scenarios are green.
T3 is earned in CP-11 only after corresponding REST T2.

### CP-09 — implement the Vue workflow

**Location:** `netbox-frontend/src/features/<module>/`, with shared API/router
infrastructure used only through typed boundaries.

**Steps**

1. Define exact snake_case wire DTOs.
2. Define typed view, form, filter, and mutation values.
3. Map read-side choices and mutation-side scalar choices deliberately.
4. Add typed feature APIs/composables; pages do not import Axios or build URLs.
5. Add permission-aware affordances while preserving backend authority.
6. Handle normalized validation, auth, forbidden, conflict, not-found, and
   cancellation errors.
7. Add accessible loading, empty, success, warning, and failure states.
8. Add colocated component/composable tests.
9. Add the clean-database real-browser positive and negative scenarios to the
   owned harness.

**Exit:** typed feature/component tests pass and the browser harness is ready
for the composed runtime. T4 is earned only by CP-11 execution.

### CP-10 — compose and contain

**Steps**

1. Wire repositories and services explicitly in
   `internal/platform/composition`.
2. Register only canonical declared REST/gRPC routes.
3. Confirm frozen deferred routes/services remain disabled.
4. Confirm frontend model/route registry includes only enabled capabilities.
5. Add architecture/import/runtime inventory assertions.

**Exit:** a missing dependency fails startup/test construction; no generic
fallback can serve the new capability. CP-10 must be green before any CP-11
oracle, deployment, parity, or browser claim.

### CP-11 — climb the verification ladder

Use the command ladder in section 15. Fix causes at the lowest responsible
layer. Do not compensate in a transport or normalizer for wrong domain/state
behavior.

Run repository/PostgreSQL/deployment checks, strict REST to T2, corresponding
gRPC to T3, and applicable browser workflows to T4 in that order.

**Exit:** all required fast and external gates are green on one unchanged
tested digest.

### CP-12 — retain evidence and promote

**Steps**

1. Retain credential-free artifacts for the tested digest in the required
   shape.
2. Link artifacts from the ledger and profile report using the two-digest
   attestation procedure in section 16.
3. Review normalizers and scenario coverage.
4. Promote only the exact capability rows exercised.
5. Keep the profile pre-publication until all of its exit conditions pass
   together.
6. Update Status conservatively and run `make check` on the final attestation
   digest.

### CP-13 — retire displaced scaffolding

This step occurs only after capability completion under ADR 0004.

- [ ] Enumerate the exact generated model/DAO/cache/service/handler/router/proto
      artifacts displaced by the completed capability.
- [ ] Prove there is no runtime registration, import, reflection entry, or test
      dependency.
- [ ] Remove only that set.
- [ ] Regenerate authoritative current inventories.
- [ ] Compile frozen remaining clients/protobufs.
- [ ] Re-run all earlier profile gates.

**Exit:** no duplicate public or runtime path remains, and unrelated deferred
scaffolding is untouched.

### CP-14 — publish or continue

- Publish a profile/API version only after its contract stability, support
  status, evidence, and upgrade implications are reviewed.
- If cross-module fields/actions remain deferred, call the profile complete
  only within its declared scope and schedule the module-closeout pass.

## 11. Recommended dependency-ordered profiles

These are execution-sized planning boundaries, not silently accepted public
contracts. Before coding, materialize each through CP-00 to CP-03. Together
they account for all deferred in-scope baseline entries; generated scaffolding
does not.

The baseline contains 132 resources and 23 custom actions. The first profile
contains 13 resources. Of what remains, 118 resources and 22 actions are
currently in-scope candidates. Two entries are explicit architectural
exclusions: `/api/extras/scripts/` and anonymous
`/api/users/tokens/provision/`.

Sixteen baseline resources do not have a corresponding frozen REST scaffold,
which is why file-based planning is invalid:

- `circuits/virtual-circuit-types`;
- `core/background-queues`, `background-tasks`,
  `background-workers`, and `object-types`;
- `dcim/cable-terminations` and `connected-device`;
- `extras/table-configs` and `tagged-objects`;
- `ipam/fhrp-group-assignments`;
- `tenancy/contact-assignments`;
- `users/config` and `permissions`;
- `virtualization/virtual-disks`;
- `vpn/ipsec-proposals`; and
- `wireless/wireless-links`.

The “critical discovery/proof topics” below are a research checklist, not
pre-accepted behavior. Verify each against the pinned source and oracle before
placing it in a Capability Profile; GOV-013 prohibits implementing these
planning hypotheses from prose alone.

### Profile 1 — `dcim-hierarchy-v1`

**Resources:** Region, SiteGroup, Location, Platform.

**Workflow:** build geographic/organizational hierarchy, place Sites and
Devices through promoted relationships, and select a platform.

**Critical discovery/proof topics:** hierarchy cycles, sibling/top-level uniqueness, nested
brief projections, location scoping, visibility, delete protection/cascade,
and promotion of existing Site/Device fields only when fully covered.

**Dependencies:** signed-off core profile.

### Profile 2 — `dcim-modular-hardware-v1`

**Resources:** ModuleTypeProfile, ModuleType, ModuleBayTemplate, ModuleBay,
Module, DeviceBayTemplate, DeviceBay, InventoryItemRole,
InventoryItemTemplate, InventoryItem, MACAddress.

**Workflow:** define modular hardware templates, instantiate/occupy component
bays, manage inventory hierarchy, and attach MAC addresses.

**Critical discovery/proof topics:** template snapshot timing, component ownership, name and
position uniqueness, bay compatibility/occupancy, recursive inventory,
protected deletion, atomic instantiation, and MAC assignment uniqueness.

**Dependencies:** DeviceType/Device/Interface from the core profile.

### Profile 3 — `dcim-ports-power-v1`

**Resources:** ConsolePortTemplate, ConsolePort, ConsoleServerPortTemplate,
ConsoleServerPort, FrontPortTemplate, FrontPort, RearPortTemplate, RearPort,
PowerPortTemplate, PowerPort, PowerOutletTemplate, PowerOutlet, PowerPanel,
PowerFeed.

**Workflow:** define and instantiate physical termination families and power
topology without yet promoting cables/path traversal.

**Critical discovery/proof topics:** typed termination endpoints, pass-through front/rear
mappings, template instantiation, feed/panel topology, phase/type/status
choices, occupancy, and delete behavior.

**Dependencies:** Profiles 1-2 and the core Device workflow.

### Profile 4 — `ipam-registry-routing-v1`

**Resources:** RIR, Aggregate, ASNRange, ASN, RouteTarget.

**Action:** `GET` and `POST` on
`/api/ipam/asn-ranges/{id}/available-asns/`.

**Workflow:** define registries/ranges, allocate ASNs safely, and manage route
targets.

**Critical discovery/proof topics:** integer ranges and overlap, private/reserved behavior,
concurrent next-available allocation, aggregate containment, RIR association,
route-target syntax/uniqueness, tenant fields explicitly deferred until owned.

**Dependencies:** core IPAM.

### Profile 5 — `ipam-vlan-v1`

**Resources:** Role, VLANGroup, VLAN, VLANTranslationPolicy,
VLANTranslationRule.

**Action:** `GET` and `POST` on
`/api/ipam/vlan-groups/{id}/available-vlans/`.

**Workflow:** create scoped VLAN groups, allocate VLAN IDs, and define
translation rules.

**Critical discovery/proof topics:** group scope types, VID range/uniqueness, concurrent
allocation, role/status choices, translation source/target integrity,
non-overlap, and delete protection.

**Dependencies:** DCIM hierarchy for scope promotion; tenancy may remain
explicitly deferred.

### Profile 6 — `ipam-address-services-v1`

**Resources:** IPRange, FHRPGroup, FHRPGroupAssignment, ServiceTemplate,
Service.

**Actions:** `GET` and `POST` on
`/api/ipam/ip-ranges/{id}/available-ips/`,
`/api/ipam/prefixes/{id}/available-ips/`, and
`/api/ipam/prefixes/{id}/available-prefixes/`.

**Workflow:** reserve/allocate addresses and prefixes, model redundant
gateways, and attach services to declared parents.

**Critical discovery/proof topics:** address/range containment, usable-space calculations,
allocation locking, prefix-length bounds, VRF/global uniqueness, FHRP
assignment constraints, protocol/port ranges, polymorphic service parents, and
atomic allocation.

**Dependencies:** core Prefix/IPAddress; Profiles 4-5 for registry/role/VLAN
relations where promoted.

### Profile 7 — `tenancy-v1`

**Resources:** TenantGroup, Tenant, ContactGroup, ContactRole, Contact,
ContactAssignment.

**Workflow:** manage tenant/contact hierarchies and assign contacts to typed
objects.

**Critical discovery/proof topics:** hierarchy cycles, scoped uniqueness, visibility,
assignment uniqueness/priority, typed content-object targets, object
authorization, and delete behavior.

**Dependencies:** signed-off target modules for each promoted assignment. This
profile may precede some consumers, but must not claim their tenant fields
until their closeout.

### Profile 8 — `users-admin-v1`

**Resources:** Users Config, User, Group, Permission, Token.

**Workflow:** expose the reviewed baseline administration surface over the
existing Go-owned identity store.

**Critical discovery/proof topics:** reuse the single identity/Principal/RBAC store; distinguish
baseline `/api/users/tokens/` wire behavior from secure
`/api/auth/tokens/`; never reintroduce the unsafe removed public
`/config` endpoint under a different route; never normalize security
divergence; hash/secret non-disclosure; group/permission/object-scope rules.

**Explicit exclusion:** anonymous
`POST /api/users/tokens/provision/` remains unsupported unless a later
accepted security/compatibility decision changes the boundary. It can never be
called T2 while intentionally rejected.

### Profile 9 — `dcim-operational-v1`

**Resources:** RackReservation, VirtualChassis, VirtualDeviceContext.

**Action:** `GET /api/dcim/racks/{id}/elevation/`.

**Workflow:** reserve rack units, group Devices into virtual chassis/VDC
contexts, and render exact rack elevation data.

**Critical discovery/proof topics:** user/tenant ownership, reservation overlap, half-unit
geometry, device position/face, chassis member/position uniqueness, VDC
assignment, SVG/JSON content and authorization, and safe rendering.

**Dependencies:** core Rack/Device, Profile 1 hierarchy, Profile 7 tenancy, and
Profile 8 users.

### Profile 10 — `virtualization-v1`

**Resources:** ClusterGroup, ClusterType, Cluster, VirtualMachine, VirtualDisk,
virtualization Interface.

**Workflow:** define clusters/VMs/disks/interfaces and use IP/VLAN assignments
with the same semantics as physical interfaces where applicable.

**Critical discovery/proof topics:** cluster scope, VM role/platform/status, name uniqueness,
vCPU/memory/disk values, disk/interface ownership, MAC/IP/VLAN relationships,
primary-IP behavior when promoted, and atomic deletion.

**Dependencies:** DCIM hierarchy/platform, Users/Tenancy, IPAM/VLAN, and
existing DeviceRole semantics.

### Profile 11 — `circuits-v1`

**Resources:** Provider, ProviderAccount, ProviderNetwork, CircuitType,
Circuit, CircuitTermination, CircuitGroup, CircuitGroupAssignment,
VirtualCircuitType, VirtualCircuit, VirtualCircuitTermination.

**Workflow:** manage providers and physical/virtual circuits with A/Z
terminations.

**Critical discovery/proof topics:** provider/account/network ownership, circuit ID uniqueness,
termination side uniqueness, termination targets, group membership,
commit/activate dates, status, rates, virtual provider-network circuits,
protection/deletion, and cross-object visibility.

**Dependencies:** Tenancy and DCIM termination owners; this profile intentionally
precedes cable traversal so circuit terminations can become typed endpoints.

### Profile 12 — `dcim-cabling-v1`

**Resources:** Cable, CableTermination, ConnectedDevice.

**Actions:** `GET` trace on ConsolePort, ConsoleServerPort, Interface,
PowerFeed, PowerOutlet, and PowerPort; `GET` paths on FrontPort and RearPort,
using their baseline `/{id}/trace/` or `/{id}/paths/` routes.

**Workflow:** connect physical termination endpoints and traverse/trace paths
across pass-through ports and circuits.

**Critical discovery/proof topics:** typed endpoint registry, termination occupancy, compatible
types, cable status/type/length, pass-through mapping, path order, cycle
detection, connected-device resolution, authorization of every visible hop,
and delete/change effects.

**Dependencies:** Profiles 3 and 11. Do not model generic endpoints with
unvalidated strings or raw maps.

### Profile 13 — `wireless-v1`

**Resources:** WirelessLANGroup, WirelessLAN, WirelessLink.

**Workflow:** organize WLANs and connect wireless Interfaces.

**Critical discovery/proof topics:** group hierarchy, SSID/status/auth/cipher choices, VLAN and
tenant relations, interface/link uniqueness, endpoint ordering, and secret
handling for authentication material.

**Dependencies:** VLAN, Tenancy, and physical/virtual Interface capabilities.

### Profile 14 — `vpn-v1`

**Resources:** IKEProposal, IKEPolicy, IPSecProposal, IPSecPolicy, IPSecProfile,
TunnelGroup, Tunnel, TunnelTermination, L2VPN, L2VPNTermination.

**Workflow:** compose ordered proposals/policies and terminate tunnels/L2VPNs
on typed objects.

**Critical discovery/proof topics:** ordered membership, cryptographic choice validation,
proposal reuse, tunnel status/encapsulation, termination uniqueness,
polymorphic endpoints, secrets, object authorization, and deletion protection.

**Dependencies:** IPAM, DCIM, Virtualization, Circuits, and Tenancy endpoints.

### Profile 15 — `core-audit-v1`

**Resources:** ObjectType, ObjectChange.

**Workflow:** discover supported content types and inspect authorized change
history already produced by shared application transactions.

**Critical discovery/proof topics:** unique `(app_label, model)`, exact action/user/request
metadata, before/after snapshots, object visibility, retention/pagination, no
secret data, and immutable audit records.

**Dependencies:** all existing changelog producers. Read behavior must not
expose an object hidden from the Principal.

### Profile 16 — `core-automation-v1`

**Resources:** BackgroundQueue, BackgroundTask, BackgroundWorker, DataFile,
DataSource, Job.

**Actions:** `POST` enqueue, requeue, stop, and delete on
`/api/core/background-tasks/{id}/<action>/`; `POST` synchronize on
`/api/core/data-sources/{id}/sync/`.

**Workflow:** run durable jobs and data-source synchronization without Python
runtime dependency.

**Critical discovery/proof topics:** explicit task state machine, idempotency, lease/heartbeat,
retry/backoff, cancellation races, ownership, input/output retention, file
integrity, SSRF/path/archive defenses, crash recovery, and object changes.

**Human-reviewed ADR prerequisite:** before CP-02, accept the durable
queue/worker state machine, lease and cancellation model, retry/idempotency
boundary, and Extension Service split for behavior formerly supplied by
Python. Luna must not choose these cross-cutting contracts inside an
implementation increment. Do not pretend a Go stub is script compatibility.

### Profile 17 — `extras-metadata-v1`

**Resources:** Tag, TaggedObject, CustomFieldChoiceSet, CustomField,
ConfigContextProfile, ConfigContext, ConfigTemplate.

**Actions:** `GET` on
`/api/extras/custom-field-choice-sets/{id}/choices/` and `POST` on
`/api/extras/config-templates/{id}/render/`.

**Workflow:** attach typed metadata/config context and render reviewed templates.

**Critical discovery/proof topics:** typed content-object registry, tag uniqueness, custom field
type/default/required/validation semantics, choice ordering, context
precedence/merge, template sandboxing, injection defenses, and visibility.

**Human-reviewed ADR prerequisite:** before CP-02, accept the closed
cross-module object registry and the template language/sandbox/resource-limit
contract that can satisfy the verified render behavior without embedding
Python. Luna must not select a rendering engine ad hoc.

### Profile 18 — `extras-productivity-v1`

**Resources:** Bookmark, CustomLink, ExportTemplate, ImageAttachment,
JournalEntry, SavedFilter, TableConfig.

**Action:** `/api/extras/dashboard/` through `GET`, `PUT`, `PATCH`,
and `DELETE`.

**Workflow:** persist per-user shortcuts/configuration, annotate objects,
attach images, export safely, and compose the dashboard.

**Critical discovery/proof topics:** ownership/visibility, typed generic targets, URL/template
sanitization, MIME/size/content validation, storage lifecycle, saved-filter
schema, table-config bounds, export authorization, and no unsafe HTML.

**Human-reviewed ADR prerequisite:** before CP-02, accept attachment blob
storage, integrity, retention/deletion, backup, and serving policy plus the
safe export-rendering boundary. Luna must not introduce a storage provider or
template engine inside a feature increment.

### Profile 19 — `extras-automation-v1`

**Resources:** EventRule, Webhook, NotificationGroup, Notification,
Subscription.

**Workflow:** subscribe to committed domain events and deliver authenticated
notifications/webhooks outside the core transaction.

**Critical discovery/proof topics:** transactional outbox, event filtering, signatures,
idempotency keys, retry/backoff, timeout, ordering boundary, replay/dead-letter,
loop prevention, redaction, endpoint allowlist/SSRF defense, and operational
ownership.

**Human-reviewed ADR prerequisite:** before CP-02, accept stable domain-event
versioning, transactional outbox ownership, delivery guarantees, replay/dead
letter behavior, and Extension Service contracts. Luna must not invent these
cross-cutting guarantees while implementing a resource.

### Inventory arithmetic check

- DCIM: 35 future resources and 9 actions.
- IPAM: 15 future resources and 5 actions.
- Tenancy: 6 resources.
- Users: 5 in-scope resources; anonymous provision action excluded.
- Virtualization: 6 resources.
- Circuits: 11 resources.
- Wireless: 3 resources.
- VPN: 10 resources.
- Core: 8 resources and 5 actions.
- Extras: 19 in-scope resources and 3 actions; Scripts resource excluded.

The profile documents must recalculate these figures from the inventory.
Numbers in this playbook are an audit cross-check, not an implementation source.

## 12. Module-closeout passes

Completing the profiles above does not automatically complete a module. Earlier
profiles deliberately defer fields and relationships until their owning
capability exists. After the dependency profiles are signed off, close modules
in this order unless the machine-readable dependency graph proves a safer
order:

1. DCIM
2. IPAM
3. Tenancy
4. Users/Identity
5. Virtualization
6. Circuits
7. Wireless
8. VPN
9. Core
10. Extras

For each module:

- [ ] Diff every baseline route/action against all accepted profiles.
- [ ] Diff every serializer field, writable flag, default, nullability state,
      brief/nested projection, filter, search/order field, and computed value.
- [ ] Promote or explicitly exclude every deferred tenant, tag, custom field,
      journal, image, content-type, assignment, termination, platform, VRF,
      VLAN, primary-IP, config-context, and template relationship.
- [ ] Cover baseline bulk create/update/delete/import/rename behavior or keep it
      explicitly out of the claimed in-scope replacement.
- [ ] Recheck generic object references against the typed object registry.
- [ ] Recheck permissions and list visibility after cross-module joins.
- [ ] Recheck delete protection/cascade/set-null across module boundaries.
- [ ] Recheck object-change snapshots and durable events.
- [ ] Run the full-module REST differential and gRPC parity matrices.
- [ ] Run all declared operator workflows.
- [ ] Prove zero runtime-enabled legacy route/service overlap.
- [ ] Retire only now-displaced frozen artifacts under ADR 0004.
- [ ] Update the baseline ledger so no accidental T0 entry is hidden.

A module is complete only when every in-scope entry meets the canonical
definition of complete. “All routes exist,” “all tables exist,” and “all
generated services compile” are not closeout conditions.

## 13. Extension Service gate

Out-of-process extensions replace Python runtime extensibility; they are not a
back door around the shared core.

### Contract checklist

- [ ] Versioned authenticated REST, gRPC, and/or event contract.
- [ ] Explicit Principal/authorization behavior for requested operations.
- [ ] Tenant/object visibility and data-minimization rules.
- [ ] Idempotency key and replay semantics.
- [ ] Delivery guarantee (at-most-once, at-least-once, or explicitly bounded
      best effort).
- [ ] Ordering boundary and concurrency behavior.
- [ ] Retry schedule, timeout, cancellation, dead-letter, and manual replay.
- [ ] Payload versioning and consumer compatibility window.
- [ ] Signature/key rotation and secret-storage procedure.
- [ ] Rate, size, destination, and SSRF controls.
- [ ] Observability, alert, support, and failure owner.
- [ ] Consumer contract suite and sandbox fixture.

### Transaction rule

Core state commits before asynchronous delivery. Required delivery intent is
written to a transactional outbox in the same application transaction. The
worker performs network I/O afterward. An extension cannot load Python into the
process, redefine a core invariant, bypass RBAC, or write PostgreSQL directly.

### Migration inventory

For every deployment-specific Python plugin, script, or report:

- [ ] identify the operator/business workflow;
- [ ] classify it as unnecessary, core product behavior, external automation,
      event consumer, or unsupported;
- [ ] define the authorized replacement contract and data boundary;
- [ ] migrate/retest outside the Go process; and
- [ ] document the intentional gap when no replacement is offered.

## 14. Production-readiness program

Production hardening can be developed in parallel when it does not destabilize
profile recovery, but no production claim closes before compatibility and the
following program are signed off.

### PROD-1 — versioned schema lifecycle

- [ ] Introduce explicit ordered schema versions with immutable checksums.
- [ ] Run migrations as a controlled deployment step, separate from ordinary
      application startup.
- [ ] Acquire a migration lock and fail on unknown, modified, partially
      applied, or out-of-order versions.
- [ ] Define install, upgrade, rollback/forward-recovery, and operator
      observability.
- [ ] Prefer expand → migrate/backfill → contract for rolling deployments.
- [ ] Test clean install, N-1 upgrade, failed step recovery, concurrent
      migrator refusal, and mixed-version application deployment.
- [ ] Disable `AutoMigrate` in production.
- [ ] Preserve missing-table-only AutoMigrate only for disposable
      development/test environments.
- [ ] Never use AutoMigrate for an existing table, column inspection, drift
      repair, data backfill, or production upgrade.

### PROD-2 — data migration and cutover

- [ ] Define a repeatable, offline-capable export/import contract for an
      existing NetBox deployment.
- [ ] Pin the supported source versions and effective settings.
- [ ] Inventory plugins/custom fields/content types and reject unsupported
      input before writes.
- [ ] Migrate in dependency order with stable ID preservation or an explicit
      auditable ID map.
- [ ] Keep passwords/tokens/session material under an accepted security
      migration policy; never dump secrets to logs.
- [ ] Validate counts, keys, relationships, hierarchy, IP/network values,
      permissions, object changes, and representative REST responses.
- [ ] Make import restartable/idempotent or define exact cleanup after failure.
- [ ] Rehearse abort, rollback, and final reconciliation.
- [ ] Define write-freeze window, delta strategy, RPO, and go/no-go authority.

Do not point the Go application at an existing NetBox database and assume table
name similarity makes an in-place upgrade safe.

### PROD-3 — security hardening

- [ ] Threat-model browser, REST, gRPC, CLI, migration/import, jobs, webhooks,
      files/templates, and Extension Services.
- [ ] Enforce TLS on public boundaries and mTLS where the deployment trust
      model requires it.
- [ ] Use external secret management, documented rotation, and least-privilege
      runtime/database identities.
- [ ] Define password, session, API-token, signing-key, certificate, and
      recovery lifecycles.
- [ ] Enforce request/header/body/upload limits, timeouts, secure headers,
      origin policy, and rate limits.
- [ ] Fuzz decoders, filter/query parsing, authentication input,
      `net/netip` values, archives, URLs, and templates.
- [ ] Scan dependencies, images, source, and infrastructure; triage findings
      with an owner and deadline.
- [ ] Complete an independent security review and close release-blocking
      findings.

### PROD-4 — operability

- [ ] Define SLIs/SLOs for availability, latency, errors, DB saturation,
      queue age, delivery failures, and data freshness.
- [ ] Emit structured credential-free logs with request/trace correlation.
- [ ] Emit metrics and traces at application, PostgreSQL, queue, and extension
      boundaries without high-cardinality secrets.
- [ ] Export append-only audit data with retention/access policy.
- [ ] Implement truthful liveness/readiness/startup probes.
- [ ] Implement graceful shutdown, connection draining, bounded work, and
      resource limits.
- [ ] Create alerts with a named responder and actionable runbook.
- [ ] Write runbooks for startup, migration, rollback, secret compromise,
      token/session revocation, DB degradation, queue recovery, failed webhook,
      backup restore, and capacity exhaustion.

### PROD-5 — reliability and performance

- [ ] Build representative production-sized and pathological datasets.
- [ ] Set per-interface latency, throughput, concurrency, and payload budgets.
- [ ] Add query-count/no-N+1 assertions for critical lists and nested views.
- [ ] Load-test reads, mutations, filters, allocation, authentication, and
      browser workflows.
- [ ] Run long soak and burst tests with leak/resource monitoring.
- [ ] Inject PostgreSQL disconnect/failover, slow queries, queue worker death,
      extension timeout, disk pressure, and network partition.
- [ ] Prove retry/idempotency behavior does not duplicate mutations/events.
- [ ] Define and test backup schedule, restore verification, RPO, RTO, and
      disaster-recovery failover.

### PROD-6 — supply chain and deployment

- [ ] Reproducible builds from pinned toolchains and lockfiles.
- [ ] Minimal non-root images with read-only filesystem and dropped
      capabilities where practical.
- [ ] SBOM, provenance, signing, vulnerability policy, and license policy.
- [ ] No compiler/package-manager credentials in final images.
- [ ] Environment configuration schema, safe defaults, and secret references.
- [ ] Staging, canary/rolling strategy, compatibility-aware migration order,
      rollback automation, and production-image smoke in connected CI.
- [ ] Capacity/resource requests, autoscaling constraints, topology, and
      disruption budgets are documented and tested.

### PROD-7 — release and cutover

- [ ] Build a release candidate and stop changing it except for reviewed
      release blockers.
- [ ] Run the full current repository, security, domain, PostgreSQL,
      compatibility, gRPC, deployment, browser, performance, supply-chain, and
      operations suites on that candidate.
- [ ] Check every published API version for an accidental breaking change.
- [ ] Complete backup and rehearse the exact migration/cutover.
- [ ] Perform shadow/differential validation where feasible.
- [ ] Execute write freeze, export/import, reconciliation, smoke, and
      limited-user workflow.
- [ ] Hold an explicit go/no-go and rollback decision point.
- [ ] Monitor the defined post-cutover window with named support ownership.
- [ ] Publish versioned release notes listing supported profiles and every
      intentional exclusion.

## 15. Verification command ladder

Run the smallest relevant layer first, then every broader gate required by the
change. Commands below assume repository root unless they `cd` explicitly.

### L0 — static local feedback

```bash
cd netbox-backend
gofmt -l path/to/changed.go
go test ./path/to/changed/package -count=1

cd ../netbox-frontend
npm run format:check
npm run lint
```

Use `gofmt -w` or `npm run format` only on intended files/areas and review
the changes. Do not use a mutating whole-tree formatter in a workspace with
unrelated edits.

### L1 — affected typed layers

```bash
cd netbox-backend
go test ./internal/domain/... ./internal/application/... -count=1
go test ./internal/adapters/rest/netbox/... ./internal/adapters/grpc/... -count=1

cd ../netbox-frontend
npm run typecheck
npm run typecheck:test
npm run test -- path/to/affected.test.ts
```

Narrow those package globs to the affected module when possible.

### L2 — backend/frontend gates

```bash
make -C netbox-backend check
make frontend-check
```

The frontend command requires pinned Node/npm and runs `npm ci`.

### L3 — deterministic repository V0

```bash
make check
```

### L4 — real PostgreSQL

```bash
cd netbox-backend
# Replace the placeholder with an actual owned disposable PostgreSQL DSN.
NETBOX_TEST_POSTGRES_DSN='<disposable-postgres-dsn>' \
  go test \
    ./internal/database \
    ./internal/adapters/postgres/bootstrap \
    ./internal/adapters/postgres/dcim \
    ./internal/adapters/postgres/dcim/row \
    ./internal/adapters/postgres/ipam \
    ./internal/adapters/postgres/ipam/row \
    ./internal/adapters/postgres/changelog \
    ./internal/application/dcim \
    ./internal/application/ipam \
    ./internal/application/identity \
    ./cmd/netbox_go_admin \
    -count=1
```

The command names the explicit DCIM, IPAM, and changelog row owners; the retired
generic PostgreSQL workflow package must not be restored. Evidence records a
redacted DSN descriptor, never the credential.

### L5 — strict REST T2

```bash
make compatibility-comparator-test
make compatibility-test
```

### L6 — gRPC T3

```bash
cd netbox-backend
go test ./test/parity -count=1
```

Run only after the corresponding REST scenarios have T2.

### L7 — standalone deployment

```bash
make deployment-smoke
```

Run `make deployment-image-smoke` in connected CI for the production
multi-stage image and pinned base images.

### L8 — real-browser T4

```bash
make browser-e2e
```

### L9 — tested-digest confirmation and claim attestation

```bash
make source-digest
make check
```

The digest before and after external gates must match; call it the
`tested_digest`. Evidence/status/tier links are then added under the
two-digest protocol below, producing an `attestation_digest`. Run
`make check` again and record both. If code, public contract behavior,
scenario, comparator, normalizer, fixture, or security policy changes between
them, the claim-only exception does not apply and affected external evidence
must be rerun.

## 16. Evidence procedure

### Two-digest claim attestation

Evidence promotion necessarily edits Status/profile metadata, and
`make source-digest` includes those files. Avoid an endless
run → document → changed digest → rerun loop as follows:

1. Compute `tested_digest`, run every required gate without changing owned
   source, and confirm the digest is unchanged afterward.
2. Retain the external artifacts against `tested_digest`.
3. Make only reviewed claim-only edits:
   - evidence summaries/ledger links under `docs/evidence/`;
   - tier/evidence-link/contract-state metadata for the exact covered
     capability;
   - generated inventory/contract documentation resulting solely from that
     metadata; and
   - conservative claim text in `docs/STATUS.md`.
4. Produce a file-and-content-hash manifest of that post-test diff. Human review
   must confirm it changes no route, field behavior, schema, code, scenario,
   fixture, comparator, normalizer, permission/security rule, or toolchain.
5. Compute `attestation_digest`, run `make check`, and record its result.
6. Write a credential-free attestation under `docs/evidence/` mapping
   `attestation_digest` to `tested_digest`, the reviewed diff manifest, and
   the external artifact locations.

The released/claimed revision is `attestation_digest`; behavioral external
results attest `tested_digest` through that reviewed mapping. Any edit outside
the claim-only list, or any ambiguity about its effect, requires rerunning the
affected external gates.

### Before a retained run

- [ ] Start from the exact intended source digest/revision.
- [ ] Use the pinned toolchain and record versions.
- [ ] Use an owned disposable process/database/project with deterministic
      fixtures.
- [ ] Record non-secret effective configuration and oracle SHA.
- [ ] Ensure old services cannot be mistaken for the owned service.
- [ ] Set start time and expected scenario/profile scope.

### During the run

- [ ] Capture exact command, exit status, start/end time, scenario totals, and
      failures.
- [ ] Capture durable-state, rollback, and side-effect results required by the
      profile.
- [ ] Redact credentials at source; do not rely only on post-processing.
- [ ] Retain useful diagnostics for failure, but do not promote a failed gate.

### After the run

- [ ] Verify no password, cookie, CSRF value, bearer/API token, hash, DSN, or
      full configuration appears.
- [ ] Store the artifact in durable CI storage or a concise repository summary,
      never solely under `/tmp`.
- [ ] Link source digest, command, toolchain, config class, scenario coverage,
      and location from the evidence ledger.
- [ ] Link the result to exact capability/profile rows.
- [ ] Promote only covered rows and keep failures/deferred work visible.
- [ ] Have a reviewer verify normalizers and artifact integrity.
- [ ] When claim metadata changes the digest, retain the two-digest attestation
      and final `make check` result.

Recommended concise summary name:
`docs/evidence/YYYY-MM-DD-<profile>-<gate>.md`. Large raw logs belong in
durable CI storage with a stable credential-free link and retention policy.

## 17. Increment specification for GPT-5.6 Luna

Give the executor one block in this form. Do not hand it “implement the next
module” without a bounded acceptance unit.

```text
Increment ID:
Profile and capability:
Outcome:

Authoritative inputs:
- Capability Profile/resource metadata:
- Baseline source locations:
- Canonical ADR/standards sections:
- Existing tests/evidence:

Entry conditions:
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
- Positive:
- Validation:
- Authorization/visibility:
- Not found/conflict:
- PATCH presence:
- Rollback/concurrency:
- State/object changes:
- REST/gRPC equivalence:
- Vue/browser, if applicable:

Commands, in order:
1.
2.

Exit conditions:
- All listed commands green.
- No broader behavior change.
- Documentation/evidence updated only to proved tier.

Completion report:
- Files changed and why.
- Commands and exact outcomes.
- Source digest.
- Deferred/skipped external gates.
- Residual risk and next increment.
```

### Required executor preamble

Place these constraints at the top of every implementation prompt:

- Do not broaden the current Capability Profile.
- Do not add features while V0 is red or unretained.
- Do not reintroduce a generic workflow package, constructor, row owner, port,
  or raw Managed Object application contract.
- Never add business logic to generated or frozen scaffolding.
- REST and gRPC call the same typed use case; neither calls the other or
  persistence directly.
- One mutation uses one application-owned PostgreSQL transaction.
- Preserve absent/null/zero/empty/concrete semantics.
- Fail closed on auth, permissions, filters, fields, identifiers, and routes.
- Never weaken a test, comparator, normalizer, security control, lint rule,
  coverage rule, or generated-drift check to pass.
- Never log or retain secrets in evidence; return only the documented hardened
  session cookie, CSRF bootstrap cookie/body value, and one-time token creation
  material.
- Never alter, inspect for repair, or backfill an existing table through
  AutoMigrate.
- Never promote status without current durable boundary evidence.
- Do not delete more legacy scaffolding before ADR 0004's completion gate.
- Stop and report if a new exception or hard-to-reverse decision is required.

## 18. Definition of done

### For one implementation increment

- [ ] Declared outcome is fully implemented with no scope creep.
- [ ] Domain/application boundaries are typed and dependency direction passes.
- [ ] Lowest-layer positive and negative regression tests pass.
- [ ] Authorization, visibility, transaction, rollback, and object changes are
      covered where applicable.
- [ ] REST/gRPC use the same operation.
- [ ] PostgreSQL-specific behavior is proved on PostgreSQL when touched.
- [ ] Vue boundary is typed and component tests pass when touched.
- [ ] Generated outputs came only from owned inputs and reproduce.
- [ ] Required command ladder is green.
- [ ] No new warning, skip, exclusion, secret, binary, build output, or
      unrelated change.
- [ ] Documentation and status say exactly what the evidence proves.

### For one Capability Profile

- [ ] Definition of ready was complete before behavior code.
- [ ] Every included baseline REST operation is T2.
- [ ] Every equivalent gRPC operation is T3 through the shared core.
- [ ] Shared identity/RBAC applies to both.
- [ ] Success, validation, permission, hidden/not-found, conflict, PATCH,
      rollback, state, and side effects are retained.
- [ ] Applicable operator workflows are T4.
- [ ] Extensions/divergences have separate accepted evidence.
- [ ] Every omission is explicit.
- [ ] Displaced legacy artifacts are retired only after completion.
- [ ] All earlier profiles remain green on the tested digest and are mapped to
      the final claim revision by the reviewed attestation protocol.

### For the complete in-scope replacement

- [ ] All 131 in-scope baseline resources and 22 in-scope custom actions are
      covered by accepted profiles.
- [ ] Every in-scope REST capability is T2.
- [ ] Every equivalent gRPC capability is T3 through shared typed application
      services.
- [ ] Every declared browser workflow is T4.
- [ ] Every module closeout is complete, including cross-module fields,
      filters, relationships, bulk shapes, and nested projections.
- [ ] Extensions retain separate contract/parity/security status.
- [ ] Frozen direct-GORM REST and table-oriented gRPC artifacts have zero
      runtime/source ownership wherever displaced.
- [ ] Runtime REST equals accepted profiles plus named
      health/readiness/schema/identity surfaces; its canonical route inventory
      contains no dedicated `GET /ping`.
- [ ] gRPC reflection lists only canonical published services and health.
- [ ] Vue exposes only declared workflows.
- [ ] GraphQL, Python runtime/plugins/scripts/reports,
      `/api/extras/scripts/`, and anonymous token provision remain explicit
      exclusions unless a later accepted decision changes them.
- [ ] No documentation claims broader behavior than evidence.

Call this outcome the **complete in-scope NetBox-compatible replacement**, not
an unqualified clone.

### For production release

- [ ] The complete in-scope replacement gate is signed off.
- [ ] PROD-1 through PROD-7 are retained green on the release candidate.
- [ ] Upgrade/import, backup/restore, rollback, DR, security, performance, and
      operations are rehearsed.
- [ ] Support ownership, SLOs, runbooks, version policy, and exclusions are
      published.
- [ ] Gate 7 has a reviewed go/no-go decision.

## 19. Forbidden shortcuts and hard-stop review findings

Reject an increment that does any of the following:

- translates Python classes or database tables one-for-one without a Capability
  Profile and behavior scenarios;
- implements a route by exposing a GORM row or generic map;
- adds a generic CRUD repository/service/router for speed;
- duplicates validation/permissions in REST and gRPC;
- makes one public transport call the other;
- guesses a status, default, validation reason, choice label, filter, or
  transaction outcome;
- collapses 401/403, absent/null, numeric types, missing/extra fields, state, or
  side effects in a normalizer;
- catches a failing test by loosening its assertion or adding a sleep;
- uses SQLite/mock results as proof of PostgreSQL behavior;
- calls AutoMigrate on an existing table or adds drift repair;
- stores password/token/session-cookie/CSRF secret material or a DSN in source,
  argv, logs, evidence, or Web Storage (documented runtime cookie/CSRF/token
  responses are the only response exception);
- stores permission or session state in browser Web Storage; non-secret
  permission identifiers remain valid in source, tests, security evidence, and
  protected CLI input;
- places authorization or semantic validity in Vue;
- hand-edits generated/frozen output;
- recreates a retired generic workflow package, constructor, row owner, port,
  or raw Managed Object application contract;
- repins a toolchain to match the current host;
- marks a capability T2/T3/T4 without the correct retained external boundary;
- starts Gate 5 while first-profile V6 is open;
- calls a module complete while cross-module behavior is still deferred; or
- deletes displaced legacy artifacts before ADR 0004 permits retirement.

When one appears, stop that increment, cite the rule ID, restore the narrow
accepted path without discarding unrelated work, and rerun the lowest test that
proves the correction.

## 20. Implementation blueprints

These blueprints constrain shape without replacing capability-specific design.
Copy the pattern, never another resource's business rules.

### One-operation data flow

```text
REST request DTO ──explicit mapper──┐
                                    │
                                    v
                            typed application use case
                            (context + Principal +
                             command/query)
                                    │
gRPC request ─────explicit mapper───┘
                                    │
                         authorize and apply visibility
                                    │
                       application-owned unit of work
                                    │
          typed repository ports + domain transitions + change recorder
                                    │
                         typed result or typed error
                                    │
              ┌─────────────────────┴─────────────────────┐
              v                                           v
       exact REST response                         typed gRPC response

Vue feature API ──REST only──> exact REST surface
```

No arrow may skip the application use case and no public transport may call the
other.

### Resource file ownership

A typical resource should have deliberately small files with one owner:

```text
internal/domain/<module>/
  <resource>.go             value objects, aggregate/entity, invariants
  <resource>_test.go

internal/application/<module>/
  <resource>_commands.go    create/replace/patch/action inputs
  <resource>_queries.go     get/list query and pagination/filter types
  <resource>_ports.go       consumer-owned repository contracts
  <resource>_service.go     authorization and orchestration
  <resource>_service_test.go

internal/adapters/postgres/<module>/
  <resource>_row.go         private row and mapping
  <resource>_repository.go  explicit projections/queries/locks
  <resource>_postgres_test.go

internal/adapters/rest/netbox/<module>/
  <resource>_dto.go
  <resource>_mapper.go
  <resource>_handler.go
  <resource>_test.go

internal/adapters/grpc/<module>/
  <resource>_mapper.go
  <resource>_handler.go
  <resource>_test.go
```

Combine files when the result is clearer; do not force one file per box or
create empty architectural ceremony. Keep ownership and dependency direction.

### Command and presence blueprint

- Create contains values/default decisions allowed at creation.
- Replace represents PUT and deliberately applies baseline replacement/default
  behavior.
- Patch contains a typed field state: omitted, explicit null where nullable, or
  present value.
- An empty string, zero, false, or empty list is a present value when allowed.
- A required relationship may reject explicit null but must still distinguish
  it from omission to produce the exact validation reason.
- REST decoding owns wire presence; protobuf presence/FieldMask owns RPC
  presence; both map to the same application field state.
- Domain values represent validated meaning. Do not carry JSON
  `null`, protobuf wrappers, pointers used as accidental flags, or
  `map[string]any` inward.

Use one precisely named typed application-presence abstraction across modules.
It must not depend on either transport, and its state cannot be forged through
an unvalidated zero value.

### List-query blueprint

A typed list query declares only supported fields:

- exact ID set and relationship identifiers;
- search term only where the baseline supports `q`;
- typed choices and domain values;
- explicit ordering enum/list, including direction;
- limit value plus whether it was supplied;
- offset; and
- any exact network/hierarchy predicates.

The use case authorizes the resource, asks an optional complete typed
list-scope authorizer for visible IDs, and passes a bounded typed query to the
repository. The repository applies exact SQL-portable predicates,
deterministic ordering, visibility, count, limit, and offset. If a predicate or
authorization rule cannot be expressed exactly in SQL, use the documented
authorization-before-pagination hybrid; never substitute an approximate
database operator.

### Mutation transaction blueprint

1. Authenticate before the use case and pass Principal explicitly.
2. Begin one unit of work.
3. Authorize the action and object scope.
4. Load current and related state using transaction-bound repositories.
5. Validate expected version/precondition when the baseline has one.
6. Apply the domain transition.
7. Persist the aggregate and required dependent objects.
8. Compute/update included derived state.
9. Write the exact required object-change entry and, only when the profile
   declares durable delivery, its outbox intent.
10. Commit.
11. Perform cache invalidation or external delivery after commit.
12. Map the typed result/error at each transport edge.

Inject a deterministic late failure in tests after the primary write but before
commit. Assert that primary, dependent, derived, audit, and any applicable
declared outbox state all rolled back.

### Typed polymorphic-reference blueprint

Capabilities such as contacts, services, cable terminations, tags, journals,
and VPN terminations require cross-object references. Use a closed typed
registry:

- each allowed target has a stable content type and typed ID;
- the profile declares the allowed target set;
- target resolution uses an owner-provided port;
- authorization checks the referenced object;
- persistence enforces uniqueness/integrity as far as PostgreSQL can;
- application validation handles cross-table existence/type rules;
- response mapping uses an explicit brief projection; and
- unknown/disabled content types fail closed.

Do not implement polymorphism with arbitrary table names, reflection over GORM
models, SQL fragments, or unchecked app/model strings from clients.

### Object-change blueprint

Each included mutation decides from the baseline:

- action kind;
- object type/ID/representation;
- actor and request/correlation metadata;
- before/after snapshots and redaction;
- creation time/order; and
- whether dependent changes receive separate records.

The application requests the record; the transaction-bound adapter persists
it. Transports never manufacture audit identity, and snapshots never include
passwords, tokens, sessions, DSNs, or unapproved internal fields.

### Error blueprint

Domain/application code returns a stable reason plus safe field violations.
Central REST and gRPC maps decide transport status/code and envelope/details.
Repositories wrap internal causes, but public mapping emits safe text.

Tests assert all three levels:

1. typed reason/field at application boundary;
2. exact REST status/envelope/reason against the oracle; and
3. equivalent gRPC code/details and unchanged durable state.

### Vue feature blueprint

A feature owns:

- exact wire DTOs;
- typed list/detail/form/filter/mutation models;
- pure wire ↔ UI mapping;
- a typed API module/composable;
- pages/components for the workflow;
- normalized error and permission-aware presentation;
- colocated adapter/component tests; and
- a real-browser scenario.

The feature does not own domain validity, authorization, endpoint strings
outside API infrastructure, Axios configuration, credential storage, or a
parallel copy of server choices/invariants that can silently drift. It may
provide immediate accessibility/usability feedback, but the server response
remains authoritative.

### Review order

Review a vertical increment in this order:

1. profile/scenarios;
2. domain invariants and types;
3. application authorization/transaction;
4. PostgreSQL integrity/query plan;
5. REST exactness;
6. gRPC equivalence;
7. Vue workflow;
8. generation/composition;
9. tests/evidence/status.

This order makes it harder for a convenient transport or storage shape to
dictate the semantic center.
