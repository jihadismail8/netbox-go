# Core Workflow source-v2 V0 — 2026-08-17

## Claim boundary

- Goal: `CW1-G00`.
- State: **retained**.
- Result: the deterministic repository-quality gate is retained for the exact
  committed source below.
- Compatibility Profile: `core-workflow-v1`, still T1 and pre-publication.

This artifact proves repository quality only. It does not prove V1
identity/security completion, V2 business behavior, V3 PostgreSQL/deployment
completion, T2 REST compatibility, T3 gRPC parity, T4 browser Workflow Parity,
V6 sign-off, module completion, complete replacement, or production readiness.
It intentionally carries no `netbox-go-evidence-v2` marker because V0 is a
repository/release gate, not a retained matrix proof dimension.

## Tested source

- Branch: `main`.
- Tested revision:
  `07aaca9264e5af9830e39fc95dd3289c0b5c9476`.
- Tested and current source digest:
  `source-v2:sha256:19957bc8191e6bc8b52d662f98e87019bf3bf00651ddb762f398f6b569cd5b82`.
- Owned entries: 2,993.
- Commit state: pushed to `origin/main` with a clean local worktree.
- Pinned upstream oracle revision:
  `fbb948d30e79ce657fac62994a22aca72c1770a9`.

The complete mode-aware manifest was reconstructed directly from the reachable
Git commit with replacement objects and grafts disabled. It matched the clean
working-tree manifest exactly, including the `100755` modes on both new oracle
materialization scripts.

## Toolchain

- Local Go: `go1.26.0 linux/amd64`.
- Local Node.js: `v24.18.0`.
- Local npm: `11.16.0`.
- CI Go/Node inputs: `.go-version` `1.26.0` and `.node-version` `24.18.0`.
- CI npm pin: `11.16.0`.
- CI protobuf compiler: `libprotoc 3.21.12`.
- CI pinned tools: golangci-lint `v2.12.2`, Buf `v1.71.0`,
  protoc-gen-go `v1.36.10`, and protoc-gen-go-grpc `v1.5.1`.

## Local exact-commit run

- Start: `2026-08-17T04:25:04Z`.
- End: `2026-08-17T04:26:40Z`.
- Exit: 0.

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

The source digest and file count were computed immediately before and after
the command and remained identical. The worktree remained clean.

## Independent GitHub run

- Workflow: `repository-gate`.
- Event: push to `main`.
- Run: [31994516467](https://github.com/jihadismail8/netbox-go/actions/runs/31994516467).
- Job: [95283696536](https://github.com/jihadismail8/netbox-go/actions/runs/31994516467/job/95283696536).
- Head SHA:
  `07aaca9264e5af9830e39fc95dd3289c0b5c9476`.
- Job start: `2026-08-17T04:27:07Z`.
- Job completion: `2026-08-17T04:38:56Z`.
- Conclusion: **success**.

GitHub independently checked out the source with submodules, installed the
pinned Go, Node, npm, protobuf, lint, Buf, and generator toolchain, and
completed the repository gate. All setup, gate, post, and completion steps
reported success.

## Result summary

- Backend formatting, vet, pinned lint, race tests, builds, architecture
  constraints, generated/frozen compile guards, protobuf/Buf checks, and the
  non-mutating coverage policy passed.
- Backend coverage was 12.4001% (36,013/290,425), exactly meeting the retained
  floor across 54 measured packages with one reviewed exclusion.
- Frontend install, pinned toolchain check, formatting, lint, application and
  test typechecks, coverage tests, and production build passed.
- Frontend tests: 28 files and 125 tests passed; aggregate statement coverage
  was 61.72%.
- Inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries.
- Oracle-source materialization self-test passed.
- Traceability validator regressions: 96/96 passed.
- Active Capability Profile: 13 resources, 3 interfaces, 17 scenarios, and 293
  traceability rows.
- OpenAPI remained valid with 33 paths and 86 operations.
- Local links were valid across 148 Markdown files.

## External gates not claimed

The V0 workflow intentionally does not run the DSN-enabled complete
PostgreSQL suite, differential REST oracle, promotion-grade gRPC report, or
real-browser workflow suite. The standalone deployment smoke passed separately
for the same committed revision and is recorded in the CORS candidate artifact;
that single result does not close V3.

## Residual risks and review

- `npm ci` reports three high-severity dependency advisories. No dependency or
  lockfile change was authorized by this increment; remediation needs a
  separately reviewed dependency update.
- V1-V6 and all T2/T3/T4 claims remain pending.
- The project owner reviewed the revision/digest, local and GitHub receipts,
  claim boundary, absence-of-secrets boundary, and conservative status wording.

Human reviewer: project owner.

Human decision: accepted. On 2026-08-17, the project owner explicitly approved
retention of the exact-SHA V0 and CORS evidence and authorized the stated
next-step execution. Recorded at `2026-08-17T05:23:24Z`. This decision makes
`CW1-G00` current for the exact revision and digest above; it does not retain or
promote any other boundary named in this artifact.
