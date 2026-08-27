# Project status

Updated: 2026-08-26

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
cookie shape. `CW1-V1-02-I4` has retained evidence. Its exact
tested password-change/session candidate is
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f`, source digest
`source-v2:sha256:3f37417bb791ac6bc97ac4a0d23c5f928062feecf81a0f8a4fb9e57445d53670`
with 3,013 entries. Its accepted-I3-based red-first, focused, race,
real-PostgreSQL, complete L4, pinned repository, feature-candidate CI, main,
claim, and receipt exact-SHA CI boundaries are green. Project-owner review
remains before the separate `evidence` to `done` transition. Its parent and the
password-policy, throttle, trusted-proxy, streaming, and aggregate rows remain
open. The 293-row
`CW1-V2-01` structural authority has human acceptance, while its 280 unresolved
and 13 contradicted behavior rows remain open. V1-V5 contain implementation and
evidence gaps. No T2, T3, T4, V6 sign-off, or publication is claimed until
corresponding current artifacts are retained.

`CW1-V2-02-I1` has an exact tested candidate at
`7acba402f0de2bd59e5b342a6f05df268bc9120b`, source digest
`source-v2:sha256:ed330b0a5bbeafd70b7b16a4ce4d1052fa9a385313a3b8827b554983571c1b43`
with 3,018 entries. It owns a bounded IPAddress scalar write-presence
correction across typed entity/application behavior, generated OpenAPI,
REST/gRPC mapping, Vue validation/serialization, and focused tests. Its
focused, race, real-PostgreSQL, complete L4, generated-contract, Vue, pinned
repository, independent exact-candidate CI, evidence-claim CI, and
pre-acceptance receipt CI boundaries are green. The project owner accepted
only this bounded result at `2026-08-23T17:25:54Z`. Its owner-accepted
closeout claim and excluded closeout receipt also passed exact-SHA CI, so I1 is
effectively `done`. The pinned differential oracle was unavailable, so REST T2
and corresponding gRPC T3 remain unearned. No parent, profile, traceability
consumer, or tier completion is claimed.

`CW1-V2-02-I2` has an exact tested candidate at
`87863efd38fe71dfa05c818b860b37b7e94d67b4`, source digest
`source-v2:sha256:c7c1b86c2bcd768bb719149a54dddbceccf2b5b2e4087dd4b79eec20bef5a37c`
with 3,022 entries. It owns only the six Site scalar create/PUT/PATCH presence
states, operation-specific generated API contracts, matching REST/gRPC
semantics, PostgreSQL durability, Vue dirty-field serialization/validation,
and eight named focused tests. Its exact-name, affected/race, real-PostgreSQL,
complete L4, generated-contract, Vue, pinned repository, and independent
exact-candidate CI boundaries are green. Its evidence-claim CI and
pre-acceptance receipt CI are also green. The project owner accepted only this
bounded result at `2026-08-24T04:46:51Z`. Its owner-accepted closeout claim
and excluded closeout receipt passed exact-SHA repository CI, so I2 is
effectively `done`. The differential harness was unavailable before oracle
execution because Docker rejected its temporary source bind, so REST T2 and
corresponding gRPC T3 remain unearned. It does not own Site uniqueness,
deletion, list/query
behavior, full CRUD, a compatibility tier, the parent, profile promotion, or
a traceability consumer.

`CW1-V2-02-I3` has an exact tested candidate at
`651d33bc3fb2c8e663b6b14320af405b8501471f`, source digest
`source-v2:sha256:09499a6618569d2dae224edfb339ac82585bf0248d20e5a4d5ff23d19221fe6f`
with 3,029 entries. It owns only Manufacturer `name`, `slug`, and
`description` create/PUT/PATCH presence, operation-specific generated API
contracts, matching REST/gRPC semantics, PostgreSQL durability, Vue
dirty-field serialization/validation, and eight fixed focused tests. Its
exact-name, affected/race, real-PostgreSQL, complete L4, generated-contract,
Vue, pinned repository, and independent exact-candidate CI boundaries are
green. Its evidence-claim CI and pre-acceptance receipt CI are also green. The
project owner's acceptance of only this bounded result was recorded at
`2026-08-24T10:28:34Z`. Its owner-accepted closeout claim and excluded
closeout receipt both passed exact-SHA repository CI, so I3 is effectively
`done`. No retained pinned differential accompanies this bounded result, so
REST T2 and corresponding gRPC T3 remain unearned. It does not own
Manufacturer uniqueness, deletion, list/query behavior, full CRUD, a
compatibility tier, the parent, profile promotion, or a traceability consumer.

`CW1-V2-02-I4` has exact tested candidate
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe` at source digest
`source-v2:sha256:68db7a9835545d66ef9a651b9c4a000c91f3d834b5503a1b89e4a122275c3bc9`
with 3,036 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner accepted only
this bounded result at `2026-08-24T18:51:56Z`. Its owner-accepted closeout
claim and excluded receipt passed exact-SHA CI, so I4 is effectively `done`.
I4 owns only RackRole `name`, `slug`, `color`, and
`description` create/PUT/PATCH presence, defaults, validation, generated API
contracts, matching REST/gRPC semantics, PostgreSQL durability, Vue
serialization/validation, and eight fixed tests. No retained differential
accompanies it, so T2/T3 remain unearned. RackRole uniqueness, deletion,
list/query behavior, full CRUD, tiers, the parent, profile promotion, and every
traceability consumer remain open.

