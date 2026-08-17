# Core Workflow token-authentication foundation — 2026-08-17

## Claim boundary

- Goal: `CW1-V1-02`.
- Increment: `CW1-V1-02-I1`.
- Executor: Codex GPT-5.6 Sol.
- State: `evidence`; human acceptance is pending.
- Result: **retained for the exact source below**.
- Capability Profile: `core-workflow-v1`, still T1 and pre-publication.

This artifact proves only the bounded token lookup, owner-load, `last_used`,
revocation, restriction-ordering, touch-failure, and affected local REST/gRPC
mapping foundation specified by the
[`CW1-V1-02` packet](../increments/CW1-V1-02.md). It does not complete the
parent credential/session matrix, close V1, earn T2/T3/T4, change traceability
rows, publish the profile, complete a module or the rewrite, prove deployment,
or claim production readiness. It intentionally carries no
`netbox-go-evidence-v2` marker because no retained compatibility or extension
proof dimension is promoted by this increment.

## Entry and tested source

- Starting revision:
  `5d6f9cf9f5a5dc5e5e09de58f4f0e76e0f7a6381`.
- Starting source digest:
  `source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82`.
- Starting owned entries: 2,993.
- Starting worktree: clean.
- Tested revision:
  `38485b08c67e11f115a5fbda5a20a1072040c5cd` on `main`.
- Tested/current implementation digest:
  `source-v2:sha256:e3ec8eb91d6e035863c63d7bf60c614dd2bf4e6967ca77add2a931ae8f514c7e`.
- Tested owned entries: 2,998.
- Commit state: pushed to `origin/main`; the implementation worktree was clean
  before and after the exact-commit gates.
- Pinned upstream oracle revision:
  `fbb948d30e79ce657fac62994a22aca72c1770a9`.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.
- Increment execution start: `2026-08-17T05:39:47Z`, after the entry
  claim-closeout CI completed.
- Evidence completion: `2026-08-17T06:45:52Z`, when the exact implementation
  commit's independent CI job completed successfully.

