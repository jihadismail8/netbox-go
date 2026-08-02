# NetBox Go Architecture

This document distinguishes the architecture that exists in the repository today from the architecture the rewrite is moving toward. The target architecture is normative. Current generated code is useful scaffolding, but its presence is not evidence that a NetBox capability is complete.

The terminology and compatibility boundary are defined in [CONTEXT.md](../CONTEXT.md). The governing decisions are:

- [ADR 0001: NetBox compatibility and interface parity](./adr/0001-netbox-compatibility-and-interface-parity.md)
- [ADR 0002: Out-of-process extension services](./adr/0002-replace-python-plugins-with-extension-services.md)
- [ADR 0003: Unified authentication and authorization](./adr/0003-unified-authentication-and-authorization.md)
- [ADR 0004: Immutable transitional generated scaffolding](./adr/0004-generated-scaffolding-is-immutable-and-transitional.md)

## Architectural objective

NetBox Go is a standalone Go implementation of the core behavior of the checked-in post-4.4.6 NetBox source snapshot at commit `fbb948d30e79ce657fac62994a22aca72c1770a9` (`v4.4.6-7-gfbb948d30`).

The checked-in [upstream NetBox source](../netbox/) is a development reference and compatibility oracle. It must not be required to build, migrate, start, or operate NetBox Go. Python, Django, Python plugins, scripts, and reports are not runtime components.

The application has two first-class public interfaces:

- NetBox-compatible HTTPS REST for existing consumers, automation, and the Vue application.
- gRPC for service-to-service use and integration into larger systems.

Every declared baseline capability is exposed through both interfaces over one
application/domain implementation. They may use transport-appropriate
representations, but they must not develop different validation,
authorization, transaction, or side-effect behavior. A one-transport extension
is permitted only when explicitly classified; if both transports expose an
extension, they share its semantics.

## Target architecture

```text
 Browser / Vue SPA          REST automation          Other services
        |                         |                         |
        +--------- HTTPS REST ---+                  gRPC + TLS
                         |                                  |
                +--------v---------+              +---------v--------+
                | REST adapter     |              | gRPC adapter     |
                | NetBox DTOs      |              | protobuf DTOs    |
                +--------+---------+              +---------+--------+
                         |                                  |
                         +---------------+------------------+
                                         |
                              authenticated Principal
                                         |
                              +----------v-----------+
                              | Application use cases |
                              | authorization         |
                              | transaction boundary  |
                              | change/event intent   |
                              +----------+-----------+
                                         |
                              +----------v-----------+
                              | Domain rules          |
                              | Managed Objects       |
                              | validation/policies   |
                              +----------+-----------+
                                         |
                         repository and integration ports
                                         |
                              +----------v-----------+
                              | Infrastructure        |
                              | GORM/PostgreSQL       |
                              | cache/events as needed|
                              +-----------------------+
```

This is a modular monolith, not a collection of per-module microservices. REST and gRPC run from the same Go application by default. The gRPC interface makes the application usable _as_ a service; it does not require splitting the domain into separate deployments.

### Dependency direction

Dependencies point inward:

1. Transport adapters depend on application contracts.
2. Application use cases depend on domain types and ports.
3. Domain code depends on neither Gin, protobuf, GORM, nor PostgreSQL.
4. Infrastructure implements ports owned by the application/domain side.

New handwritten backend code uses this layer-first package structure:

```text
internal/
  domain/
    shared/
    dcim/
    ipam/
    identity/
  application/
    transaction/
    presence/
    identity/
    dcim/
    ipam/
    authz/
    changelog/
  adapters/
    postgres/
    rest/netbox/
    grpc/
  platform/
    composition/
    config/
    logging/
    server/
```

Domain-specific packages may be added under these layers as slices expand. Repository and integration interfaces live beside the application use cases that consume them; PostgreSQL/GORM implementations live under `adapters/postgres`. `platform` contains process-level technical wiring and must not become a generic home for domain or application helpers.

