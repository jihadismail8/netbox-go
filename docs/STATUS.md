# Project status

Updated: 2026-08-19

Compatibility baseline: checked-in NetBox source snapshot at commit
`fbb948d30e79ce657fac62994a22aca72c1770a9`
(`v4.4.6-7-gfbb948d30`).

This is an evidence ledger, not a single percentage estimate. It includes
bounded ratios only where the denominator and tier are explicit. Source code
and test harnesses show what can be exercised; a gate is passed only by a
retained artifact from the current revision. The tier definitions in
[Compatibility](COMPATIBILITY.md) remain authoritative.

## Summary

The repository now contains a standalone Go/Vue implementation of the first
13-resource [Capability Profile](../CONTEXT.md). The deployed application does
not build, migrate, start, or run through Python, Django, or the checked-in
upstream source. That source is used only by the differential development
oracle.

The default Go process exposes profile-declared REST intended to become
exact-in-profile and versioned gRPC over one shared application path. The 13
Managed Objects have private typed
PostgreSQL tables, the Vue surface has typed per-resource adapters, and users,
groups, memberships, model grants, and object-scoped grants are persisted by
the Go-owned identity implementation. The displaced legacy stacks for all 13
resources have been physically removed.

The profile nevertheless remains **T1 and pre-publication**. The source-digest
algorithm is versioned `source-v2` and includes executable mode plus closed
owned symlinks. The exact entry revision now has a human-reviewed retained
`CW1-G00` result, and `CW1-V1-01` trusted-origin CORS is done. Bounded
`CW1-V1-02-I1` is accepted and done for token lookup classification, strict
touch ordering, revocation containment, and durable PostgreSQL touch semantics;
`CW1-V1-02-I2` is accepted and done for exact baseline REST token
grammar/outcomes, typed application credential causes, strict unary gRPC
bearer parsing, and fail-closed unary method safety. `CW1-V1-02-I3` is accepted
and done for typed browser-session outcomes, valid-session-first REST
arbitration, exact CSRF pairing/recovery, transactional login/logout, and
cookie shape. `CW1-V1-02-I4` is now the sole active identity child. Its exact
tested password-change/session candidate is
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f`, source digest
`source-v2:sha256:3f37417bb791ac6bc97ac4a0d23c5f928062feecf81a0f8a4fb9e57445d53670`
with 3,013 entries. Its accepted-I3-based red-first, focused, race,
real-PostgreSQL, complete L4, pinned repository, feature-candidate CI, and main
exact-SHA CI boundaries are green. The current nine-path claim-only revision
conditionally moves I4 to `evidence`: that state becomes effective only after
this exact revision passes repository CI. The digest-excluded receipt and
project-owner review then remain. Its parent and the
password-policy, throttle, trusted-proxy, streaming, and aggregate rows remain
open. The 293-row
`CW1-V2-01` structural authority has human acceptance, while its 275 unresolved
and 18 contradicted behavior rows remain open. V1-V5 contain implementation and
evidence gaps. No T2, T3, T4, V6 sign-off, or publication is claimed until
corresponding current artifacts are retained.

> **Operational boundary:** this is a development build for disposable data.
> Production TLS, schema upgrades, backup/restore, operational hardening, and
> V1-V6 evidence are not complete. Do not expose it as a production service.

## Quantitative rewrite audit

There is no defensible single completion percentage. Catalogue breadth,
implemented structure, verified behavior, and production readiness are
different dimensions and must remain separate.

| Dimension                     | Accepted denominator                    | Current state                                             | Remaining                                              |
| ----------------------------- | --------------------------------------- | --------------------------------------------------------- | ------------------------------------------------------ |
| High-level baseline catalogue | 155 resource/action entries             | 155 catalogued (100%)                                     | No route/action discovery entries                      |
| Accepted in-scope baseline    | 153 entries                             | 13 T1 entries (8.50%)                                     | 140 T0 entries (91.50%)                                |
| In-scope resources            | 131 resources                           | 13 T1 resources (9.92%)                                   | 118 T0 resources                                       |
| In-scope custom actions       | 22 actions                              | 0 promoted actions                                        | 22 T0 actions                                          |
| REST compatibility            | Every promoted in-scope capability      | 0 retained T2 capabilities                                | All accepted capabilities as their profiles are opened |
| gRPC semantic parity          | Every corresponding promoted capability | 0 retained T3 capabilities                                | All corresponding accepted capabilities                |
| Browser Workflow Parity       | Every declared browser workflow         | 0 retained T4 workflows                                   | First-profile and all later applicable workflows       |
| Production readiness          | PROD-1 through PROD-7                   | 0 signed-off programs; foundations exist in several areas | All seven exit programs                                |

The catalogue's 100% is discovery at resource/action level, not business-rule
coverage. Frozen generated artifacts are not counted as implementation. The
two explicit exclusions are `/api/extras/scripts/` and anonymous
`/api/users/tokens/provision/`; GraphQL and Python plugin/script/report runtime
compatibility are also outside the accepted runtime boundary.

### Remaining breadth by module

Counts below combine resources and custom actions because those are the units
in the authoritative 155-entry baseline inventory.

| Module         | In-scope entries | T1 now | T0 remaining |
| -------------- | ---------------: | -----: | -----------: |
| Circuits       |               11 |      0 |           11 |
| Core           |               13 |      0 |           13 |
| DCIM           |               54 |     10 |           44 |
| Extras         |               22 |      0 |           22 |
| IPAM           |               23 |      3 |           20 |
| Tenancy        |                6 |      0 |            6 |
| Users/Identity |                5 |      0 |            5 |
| Virtualization |                6 |      0 |            6 |
| VPN            |               10 |      0 |           10 |
| Wireless       |                3 |      0 |            3 |
| **Total**      |          **153** | **13** |      **140** |

The 140 deferred entries break down as 102 frozen runtime-disabled REST
configurations, 16 in-scope resources with no current REST configuration, and
22 custom actions absent from the current REST inventory. Generated
counterparts do not advance a tier.

No module is complete. Passing `core-workflow-v1` will verify only its declared
slice; it will still leave 44 DCIM and 20 IPAM entries for later profiles and
module closeout.

### Business-logic coverage

The first profile has accepted metadata for 13 resources: 78 CRUD
resource/operation combinations, 2 gRPC assignment actions, 107 writable
fields, 84 response-only fields, 65 filters, 81 ordering fields, 15
relationships, 19 choice fields, and 17 coarse scenarios. Those are declared
scope, not pass evidence. None of the 17 scenario IDs currently maps through a
complete traceability chain from pinned source to Go test, differential/parity
case, applicable browser case or explicit not-applicable classification, and
retained artifact.

The other 140 in-scope entries still need profile-ready field, presence,
filter, ordering, relationship, validation, permission, transaction, error,
side-effect, and workflow discovery. The narrative files under
[`business-logic/`](business-logic/), [`entities/`](entities/), and
[`CRUD_PARADIGM.md`](CRUD_PARADIGM.md) are partial derived references with
known stale paths and inaccurate relationship/delete claims; they are not an
accepted specification. A precise field/rule-level completion percentage is
therefore not yet measurable and cannot be inferred from the 8.50% T1 breadth
ratio.

### Technical-logic coverage

The 13 first-profile resources have typed domain/application contracts,
application services, private PostgreSQL rows/repositories, canonical REST and
gRPC adapters, and typed Vue adapters. REST and gRPC share the same use cases;
the generic raw-map workflow boundary is retired. This is strong structure for
the first slice, not whole-system completion.

Cross-cutting first-profile work still includes review and retained evidence
for the implemented explicit CORS allowlist, dependency-aware HTTP/gRPC
readiness, rule/test/evidence traceability, missing PostgreSQL concurrency
cases, durable external-gate retention, and protobuf publication controls.
Outside the first slice, 102 direct-GORM REST
configurations, 176 generated table-oriented gRPC services, and their retained
model/DAO/cache/service/protobuf layers are frozen and runtime-disabled. They
may be reused only through reviewed typed replacements and are retired only
after capability completion.

Production technical logic is earlier still: there is no versioned upgrade
engine or import/cutover system, production-safe TLS refusal and secret/CORS
program is incomplete, probes are not dependency-truthful, and the required
operability, performance, backup/restore, supply-chain, and release controls
have not passed their programs.

## Current execution recovery

The interrupted-cutover conditions preserved in the
[historical handoff snapshot](IMPLEMENTATION_EXECUTION_PLAYBOOK.md#historical-handoff-snapshot)
have been structurally recovered. Current implementation observations are:

- production and parity composition use the same typed per-resource services;
- the generic domain, application, and PostgreSQL workflow packages,
  transitional constructors, and `composition.Core.Workflow` are gone, with a
  repository-wide architecture prohibition against their return;
- DCIM, IPAM, and changelog adapters explicitly own their rows, and the moved
  bootstrap, constraint, uniqueness, transaction, and audit-chain suites pass
  against disposable PostgreSQL 16 in recovery diagnostics;
- the canonical router has no dedicated `GET /ping`; frozen runtime-disabled
  legacy source still contains its historical route, and SPA history fallback
  is not a diagnostic endpoint; and
- backend coverage is non-mutating, records an exact package/count baseline,
  rejects regression or exclusion drift, and is part of `make check`.

The historical V0 result covers tested digest
`sha256:a55fab792cea1100e5fd2cc641fad02345189dd27d0b28c3b7ed2b1e1dcc22e1`
and is recorded in the
[recovery artifact](evidence/2026-08-01-core-workflow-v1-v0.md). Because the
current source includes non-claim-only implementation and test changes, the
old claim-attestation mapping was not carried forward. The later
[post-cleanup V0 artifact](evidence/2026-08-03-post-cleanup-v0.md) recorded a
fresh exact-digest pass under the superseded v1 algorithm. It remains useful
historical evidence but cannot attest a source-v2 revision. The
[source-v2 evidence](evidence/2026-08-17-core-workflow-v1-source-v2-v0.md)
retains the exact entry revision, and the
[CORS evidence](evidence/2026-08-03-core-workflow-v1-cors-v0.md) closes
`CW1-V1-01`. V1-V6 remain separate first-profile work.

## Current surface inventory

The generated files under
[`contracts/netbox/v4.4.6-post7/inventory/`](../contracts/netbox/v4.4.6-post7/inventory/)
are authoritative for counts.

| Inventory     | Count | Composition                                                                                    | Runtime meaning                                                               |
| ------------- | ----: | ---------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Baseline REST |   155 | Pinned baseline resources and actions                                                          | Catalogue only; not all are in the first profile                              |
| Current REST  |   123 | 102 frozen legacy resources + 13 canonical profile resources + 8 identity extension operations | Only the 13 canonical resources and 8 identity operations are runtime-enabled |
| Current gRPC  |   179 | 176 frozen legacy services + 3 canonical services                                              | Only `IdentityService`, `DCIMService`, and `IPAMService` are runtime-enabled  |
| Current Vue   |    13 | 10 DCIM + 3 IPAM profile resources                                                             | Runtime-enabled, pre-publication T1                                           |

The 198-table startup registry is a separate persistence inventory:

- 176 deferred legacy row types;
- 8 Go-owned identity rows;
- 13 typed first-profile rows; and
- 1 typed append-only object-change row.

Those counts describe artifacts, not compatible capabilities.

## Verified implementation structure

### Standalone runtime and interfaces

- REST and gRPC start in one Go process and invoke the same typed per-resource
  application services.
- REST is the exact-in-profile compatibility surface. gRPC uses
  transport-native messages while preserving authorization, validation,
  transaction, state, and error meaning.
- The runtime router exposes profile resources, identity, schema, health,
  readiness, and the Vue application. It has no dedicated canonical
  `GET /ping`; frozen runtime-disabled source retains the historical route, and
  `/ping` may only be an ordinary SPA history fallback. Frozen direct-GORM REST
  and generated table-oriented gRPC services are not registered.
- Python/Django is confined to [`tests/compatibility/`](../tests/compatibility/)
  and is never a production or development-runtime dependency of the rewrite.

### PostgreSQL and schema lifecycle

- [`migrate.go`](../netbox-backend/internal/database/migrate.go) builds an
  ordered 198-entry registry.
- Bootstrap checks `HasTable` and calls GORM `AutoMigrate` only for an absent
  table. Existing tables are not passed to `AutoMigrate`, inspected for missing
  columns, repaired, or backfilled.
- [`dcim/row/rows.go`](../netbox-backend/internal/adapters/postgres/dcim/row/rows.go),
  [`ipam/row/rows.go`](../netbox-backend/internal/adapters/postgres/ipam/row/rows.go),
  and [`changelog/row.go`](../netbox-backend/internal/adapters/postgres/changelog/row.go)
  own the ten DCIM rows, three IPAM rows, and typed change row. Explicit
  mappings, constraints, indexes, and PostgreSQL-focused tests replace the
  former JSON-table persistence design.
- Profile list reads push exact portable filters, visibility constraints,
  count, ordering, limit, and offset into SQL. Predicates whose semantics are
  not portable (including free text, network containment, and collation-bound
  ordering) use an authorization-before-pagination hybrid path rather than an
  approximate SQL shortcut.
- Development and test databases are disposable. A shape change is handled by
  dropping and recreating the database; production upgrade migrations remain
  deferred.

### Identity and authorization

- Go-owned rows persist users, groups, memberships, permission grants,
  user/group grant links, sessions, and API tokens.
- The shared principal combines direct, group, global, and object-scoped
  grants. List visibility is applied before count and pagination.
- Browser sessions use cookie/CSRF mechanics; REST automation uses NetBox-style
  API tokens; gRPC resolves bearer metadata through the same identity path.
- Administrator bootstrap, password reset, non-superuser creation, and global
  model-permission grants are protected CLI operations, not anonymous HTTP or
  RPC endpoints. Provisioning commands authenticate an existing active
  superuser and keep passwords off argv and SQL logs.

### Vue boundary

- The supported route/model registry contains exactly the 13 profile resources.
- [`features/core/adapters.ts`](../netbox-frontend/src/features/core/adapters.ts)
  owns typed DTO, form, mutation, and filter conversion for every resource.
- Read-side choice DTOs preserve NetBox's `{value, label}` wire envelope while
  forms and mutations deliberately unwrap to scalar choice values.
- Session/CSRF authentication is used and browser credentials are not stored in
  `localStorage`.
- A real-Chrome/CDP harness exists under
  [`tests/browser/`](../tests/browser/), but its current successful execution is
  pending evidence and therefore does not earn T4.

### Physical legacy retirement and wrapper cleanup

The first-profile model, DAO, cache, service, handler, router, and legacy
protobuf/generated stacks are removed. This retirement happened historically
while the profile was still T1, earlier than ADR 0004's capability-completion
condition. Do not restore those paths solely to recreate the intended ordering,
and do not use this deviation as precedent.

[ADR 0005](adr/0005-retire-dormant-sponge-http-wrappers.md) separately
authorized a narrow source-hygiene cleanup: exactly 118 untouched,
runtime-dormant Sponge per-resource handler delegates and their 118 matching
top-level route wrappers were removed. They contained no hand-owned behavior,
were absent from production composition, and did not define the authoritative
REST inventory. Their removal changes no capability classification, tier,
completion, or publication claim.

The separate 102 direct-GORM REST configurations and 176 generated gRPC
services still correspond only to deferred resources. Their models, DAOs,
caches, services, and protobufs also remain frozen and runtime-disabled for
reuse or reference. No additional displaced stack may be deleted before its
replacement completes ADR 0004's governing gate and CP-13.

## Evidence status

| Checkpoint               | Present foundation                                                                                                                                         | Missing before exit                                                                                                                                         | Status                                     |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| V0 repository gate       | Retained exact-entry source-v2 result plus deterministic quality policy                                                                                    | Refresh for the active source and every later relevant merged digest                                                                                        | Continuous; entry source retained          |
| V1 identity/RBAC         | Persisted identity, retained CORS, accepted bounded I1/I2/I3 evidence, and an exact I4 candidate green through local/PostgreSQL/feature/main CI boundaries | Pass the I4 claim revision's CI, retain its receipt, obtain owner review, and complete password-policy/throttle/proxy/streaming plus RBAC/admin/secret work | I4 evidence claim conditional; parent open |
| V2 domain behavior       | Typed shared services, broad focused tests, and reviewed machine-readable 293-row traceability                                                             | Uncovered presence, invariant, rollback, side-effect, and concurrency cases                                                                                 | V2-01 done; behavior/evidence pending      |
| V3 PostgreSQL/deployment | Typed schema, real-PostgreSQL suites, and Compose harness                                                                                                  | Dependency-aware HTTP/gRPC readiness, dependency loss/recovery, remaining locking cases, retained external run                                              | Implementation and evidence pending        |
| V4 REST/gRPC             | Strict comparator/orchestrator and broad CRUD/parity suites                                                                                                | Required negative/invariant/permission/presence scenarios, durable T2 report, then corresponding T3 report                                                  | T1; implementation/evidence pending        |
| V5 Vue                   | Typed adapters and real-Chrome workflow harness                                                                                                            | Session refresh, edit/filter, rollback, null/conflict/not-found, reassignment, and exact-state browser outcomes                                             | T1; implementation/evidence pending        |
| V6 sign-off              | First-profile legacy stacks physically retired                                                                                                             | V0-V5 retained green together, per-capability review, protobuf freeze/breaking baseline, joint sign-off                                                     | Not earned                                 |

See [Evidence](evidence/README.md) for the artifact policy and commands.

## Remaining first-profile work

1. Keep `CW1-G00` green with pinned Go, Node, and npm.
2. Complete the exact `CW1-V1-02-I4` claim CI, receipt, and owner review, then complete
   the later `CW1-V1-02` children through `CW1-V1-04`: close the credential and remaining
   identity/security matrix, then retain V1.
3. Complete `CW1-V2-02` through `CW1-V2-08`: use the accepted traceability for
   every profile scenario and plan rule;
   implement every uncovered domain, presence, rollback, object-change, and
   PostgreSQL case.
4. Complete `CW1-V3-01` through `CW1-V3-05`: make HTTP and gRPC readiness
   dependency-aware, add loss/recovery and missing concurrency cases, then
   retain the real-PostgreSQL and deployment V3 bundle.
5. Complete `CW1-V4-01` through `CW1-V4-06`: expand the strict REST oracle from
   broad CRUD/filter coverage to every
   declared positive and negative rule; earn per-capability T2, then prove the
   same outcomes through gRPC for T3.
6. Complete `CW1-V5-01` through `CW1-V5-05`: expand the browser suite to all V5
   outcomes, wire durable artifact retention, and earn T4 only for the
   workflows actually exercised.
7. Complete `CW1-V6-01` through `CW1-V6-03`: review V0-V5 together, freeze the
   published protobuf baseline, record V6, and only then open the next
   Capability Profile.

## Remaining complete-replacement program

After first-profile V6, the accepted queue still contains 19 candidate
Capability Profiles covering 118 resources and 22 custom actions. Each must
pass discovery, typed implementation, REST T2, corresponding gRPC T3,
applicable browser T4, evidence retention, and safe legacy displacement. Ten
module-closeout passes must then reconcile every deferred field, action, bulk
shape, cross-module relation, nested projection, and exclusion before any
module is complete.

Production release is a separate final program. PROD-1 through PROD-7 cover
versioned schema lifecycle, data migration/cutover, security, operability,
reliability/performance, supply chain/deployment, and release/cutover. All
seven remain unsigned; current `AutoMigrate` bootstrap is only a disposable
development convenience.

Passing the first profile will not make all of DCIM, IPAM, or NetBox complete.
