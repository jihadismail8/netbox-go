# Trusted-origin CORS evidence — 2026-08-17

## Claim boundary

- Goal: `CW1-V1-01`.
- Increment: `CW1-V1-01-I1`.
- Executor: Codex GPT-5.6 Sol.
- State: `done` after separate human review.
- Result: **retained**.

This artifact proves only the reviewed trusted-origin CORS increment. It does
not close V1, earn T2/T3/T4, complete the Capability Profile, prove V3, claim
complete replacement, or claim production readiness.

## Tested source and entry

- Tested revision:
  `07aaca9264e5af9830e39fc95dd3289c0b5c9476` on `main`.
- Tested/current source digest:
  `source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82`.
- Owned entries: 2,993.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.
- Exact commit was pushed to `origin/main` and independently passed
  [GitHub repository-gate run 31994516467](https://github.com/jihadismail8/netbox-go/actions/runs/31994516467).

The committed Git manifest and clean worktree produced the same source-v2
digest. No generated configuration, dependency, lockfile, route inventory,
OpenAPI, protobuf, profile, scenario, fixture, comparator, normalizer,
persistence schema, or legacy-wrapper restoration occurred.

## Implemented boundary

- `NETBOX_CORS_ALLOWED_ORIGINS` is the sole CORS configuration input.
- Unset or empty input produces an empty allowlist.
- Exact canonical HTTP(S) origins are constructor-injected through startup,
  service composition, the HTTP server, and the canonical router.
- Wildcards, regexes, duplicates, unsafe/malformed hosts and ports, paths,
  credentials, control bytes, non-ASCII hosts, and multiple-origin forms fail
  closed without disclosing input material.
- Trusted actual and preflight responses use a fixed credentialed wire
  contract; untrusted/malformed/missing origins receive no CORS permission and
  preserve the existing downstream result.
- Existing session, CSRF, token, route, cookie, and authentication semantics
  remain unchanged.
- Root Compose passes the setting through with an empty safe default.
- The reviewed L7 addendum removes the contradictory Docker-context exclusion
  so the existing offline smoke Dockerfile can package only its run-selected
  host artifacts. Git and source-v2 continue to exclude those transient bytes.

## Focused committed matrix

Fail-fast focused run: `2026-08-17T04:40:31Z` through
`2026-08-17T04:40:33Z`, exit 0.

- Parser accepted cases: 10/10.
- Parser top-level rejection cases: 58/58.
- Explicit rejected C0 control-byte cases: 31/31.
- Environment runtime cases: 5/5.
- Configuration defensive-copy proof: passed.
- Startup-before-database proof: passed.
- HTTP server option defensive-copy proof: passed.
- Actual-request cases: 21/21.
- Preflight cases: 11/11.
- `Vary` merge cases: 2/2.
- Runtime CSRF/token cases: 6/6.
- Route/inventory containment proof: passed.
- Root Compose non-secret pass-through assertion: passed.

Commands:

```bash
env GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/platform/config \
  -run '^Test(CORSAllowedOrigins|CORSAllowedOriginsDefensiveCopy|LoadHTTPRuntimeFromEnvironment)$' \
  -count=1

env GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
  go test ./cmd/netbox_go/initial \
  -run '^TestCORSConfigurationLoadsBeforeDatabaseInitialization$' -count=1

env GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/server \
  -run '^TestCORSHTTPOptionDefensiveCopy$' -count=1

env GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
  go test ./internal/adapters/rest/netbox/router \
  -run '^Test(CORSActualRequests|CORSPreflightRequests|CORSVaryHeader|RuntimeCORSCSRFAndToken|RuntimeCORSRouteInventory)$' \
  -count=1
```

The root Compose assertion rendered JSON into a checking process and did not
print or retain the complete effective configuration. The digest remained
unchanged and the worktree remained clean.

One earlier composite diagnostic ran the focused tests successfully but then
invoked root `make source-digest` from `netbox-backend`, where that target does
not exist. Its trailing shell command masked that bookkeeping exit. This
artifact does not rely on that run: the fail-fast command above reran the
matrix and correct root-level digest/cleanliness checks successfully.

## Full repository L3

Exact committed local run: `2026-08-17T04:25:04Z` through
`2026-08-17T04:26:40Z`, exit 0.

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

The before/after digest was identical. The independent GitHub run also passed
for the exact commit. Backend coverage was 12.4001% (36,013/290,425) across 54
measured packages with one reviewed exclusion; frontend tests were 125/125.

## Deployment L7

Exact committed run: `2026-08-17T04:37:19Z` through
`2026-08-17T04:37:41Z`, exit 0.

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make deployment-smoke
```

The disposable Compose project verified a named PostgreSQL volume, empty
database bootstrap, application health, content-type seed presence, and
restart idempotence. The unique project, volume, image, and run directory were
cleaned. The source digest remained identical before and after.

## Skipped gates

- L4 complete real-PostgreSQL identity/security suite: belongs to the broader
  V1/V3 evidence bundle.
- L5 differential REST oracle: required for T2, not this extension increment.
- L6 promotion-grade gRPC report: requires corresponding T2 first.
- L8 real-browser workflow: belongs to V5/T4.

None of these skipped gates is implied by L3 or L7.

## Residual risks

- `SameSite=Lax` remains unchanged; this increment does not enable arbitrary
  third-party-site cookies.
- Full V1 credential/RBAC/admin/secret evidence remains open.
- A host crash or `SIGKILL` can bypass the smoke cleanup trap and leave ignored
  artifacts visible to a later Docker build context until removed. Normal
  completion cleans the unique directory, and production images do not copy
  that path.
- `npm ci` reports three high-severity advisories; dependency remediation is a
  separate reviewed increment.

## Human review

Reviewer: project owner.

Decision: accepted. On 2026-08-17, the project owner explicitly approved
retention of the exact-SHA V0 and CORS evidence and authorized the stated
next-step execution. Recorded at `2026-08-17T05:23:24Z`. `CW1-V1-01` is done
for the exact revision and digest above. V1 as a whole and every compatibility
tier remain unchanged.

## Historical claim-only reconciliation — 2026-08-23

The accepted result was retained by the separate claim-only revision
`5d6f9cf9f5a5dc5e5e09de58f4f0e76e0f7a6381`, whose parent is the tested
revision `07aaca9264e5af9830e39fc95dd3289c0b5c9476`. That claim changed exactly
the two new evidence artifacts and the evidence ledger, all beneath the
source-v2-excluded `docs/evidence/` boundary. It changed no owned source entry.

Reconstruction from both reachable Git trees on 2026-08-23 produced the same
2,993-entry source manifest, SHA-256
`8f82bbc5b6ba5fe0c6acefe17cd552fe5bd9523414a5f224fec0b7040e45ea49`,
and the same source digest recorded above. The claim revision's repository gate
also passed at exact SHA: workflow run
[31997963053](https://github.com/jihadismail8/netbox-go/actions/runs/31997963053),
job
[95292950293](https://github.com/jihadismail8/netbox-go/actions/runs/31997963053/job/95292950293),
`2026-08-17T05:27:31Z`–`2026-08-17T05:39:47Z`, with every step successful.

The evidence ledger retains the exact three-path mode/size/content-hash mapping
and its canonical SHA-256. Once this reconciliation revision passes its own
exact-SHA repository gate, it closes the historical claim-manifest
documentation gap without changing CORS behavior, V1, the Capability Profile,
or any compatibility tier.