Existing generated and mixed-responsibility directories remain in place only until their capabilities are replaced. They must not be renamed en masse or used as precedents for new code. The dependency rule remains authoritative if a future package subdivision is needed.

### Layer responsibilities

| Layer          | Owns                                                                                                               | Must not own                                                              |
| -------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| REST adapter   | Routes, NetBox request/response shapes, query parsing, HTTP status and error mapping, REST credential extraction   | Domain validation, authorization rules, transactions, direct GORM queries |
| gRPC adapter   | Protobuf conversion, metadata credential extraction, gRPC status mapping                                           | Domain validation, authorization rules, transactions, direct GORM queries |
| Application    | Use-case orchestration, authorization, transaction boundaries, repository calls, change-log and event coordination | HTTP or protobuf-specific behavior                                        |
| Domain         | Managed Objects, invariants, cross-object rules, domain errors                                                     | Transport, framework, or storage concerns                                 |
| Infrastructure | PostgreSQL repositories, GORM mapping, optional cache and event delivery implementations                           | Public API behavior or independent business rules                         |
| Vue frontend   | Operator workflows and presentation through REST                                                                   | Direct database access, gRPC calls, duplicated backend authorization      |

An adapter may reject malformed transport input before invoking a use case. Once input has been translated, all business decisions belong to the shared application/domain path.

### Generated source boundary

Generated files are immutable outputs, not extension points. Business rules, authorization, transactions, and compatibility behavior must never be added directly to generated models, DAOs, caches, services, handlers, routers, or protobuf Go files. A generated file is changed through its owned source definition or generator and then regenerated.

The existing Sponge-generated per-table contracts and execution paths are transitional and have no backward-compatibility guarantee. New public `.proto` files are handwritten, versioned contract sources in packages such as `netbox.dcim.v1` and `netbox.ipam.v1`; the corresponding generated Go code is reproducible output. A new contract becomes a compatibility commitment only when it is explicitly declared published. New persistence mappings and adapters for a promoted capability are handwritten unless and until a deterministic generator has an explicitly owned input and output boundary. Required generated source may remain committed, but generation must use pinned tools, produce standard generated-file markers, and be cleanly reproducible in CI.

## Interface parity

For each supported operation, REST and gRPC must reach the same use case and produce equivalent domain outcomes:

- the same accepted and rejected state transitions;
- the same view, add, change, and delete authorization decisions, including object constraints;
- the same transaction boundary and persisted state;
- the same change-log and event effects;
- equivalent not-found, conflict, validation, and permission errors;
- the same visibility rules on reads and lists.

REST remains responsible for matching the pinned NetBox REST contract. gRPC does not need to copy REST JSON shapes into protobuf, but its messages must preserve the information needed to perform the same operation. A gRPC-only extension is allowed only when it is explicitly documented as an extension rather than parity coverage.

The Vue application is a REST client only. It does not call gRPC from the browser, and frontend code is not a third implementation of domain rules. Visual styling may differ from upstream NetBox, while supported workflows and their outcomes remain compatible.

## Application and transaction boundary

One application use case represents one externally visible operation. The use case should:

1. receive a transport-neutral command or query plus an authenticated `Principal`;
2. authorize the action and, for reads, constrain object visibility;
3. load the required Managed Objects through repository ports;
4. run domain validation and apply the state transition;
5. persist the transition in one PostgreSQL transaction;
6. record the compatible object change and stage any baseline-required events in that transaction;
7. return a transport-neutral result or typed application/domain error.

Transport adapters map that result to REST or gRPC. They must not repeat the operation with their own database statements. Cross-object workflows such as assigning an IP address to an interface are application operations, not a sequence of unrelated generic CRUD calls.

The precise event-delivery mechanism is deferred. Whatever implementation is chosen must not allow an extension to bypass authorization or write core tables directly.

## Persistence and development schema lifecycle

PostgreSQL is the authoritative persistence store. GORM is an infrastructure detail, not the domain model. Redis may be retained for disposable caches or delivery mechanisms, but correctness must not depend on a stale cache and Redis is not a second source of truth.

