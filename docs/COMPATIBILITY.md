# Compatibility contract

This document defines what NetBox compatibility means, how it is measured, and
when a capability may be called complete. Artifact counts are inventory; they
are never compatibility evidence by themselves.

## Baseline and runtime boundary

The [Compatibility Baseline](../CONTEXT.md) is the checked-in NetBox source at
commit `fbb948d30e79ce657fac62994a22aca72c1770a9`. It reports package version
4.4.6 and `git describe` value `v4.4.6-7-gfbb948d30`; it is not the official
`v4.4.6` tag.

The upstream tree is a development reference and differential oracle only.
The delivered system is a [Standalone Runtime](../CONTEXT.md): Python, Django,
and upstream source are not build, migration, startup, or deployment
dependencies. The oracle job must assert the exact Git SHA and the committed
configuration, authentication policy, timezone, plugins, database, and fixture
set before comparison.

Compatibility belongs to a declared capability: operations, fields, filters,
relationships, validation, permissions, transactions, errors, and durable side
effects. A similarly named table, route, RPC, or page proves nothing on its own.

## Public interfaces

### REST: exact in profile

For a capability declared in a [Capability Profile](../CONTEXT.md), HTTPS REST
must preserve the baseline's observable contract, including where applicable:

- URL, method, trailing slash, content type, and status;
- authentication and object-level authorization;
- writable and response fields, defaults, nullability, and PATCH presence;
- scalar choice inputs and exact read-side `{value, label}` choice envelopes;
- validation reasons, error shape, nested objects, and pagination;
- filters, search, ordering, actions, and bulk semantics;
- transactions, relationships, object changes, and other durable effects.

Only explicit normalizers may account for environment-specific origins,
generated identifiers, and volatile timestamps. The committed
[`normalizers.yaml`](../contracts/netbox/v4.4.6-post7/normalizers.yaml)
forbids normalizing status codes, validation reasons, authorization outcomes,
missing or extra fields, committed state, or side effects.

The first profile contains a secure identity REST extension under
`/api/auth/`. Extensions are documented and tested separately; they cannot
earn T2 against an upstream route they intentionally replace or omit.

### gRPC: semantic parity

gRPC is a first-class service-to-service interface over the same supported
application capabilities. It has semantic, not JSON-wire, parity:

- both adapters call the same application path;
- the same Principal receives the same authorization and visibility;
- equivalent intent receives the same validation and state transition;
- success commits equivalent state and side effects;
- failure leaves equivalent state and maps to a documented gRPC status; and
- pagination, concurrency, and transaction semantics do not drift.

The canonical services are `netbox.identity.v1.IdentityService`,
`netbox.dcim.v1.DCIMService`, and `netbox.ipam.v1.IPAMService`. The 176 frozen
table-oriented services are unpublished, runtime-disabled scaffolding.

### Vue: workflow parity

The Vue application uses REST only. It must let an operator complete every
declared browser workflow with the same Managed Objects, constraints, and
outcomes as the baseline. Visual composition may differ; pixel parity is not a
goal. Typed frontend adapters do not earn workflow parity without a real
browser execution against the application and database.

## First Capability Profile

[`core-workflow-v1`](../contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml)
is pre-publication and includes six single-object operations—list, get, create,
replace, update, and delete—for:

- DCIM: Site, Manufacturer, RackRole, RackType, Rack, DeviceRole, DeviceType,
  InterfaceTemplate, Device, and Interface; and
- IPAM: VRF, Prefix, and IPAddress, including Interface assignment and
  unassignment.

It declares exact fields, filters, projections, and scenarios. Deferred fields,
bulk operations, rack elevation, automatic allocation, cabling, GraphQL, and
Python plugin/script/report behavior are not silently included. Completing the
profile will not imply that all of DCIM or IPAM is complete.

## Compatibility tiers

The tier belongs to one declared capability, not to a whole repository or
module.

| Tier | Name                    | Required evidence                                                                     | Meaning                                 |
| ---- | ----------------------- | ------------------------------------------------------------------------------------- | --------------------------------------- |
| T0   | Catalogued              | Baseline capability and expected scenarios recorded                                   | Scope known; no implementation claim    |
| T1   | Scaffolded              | Runtime surface and supporting smoke/unit checks exist                                | Plumbing exists; compatibility unproved |
| T2   | REST verified           | Positive and negative strict differential scenarios pass against the pinned oracle    | Declared REST behavior is compatible    |
| T3   | Dual-interface verified | Equivalent gRPC scenarios pass through the shared core with matching durable outcomes | Interface Parity is verified            |
| T4   | Workflow verified       | Real-browser workflow passes, including applicable permissions and effects            | Workflow Parity is verified             |

