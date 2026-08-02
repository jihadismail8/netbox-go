# Coding Standards

These rules are normative for new and modified handwritten code. Existing generated and mixed-responsibility code is transitional: it may violate these rules temporarily, but it must not be copied, extended with business logic, or used to justify a new exception.

`MUST`, `MUST NOT`, `SHOULD`, and `MAY` express requirement strength. An exception must be narrow and record the rule, exact path or symbol, reason, owner, and removal milestone. Project-wide suppressions and unowned exclusions are prohibited.

## Architecture and package boundaries

New backend code uses the layer-first structure defined in [Architecture](ARCHITECTURE.md):

```text
internal/
  domain/{shared,dcim,ipam,identity}
  application/{transaction,presence,identity,authz,changelog,dcim,ipam}
  adapters/{postgres,rest/netbox,grpc}
  platform/{composition,config,logging,server}
```

- Dependencies MUST point inward: adapter → application → domain.
- Domain packages MUST NOT import Gin, protobuf, GORM, PostgreSQL drivers, generated API types, or global configuration/database packages.
- Application packages MUST NOT import transport or storage implementations.
- REST and gRPC adapters MAY authenticate, decode, perform transport-shape validation, invoke a use case, and encode its result. They MUST NOT own authorization policy, domain validation, transactions, or persistence.
- Only PostgreSQL adapters may use GORM or SQL for domain persistence. New code MUST NOT call the global `database.GetDB()` outside legacy wiring slated for retirement.
- Repository and integration interfaces belong beside the application use case that consumes them. Interfaces MUST be narrow and capability-oriented; generic CRUD repositories are prohibited.
- Dependencies use constructor injection. New package-global mutable state, service locators, hidden `init()` registration, and singleton database/cache access are prohibited.
- Packages named `util`, `common`, `helper`, or similar dumping grounds are prohibited. Put code with the capability that owns it or give a shared abstraction a precise name.
- The retired `internal/domain/workflow`, `internal/application/workflow`, and
  `internal/adapters/postgres/workflow` packages MUST NOT be recreated or
  imported. Transitional constructors or ports that accept a generic workflow
  service or raw Managed Object map are equally prohibited.

### Exception ledger

There are no active handwritten architecture exceptions. The former generic
workflow exception closed during recovery R2-R5: production and parity
composition now require typed per-resource services, the three retired
workflow packages and transitional constructors are absent, and PostgreSQL row
ownership belongs explicitly to the DCIM, IPAM, and changelog adapters. The
[`TestGenericWorkflowPackagesAreRetired`](../netbox-backend/internal/architecture/import_boundary_test.go)
architecture check prohibits recreating the retired package directories or
importing them from production or test code.

The former frontend exception is also closed: all 13 promoted resources use
typed DTO/form/filter adapters under `src/features/core`. The first-profile
legacy model/DAO/cache/service/handler/router/protobuf stacks have also been
physically removed. Frozen generated source for deferred resources remains
immutable and runtime-disabled under ADR 0004; it is not an exception for new
work. Any future exception must satisfy the narrow ownership and removal
requirements above before affected code is merged.

## Go

### Formatting and naming

- Handwritten Go MUST pass `gofmt`; imports SHOULD be organized with `goimports` once the pinned tool is installed.
- Package and file names use short lowercase words. Exported identifiers use standard initialisms such as `ID`, `HTTP`, `API`, `URL`, `IP`, and `RPC`.
- Avoid package stutter: use `dcim.Site`, not `dcim.DcimSite`.
- Export only stable package contracts. Comments explain intent, invariants, and non-obvious trade-offs rather than restating code.
- A `TODO` or `FIXME` MUST include an issue or named removal milestone. `//nolint:<linter>` MUST name the linter and explain the local reason.

### Types and commands

- Use one signed identifier representation across the handwritten core: typed IDs backed by `int64`. Zero MUST NOT represent database `NULL`.
- Use explicit nullable relationship types. Do not overload zero values, empty strings, or empty slices to mean absent.
- Create, replace, and patch are distinct typed commands. An update MUST preserve absent, explicit `null`, zero, and concrete values whenever the public contract distinguishes them.
- Raw maps, protobuf messages, REST DTOs, and GORM rows MUST NOT enter the application/domain boundary.
- Use `net/netip`-backed value objects for prefixes and IP addresses. The public Prefix contract rejects host bits and suggests the canonical network; persistence still enforces canonical storage. An IP address assignment preserves the host address and mask.
- Use domain value types for slugs, status/choice values, DNS names, timestamps, and rack positions. Rack positions MUST represent valid half-unit values without binary floating-point comparisons.
- Clients MUST NOT set IDs, creation/update timestamps, derived counters, cached hierarchy fields, or audit identity unless the compatibility contract explicitly makes a field writable.