During this early rewrite phase, development and test databases are disposable:

- developers may drop and recreate the database for clean testing;
- the agreed current-phase scope for GORM `AutoMigrate` is disposable bootstrap and creation of missing tables only;
- when an existing table shape changes, the normal development response is to rebuild the database rather than backfill it in place;
- upgrade migrations, data backfills, and comprehensive schema-drift repair are intentionally deferred;
- code should not accumulate compatibility branches for arbitrary partially upgraded development databases;
- no production upgrade guarantee is implied by the current schema bootstrap.

Startup invokes migration and Go-owned content-type seeding in
[`initial.InitApp`](../netbox-backend/cmd/netbox_go/initial/initApp.go). The
current [`AutoMigrate`](../netbox-backend/internal/database/migrate.go) builds an
ordered registry and delegates to
[`bootstrap.Run`](../netbox-backend/internal/adapters/postgres/bootstrap/bootstrap.go).
For each row type, bootstrap checks `HasTable` and invokes GORM `AutoMigrate`
only when that table is absent. Existing tables are never passed to
`AutoMigrate`; their columns and indexes are neither inspected nor repaired.
Bootstrap or content-type seeding failure is fatal.

Table names inherited from the compatibility baseline, including names containing `django`, do not create a Django runtime dependency. They are storage compatibility details until deliberately replaced.

## Authentication and authorization

Authentication is required by default for all Managed Object and domain operations, including reads. Anonymous mutation is never permitted. The only public exceptions are narrowly defined system/authentication entry points such as health, readiness, login, and CSRF bootstrap; they cannot expose or mutate Managed Objects. Administrator bootstrap/password recovery, local non-superuser creation, local global-model-permission grants, and token provisioning are never anonymous HTTP or RPC endpoints. Optional anonymous read-only behavior is deferred and must be explicitly configured if introduced.

As specified by [ADR 0003](./adr/0003-unified-authentication-and-authorization.md):

- the Vue application uses a secure `HttpOnly`, `SameSite` session cookie;
- state-changing cookie-authenticated REST requests use CSRF protection;
- REST automation supports NetBox-compatible API tokens;
- authenticated token management uses the profile's secure REST extension and paired identity RPCs; the baseline anonymous username/password provision action is deferred and rejected;
- gRPC callers send bearer credentials in request metadata;
- all credential types resolve to the same `Principal`, users, groups, token restrictions, and object-level RBAC evaluation;
- browser credentials are not stored in `localStorage`.

Authentication at an adapter establishes identity. Authorization remains a shared application concern and must be enforced identically after either adapter has resolved the principal.

## Deployment boundary

The default deployable consists of:

- one standalone Go application exposing REST and gRPC;
- the built Vue SPA, currently served by the Go HTTP process;
- PostgreSQL;
- optional infrastructure such as Redis when a concrete cache or delivery use case requires it.

The current process wiring already starts HTTP and gRPC together from [`main.go`](../netbox-backend/cmd/netbox_go/main.go) through [`CreateServices`](../netbox-backend/cmd/netbox_go/initial/createService.go). The local topology is illustrated by [`docker-compose.yml`](../docker-compose.yml). Production REST is HTTPS and production gRPC must use an appropriately secured channel, whether TLS terminates in the Go process or at trusted ingress infrastructure.

Extension capabilities run out of process and integrate through authenticated REST, gRPC, webhook, or event contracts. They cannot load Python into the core process or mutate PostgreSQL directly.

## Current architecture audit

The production REST/gRPC runtime follows the target dependency direction and
uses typed per-resource DCIM/IPAM application services, typed per-table
PostgreSQL persistence, and typed Vue adapters. Production construction uses
`composition.NewCore`; parity fixtures use the narrow
`composition.NewCoreWithAuthorizer` seam to inject deterministic typed
authorization without changing the service graph. The generic domain,
application, and PostgreSQL workflow packages and their transitional
constructors have been retired, and an architecture test prohibits their
recreation or import. The strict compatibility, PostgreSQL, and browser
harnesses exist, but a current post-hardening retained artifact is not yet
recorded, so the profile remains T1.