T4 applies only to browser-facing workflows. An intentionally programmatic
capability may stop at T3.

## Definition of complete

A capability is complete only when:

1. supported and deferred baseline behavior is catalogued;
2. REST is T2 and its corresponding gRPC behavior is T3;
3. both transports use one shared application/domain path;
4. users, groups, token constraints, and object-level RBAC apply equally;
5. success, validation, permission, not-found, conflict, and rollback cases are
   automated;
6. committed state and required side effects match the baseline;
7. its browser workflow is T4 where applicable; and
8. the evidence report and public contract identify every limitation and
   extension.

A module is complete only when every in-scope capability in its authoritative
inventory satisfies this definition.

## Current classification

The generated inventories currently contain:

- baseline REST: 155 entries;
- current REST: 123 entries—102 frozen legacy, 13 canonical profile, and 8
  identity extension operations;
- current gRPC: 179 entries—176 frozen legacy and 3 canonical services; and
- current Vue: 13 profile resources.

The 13 baseline resources remain **T1**. Repository V0 and the recovery-scoped
PostgreSQL replay are retained under [Evidence](evidence/README.md), while typed
per-table persistence, persisted RBAC, typed Vue adapters, gRPC parity tests, a
strict oracle comparator, and a real-browser harness also exist. No complete
T2, T3, or T4 boundary bundle is retained, so none of those implementation
facts advances a capability tier and the contract remains pre-publication.

The first-profile legacy model/DAO/cache/service/handler/router/protobuf stacks
have been physically removed. This reduces ambiguity and maintenance burden;
it does not substitute for differential or browser evidence.

## Differential proof

[`tests/compatibility/`](../tests/compatibility/) starts the pinned oracle and
standalone Go implementation with isolated databases and a fail-closed
comparator. A valid T2 report must:

1. refuse a source or oracle-configuration mismatch;
2. apply equivalent authenticated scenarios and deterministic fixtures;
3. compare exact status, declared fields, validation reasons, authorization,
   ordering, pagination, committed state, and side effects;
4. preserve path, query, and trailing slash while normalizing only origins;
5. bind every generated identifier to an explicit scenario object and reject
   unbound or cross-resource identity;
6. exercise positive, invalid, conflict, permission, partial-update, unknown
   field/filter, and rollback behavior; and
7. fail on deliberate choice-shape, numeric-type, query-order, authorization,
   validation-reason, path, and state divergences.

The equivalent gRPC suite must then compare the same intent and durable outcome
through the shared core. OpenAPI equality, route enumeration, compilation,
single-implementation snapshots, and happy-path CRUD are supporting checks
only.

The harness is implemented. Its latest post-hardening real execution is pending
and no pass is claimed until its report is retained and reviewed.

## Development database policy

Development and test databases are disposable. Startup calls GORM
`AutoMigrate` for an individual row type only when `HasTable` says its table is
absent. Existing table shapes are not passed to `AutoMigrate`, inspected for
missing columns, repaired, or backfilled. A schema change requires dropping and
recreating the development database.

The current startup registry has 198 tables: 176 deferred legacy rows, 8
Go-owned identity rows, 13 typed profile rows, and 1 typed object-change row.
This is not upstream schema compatibility and not a production migration
strategy.

## Extensions and non-goals

- Extension services run out of process and use authenticated REST, gRPC,
  webhooks, or future event contracts; they never write the core database
  directly.
- A deliberate safe replacement for baseline behavior remains an explicit
  extension or compatibility gap and may not be normalized into a pass.
- GraphQL, Python/Django runtime integration, Python plugins, custom scripts
  and reports, pixel-perfect UI reproduction, and production upgrade/backfill
  support are outside the current target.
- REST JSON and protobuf wire representations need not be identical; their
  domain meaning and outcomes must be.

See [ADR 0001](adr/0001-netbox-compatibility-and-interface-parity.md),
[ADR 0002](adr/0002-replace-python-plugins-with-extension-services.md),
[ADR 0003](adr/0003-unified-authentication-and-authorization.md), and
[ADR 0004](adr/0004-generated-scaffolding-is-immutable-and-transitional.md).