### Context, errors, and logging

- `context.Context` is the first parameter of I/O-capable operations. Propagate it, honor cancellation, and call every returned cancellation function.
- Domain/application errors are typed and transport-neutral, with stable reasons for validation, unauthenticated, forbidden, not found, conflict, and internal failure. Validation errors carry field details where applicable.
- Wrap causes with `%w`. Do not expose raw SQL, GORM, filesystem, or internal error text through a public interface.
- REST and gRPC each map typed errors centrally; individual handlers MUST NOT invent status behavior.
- Emit structured logs once at a process or transport boundary. Never log full requests, password material, hashes, API/bearer tokens, cookies, CSRF secrets, DSNs, or complete configuration objects.

### Transactions and side effects

- One externally visible mutation maps to one application use case and one application-owned PostgreSQL transaction.
- Authorization, current-state loading, domain validation, persistence, derived state, required `core_objectchange` records, and baseline-required durable event intent MUST commit or roll back together.
- Repositories participate in the active unit of work and MUST NOT independently commit a nested transaction.
- Do not make network calls inside a database transaction. Event or extension delivery that must survive a crash uses a transactional outbox when that feature is introduced.
- Cache updates happen after commit. Correctness MUST NOT depend on cache state, and the first shared vertical slice has no cache by default.
- Bulk operations MUST reproduce the baseline atomicity; they MUST NOT silently loop over separately committed CRUD calls.

## REST

- REST is the exact in-scope compatibility interface for the pinned baseline. Preserve paths, methods, trailing slashes, request/response fields, nullability, errors, pagination, filters, ordering, bulk semantics, actions, permissions, and durable effects.
- Public NetBox DTOs intentionally use baseline `snake_case` names. They are transport types and are mapped explicitly to application commands/results.
- `PUT` and `PATCH` are distinct. PATCH decoding MUST preserve omitted versus explicit-null fields.
- Query filters, search fields, ordering fields, and dynamic identifiers use explicit allowlists. All values are parameterized.
- Handlers MUST use explicit projections and safe response DTOs. Public `SELECT *`, raw row serialization, and raw-map writes are prohibited.
- Browser requests use the agreed session-cookie and CSRF flow. REST automation tokens are accepted only through the documented API-token path.
- Credentialed CORS uses an explicit configuration allowlist; wildcard origins with cookies or authorization credentials are prohibited. Authentication and token-creation endpoints require bounded throttling without logging submitted secrets.

## gRPC and protobuf

- New contracts are handwritten source in versioned packages such as `netbox.dcim.v1`, `netbox.ipam.v1`, and `netbox.identity.v1`.
- Filenames and fields use `lower_snake_case`; packages use lowercase dotted names; messages, enums, services, and RPCs use PascalCase.
- Services are organized around bounded modules and capabilities, not generated one-for-one from database tables.
- Partial updates use presence-aware fields plus `google.protobuf.FieldMask`. Timestamps use `google.protobuf.Timestamp`; structured values use typed messages rather than encoded strings.
- Removed field numbers and names MUST both be reserved and never reused. A v1 capability freezes when its T3 scenarios pass; afterward it evolves additively and breaking changes require a new API version even if broader profile publication occurs later.
- Public messages MUST NOT expose table names, columns, GORM models, SQL operators, generic condition trees, password hashes, audit internals, or server-owned fields as client inputs.
- Protobuf validation may reject malformed shapes and basic ranges. Shared application/domain code owns semantic validation and authorization.
- REST is not implemented as a gRPC gateway and gRPC is not implemented by calling REST. HTTP annotations are omitted unless they describe an actually served and tested gateway.
- The legacy `api/netbox_go/v1` contracts are frozen transitional scaffolding, not a naming or compatibility precedent.

## PostgreSQL and schema lifecycle