`CW1-V2-02-I5` has exact tested candidate
`89507d95d2743de7f97d64ca14cc43f6b834770b` at source digest
`source-v2:sha256:a8325eaae703aa801ed587deefae7e8d08d9c9e0189c80ff7569da95c36d6f90`
with 3,043 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim,
pre-acceptance receipt, owner-accepted closeout claim, and excluded receipt
boundaries are green. The project owner accepted only this bounded result at
`2026-08-25T03:56:48Z`, so I5 is effectively `done`. I5 owns only RackType common writable-field presence,
including the numeric-ID Manufacturer envelope and nine scalar fields across
POST/PUT/PATCH, operation-specific generated contracts, typed
REST/gRPC/application semantics, PostgreSQL durability, Vue dirty-field
handling, and the eight fixed tests. No retained differential accompanies the
result, so T2/T3 remain unearned. RackType uniqueness, propagation, deletion,
list/query behavior, alternate nested Manufacturer inputs, every
tier/consumer, the parent, profile promotion, and every later child remain
open. `CW1-V2-02-I6` has exact tested candidate
`dddd7adbda72f5dd760202c4862ce23b17cdf180` at source digest
`source-v2:sha256:be2168180bfbbb10406772d57e75210d32e5084051c4b765d4cf2954d958e621`
with 3,050 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner accepted only
this bounded result at `2026-08-25T08:13:40Z`. Its owner-accepted closeout
claim and excluded receipt passed exact-SHA repository CI, so I6 is effectively
`done`. I6 owns only DeviceRole `parent`, `name`, `slug`,
`color`, `vm_role`, `description`, and `comments` common writable-field
presence across POST/PUT/PATCH, operation-specific generated contracts,
typed REST/gRPC/application semantics, PostgreSQL durability, Vue dirty-field
handling, and the eight fixed tests. No retained differential accompanies the
result, so T2/T3 remain unearned. DeviceRole hierarchy, uniqueness, deletion,
list/query behavior, full CRUD, every tier/consumer, the parent, profile
promotion, and every later child remain open. `CW1-V2-02-I7` has exact tested
candidate `e2ad1acc33b84f20f24418d89b3b881b897b7ed3` at source digest
`source-v2:sha256:8930fa5fbe487df4c225a301c674db096664cccaf3f0c53df93c51e59edaeba7`
with 3,057 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner accepted only
this bounded result at `2026-08-25T14:12:13Z`. Its owner-accepted closeout
claim and excluded receipt passed exact-SHA repository CI, so I7 is effectively
`done`. I7 owns only DeviceType `manufacturer`, `model`, `slug`,
`part_number`, `u_height`, `exclude_from_utilization`, `is_full_depth`,
`airflow`, `description`, and `comments` common writable-field presence across
POST/PUT/PATCH, operation-specific generated contracts, typed
REST/gRPC/application semantics, PostgreSQL durability, Vue dirty-field
handling, and the eight fixed tests. No retained differential accompanies the
result, so T2/T3 remain unearned and the Vue unit/type boundary earns no T4.
DeviceType uniqueness, positioned-Device height transitions,
deletion/cascades, list/query behavior, full CRUD, every tier/consumer, the
parent, profile promotion, and every later child remain open.
`CW1-V2-02-I8` has exact tested candidate
`b216d4c217cf863a8760494fd6499e54899ef368` at source digest
`source-v2:sha256:da2b93d51dbdc2dfbcb6e348e2f5b23f42439b5973169f683ddc39de285c5048`
with 3,064 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner accepted only
this bounded result at `2026-08-26T02:27:06Z`. Its owner-accepted closeout
claim and excluded receipt passed exact-SHA repository CI, so I8 is effectively
`done`. I8 owns only InterfaceTemplate `device_type`, `name`,
`label`, `type`, `enabled`, `mgmt_only`, and `description` common writable-field
presence across POST/PUT/PATCH, operation-specific generated contracts, typed
REST/gRPC/application semantics, PostgreSQL durability, Vue dirty-field
handling, and the eight fixed tests. No retained differential accompanies the
result, so T2/T3 remain unearned and the Vue unit/type boundary earns no T4.
Existing DeviceType-owner containment is exercised without closing owner
immutability. InterfaceTemplate uniqueness, ModuleType ownership, bridge
behavior, Device instantiation/snapshot/rollback, non-retroactivity,
deletion, list/query behavior, full CRUD, every tier/consumer, the parent,
profile promotion, and every later child remain open.
`CW1-V2-02-I9` has exact tested candidate
`9c257b04b7cf798199c5aa4b7ae076cebbbbdff1` at source digest
`source-v2:sha256:343a1767534de69acad81e7fdfbf8bd23cf1e7a22450c00df70b038c4c54c152`
with 3,071 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner accepted only
this bounded result at `2026-08-26T11:01:23Z`. Its owner-accepted closeout
claim and excluded receipt passed exact-SHA repository CI, so bounded I9 is
effectively `done`. I9 owns only Rack `site`, `name`, `facility_id`,
`rack_type`, `status`, `role`, `serial`, `asset_tag`, `form_factor`, `width`,
`u_height`, `starting_unit`, `desc_units`, `airflow`, `description`, and
`comments` common writable-field presence across POST/PUT/PATCH,
operation-specific generated contracts, typed REST/gRPC/application
semantics, direct-save RackType copy precedence, PostgreSQL durability, Vue
dirty-field handling, and the eight fixed tests. No retained differential
accompanies the result, so T2/T3 remain unearned and the Vue unit/type
boundary earns no T4. Rack uniqueness, mounted-device and placement rules,
RackType-update propagation, Device site propagation, deletion, list/query
behavior, full CRUD, every tier/consumer, the parent, profile promotion, and
every later child remain open.

