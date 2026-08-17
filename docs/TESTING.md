# Testing

Verified structurally: 2026-08-03

A test file or harness is not a passed gate. Results must come from the current
revision, use the required external boundary, and be retained according to the
[evidence policy](evidence/README.md). The latest full post-hardening external
runs are pending, so this document does not claim T2, T3, or T4.

## Fast repository gate

From the repository root:

```bash
make check
```

This runs:

1. backend format, vet, lint, race-unit, non-mutating coverage non-regression,
   build, frozen legacy-client compile, frozen proto compile, and canonical Buf
   checks;
2. pinned frontend install, formatting, lint, application/test typechecks,
   coverage tests, and production build;
3. deterministic current-surface inventory drift checks;
4. Capability Profile, inventory, protobuf descriptor, generated contract,
   OpenAPI, and Markdown link validation.

The obsolete Vue-registry API catalogue and DRF registry generators have been
retired. `generated-check` now validates only the authoritative contract
inventory. Frozen live-client tests still require an unmanaged service and are
compile-checked under the owned exclusion in
[`legacy-exclusions.yml`](../netbox-backend/quality/legacy-exclusions.yml);
they are not parity evidence.

Backend `coverage-check` and its backward-compatible `cover` alias write the
atomic profile only beneath `/tmp`, validate it, and remove it. The reviewed
[`coverage-baseline.tsv`](../netbox-backend/quality/coverage-baseline.tsv)
records 54 measured module packages, the one owned generated live-client
exclusion, and the current statement baseline of 36,013/290,425 (12.4001%).
The prior post-recovery floor was 35,987/292,320 (12.3108%). Package-set drift,
an unowned exclusion, or a lower exact ratio fails the gate; lowering the
baseline is not a permitted repair. Durable V0 status is recorded only in the
[evidence ledger](evidence/README.md).

The [2026-08-03 artifact](evidence/2026-08-03-post-cleanup-v0.md) is historical:
it used the superseded mode-blind v1 digest and cannot attest the current
source-v2 tree. A fresh committed source-v2 V0 run and human evidence review
remain pending. GitHub CI currently runs the V0 command only. The owned
PostgreSQL, deployment, differential, gRPC-promotion, and browser gates remain
manual until connected jobs retain durable artifacts.

## Owned external gates

### Differential REST oracle

```bash
make compatibility-comparator-test
make compatibility-test
```

[`tests/compatibility/`](../tests/compatibility/) owns the pinned NetBox oracle,
isolated oracle/Go databases, deterministic fixtures, strict comparator, and
artifact directory. It refuses a dirty or mismatched oracle SHA and mismatched
effective configuration.

The comparator permits only the normalizers declared in
[`normalizers.yaml`](../contracts/netbox/v4.4.6-post7/normalizers.yaml). It must
fail on status, validation-reason, authorization, path/trailing-slash,
query-order, choice-envelope/label, numeric JSON type, missing/extra-field,
committed-state, or side-effect divergence. Generated IDs are bound to named
scenario objects; an unbound ID is a failure. A successful current report is
required for T2; the harness's existence is only T1 evidence. The current
driver has broad CRUD/filter/ordering coverage but does not yet cover every
declared permission, presence, conflict, invariant, rollback, and side-effect
case. Build the scenario/rule traceability matrix and add the uncovered cases
before treating `make compatibility-test` as a complete first-profile T2 gate.

### gRPC semantic parity

The backend parity packages cover canonical resource lifecycle, error mapping,
authorization, Device-instantiation rollback, and IPAddress assign/unassign
through shared REST-visible state:

```bash
cd netbox-backend
go test ./test/parity -count=1
```

These tests support T3 only after the corresponding REST capabilities have
earned T2 and the current full parity run is retained. A separate transport
shape does not permit separate business behavior.

### Real PostgreSQL

Provide a disposable PostgreSQL DSN; each suite creates and removes its own
schema:

```bash
cd netbox-backend
NETBOX_TEST_POSTGRES_DSN='postgres://...' \
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

The checked-in suites exercise missing-table-only bootstrap, the 198-entry
startup registry, the ten DCIM rows owned by
[`dcim/row/rows.go`](../netbox-backend/internal/adapters/postgres/dcim/row/rows.go),
the three IPAM rows owned by
[`ipam/row/rows.go`](../netbox-backend/internal/adapters/postgres/ipam/row/rows.go),
the change row owned by
[`changelog/row.go`](../netbox-backend/internal/adapters/postgres/changelog/row.go),
constraints/indexes, concurrent duplicate and conditional-IP handling,
assignment/audit serialization, one-time administrator bootstrap,
authenticated non-superuser creation, and authenticated model-permission
grants. SQLite tests may support orchestration, but cannot prove PostgreSQL
constraints, locking, or concurrency.

A current retained run is still required for V3.

### Deployment smoke

```bash
make deployment-smoke
```

[`compose_smoke.sh`](../tests/deployment/compose_smoke.sh) owns a uniquely named
Compose project and volume, starts a clean PostgreSQL and standalone Go/Vue
application, verifies health and schema/content-type bootstrap, restarts the
application, checks idempotence, and tears everything down. It does not use a
Django migration or initialization-SQL mount.

The current smoke checks `/health`, while HTTP `/ready` and gRPC health do not
yet reflect PostgreSQL dependency state. V3 therefore requires implementation
of dependency-aware readiness plus dependency loss/recovery assertions before
the deployment run can satisfy its exit condition.

### Real-browser workflow

```bash
make browser-e2e
```

[`tests/browser/`](../tests/browser/) drives an installed headless Chrome over
the disposable deployment using Chrome DevTools Protocol. It retains
credential-free diagnostics on failure. The harness is the required boundary
for T4; component tests or a typed adapter alone cannot earn workflow parity.

A successful current run must cover the declared DCIM/IPAM workflow and the
applicable authentication, validation, real limited-user permission denial,
rollback, assignment/unassignment, and delete-warning outcomes before T4 is
considered.

The current browser driver covers the main creation chains and a subset of
negative cases. Session refresh, edit/list/filter, late template rollback,
explicit-null update, reassignment/assignment rollback, conflict, not-found,
and exact change-state outcomes remain to be added and retained.

## Implemented test layers

| Layer                      | Checked-in boundary                                    | What it can prove                                | What it cannot prove alone                              |
| -------------------------- | ------------------------------------------------------ | ------------------------------------------------ | ------------------------------------------------------- |
| Domain unit                | Table-driven Go tests without I/O                      | Invariants and stable reasons                    | SQL, transport, or oracle behavior                      |
| Application integration    | Typed resource/identity services and transaction tests | Orchestration, authorization, rollback intent    | External PostgreSQL semantics unless a real DSN is used |
| REST contract              | In-process canonical router and OpenAPI equality       | Runtime surface and local mappings               | Baseline compatibility                                  |
| gRPC contract/parity       | In-process canonical services over shared state        | RPC mapping and same-core outcomes               | T3 before REST reaches T2                               |
| PostgreSQL integration     | Disposable schemas on a real server                    | Tables, constraints, indexes, locks, concurrency | Oracle or browser parity                                |
| Differential compatibility | Pinned NetBox and Go systems with isolated databases   | Exact-in-profile REST compatibility              | Vue workflow parity                                     |
| Vue component              | Typed adapters/components with controlled API input    | Hydration, payloads, filters, UI behavior        | Real session/backend/database workflow                  |
| Browser E2E                | Built Vue + Go + PostgreSQL in real Chrome             | Operator Workflow Parity                         | Undeclared capabilities                                 |
| Deployment smoke           | Fresh Compose project and restart                      | Standalone startup/idempotence                   | Production upgrade/operations                           |

## Verification checkpoints

| Checkpoint               | Implementation state                                                  | Exit condition                                                                  | Current status                                                  |
| ------------------------ | --------------------------------------------------------------------- | ------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| V0 repository baseline   | Complete deterministic gate, including non-mutating backend coverage  | `make check` retained green on the exact current source digest                  | Historical v1 pass; source-v2 pending                           |
| V1 identity/RBAC         | Persisted users/groups/grants, sessions/tokens, CLI and focused tests | Trusted-origin CORS plus complete principal/visibility/CLI/session/token matrix | CORS source present; review/evidence and broader matrix pending |
| V2 profile behavior      | Shared 13-resource service and broad positive/negative tests          | Trace every declared invariant, field/filter/relationship, error, and rollback  | Implementation/evidence gap                                     |
| V3 PostgreSQL/deployment | Typed tables, bootstrap/concurrency tests, Compose harness            | Truthful readiness, loss/recovery, real PostgreSQL and deployment artifacts     | Implementation/evidence gap                                     |
| V4 REST/gRPC             | Strict comparator/orchestrator and broad resource/parity suites       | Complete T2 scenario report, then equivalent T3 report                          | T1; implementation gap                                          |
| V5 Vue workflow          | Typed adapters and real-browser harness                               | Both workflows and every required negative/state case retained green            | T1; implementation gap                                          |
| V6 sign-off              | First-profile legacy stacks retired                                   | V0-V5 green together with linked evidence                                       | Not earned                                                      |

## Merge rules

- Tests own their process, database, clock, randomness, ports, and fixtures.
- Do not depend on an already-running developer service, execution order, or
  arbitrary sleeps as synchronization.
- Every defect fix adds the lowest-layer regression that proves it.
- A tier changes only when its required boundary passes and the evidence is
  retained. Compilation, route enumeration, generation, SQLite rollback, and
  one happy path cannot substitute for a stronger boundary.
- Coverage is a non-regression gate. Security and compatibility-critical code
  cannot be excluded to improve the number.
- External artifacts must contain no credentials, cookies, token secrets,
  password material, or DSNs.
