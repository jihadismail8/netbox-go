# Core Workflow browser-session and CSRF lifecycle — 2026-08-17

## Claim boundary

- Goal: `CW1-V1-02`.
- Increment: `CW1-V1-02-I3`.
- Executor: Codex GPT-5.6 Sol.
- State: `done`; the project owner accepted this bounded result at
  `2026-08-19T12:31:02Z`.
- Result: **accepted bounded browser-session and CSRF source passed for the
  exact revision and digest below**.
- Capability Profile: `core-workflow-v1`, still T1 and pre-publication.

This artifact proves only the bounded browser-session slice specified by the
[`CW1-V1-02-I3` packet](../increments/CW1-V1-02-I3.md). It does not complete
the parent credential/session matrix, V1, a compatibility tier, the Capability
Profile, the rewrite, deployment, or production readiness.

The artifact intentionally carries no `netbox-go-evidence-v2` marker. I3
promotes no compatibility tier or identity-extension verification set, and no
profile, scenario, or traceability source changed. The project owner accepted
only this bounded result; the browser-session behavior remains a classified
secure identity extension except for the explicitly pinned REST
authenticator-order and cookie-name alignment.

## Entry and tested source

- Starting revision:
  `08571fe043af7439f01b58889f6e0c8a989503fa` on `main`.
- Starting source digest:
  `source-v2:sha256:61e2821bb85036df3f62219c807086086cf42a0b8b62b149a87cf0c0262e8b03`.
- Starting owned entries: 3,002.
- Final packet-entry revision:
  `7f70b577e6b86841e85e87b832f1da801fe53d57` on `main`.
- Packet-entry source digest:
  `source-v2:sha256:e4c2ac82db77cb5251682c7b49829018b2b2d9b0789ce3f905766f64e33d9bd4`.
- Packet-entry owned entries: 3,003.
- Tested revision:
  `d84717cb09c02aa43dcd79a5f0dc98c0a0b826d1` on `main`.
- Tested source digest:
  `source-v2:sha256:bcb4ef437cd6373abb56cb93980df61234b1a6d9cf10f610fb02ad0f8c9c2f97`.
- Tested owned entries: 3,007.
- Tested source-manifest output SHA-256:
  `dc0f0454d76959957cde961db46a0fca5413211368c21563c71a00af37d2c7ce`.
- Commit state: pushed to `origin/main`; the implementation worktree was clean
  for the exact-commit local, PostgreSQL, and independent CI gates.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.

The final packet-entry revision independently passed:

