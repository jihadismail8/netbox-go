# Core Workflow V0 recovery evidence — 2026-08-01

Status: **PASS**

Scope: recovery checkpoints R0-R8 and repository-quality checkpoint V0 for
`core-workflow-v1`. This artifact does not claim V1-V6, REST T2, gRPC T3, Vue
T4, profile publication, deployment readiness, or production readiness. The
profile remains T1 and pre-publication.

## Source and attestation

- Tested source digest:
  `sha256:a55fab792cea1100e5fd2cc641fad02345189dd27d0b28c3b7ed2b1e1dcc22e1`
  (`3210` owned files).
- Attestation digest:
  `sha256:d4231b3587b42d939a891030c6b7e45c8492e5063e8cc2bd31df1ccca3f9c339`
  (`3210` owned files).
- Git revision: unavailable because this workspace has no usable repository
  metadata; `make source-digest` is the immutable source identifier.
- Pinned NetBox oracle: `fbb948d30e79ce657fac62994a22aca72c1770a9`.
- The source digest was identical immediately before and after both the full V0
  gate and the timestamped PostgreSQL replay.

## Pinned toolchain

| Tool               | Version                |
| ------------------ | ---------------------- |
| Go                 | `go1.26.0 linux/amd64` |
| Node.js            | `24.18.0`              |
| npm                | `11.16.0`              |
| golangci-lint      | `2.12.2`               |
| protoc             | `3.21.12`              |
| Buf                | `1.71.0`               |
| protoc-gen-go      | `1.36.10`              |
| protoc-gen-go-grpc | `1.5.1`                |

## Repository V0 run

- Start: `2026-08-01T11:31:49Z`
- End: `2026-08-01T11:33:42Z`
- Command: pinned-toolchain `make check` from the repository root.
- Configuration: `GOCACHE=/tmp/netbox-go-v0-cache`,
  `GOFLAGS=-buildvcs=false`; no service credential, DSN, cookie, token, or
  secret was supplied.
- Result: pass, with the source digest unchanged.

Observed results:

- backend formatting, vet, pinned lint (`0 issues`), race-unit suites, build,
  frozen live-client compile, frozen protobuf compile, and canonical Buf
  generation checks passed;
- backend coverage policy self-tests passed; 53 packages measured and one
  generated live-client package remained in the reviewed exclusion manifest;
  exact atomic coverage was `35,987/292,320` statements (`12.3108%`), equal to
  the first trustworthy baseline;
- clean `npm ci` installed 407 packages, audited 408 packages, reported zero
  vulnerabilities, and emitted no deprecation warning;
- frontend Prettier, ESLint with zero warnings, application/test typechecks,
  all 28 test files and 125 tests, and the production build passed;
- frontend V8 coverage was `61.72%` statements/lines, `75.07%` branches, and
  `59.66%` functions;
- inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries;
- the Capability Profile remained 13 resources, 3 interfaces, and 17
  scenarios; OpenAPI remained 33 paths and 86 operations; and
- all 141 Markdown files passed local-link validation.

The dependency refresh preceding this retained run stayed within declared
ranges for patched PostCSS/brace-expansion releases and adds a reviewed
`glob@13.0.6` transitive override. The override supports the pinned Node 24
runtime; the two consumers use the retained `glob`/`globSync` API, and the full
frontend test/build gate passed afterward.

## Real PostgreSQL replay

- Start: `2026-08-01T11:50:18Z`
- End: `2026-08-01T11:50:26Z`
- Boundary: disposable `postgres:16-alpine` container at image digest
  `sha256:e013e867e712fec275706a6c51c966f0bb0c93cfa8f51000f85a15f9865a28cb`.
- Configuration class: loopback-only, trust-authenticated, uniquely named,
  without a persistent volume; the container was stopped and auto-removed
  after the run.
- Command: the canonical DSN-enabled package list in
  [Testing](../TESTING.md#real-postgresql), with the DSN represented here only
  as `<owned-disposable-local-postgresql>`.
- Result: all database, bootstrap, changelog, DCIM/DCIM-row, IPAM/IPAM-row,
  application DCIM/IPAM/identity, and administrator CLI packages passed.

This replay covers missing-table-only bootstrap, owned row shape/constraints,
concurrent Site and conditional IP uniqueness, serialized IP assignment audit
history, identity bootstrap, and administrative process behavior. It is
recovery evidence for the typed persistence cutover; it does not by itself
earn V3, which still requires the complete retained PostgreSQL/deployment
boundary.

## Recovery-boundary audit

- The generic domain, application, and PostgreSQL workflow package directories
  are absent. The only old import-path strings are the architecture test's
  permanent prohibition.
- No transitional REST/gRPC constructor or `composition.Core.Workflow` caller
  remains.
- Canonical route inventory contains health, readiness, identity, schema,
  profile, and Vue routes, with no dedicated `GET /ping`. Frozen
  runtime-disabled source still contains the historical route; SPA history
  fallback is not a diagnostic endpoint.
- Canonical gRPC registration contains Identity, DCIM, IPAM, health, and
  reflection only.
- REST contract, gRPC contract, typed parity, inventory, Capability Profile,
  protobuf, OpenAPI, and documentation checks remained green. No public field,
  protobuf service, inventory count, table name/shape, or Vue profile resource
  was intentionally expanded.

## Claim-only mapping

The conservative claim updates were followed by a second pinned-toolchain
`make check`:

- Start: `2026-08-01T11:52:14Z`
- End: `2026-08-01T11:53:50Z`
- Result: pass, with 142 Markdown files validated and the attestation source
  digest unchanged before and after the gate.

The attestation digest maps back to the tested digest through exactly these
reviewed claim-file changes:

| Claim file                | Tested-run SHA-256                                                 | Attestation SHA-256                                                |
| ------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `docs/STATUS.md`          | `8e1f9c6383f529e4a4db3fc7117d8fcd361020432426d4a4f4ef8dd10293229a` | `e8c808be540b68668b08843eb55c13bf6c1a445b27c54438e06f95ea3a25d0b8` |
| `docs/COMPATIBILITY.md`   | `81bd2e6c613e902891b0fbc1e8119161c294c5a920a6902963a48a4f4920355f` | `cb735501bc03c58d0f30959598680de35c868cbecb8b6f266ab61c4134360c24` |
| `docs/evidence/README.md` | `d89a49fb246dc78778b1e33db08f8b3a5a59fe487a6f533f454b6ef3558daa44` | `63bdacf95d6df23d7fedb5e9ae44533131b6c50919b8cfa85ac7fc051054c141` |

This evidence file was added after the tested run and is deliberately excluded
from `make source-digest`, as are all files under `docs/evidence/`. Review
confirmed that the mapped changes only state already-proven recovery and V0
claims: no code, behavior, public schema, scenario, fixture, comparator,
normalizer, security policy, dependency, or toolchain changed between the
tested and attestation digests.