`CW1-V3-01` has an independently reviewed packet and green exact packet-claim
CI. Its implementation candidate owns process-only HTTP liveness and
request-time PostgreSQL-aware HTTP/gRPC readiness, including exact `/health` and
`/ready` state mapping, empty-service gRPC Health Check, fail-closed
named-service and Watch behavior, one injected checker, seven fixed named tests,
and auxiliary constructor coverage. Focused, affected, race, coverage, backend,
root repository, and complete L4 PostgreSQL local gates are green; exact
committed replay, candidate CI, and evidence remain. It changes no Managed
Object behavior, profile metadata, compatibility tier, deployment
claim, V3 parent, rewrite state, or production-readiness boundary.

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

| Checkpoint               | Present foundation                                                                                      | Missing before exit                                                                                                              | Status                                     |
| ------------------------ | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| V0 repository gate       | Retained exact-entry source-v2 result plus deterministic quality policy                                 | Refresh for the active source and every later relevant merged digest                                                             | Continuous; entry source retained          |
| V1 identity/RBAC         | Persisted identity, retained CORS, accepted I1/I2/I3, and retained I4 evidence                          | Obtain I4 owner review; complete password-policy/throttle/proxy/streaming plus RBAC/admin/secret work                            | I4 evidence; parent open                   |
| V2 domain behavior       | Typed shared services, reviewed 293-row traceability, and accepted IPAddress I1/Site I2/Manufacturer I3/RackRole I4/RackType I5/DeviceRole I6/DeviceType I7/InterfaceTemplate I8/Rack I9 | Complete remaining invariant, rollback, side-effect, concurrency, and tier cases | I1-I9 done; parent open |
| V3 PostgreSQL/deployment | Typed schema, real-PostgreSQL suites, Compose harness, and green V3-01 packet gate                    | Finish and retain dependency-aware HTTP/gRPC readiness, then dependency loss/recovery, locking cases, and retained external run | V3-01 implementation candidate in progress |
| V4 REST/gRPC             | Strict comparator/orchestrator and broad CRUD/parity suites                                             | Required negative/invariant/permission/presence scenarios, durable T2 report, then corresponding T3 report                       | T1; implementation/evidence pending        |
| V5 Vue                   | Typed adapters and real-Chrome workflow harness                                                         | Session refresh, edit/filter, rollback, null/conflict/not-found, reassignment, and exact-state browser outcomes                  | T1; implementation/evidence pending        |
| V6 sign-off              | First-profile legacy stacks physically retired                                                          | V0-V5 retained green together, per-capability review, protobuf freeze/breaking baseline, joint sign-off                          | Not earned                                 |

See [Evidence](evidence/README.md) for the artifact policy and commands.

## Remaining first-profile work

1. Keep `CW1-G00` green with pinned Go, Node, and npm.
2. Obtain project-owner review for retained `CW1-V1-02-I4`, then complete the
   later `CW1-V1-02` children through `CW1-V1-04`: close the credential and remaining
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