> **Current operational boundary:** Managed Object requests fail closed and the
> prior generic users/configuration exposure is not registered by the default
> runtime. The build nevertheless remains development-only until the complete
> identity security matrix, differential compatibility, real-PostgreSQL
> behavior, TLS deployment, and production hardening gates pass.

| Area              | Current implementation                                                                                                                                                                                                                                                                                                                                                                                  | Architectural assessment                                                                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Process           | [`http.go`](../netbox-backend/internal/server/http.go) and [`grpc.go`](../netbox-backend/internal/server/grpc.go) compose one Go-owned core in one process and expose separate HTTP and gRPC listeners.                                                                                                                                                                                                 | Correct modular-monolith deployment shape.                                                                                                                  |
| REST runtime      | [`NewRuntimeRouter`](../netbox-backend/internal/adapters/rest/netbox/router/router.go) registers Go-owned identity, the 13 profile resources, and authenticated schema access. REST maps the six declared operations into typed per-resource application services.                                                                                                                                      | Exact-in-profile REST is the intended boundary; T2 still requires a current successful strict oracle artifact.                                              |
| Frozen REST       | The current inventory contains 102 frozen, runtime-disabled direct-GORM resource configurations. The displaced stacks for all 13 promoted resources have been physically removed.                                                                                                                                                                                                                       | Remaining legacy source is deferred inventory, not public behavior or compatibility evidence.                                                               |
| gRPC runtime      | `registerCanonicalServices` registers only versioned Identity, DCIM, and IPAM services. DCIM/IPAM adapters and all parity fixtures invoke the same typed per-resource services as REST; focused lifecycle, error, RBAC, rollback, and assignment parity diagnostics pass.                                                                                                                               | Correct production dual-adapter shape; corresponding REST T2 and a current retained parity run are still required before T3.                                |
| Shared core       | Typed DCIM/IPAM services own authorization, validation, transactions, relationships, projections, and change recording. `composition.Core` exposes only Identity and the 13 typed resource services; the production constructor selects the fail-closed permission authorizer.                                                                                                                          | The shared path is typed. The retired generic packages and constructors are protected by a total architecture prohibition.                                  |
| Persistence       | [`dcim/row/rows.go`](../netbox-backend/internal/adapters/postgres/dcim/row/rows.go), [`ipam/row/rows.go`](../netbox-backend/internal/adapters/postgres/ipam/row/rows.go), and [`changelog/row.go`](../netbox-backend/internal/adapters/postgres/changelog/row.go) own the 10 DCIM, 3 IPAM, and 1 append-only change row respectively, with explicit mappings, foreign keys, indexes, and locking tests. | The obsolete JSON-table design is gone. Real-PostgreSQL suites exist; their latest post-hardening execution remains pending in the durable evidence ledger. |
| Identity and RBAC | Go-owned users, groups, group memberships, direct/group model grants, object-scoped grants, sessions, tokens, bearer authentication, CLI bootstrap/reset/non-superuser creation/global permission grant, and shared authorization are persisted.                                                                                                                                                        | The common principal/RBAC path is implemented. Current full security and cross-transport execution evidence is still pending.                               |
| Diagnostics       | The canonical router publishes health, readiness, identity, profile resources, authenticated schema, and the SPA. It registers no dedicated `GET /ping`; when the SPA is enabled, `/ping` is merely an ordinary history-fallback path. Frozen ADR 0004 legacy wiring still contains a runtime-disabled `/ping` route.                                                                                   | Configuration and the canonical process-probe surface are contained. Production observability still needs a protected design.                               |
| Vue runtime       | The profile registry publishes only 13 resources. [`adapters.ts`](../netbox-frontend/src/features/core/adapters.ts) supplies typed DTO/form/filter adapters consumed by the profile pages; authentication uses session/CSRF and stores no credential in `localStorage`.                                                                                                                                 | Typed transport boundaries and a real-Chrome harness exist. T4 remains unearned until a successful current browser artifact is retained.                    |
| Schema            | Startup registers 198 tables: 176 deferred legacy rows, 8 Go-owned identity rows, 13 typed profile rows, and 1 typed object-change row. Missing tables are created individually; existing table shapes are untouched.                                                                                                                                                                                   | Correct for disposable development, not a production upgrade system. A current real-PostgreSQL/Compose artifact is still required for sign-off.             |