- PostgreSQL is authoritative. Storage rows are private adapter types, not domain objects or public DTOs.
- Queries MUST use context, parameterized values, deterministic ordering, bounded pagination, and explicit projections.
- Constraints, foreign-key deletion behavior, unique and partial indexes, checks, locking, CIDR behavior, JSONB, arrays, and transactions are proved on real PostgreSQL. SQLite and SQL mocks may support narrow unit tests but cannot prove PostgreSQL behavior.
- Concurrency-sensitive rack placement, prefix/address uniqueness, and allocation require an explicit locking or isolation strategy plus concurrent tests.
- Content types enforce uniqueness on `(app_label, model)`.
- During the disposable-development phase, `AutoMigrate` may create missing tables only. It MUST NOT backfill data, alter an existing table shape, repair drift, or claim production upgrade support. Schema/bootstrap failure is fatal.

## Vue and TypeScript

- Use Vue 3 Composition API with `<script setup lang="ts">`, typed props/emits, PascalCase component files, and `useX` composables.
- TypeScript remains strict. Production `any` is prohibited by default; accept untyped input as `unknown`, narrow or validate it at the boundary, and return a typed value. A framework-boundary exception must be local and justified.
- Only API infrastructure may import Axios. Pages and components call typed feature API modules or composables, never the raw client or string-built endpoint paths.
- NetBox wire DTOs remain explicit `snake_case` types. UI view models may use idiomatic names only through deliberate adapters.
- Normalize API failures into one discriminated error type. Catch values as `unknown` and centralize cancellation, query serialization, trailing slashes, and 401/403 behavior.
- UI code controls presentation and affordances, not authorization or domain validity. The backend remains authoritative.
- Credentials, permissions, CSRF data, and session state MUST NOT be stored in `localStorage`. It is limited to non-sensitive display preferences.
- Demo identities and fake login behavior belong only in explicit tests or demo builds, never the normal runtime path.
- Raw `v-html` is prohibited except inside one audited sanitized-markdown component. Sanitization has adversarial tests.
- Component/composable tests are colocated as `*.test.ts`; browser-facing
  workflows receive owned real-browser coverage against the built application
  and disposable backend/database.

## Generated code and dependencies

- Follow [ADR 0004](adr/0004-generated-scaffolding-is-immutable-and-transitional.md): generated files are immutable outputs and never contain handwritten fixes or business rules.
- Every newly generated file identifies its generator and owned source and carries the standard generated marker when comments are supported.
- Generators use sorted inputs, fail on missing or unmatched inputs, produce no environment-specific content, and are idempotent.
- Generation MUST NOT patch a database, merge into handwritten files, modify `go.mod`, or perform unrelated scaffolding. Split one-time scaffolding from deterministic generation.
- Pin the Go, Node, protobuf, generator, linter, and formatter versions used by CI. Commit module and package lockfiles; CI uses `npm ci`.
- Vendored `third_party` code is immutable except through a reviewed dependency update. It must compile but is excluded from handwritten lint and coverage.

## Tests and merge gates

The agreed policy is strict but staged. Before feature work, checkpoint V0 in [Testing](TESTING.md) must establish one green, self-contained default gate.

- The complete backend and frontend MUST build.
- Formatting, linting, typechecking, unit tests, and deterministic generation checks are warning-free.
- Go unit tests are hermetic and run with the race detector. A test owns its server, database, clock, randomness, and fixture dependencies appropriate to its layer.
- Generated tests that require an unmanaged live service are explicitly classified as legacy/integration tests and excluded from the default unit suite until replaced. Every exclusion has an owner and removal milestone; new exclusions are prohibited by default.
- PostgreSQL, differential REST, gRPC parity, and browser suites are separate owned jobs with deterministic setup and teardown.
- Do not use sleeps, execution order, wall-clock assumptions, or an already-running developer service as synchronization.
- Every defect fix adds a regression test at the lowest layer that proves the behavior.
- Coverage collection is mandatory. Do not invent a percentage target: record a trustworthy baseline, prohibit regression, and raise the floor as vertical slices complete. Security and compatibility-critical code cannot be excluded to improve the number.
- A T3 backend increment includes domain positive/negative tests, PostgreSQL constraint/rollback tests, authenticated REST differential tests, in-process gRPC parity tests, RBAC allow/deny/list-visibility cases, and change-log assertions. It must remain labelled non-T4 until an operator-facing capability also has Vue evidence.

Successful compilation, route enumeration, proto generation, snapshots from one implementation, or a happy-path CRUD test are supporting checks only. They do not prove compatibility or interface parity.
