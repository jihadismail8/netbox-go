# Project status

Updated: 2026-08-01

Compatibility baseline: checked-in NetBox source snapshot at commit
`fbb948d30e79ce657fac62994a22aca72c1770a9`
(`v4.4.6-7-gfbb948d30`).

This is an evidence ledger, not a percentage estimate. Source code and test
harnesses show what can be exercised; a gate is passed only by a retained
artifact from the current revision. The tier definitions in
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

The profile nevertheless remains **T1 and pre-publication**. Repository V0 and
the recovery PostgreSQL replay are now retained, but the complete V1-V5
security, behavior, differential REST, corresponding gRPC, deployment, and
real-browser boundaries are not. No T2, T3, T4, V6 sign-off, or publication is
claimed.

> **Operational boundary:** this is a development build for disposable data.
> Production TLS, schema upgrades, backup/restore, operational hardening, and
> V1-V6 evidence are not complete. Do not expose it as a production service.

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

The retained V0 result covers tested digest
`sha256:a55fab792cea1100e5fd2cc641fad02345189dd27d0b28c3b7ed2b1e1dcc22e1`
and is recorded in the
[recovery artifact](evidence/2026-08-01-core-workflow-v1-v0.md). Its
claim-attestation mapping is authoritative in [Evidence](evidence/README.md).
V1-V6 remain separate first-profile evidence work.

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

### Physical legacy retirement

The first-profile model, DAO, cache, service, handler, router, and legacy
protobuf/generated stacks are removed. This retirement happened historically
while the profile was still T1, earlier than ADR 0004's capability-completion
condition. Do not restore those paths solely to recreate the intended ordering,
and do not use this deviation as precedent: no additional displaced stack may
be deleted before its replacement completes the governing gate. The remaining
102 REST configurations and 176 gRPC services correspond only to deferred
resources and stay frozen, runtime-disabled, and unpublished.

## Evidence status

| Checkpoint               | Implementation present                                                            | Current durable result                                                        | Status                    |
| ------------------------ | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ------------------------- |
| V0 repository gate       | Backend/frontend quality gates, coverage policy, and deterministic checks exist   | [2026-08-01 recovery artifact](evidence/2026-08-01-core-workflow-v1-v0.md)    | Passed on retained digest |
| V1 identity/RBAC         | Persisted groups, memberships, grants, object visibility, CLI/session/token tests | No retained full security run                                                 | Pending                   |
| V2 domain behavior       | Shared validation/workflows and focused positive/negative tests                   | No complete profile evidence bundle                                           | Pending                   |
| V3 PostgreSQL/deployment | Typed schema, concurrency/bootstrap suites, Compose smoke harness                 | Recovery PostgreSQL replay retained; no complete PostgreSQL/Compose V3 bundle | Pending                   |
| V4 REST/gRPC             | Strict fail-closed oracle comparator and all-resource gRPC parity suites          | No retained current oracle/parity run                                         | T1; pending               |
| V5 Vue                   | Typed adapters and real-browser harness                                           | No retained current browser run                                               | T1; pending               |
| V6 sign-off              | First-profile legacy stacks physically retired                                    | V0-V5 have not been retained green together                                   | Not earned                |

See [Evidence](evidence/README.md) for the artifact policy and commands.

## Remaining first-profile work

- Keep V0 green and retain every result against the exact source digest with
  the pinned Go/Node/npm toolchains.
- Run the strict differential oracle after comparator hardening and retain its
  report. The comparator must not collapse 401/403, validation reasons,
  trailing slashes, missing/extra fields, state, or side effects.
- Run and retain the complete real-PostgreSQL bootstrap, constraint,
  concurrency, rollback, and deployment-smoke results.
- Run and retain the complete gRPC lifecycle/error/RBAC/rollback/assignment
  parity set.
- Run and retain the clean-database real-browser DCIM/IPAM workflows and their
  permission, rollback, validation, assignment, and delete-warning cases.
- Keep the profile pre-publication and at T1 until the evidence above is
  complete and reviewed.

Passing the first profile will not make all of DCIM, IPAM, or NetBox complete.
