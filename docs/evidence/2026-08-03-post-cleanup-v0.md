# Post-cleanup repository V0 — 2026-08-03

## Scope

This artifact retains the deterministic V0 repository gate after ADR 0005's
dormant Sponge handler/router cleanup, its regression tests, the rewrite-status
audit, and the identity-scenario wording correction.

It proves repository quality only. It does not prove V1 identity/security, V2
business-rule completeness, V3 real-PostgreSQL/deployment behavior, T2 REST
compatibility, T3 gRPC parity, T4 browser Workflow Parity, V6 sign-off, module
completion, or production readiness.

## Tested source and toolchain

- Base Git revision: `05dcc21` on `main`, with the audited working source
  identified by the digest below.
- Tested source digest:
  `sha256:b3ab64c99a2698ac4325bce835d4d1bd51bc6fe6b48be0442cffa73b510344aa`
- Owned files in digest: 2,978.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.
- Start: `2026-08-03T04:28:18Z`.
- End: `2026-08-03T04:29:49Z`.

Command:

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

Result: **passed**, exit code 0. The digest was recomputed immediately after
the run and remained identical.

## Result summary

- Backend formatting, vet, lint, race tests, build, frozen-client/protobuf
  compile checks, canonical Buf checks, and architecture assertions passed.
- Backend coverage was 12.3261% (35,767/290,174), above the retained 12.3108%
  floor (35,987/292,320), with 53 measured packages and 1 reviewed exclusion.
- Frontend pinned install, toolchain, formatting, lint, application/test
  typechecks, coverage tests, and production build passed.
- Frontend tests: 28 files and 125 tests passed; aggregate statement coverage
  was 61.72%.
- Generated inventory checks passed: 155 baseline REST, 123 current REST, 179
  current gRPC, and 13 current Vue entries.
- Capability Profile validation passed: 13 resources, 3 interfaces, and 17
  scenarios.
- Generated contract documentation matched; OpenAPI validation passed with 33
  paths and 86 operations.
- Markdown local-link validation passed across 144 files.

## Final-digest verification and claim audit

After the tested run, only documentation claims, evidence links, audit wording,
and V0 status were updated. No Go, Vue, schema, profile/scenario, fixture,
comparator, normalizer, security behavior, dependency, generator, or toolchain
content changed.

- Final source digest:
  `sha256:52898fa3caf739765eb8b95cddd3f832c875933ae057fa29317292aae341cec3`
- Owned files in digest: 2,978.
- Direct final-digest gate: `make check`, passed with exit code 0.
- Final-digest start: `2026-08-03T04:34:59Z`.
- Final-digest end: `2026-08-03T04:36:30Z`.

The complete gate was rerun directly on the attestation digest, so the current
V0 claim does not depend on treating the broader documentation audit as a
claim-only exception. The hash mapping below remains as an audit trail between
the first passing run and the final self-attesting run.

Audited claim-file hashes (`tested` → `attested`):

| File                                        | Tested SHA-256                                                     | Attested SHA-256                                                   |
| ------------------------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `README.md`                                 | `1b3280cdf1d939986a1654029903f00c61be7b89d86d6ec6530b1f93d538bf5e` | `58e58fb95b25ad1eba348dbd61b89df03ef38d256fba1ee38ac4ab3c3842c045` |
| `CONTEXT.md`                                | `6cb872ab73ec0a72fa3a85ade40d65a50741732b96560af0c4dde95f78422ec3` | `69eeb08551d321e70fdef348fdde875b3e9f254db13a46e08ac78931725352b9` |
| `AGENTS.md`                                 | `282f76ac81a1ab30b2636b8485e3fb5134dd6c637da4558d910692d5ed1babf0` | `c0b547c471534fa2235113dd0ac2e115fac788620f7b1b535f039001ed0e1b83` |
| `docs/STATUS.md`                            | `4eb8dd69b2ce41bc88775a93bfc54976d63fe27981c3eef02931435faff0ef31` | `7347220ba848e3ab4e0b53e34b692e7d4a34ca72e52c0d528a3ffc13a4307e3f` |
| `docs/ROADMAP.md`                           | `205364b99f55d93931b69ffe732d2083a83a02b50126e23128a0536ada4d661d` | `eab21120a4c39e576c7137ceb4fb527d437331f446c7746cbf89524b2bf6f657` |
| `docs/ARCHITECTURE.md`                      | `4b0377407e6c0b88f2d95ae22ec8c6ed9fe1b7b2a0f9a4311d2da132e463870b` | `2aeac6666261c7fd9b75c68ac3b03f9e7ad3ee34c561dada9e5585f86d6ca06e` |
| `docs/COMPATIBILITY.md`                     | `b97169c43cb91b4a071e7db77586ddabdfc6280cfe3e421ddfca7c809f315f63` | `6b730b83503bcd8bb848ec3d4b7b4c60de1e9a79827e9cc340717c4ae1dacc09` |
| `docs/TESTING.md`                           | `3b7f5e1cb8905587bec5f33690ff2ed6e2fed557827368096166dbefc70ae424` | `33170f89b9566bd3a54167e1d298fc9ab0c401101a956f942dc0277225a8220b` |
| `docs/IMPLEMENTATION_PLAN.md`               | `29dcc3fdaa430d0e007fd0cf0107803cf4712ec7062fb028b133f60b55c0d537` | `b8a86bcb0945f53d601fedb814d17d63956aaa58da6289f55cda6af9f6606a51` |
| `docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md` | `314c4c5e08febbd898c3c910e7e9c4b34dd55eff4fe30b246e40fed579dd5db8` | `a7508ef7ac800312bf1b317b0ba687d3fa762dee659edab66fd37bae8a0deaa7` |
| `docs/evidence/README.md`                   | `17b0cb126e7fbf958fd1ae296a1e41f9eabc2af51b7a5984552f26e0dffd83fc` | `14dda06adbaf19fd00cb249a14080485ea6faaa604350b6e304be8e698b0750b` |