The entry revision had retained
[`CW1-G00` source-v2 evidence](2026-08-17-core-workflow-v1-source-v2-v0.md),
completed
[`CW1-V1-01` CORS evidence](2026-08-03-core-workflow-v1-cors-v0.md), and a
successful claim-closeout
[GitHub run 31997963053](https://github.com/jihadismail8/netbox-go/actions/runs/31997963053)
before this increment changed production source.

## Implemented boundary

- Empty and unknown token material remains unauthenticated and causes no
  token-row touch.
- Only an absent token key maps to application `ErrNotFound`. A persisted token
  whose owner row cannot be loaded is an internal storage failure; its cause is
  retained server-side and no user or durable token update is returned.
- A soft-revoked token behaves like the baseline's deleted row and cannot be
  touched, including by a stale caller which loaded it before revocation.
- A recognized non-revoked token is eligible for `last_used` only when the
  stored timestamp is absent or more than 60 seconds old. Exactly 60 seconds is
  not eligible.
- The conditional PostgreSQL update independently enforces the strict time
  predicate and `revoked_at IS NULL`.
- Eligible touch occurs before expiry, inactive-owner, allowed-IP, and unsafe
  write-permission rejection. Source-IP denial precedes write denial.
- A touch failure stops authentication as typed internal failure instead of
  being ignored.
- Affected REST failures are safe HTTP 500 responses without invoking the
  protected handler. Affected gRPC failures are safe `Internal` responses;
  combined IP/write denial is `Unauthenticated` and does not invoke the RPC.

No public route, middleware production path, protobuf, OpenAPI, Vue source,
profile, scenario, inventory, comparator, normalizer, migration, schema model,
dependency, toolchain, generated file, coverage exclusion, or legacy stack was
changed.

## Changed files

- `netbox-backend/internal/application/identity/service.go` — token lookup
  classification, revocation containment, strict touch boundary, failure
  propagation, and restriction ordering.
- `netbox-backend/internal/adapters/postgres/identity/store.go` — distinct
  token/owner miss handling and the strict revocation/time update predicate.
- `netbox-backend/internal/application/identity/credential_matrix_test.go` —
  shared application matrix.
- `netbox-backend/internal/application/identity/credential_postgres_test.go` —
  real-PostgreSQL classification, precision, durability, concurrency, and
  failure proof.
- `netbox-backend/internal/adapters/rest/netbox/identity/credential_matrix_test.go`
  — affected HTTP mapping and handler-containment proof.
- `netbox-backend/internal/adapters/grpc/identity/credential_matrix_test.go` —
  affected gRPC mapping and handler-containment proof.
- `AGENTS.md`, `docs/IMPLEMENTATION_PLAN.md`, `docs/ROADMAP.md`,
  `docs/STATUS.md`, `docs/TESTING.md`, `docs/increments/CW1-V1-01.md`,
  `docs/increments/CW1-V1-02.md`, and `docs/increments/README.md` — reconcile
  only the accepted entry state and bounded I1 evidence state.

## Red-first and review evidence

The application matrix was run against the pre-correction implementation. It
failed the intended lookup-infrastructure, exact-60-second, revoked-touch,
IP/write-precedence, and ignored-touch-error rows. Supporting REST diagnostics
showed infrastructure failures taking the wrong 403/success paths; supporting
gRPC diagnostics showed the wrong `Unauthenticated`/handler-success and
`PermissionDenied` paths. No credential material was printed. These failures
were diagnostic red baselines, not retained pass evidence.

Two independent read-only reviews then checked the complete candidate. Review
found and caused correction of two issues before commit:

1. owner-row not-found had been indistinguishable from an unknown token key;
2. the first orphan fixture depended on superuser-only trigger disabling.

The final code reserves application `ErrNotFound` for an absent token, and the
final fixture uses a catalog-discovered foreign-key drop inside a transaction
owned by the test role. The orphan update, authentication, audit assertion,
constraint change, and row change share that transaction; rollback restores
the row and foreign key. The final independent re-review reported no blocker.

## Focused exact-commit matrix

The retained exact-commit focused/race/PostgreSQL interval ran from
`2026-08-17T06:29:35Z` through `2026-08-17T06:32:46Z`, with every command at
exit 0.

- Application: 4 top-level tests and 29 subcases passed.
- REST: 2 top-level tests and 3 exercised cases passed.
- gRPC: 1 top-level test and 3 exercised cases passed.
- Existing affected application, REST, gRPC, and parity packages passed.
- Race results: application 52.019s, REST 28.597s, gRPC 1.008s, and parity
  22.355s.
- Exact 60-second application and PostgreSQL attempts performed zero touch;
  later-than-60-second attempts performed one.
- Touch failures returned internal and stopped before later restrictions.
- Unknown, orphaned-owner, and revoked cases performed zero actual token-row
  updates.

Commands:

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/application/identity \
  -run '^TestTokenCredentialMatrix' -count=1

env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/grpc/identity \
  ./test/parity \
  -run 'TokenCredentialMatrix|Authentication|UnaryAuthenticator|RESTAndGRPC' \
  -count=1

env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test \
  ./internal/application/identity \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/grpc/identity \
  ./test/parity -count=1

env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test -race \
  ./internal/application/identity \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/grpc/identity \
  ./test/parity -count=1
```

Test-discovery commands separately found all 4 application tests, both REST
tests, the gRPC test, and the PostgreSQL suite before execution.

## Real PostgreSQL L4

The exact-commit suite passed 9/9 subtests in 0.221s against an owned disposable
PostgreSQL 16 database:

1. unknown key performs no durable write;
2. orphaned owner is internal without durable touch;
3. recognized key initializes `last_used`;
4. application boundary is strict;
5. store boundary is strict;
6. stored microsecond precision is preserved;
7. coordinated stale callers produce one actual update;
8. revoked key performs no durable touch; and
9. touch failure is fail closed.

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  NETBOX_TEST_POSTGRES_DSN='<owned-disposable-postgres-dsn>' \
  go test ./internal/application/identity \
  -run '^TestPostgresTokenCredentialDurability$' -count=1 -v
```

The suite created a uniquely named empty schema, ran the explicit eight-entry
missing-table-only bootstrap registry, installed an all-UPDATE audit trigger,
and dropped only its schema. A separate same-source run passed all 9 subtests
using a temporary login role with `NOSUPERUSER`, `NOCREATEDB`, `NOCREATEROLE`,
and `NOINHERIT`, granted only database-level schema creation. That role was
revoked and dropped. The dedicated container used no volume and was stopped
and auto-removed after the final run; its temporary data is not recoverable.

## Repository L2/L3 and independent CI

The final exact-commit local `make check` passed at exit 0 within the retained
interval above:

```bash
env \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  GOCACHE=/tmp/go-cache \
  GOFLAGS=-buildvcs=false \
  make check
```

Results included:

- backend lint, tests/race, builds, architecture, generated/frozen checks, and
  coverage policy passed;
- backend coverage 12.4053% (36,029/290,433) met the retained 12.4001%
  (36,013/290,425) floor across 54 measured packages and one reviewed
  exclusion;
- frontend formatting, lint, application/test typechecks, coverage tests, and
  production build passed; 28 files and 125 tests passed at 61.72% statement
  coverage;
- inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries;
- oracle-source materialization self-test passed;
- traceability validator regressions passed 96/96;
- the active profile remained 13 resources, 3 interfaces, 17 scenarios, and
  293 traceability rows;
- OpenAPI remained 33 paths and 86 operations; and
- all local links were valid across 151 Markdown files.

Independent GitHub result:

- Workflow: `repository-gate`.
- Run:
  [32002130248](https://github.com/jihadismail8/netbox-go/actions/runs/32002130248).
- Job:
  [95304337018](https://github.com/jihadismail8/netbox-go/actions/runs/32002130248/job/95304337018).
- Head SHA:
  `38485b08c67e11f115a5fbda5a20a1072040c5cd`.
- Start: `2026-08-17T06:33:18Z`.
- Completion: `2026-08-17T06:45:52Z`.
- Conclusion: **success**.

Every setup, repository-gate, post, and completion step reported success.

## Digest, secret, and containment checks

- The implementation digest remained
  `source-v2:sha256:e3ec8eb91d6e035863c63d7bf60c614dd2bf4e6967ca77add2a931ae8f514c7e`
  with 2,998 entries before and after all non-writing gates.
- Go formatting, Prettier, trailing-whitespace checks, and local links passed.
- A scoped filename-only scan found no private-key marker, credential-bearing
  PostgreSQL URI, long literal authorization credential, or JWT pattern.
- No raw password, cookie, CSRF value, API/bearer token, secret hash, DSN, or
  complete configuration was retained.
- An initial sandbox run could not open loopback sockets. The exact commands
  were rerun in the approved environment and passed; the sandbox denial was not
  treated as a behavioral failure.

## External gates not claimed

- L5 differential REST is deferred because I1 promotes no REST T2 row.
- L6 promotion-grade gRPC evidence is deferred because no T3 row is promoted;
  affected local gRPC and existing parity tests were still mandatory and
  passed.
- L7 deployment smoke is not applicable to this source slice because no
  deployment wiring changed and it cannot substitute for L4.
- L8 real-browser evidence is deferred because no Vue/browser behavior or T4
  row changed.

None of these skipped boundaries is implied by the local or CI results.

## Residual work and review state

- The full REST and gRPC credential-header/metadata shape and outcome matrix is
  still open.
- Valid, invalid, expired, refreshed, rotated, logged-out, and
  password-invalidated browser sessions remain open.
- Cookie-authenticated mutation CSRF, session fixation/rotation, login
  throttling, and token-creation throttling remain open.
- `CW1-V1-03`, `CW1-V1-04`, V2 behavior gaps, V3 deployment readiness, and all
  T2/T3/T4/V6 boundaries remain open.
- `npm ci` continues to report three high-severity dependency advisories. No
  dependency or lockfile change was authorized; remediation requires a
  separate reviewed increment.
- GitHub CI runs V0 only; the credential-free committed summary retains the
  operator-run L4 receipt.

Human reviewer: pending.

Human decision: pending. `CW1-V1-02-I1` remains `evidence`, not `done`, and the
parent goal remains open.

Next bounded increment: `CW1-V1-02-I2`, the REST/gRPC API-token
header/metadata-shape and stable public-outcome matrix.
