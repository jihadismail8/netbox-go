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
it used the superseded mode-blind v1 digest. The
[source-v2 artifact](evidence/2026-08-17-core-workflow-v1-source-v2-v0.md)
retains the exact entry revision after human review. Every later source change
must refresh that continuous gate. Accepted bounded `CW1-V1-02-I1` evidence
separately proves token lookup classification, strict `last_used` ordering, affected local
REST/gRPC mapping, and actual-update behavior on an owned disposable
PostgreSQL schema. It does not complete the parent credential matrix or promote
T2/T3. Accepted bounded `CW1-V1-02-I2` evidence adds the exact baseline REST Token
grammar and details, typed credential-state causes, strict unary gRPC Bearer
grammar, an explicit 28-read/57-write method classification over 85 protected
RPCs, malformed-input zero-lookup containment, and cross-transport
principal/cause/effect parity. It remains direct-peer-only and does not close
session, CSRF, throttle, trusted-proxy, simultaneous-credential, or streaming
rows. Accepted bounded `CW1-V1-02-I3` evidence adds typed browser-session outcomes,
valid-session-first REST arbitration, exact double-submit CSRF pairing,
transactional login/logout, deterministic active-session CSRF recovery, fixed
session-cookie shape, and real-PostgreSQL rollback/advisory-lock/concurrency
proof. It remains a classified browser-identity extension and does not close
password-session, throttle, trusted-proxy, streaming, aggregate all-transport,
REST T2, gRPC T3, or browser T4 work. GitHub CI currently runs the V0 command
only. `CW1-V1-02-I4` has an exact tested candidate at
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f`, source digest
`source-v2:sha256:3f37417bb791ac6bc97ac4a0d23c5f928062feecf81a0f8a4fb9e57445d53670`
over 3,013 entries. Its rebased red-first, focused, PostgreSQL,
affected-package, race, backend, root, feature-candidate CI, and main exact-SHA
CI boundaries passed. The current nine-path claim-only transition records
bounded `evidence` only if the exact attestation revision passes repository CI;
its digest-excluded receipt and project-owner review then remain. The owned
PostgreSQL, deployment, differential, gRPC-promotion, and
browser gates remain manual until connected jobs retain durable artifacts.

`CW1-V2-02-I1` has an exact tested candidate at
`7acba402f0de2bd59e5b342a6f05df268bc9120b`, source digest
`source-v2:sha256:ed330b0a5bbeafd70b7b16a4ce4d1052fa9a385313a3b8827b554983571c1b43`
over 3,018 entries. Its eight named domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the affected race,
generated-contract, Vue adapter/form, complete L4, pinned repository, and
independent exact-candidate CI boundaries. Its evidence claim and
pre-acceptance receipt also passed exact-SHA repository CI. The project owner
accepted only this bounded result at `2026-08-23T17:25:54Z`; the
owner-accepted closeout claim and excluded closeout receipt also passed
exact-SHA CI, so I1 is effectively `done`. The
pinned REST oracle remained unavailable. The local transport/parity and Vue
results do not substitute for that oracle or promote T2, T3, or T4.

`CW1-V2-02-I2` has an exact tested candidate at
`87863efd38fe71dfa05c818b860b37b7e94d67b4`, source digest
`source-v2:sha256:c7c1b86c2bcd768bb719149a54dddbceccf2b5b2e4087dd4b79eec20bef5a37c`
over 3,022 entries. Its eight named Site domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 17-test Vue adapter/form, complete L4, pinned repository,
and independent exact-candidate CI boundaries. Its evidence claim and
pre-acceptance receipt also passed exact-SHA repository CI. The project owner
accepted only this bounded result at `2026-08-24T04:46:51Z`; the
owner-accepted closeout claim and excluded closeout receipt both passed
exact-SHA repository CI, so I2 is effectively `done`. Docker
rejected the differential harness's temporary source bind before oracle
execution. The local transport/parity and Vue results do not substitute for
that oracle or promote T2, T3, or T4.

`CW1-V2-02-I3` has an exact tested candidate at
`651d33bc3fb2c8e663b6b14320af405b8501471f`, source digest
`source-v2:sha256:09499a6618569d2dae224edfb339ac82585bf0248d20e5a4d5ff23d19221fe6f`
over 3,029 entries. Its eight named Manufacturer domain, application, REST,
gRPC, real-PostgreSQL, and parity tests passed, as did the exact
affected/race, generated-contract, 23-test Vue adapter/form, complete L4,
pinned repository, and independent exact-candidate CI boundaries. Its evidence
claim and pre-acceptance receipt also passed exact-SHA repository CI. The
project owner's acceptance of only this bounded result was recorded at
`2026-08-24T10:28:34Z`; its owner-accepted closeout claim and excluded
closeout receipt both passed exact-SHA repository CI, so I3 is effectively
`done`. No retained pinned differential accompanies this result. The local
transport/parity and Vue results do not substitute for that oracle or promote
T2, T3, or T4.

`CW1-V2-02-I4` has exact tested candidate
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe` at source digest
`source-v2:sha256:68db7a9835545d66ef9a651b9c4a000c91f3d834b5503a1b89e4a122275c3bc9`
over 3,036 entries. Its exact-name, affected/race, real-PostgreSQL, complete
L4, generated-contract, Vue, backend/frontend/root, candidate-CI,
evidence-claim, and pre-acceptance receipt boundaries are green. The project
owner accepted only this bounded result at `2026-08-24T18:51:56Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I4 is effectively `done`. The matrix remains limited to RackRole
`name`, `slug`, `color`, and
`description` create/PUT/PATCH omitted/null/blank/concrete states across eight
fixed domain, application, REST, gRPC, real-PostgreSQL, and parity tests,
generated OpenAPI, and Vue adapter/form checks. No differential was run, so
T2/T3 remain unearned; the Vue boundary earns no T4.

`CW1-V2-02-I5` has exact tested candidate
`89507d95d2743de7f97d64ca14cc43f6b834770b` at source digest
`source-v2:sha256:a8325eaae703aa801ed587deefae7e8d08d9c9e0189c80ff7569da95c36d6f90`
over 3,043 entries. Its eight named RackType domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 35-test Vue adapter/form, complete L4, backend/frontend/
root, candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-25T03:56:48Z`.
`Done` becomes effective when the current owner-accepted closeout claim passes
exact-SHA CI; its excluded receipt then remains. The matrix is limited to the
ten declared RackType writable fields across create/PUT/PATCH presence,
numeric-ID Manufacturer input, generated OpenAPI, shared REST/gRPC semantics,
PostgreSQL durability, and Vue dirty-field handling. No differential was run,
so T2/T3 remain unearned; the Vue boundary earns no T4. RackType uniqueness,
propagation, deletion, list/query behavior, alternate Manufacturer inputs,
full CRUD, the parent, profile promotion, and every later child remain open.

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