- Workflow: `repository-gate`.
- Run:
  [32055811194](https://github.com/jihadismail8/netbox-go/actions/runs/32055811194).
- Job:
  [95465581840](https://github.com/jihadismail8/netbox-go/actions/runs/32055811194/job/95465581840).
- Head SHA: `7f70b577e6b86841e85e87b832f1da801fe53d57`.
- Job start: `2026-08-17T18:37:40Z`.
- Job completion: `2026-08-17T18:50:04Z`.
- Run completion: `2026-08-17T18:50:05Z`.
- Conclusion: **success**; every job step succeeded.

## Implemented boundary

- Application session failures now distinguish missing, unknown, expired, and
  inactive-owner credentials from infrastructure, decode, and orphan-owner
  failures without exposing transport concepts.
- Password verification preserves lookup infrastructure failures, treats only
  a password mismatch as an ordinary credential rejection, and performs the
  password comparison before inactive-owner rejection.
- Login always owns one application transaction. It re-reads and revalidates
  user identity, active state, and the verified password hash before deleting
  any presented session and creating the replacement, preventing a stale
  password from minting a session across a concurrent reset.
- Session issue uses one UTC clock sample, a fixed non-sliding 12-hour
  lifetime, fresh server-generated session material, and a deterministic
  session-bound CSRF value. The value is HMAC-SHA-256 over the exact declared
  domain tag with the opaque session string as key; only its SHA-256 digest is
  stored.
- Active-session CSRF bootstrap converges legacy rows transactionally and is
  stable under concurrent responses. A disappearing row maps to the typed
  unknown-session state; infrastructure failures remain internal.
- Logout is session-only, verifies the same session and CSRF state within one
  application transaction, and clears cookies only after committed revocation.
- Baseline and identity-extension REST middleware are valid-session-first. A
  valid session ignores Authorization; CSRF denial never falls through to a
  token. Only expected invalid-session states may fall through, while session
  infrastructure failures stop the request.
- Cookie-authenticated mutation requires exactly one nonempty `csrftoken`
  cookie and `X-CSRFToken` header with literal constant-time equality. Session
  CSRF-safe methods are `GET`, `HEAD`, `OPTIONS`, and `TRACE`; token-safe
  methods remain `GET`, `HEAD`, and `OPTIONS`.
- The canonical session cookie is `sessionid`. Successful issue and deletion
  use the fixed 43,200-second lifetime, declared path, SameSite, HttpOnly, and
  environment-sensitive Secure attributes; the CSRF cookie remains readable.
- PostgreSQL updates only the existing session `csrf_hash` selected by
  `secret_hash`, returns `ErrNotFound` for zero rows, and changes no schema,
  row type, migration, bootstrap rule, or AutoMigrate boundary.

I1 token lookup/touch behavior and I2 baseline REST Token and unary gRPC
behavior remain unchanged. I3 changed no route registration, protobuf,
generated output, OpenAPI, Vue behavior, profile, scenario, traceability,
dependency, toolchain, inventory, comparator, normalizer, or persistence
schema.

## Changed source files

Production and tests:

1. `netbox-backend/internal/application/identity/service.go`
2. `netbox-backend/internal/application/identity/session_errors.go`
3. `netbox-backend/internal/application/identity/session_matrix_test.go`
4. `netbox-backend/internal/application/identity/session_postgres_test.go`
5. `netbox-backend/internal/adapters/postgres/identity/store.go`
6. `netbox-backend/internal/adapters/rest/netbox/identity/http.go`
7. `netbox-backend/internal/adapters/rest/netbox/identity/session_matrix_test.go`
8. `netbox-backend/internal/adapters/rest/netbox/router/router_test.go`
9. `tests/compatibility/run.mjs`
10. `netbox-backend/internal/application/identity/service_integration_test.go`
11. `netbox-backend/internal/adapters/rest/netbox/identity/http_test.go`

Source-included governance:

12. `docs/increments/CW1-V1-02-I3.md`
13. `docs/increments/README.md`
14. `docs/ROADMAP.md`
15. `docs/STATUS.md`
16. `docs/TESTING.md`
17. `docs/IMPLEMENTATION_PLAN.md`
18. `AGENTS.md`

The tested commit contains 4,596 insertions and 199 deletions across exactly
these 18 regular mode-`100644` paths, including four new test/application
files. No file was renamed, retargeted, or made executable. This artifact and
the evidence-ledger link are the two predeclared digest-excluded receipt paths.

## Red-first and independent review evidence

All fourteen packet-named matrices were added before the corresponding
production correction:

```text
TestSessionCredentialOutcomeClassification
TestSessionCSRFOutcomeClassification
TestLoginUsesOneTransactionForPasswordRevalidationAndRotation
TestSessionCSRFDerivationAndRecovery
TestLogoutRevalidatesAndRevokesInOneTransaction
TestRESTSessionFirstCredentialArbitration
TestRESTSessionCSRFPairsAndMethodSafety
TestRESTCSRFBootstrapRecovery
TestRESTSessionCookieLifecycleContract
TestRESTLogoutIsSessionOnly
TestPostgresBrowserSessionRollback
TestPostgresSessionCSRFRecoveryDurability
TestPostgresLoginRevalidatesAgainstConcurrentPasswordReset
TestPostgresConcurrentLogoutHasSingleRevocation
```

The red runs exposed the intended missing typed outcomes and APIs,
nontransactional login/logout, token-first arbitration, incomplete CSRF pair,
incorrect TRACE safety, active-bootstrap poisoning, old cookie name/lifetime,
and token-capable logout. The PostgreSQL suite initially compile-failed only on
the deliberately absent store/service operations. Diagnostics used semantic
labels and did not print reusable material.

Independent read-only reviews then checked the complete source candidate and
caused bounded evidence strengthening before commit: corrupt password verifiers
remain internal, second-lookup CSRF races retain typed application causes while
rendering 403 after session authentication, login negative/order rows prove no
throttle consumption or cookie issue, exact HMAC binding is checked without
printing values, PostgreSQL competitors are observed waiting on the actual
advisory lock, and pre-existing credential-bearing assertions use generic
constant-time/boolean diagnostics. Final code, governance, and secret/scope
reviews found no blocker.

## Focused and race results

The clean tested commit passed the exact focused application and REST commands
from `2026-08-19T06:39:12Z` through `2026-08-19T06:39:15Z`, exit 0:

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/application/identity \
  -run '^(TestSessionCredentialOutcomeClassification|TestSessionCSRFOutcomeClassification|TestLoginUsesOneTransactionForPasswordRevalidationAndRotation|TestSessionCSRFDerivationAndRecovery|TestLogoutRevalidatesAndRevokesInOneTransaction)$' \
  -count=1

env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/adapters/rest/netbox/identity \
  -run '^(TestRESTSessionFirstCredentialArbitration|TestRESTSessionCSRFPairsAndMethodSafety|TestRESTCSRFBootstrapRecovery|TestRESTSessionCookieLifecycleContract|TestRESTLogoutIsSessionOnly)$' \
  -count=1
```

The exact affected-package race command passed from
`2026-08-19T06:39:15Z` through `2026-08-19T06:40:35Z`, exit 0:

```bash
env GOCACHE=/tmp/go-cache GOFLAGS=-buildvcs=false \
  go test -race \
  ./internal/application/identity \
  ./internal/adapters/postgres/identity \
  ./internal/adapters/rest/netbox/identity \
  ./internal/adapters/rest/netbox/router \
  ./internal/adapters/grpc/identity \
  ./test/parity -count=1
```

Race results were application identity 79.670s, REST identity 28.833s,
composed REST router 11.000s, gRPC identity 1.012s, and REST/gRPC parity
22.281s. The PostgreSQL identity adapter has no direct test files; it is
covered through the application PostgreSQL matrices and complete L4 run.

The corresponding non-race affected-package command also passed on the same
digest before commit.

## Real-PostgreSQL L4

The clean tested commit passed the required focused and complete L4 commands
from `2026-08-19T06:22:09Z` through `2026-08-19T06:22:22Z`, exit 0. The target
was an owned disposable PostgreSQL 17 database on local host port 5433. The
generated password and complete DSN were never printed or retained.

Focused command, with the credential deliberately redacted:

```bash
NETBOX_TEST_POSTGRES_DSN='<owned-disposable-postgres-dsn>' \
  go test ./internal/application/identity \
  -run '^TestPostgres(BrowserSessionRollback|SessionCSRFRecoveryDurability|LoginRevalidatesAgainstConcurrentPasswordReset|ConcurrentLogoutHasSingleRevocation)$' \
  -count=1 -v
```

All four top-level tests were discovered, none skipped, and every test and
subtest passed. They prove rollback after delete/create and CSRF update faults,
durable/no-op/zero-row CSRF repair, both recovery/logout lock orders, stale
password rejection across a concurrent reset, and a single committed deletion
under concurrent logout.

The complete playbook L4 command then passed all eleven explicit packages:

```bash
NETBOX_TEST_POSTGRES_DSN='<owned-disposable-postgres-dsn>' \
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
  ./cmd/netbox_go_admin -count=1
```

The runner first proved the exact role and database names absent, created only
those disposable objects without elevated role attributes, and installed a
cleanup trap. After the tests it terminated only exact-target connections,
dropped the exact database and role, and proved role, database, and matching
connection counts were all zero. A sorted fingerprint over all non-target
database, role, membership, and shared-dependency rows remained
`5984bbaa79b8facce49aaaa291c2b11e15c7f061e1d5a2d9cfc72563fb8634ba`
before and after cleanup.

## Repository L0-L3 and independent CI

The final timestamped exact-commit root gate passed from
`2026-08-19T06:42:27Z` through `2026-08-19T06:44:38Z`, exit 0:

```bash
env \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

Results included:

- backend lint, tests, builds, architecture, generated/frozen checks, and
  coverage policy passed;
- backend coverage 12.4997% (36,338/290,712) met the retained 12.4001%
  (36,013/290,425) floor;
- application identity coverage was 84.1%, REST identity 88.1%, and the
  composed REST router 86.8%;
- frontend formatting, lint, application/test typechecks, coverage tests, and
  production build passed; 28 files and 125 tests passed at 61.72% statement
  coverage;
- inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries;
- traceability validator regressions passed 96/96;
- the active profile remained 13 resources, 3 interfaces, 17 scenarios, and
  293 traceability rows;
- OpenAPI remained 33 paths and 86 operations; and
- all local links were valid across 155 Markdown files.

An earlier shell attempt selected host Node 22/npm 10. The repository correctly
rejected it at the pinned frontend toolchain check. The source digest remained
unchanged, and that rejected attempt is not used as pass evidence.

Independent candidate result:

- Workflow: `repository-gate`.
- Run:
  [32223160072](https://github.com/jihadismail8/netbox-go/actions/runs/32223160072).
- Job:
  [95977463498](https://github.com/jihadismail8/netbox-go/actions/runs/32223160072/job/95977463498).
- Head SHA: `d84717cb09c02aa43dcd79a5f0dc98c0a0b826d1`.
- Run created: `2026-08-19T06:22:44Z`.
- Job start: `2026-08-19T06:22:47Z`.
- Job completion: `2026-08-19T06:37:35Z`.
- Run completion: `2026-08-19T06:37:36Z`.
- Conclusion: **success**; every setup, repository-gate, post, and completion
  step succeeded.

Evidence-receipt result before project-owner acceptance:

- Workflow: `repository-gate`.
- Evidence-receipt revision:
  `a30a60b2cbc2445173fcd92ecbcd8b40f5370674` on `main`.
- Source digest:
  `source-v2:sha256:bcb4ef437cd6373abb56cb93980df61234b1a6d9cf10f610fb02ad0f8c9c2f97`.
- Owned entries: 3,007.
- Run:
  [32225681378](https://github.com/jihadismail8/netbox-go/actions/runs/32225681378).
- Job:
  [96064318118](https://github.com/jihadismail8/netbox-go/actions/runs/32225681378/job/96064318118).
- Head SHA: `a30a60b2cbc2445173fcd92ecbcd8b40f5370674`.
- Job start: `2026-08-19T12:16:36Z`.
- Job completion: `2026-08-19T12:30:00Z`.
- Run completion: `2026-08-19T12:30:01Z`.
- Conclusion: **success**; every job step succeeded.

## Gate disposition

- L0-L3: passed locally on the exact tested digest; repository CI passed on
  the exact tested commit.
- L4 real PostgreSQL: passed locally with actual rollback, durability,
  advisory-lock wait, reset race, and concurrent logout evidence.
- L5 differential REST: skipped because I3 does not close a complete REST T2
  owner-row boundary.
- L6 gRPC: affected local gRPC and parity regressions passed; promotion was
  skipped because cookie/CSRF has no gRPC equivalent and no corresponding T2
  boundary exists.
- L7 deployment: not applicable because deployment wiring did not change.
- L8 browser: skipped; the compatibility harness cookie-name correction is
  syntax-checked only and no T4 browser result is claimed.

No skipped boundary is implied by the local or CI results.

## Digest, secret, and containment checks

- The tested source digest remained
  `source-v2:sha256:bcb4ef437cd6373abb56cb93980df61234b1a6d9cf10f610fb02ad0f8c9c2f97`
  with 3,007 entries across focused, affected, race, PostgreSQL, backend, root,
  exact-commit, push, and candidate-CI gates.
- Go formatting, Prettier, compatibility JavaScript syntax, local links,
  trailing-whitespace, scoped-diff, generated/frozen, and source-manifest
  checks passed.
- Three independent final reviews found no code, transaction, concurrency,
  compatibility, governance, scope, or secret-containment blocker.
- A scoped high-risk signature scan found no private-key marker,
  credential-bearing database URI, JWT, provider token, or live credential.
- Test fixtures use synthetic opaque values. Credential comparisons use
  constant-time or boolean helpers with generic diagnostics, and no reusable
  value appears in test names, logs retained here, or evidence links.
- This artifact and its ledger link are excluded from `source-v2`; their
  evidence-only commit preserved the tested digest, entry count, and manifest
  and passed its exact-SHA repository CI before project-owner acceptance.

## Human closeout attestation

The project owner accepted this bounded result in the project thread at
`2026-08-19T12:31:02Z`, after the evidence-receipt exact-SHA CI completed
successfully. The acceptance changes only the state of `CW1-V1-02-I3` from
`evidence` to `done`; it does not close `CW1-V1-02`, change a traceability row,
complete an identity verification axis, or promote V1, T2, T3, T4, the
Capability Profile, publication, deployment, or production readiness.

- Tested implementation revision:
  `d84717cb09c02aa43dcd79a5f0dc98c0a0b826d1`.
- Tested implementation digest:
  `source-v2:sha256:bcb4ef437cd6373abb56cb93980df61234b1a6d9cf10f610fb02ad0f8c9c2f97`
  with 3,007 entries.
- Pre-closeout evidence revision:
  `a30a60b2cbc2445173fcd92ecbcd8b40f5370674`; its source digest remained the
  tested implementation digest.
- Accepted closeout revision:
  `0a4043e3381fc34d3a226c624c8e0b02df73f45c`.
- Accepted closeout digest:
  `source-v2:sha256:f98db5a0e1beb0590ad8eac6be5dbdcc0c6a962c85f592a005bfe8d448848b6d`
  with 3,007 entries.
- Accepted closeout manifest-output SHA-256:
  `fbc906a3f82a9dd5ae8b8bc1337aadd95e156fc73dbdc9b9b05186bad9422421`.
- Exact-commit local result: the pinned-environment `make check` ran from
  `2026-08-19T12:49:07Z` through `2026-08-19T12:51:13Z` at exit 0. Backend
  coverage was 12.4997% (36,338/290,712) against the retained 12.4001%
  (36,013/290,425) floor; 28 frontend files and 125 tests passed at 61.72%
  statement coverage; 96 traceability-validator tests passed; the active
  profile remained 13 resources, 3 interfaces, 17 scenarios, and 293 rows;
  inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries; OpenAPI remained 33 paths and 86 operations; and
  all local links were valid across 156 Markdown files.
- Exact-commit independent result: repository-gate
  [run 32254507381](https://github.com/jihadismail8/netbox-go/actions/runs/32254507381),
  [job 96072971766](https://github.com/jihadismail8/netbox-go/actions/runs/32254507381/job/96072971766),
  started `2026-08-19T12:47:33Z`, completed
  `2026-08-19T13:01:01Z`, and reported success for every step at the accepted
  closeout SHA.

The exact local command was:

```bash
env \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

Replacement- and graft-disabled Git reads reconstructed the reachable before
and after trees. Every changed entry retained mode `100644`. The exact
`a30a60b2` to `0a4043e3` file mapping is below; hashes are SHA-256 over the raw
blob bytes.

| Path                                                                 | Before bytes | Before SHA-256                                                     | After bytes | After SHA-256                                                      |
| -------------------------------------------------------------------- | -----------: | ------------------------------------------------------------------ | ----------: | ------------------------------------------------------------------ |
| `AGENTS.md`                                                          |        5,452 | `7406aaffcd87e876dfb9098e63b533ec317ffd01db1a2125be362862d81d2ecb` |       5,427 | `5356667f45c88ee1f5fc0556f04c93de55b3ee74c02e1366cca564ef69839971` |
| `docs/IMPLEMENTATION_PLAN.md`                                        |       85,313 | `54efa05c63b0ffbe6bf45d741ffeaeb6458058e95ba7c5f3f08239a578e615ee` |      85,430 | `854c60ead34f0182749a83449c2b4c69bf3f3244dcb168a356a42cdb8b17acb3` |
| `docs/ROADMAP.md`                                                    |       75,693 | `afd97037bd3dc8eab53e09b38e39d714318a2d397f2049833d2611a8833fd4b7` |      75,664 | `4e1e14b47b0c54b9fef55f25d12a38390e4e215341e298b9f8bce4ba26e43787` |
| `docs/STATUS.md`                                                     |       22,781 | `9e7b59f28b5c46da945886fb9e4351311684f9919ce5309e0eba49712a630af9` |      22,364 | `7241eeca929d3b53baac02635b4b603a488247bb83e5c638a54fff375211c2f1` |
| `docs/TESTING.md`                                                    |       13,509 | `b60b4be513b19123bef2a58dfe036b947fb7e1ae03259c0de0d55f1ea25a1666` |      13,518 | `a4f1d051643f50d136b199cf91a322042595f576f974b2bd4ae9819b8bd057fb` |
| `docs/increments/CW1-V1-02-I3.md`                                    |       37,678 | `39354cf861f2909364241ab873088f3b360c1bea0c866e14413b775d90d3c930` |      37,995 | `47f3402ec90cda4dc82d0d331e2d17eb47a805ebdd0981024d52afe90f1d85ed` |
| `docs/increments/README.md`                                          |        3,208 | `e1d7f07a8947ed4cd86a2204e9c6d7dd82d9278402d17cee61c4da7b1fb06301` |       3,180 | `36920c8a4b60bb942d9b2fc371de31e4d16d2d789555ab840a803cbc1c86d788` |
| `docs/evidence/2026-08-17-core-workflow-v1-browser-session-i3-v0.md` |       17,436 | `3946219e68ca949671138dedc47bd38d424d5062710bca2993f229b78d5e241e` |      18,109 | `688855711a61167b02e074eb4c216aef370a1ab0cc345b5fafff89fec0321b44` |
| `docs/evidence/README.md`                                            |        6,884 | `a53d568d7356f78aae0c3a705d40ab01bcfd39de7cbebaf5acae5128e57909d4` |       6,802 | `4d123afaaff28da26bb4e0ffaa37917da072009ae666f7988989b3d591d96a4c` |

The first seven rows are source-v2 inputs and constitute the complete
claim-only source diff. The final two rows are excluded evidence paths. No
entry was added, removed, renamed, retargeted, or made executable, and no
other byte changed. The complete manifests retained 3,007 entries before and
after the claim. The pre-closeout manifest-output SHA-256 was
`dc0f0454d76959957cde961db46a0fca5413211368c21563c71a00af37d2c7ce`.

The claim-only transition changed no application behavior, route, schema,
migration, security rule, scenario, fixture, comparator, normalizer,
dependency, toolchain, coverage policy, verification set, or parent-goal
state. This artifact still intentionally carries no `netbox-go-evidence-v2`
marker because I3 promotes no traceability consumer or extension verification
set. The evidence-only receipt containing this attestation is excluded from
`source-v2`; it must preserve the accepted closeout digest and pass its own
exact-SHA repository CI.

Human reviewer: project owner.

Human decision: accepted only `CW1-V1-02-I3` in the project thread at
`2026-08-19T12:31:02Z`; the parent goal remains open.

## Residual work and risk

The following `CW1-V1-02` work remains open and unclaimed:

- password-change current-session policy and password/session invalidation;
- login and token-creation throttle policy, storage, boundary, and parity;
- trusted-proxy and forwarded-client-source resolution;
- gRPC Health Watch, reflection, and unknown-stream authentication;
- Django Origin/HTTPS Referer checks, masked-CSRF grammar, configurable or
  sliding session lifetime, and stale-cookie cleanup;
- simultaneous-credential provenance beyond the valid-session-first boundary;
- aggregate all-transport credential closeout; and
- real-browser T4 behavior.

The current frontend dependency audit still reports three high-severity
advisories. They predate I3 and were not changed or waived; dependency changes
were outside this increment.

The parent `CW1-V1-02`, V1, T2/T3/T4, Capability Profile publication, full
rewrite, deployment, and production-readiness claims all remain open. The
project owner accepted the bounded I3 result, so I3 is `done`; no broader claim
changes.