The production runtime behavior path is now typed and shared:

```text
REST adapter --+
               +-> typed resource application service -> authorization/domain policy
gRPC adapter --+                                  -> transaction/ports -> PostgreSQL
```

The first-profile typed-boundary migration is closed: composition, parity
fixtures, transport mapping, application/domain contracts, and PostgreSQL row
ownership now have precise typed owners. Remaining work is retained V0 and
V1-V6 evidence; this structural recovery does not promote a compatibility
tier.

## Migration direction

Migration proceeds as a strangler refactor inside the Go application. The
first-profile legacy stacks were removed earlier than ADR 0004's intended
completion order; that historical deviation is not precedent. Deferred
resources remain frozen until a later reviewed Capability Profile completes.

1. **Introduce transport-neutral boundaries.** Define the principal, typed errors, transaction runner, repository ports, and application contracts needed by a selected workflow. Keep Gin, protobuf, and GORM types outside those contracts.
2. **Implement a vertical slice.** For later profiles, select the next
   dependency-correct operator workflow. Supporting reference objects may be
   fixtures or narrowly implemented dependencies, but must not be reported as
   complete modules. The first DCIM/IPAM slices already exist with their typed
   boundary closed and now require retained evidence.
3. **Put domain behavior in one place.** Port validation, relationship rules, RBAC, transactional effects, and change logging from the pinned baseline into the shared use cases/domain.
4. **Rewire both adapters.** Make the NetBox-compatible REST handlers and new handwritten-contract gRPC adapters translate into the same use cases. Neither adapter may query GORM after conversion.
5. **Prove parity.** Run baseline REST compatibility scenarios and equivalent gRPC scenarios against the same fixtures and expected persisted effects. Exercise the operator workflow through Vue over REST.
6. **Retire displaced scaffolding.** The first-profile stacks were already
   removed as the historical deviation above. Retire each later displaced path
   only after capability completion under ADR 0004.
7. **Expand by workflow.** Repeat for the next coherent vertical slice rather than declaring broad completion from generated CRUD coverage.

Deferred scaffolding may remain in source while later workflows are replaced
by an auditable shared implementation. Direct-GORM REST and generated gRPC
business surfaces are disabled in the default runtime and excluded from
compatibility evidence.

The file-specific recovery and repeatable profile method are in the
[execution playbook](IMPLEMENTATION_EXECUTION_PLAYBOOK.md).

## Completion gate for a capability

A capability is complete only when all of the following are true:

- its behavior is defined against the pinned Compatibility Baseline;
- REST satisfies the compatible contract for the supported operation;
- gRPC reaches the same use case and demonstrates equivalent outcomes;
- validation, authorization, transaction handling, change logging, and applicable baseline side effects, including events where required, are shared;
- adapters contain no direct persistence or competing business policy;
- authenticated REST and gRPC tests cover success, validation, permission, conflict, and not-found cases;
- the Vue workflow, when applicable, works through REST without browser-stored bearer/API tokens;
- persisted results and side effects are checked, not only response shapes.

## Explicit non-goals

- Running or embedding Python or Django.
- Compatibility with the Python plugin, script, or report runtime.
- GraphQL as a public interface.
- A browser gRPC client.
- Pixel-perfect reproduction of the upstream NetBox UI.
- Production database upgrade/backfill machinery during the disposable-database phase.
- Treating generated CRUD files, routes, or protobuf methods as proof of behavioral completeness.
