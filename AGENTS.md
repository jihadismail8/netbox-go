# Implementation-agent instructions

These instructions apply to the entire repository. They are a compact entry
point, not a substitute for the canonical documents.

## Mandatory reading

Before changing code, read:

1. [Project language](CONTEXT.md)
2. [Accepted ADRs](docs/adr/README.md)
3. [Compatibility contract](docs/COMPATIBILITY.md)
4. [Coding standards](docs/CODING_STANDARDS.md)
5. [Whole-project execution playbook](docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md)
6. [Goal and dependency roadmap](docs/ROADMAP.md)
7. [Current status](docs/STATUS.md)
8. The active machine-readable Capability Profile and relevant resource/scenario
   metadata

When sources disagree, follow the authority order in section 1 of the execution
playbook. Generated code and prior agent summaries are never architectural or
behavioral precedent.

## Current operating mode

At the 2026-08-17 audit, `core-workflow-v1` is T1 and pre-publication. The
interrupted typed-boundary recovery is structurally closed, ADR 0005's
dormant-wrapper cleanup is complete, and the new mode-aware `source-v2` digest
has made the older V0 artifacts historical. The exact source-v2 entry revision
has a retained `CW1-G00` result, `CW1-V1-01` is human-reviewed and done, and
`CW1-V2-01` is structurally accepted. `CW1-V1-02-I1` is accepted and done for
token lookup classification, strict touch ordering, and durable PostgreSQL
touch semantics. `CW1-V1-02-I2` is accepted and done only for baseline REST
token grammar/outcomes and unary gRPC bearer/method safety.
`CW1-V1-02-I3` has bounded evidence awaiting project-owner review for
browser-session classification, valid-session-first REST arbitration, exact
CSRF pairing, transactional login/logout, active-session CSRF recovery, and
session-cookie shape. The `CW1-V1-02` parent and its password, throttle,
trusted-proxy, gRPC streaming, Django Origin/Referer/masking, and aggregate
rows remain open. Do not begin or merge another
Capability Profile while V0 is red or before `CW1-V6-03`. Once those gates are
durably closed, follow the next accepted profile in the roadmap and playbook
rather than this dated snapshot.

## Non-negotiable rules

- Scope every change to a reviewed Capability Profile.
- Replace observable behavior, not Python files or database tables one for one.
- REST is exact for declared baseline behavior; gRPC preserves the same
  semantics through the same typed application use cases; Vue uses REST.
- Dependencies point adapter → application → domain.
- Domain/application contracts contain no Gin, protobuf, GORM row, SQL driver,
  raw Managed Object map, or global database/config dependency.
- One public mutation uses one application-owned PostgreSQL transaction,
  including authorization, validation, persistence, derived state, object
  change, and required durable event intent.
- Preserve absent, explicit null, zero, empty, and concrete values when the
  contract distinguishes them.
- Fail closed for authentication, authorization, visibility, fields, filters,
  identifiers, enum values, and undeclared routes.
- Never hand-edit generated/frozen output or add business behavior to legacy
  scaffolding.
- Never weaken tests, comparators, normalizers, security controls, lint,
  coverage, generated checks, or evidence requirements to make a command pass.
- Never log or retain secrets in evidence. Return only the documented hardened
  session cookie, CSRF bootstrap cookie/body value, and one-time token-creation
  secret; never return reusable secret material later.
- Call AutoMigrate only after confirming a table is absent. Never use it to
  inspect, alter, repair, backfill, or drift-correct an existing table.
- Do not change pinned toolchains/dependencies merely to match the host.
- Do not promote T2/T3/T4 or “complete” without current retained evidence at
  the required boundary.
- Do not delete another displaced legacy stack until ADR 0004's capability
  completion condition is satisfied. ADR 0005 is the sole narrow exception:
  it authorizes only the recorded 118 dormant handler wrappers and their 118
  matching Sponge router wrappers, with all retained legacy layers still
  frozen.
- Never restore or regenerate the HTTP wrapper pairs retired by ADR 0005.
  Preserve `internal/handler/auth.go`, canonical adapter handlers, hand-owned
  router code, and the separately retained DRF registry.
- Preserve unrelated user changes and avoid destructive Git/filesystem actions.

## Increment protocol

For each increment:

1. Claim one roadmap goal ID and state the capability, outcome, entry
   condition, permitted files, forbidden scope, tests, and exit condition.
2. Read complete target files and the nearest tests before editing.
3. Add the lowest-layer regression before or with behavior.
4. Make one coherent vertical change with no unrelated cleanup.
5. Run focused tests, then climb the command ladder in section 15 of the
   execution playbook.
6. Update contracts/docs/status only to the boundary actually proved.
7. Report changed files, exact commands/results, source digest, skipped
   external gates, residual risk, and next increment.

Stop and ask only for a genuinely absent hard-to-reverse product/security/data
decision, unavailable authority/credential, destructive action, irreconcilable
overlap with user changes, required profile expansion, or a new exception.
Do not ask questions answered by the pinned source, profile, standards, tests,
or ADRs.