The evidence directory is excluded from `make source-digest`; its ledger and
this summary do not alter either owned-source digest.

## Roadmap goal-ledger verification

The canonical roadmap was subsequently expanded into the stable goal,
dependency, checklist, and implementation-handoff ledger, with corresponding
navigation/status wording and formatting corrections to already-modified
reference documentation. This section is the newest V0 result and supersedes
the earlier final digest above for the current-source claim; the earlier runs
remain part of the audit trail.

- Tested source digest:
  `sha256:957e4cd966bc7052f47378d187708d66571f0ab620226917616a7adb0a8fcf38`
- Owned files in digest: 2,978.
- Start: `2026-08-03T04:57:42Z`.
- End: `2026-08-03T04:59:18Z`.

Command:

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

Result: **passed**, exit code 0. The digest was recomputed immediately after
the run and remained identical.

- Backend formatting, vet, lint, race tests, build, architecture, frozen
  compile, and coverage checks passed.
- Backend coverage remained 12.3261% (35,767/290,174), above the retained
  12.3108% floor, with 53 measured packages and 1 reviewed exclusion.
- Frontend toolchain, formatting, lint, application/test typechecks, coverage,
  and production build passed; 28 files and 125 tests passed with 61.72%
  aggregate statement coverage.
- Generated inventories remained 155 baseline REST, 123 current REST, 179
  current gRPC, and 13 current Vue entries.
- Capability Profile validation remained 13 resources, 3 interfaces, and 17
  scenarios; generated contracts matched and OpenAPI remained 33 paths/86
  operations.
- Markdown local-link validation passed across 145 files.
- A separate explicit Prettier check covered every modified Markdown file; the
  final evidence-only formatting does not change the owned-source digest.

## `CW1-V1-01` implementation-handoff verification

The next-step planning pass added the bounded GPT-5.6 Luna execution packet for
trusted-origin CORS, its handoff index, and navigation from the canonical
roadmap/documentation index. It made no Go, Vue, deployment, profile, scenario,
fixture, comparator, normalizer, schema, dependency, route, security-runtime,
or application-behavior change.

The packet was checked against the active Capability Profile, current runtime
composition, SEC-009, CSRF behavior, generated-code rules, coverage-package
policy, the pinned NetBox source, and `django-cors-headers` 4.9.0. Independent
read-only code, contract/security, command, and handoff reviews found no
remaining blocking contradiction after corrections. In particular, the final
packet requires immutable constructor injection rather than mutable global
configuration, exact origin parsing and pinned OPTIONS behavior, enforced
focused-test discovery, a non-regressing 54-package coverage refresh when the
new platform package is implemented, and separate evidence/claim closeout.

- Tested source digest:
  `sha256:c5c9783d20f003482b6d9912ee84dcd4ce5c9c4675376cea46956200722f85d0`
- Owned files in digest: 2,980.
- Go: `go1.26.0 linux/amd64`.
- Node.js: `v24.18.0`.
- npm: `11.16.0`.
- Start: `2026-08-03T05:57:16Z`.
- End: `2026-08-03T05:58:36Z`.

Command:

```bash
env GOCACHE=/tmp/netbox-go-cache \
  GOFLAGS=-buildvcs=false \
  PATH=/home/jihad/.nvm/versions/node/v24.18.0/bin:/home/jihad/.local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  make check
```

Result: **passed**, exit code 0. The source digest was computed immediately
before and after the timed run and remained identical.

- Backend formatting, vet, lint, race tests, build, architecture, frozen
  compile, generated-contract, protobuf, and coverage checks passed.
- Backend coverage remained 12.3261% (35,767/290,174), above the retained
  12.3108% floor, with 53 currently measured packages and 1 reviewed exclusion.
  The 54-package refresh is an implementation-time requirement, not a claim
  that the new package already exists.
- Frontend toolchain, install, formatting, lint, application/test typechecks,
  coverage, and production build passed; 28 files and 125 tests passed with
  61.72% aggregate statement coverage.
- Inventories remained 155 baseline REST, 123 current REST, 179 current gRPC,
  and 13 current Vue entries.
- The Capability Profile remained 13 resources, 3 interfaces, and 17 scenarios;
  generated contracts matched and OpenAPI remained 33 paths/86 operations.
- Pinned Prettier checks passed for the new/edited handoff documents,
  `git diff --check` passed, and local-link validation passed across 147
  Markdown files.

`CW1-V1-01` remains `ready`; CORS is not implemented by this artifact. V1,
T2, T3, T4, profile completion, replacement completion, and production
readiness all remain pending.