| Checkpoint               | Implementation state                                                                       | Exit condition                                                                                                            | Current status                                           |
| ------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| V0 repository baseline   | Complete deterministic gate, including non-mutating backend coverage                       | `make check` retained green on each exact relevant source digest                                                          | Entry source-v2 retained; active-source refresh required |
| V1 identity/RBAC         | Persisted users/groups/grants, sessions/tokens, CLI and focused tests                      | Pass I4's exact claim CI, retain its receipt, obtain owner review, then complete the remaining identity matrix            | I4 evidence claim conditional; parent open               |
| V2 profile behavior      | Shared 13-resource service, broad tests, and accepted IPAddress I1/Site I2/Manufacturer I3/RackRole I4 plus owner-accepted RackType I5 | Retain I5 closeout, then trace every remaining invariant, rollback, and external-tier boundary | I1/I2/I3/I4 done; I5 done claim conditional; parent open |
| V3 PostgreSQL/deployment | Typed tables, bootstrap/concurrency tests, Compose harness                                 | Truthful readiness, loss/recovery, real PostgreSQL and deployment artifacts                                               | Implementation/evidence gap                              |
| V4 REST/gRPC             | Strict comparator/orchestrator and broad resource/parity suites                            | Complete T2 scenario report, then equivalent T3 report                                                                    | T1; implementation gap                                   |
| V5 Vue workflow          | Typed adapters and real-browser harness                                                    | Both workflows and every required negative/state case retained green                                                      | T1; implementation gap                                   |
| V6 sign-off              | First-profile legacy stacks retired                                                        | V0-V5 green together with linked evidence                                                                                 | Not earned                                               |

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
