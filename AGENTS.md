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

At the 2026-08-26 audit, `core-workflow-v1` is T1 and pre-publication. The
interrupted typed-boundary recovery is structurally closed, ADR 0005's
dormant-wrapper cleanup is complete, and the new mode-aware `source-v2` digest
has made the older V0 artifacts historical. The exact source-v2 entry revision
has a retained `CW1-G00` result, `CW1-V1-01` is human-reviewed and done, and
`CW1-V2-01` is structurally accepted. `CW1-V1-02-I1` is accepted and done for
token lookup classification, strict touch ordering, and durable PostgreSQL
touch semantics. `CW1-V1-02-I2` is accepted and done only for baseline REST
token grammar/outcomes and unary gRPC bearer/method safety.
`CW1-V1-02-I3` is accepted and done only for browser-session classification,
valid-session-first REST arbitration, exact CSRF pairing, transactional
login/logout, active-session CSRF recovery, and session-cookie shape.
`CW1-V1-02-I4` has retained evidence. Its tested candidate at
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f` passed the accepted-I3-based
red-first proof, focused and race suites, real-PostgreSQL matrices and complete
L4, the pinned repository gate, feature-candidate CI, and main exact-SHA CI on
unchanged source digest
`source-v2:sha256:3f37417bb791ac6bc97ac4a0d23c5f928062feecf81a0f8a4fb9e57445d53670`.
The nine-path claim-only revision and digest-excluded receipt both passed
their exact-SHA repository CI boundaries. Project-owner review remains before
the separate `evidence` to `done` transition. The
`CW1-V1-02` parent and its password-policy, throttle,
trusted-proxy, gRPC streaming, Django Origin/Referer/masking, and aggregate
rows remain open. `CW1-V2-02-I1` has an exact tested candidate at
`7acba402f0de2bd59e5b342a6f05df268bc9120b`, source digest
`source-v2:sha256:ed330b0a5bbeafd70b7b16a4ce4d1052fa9a385313a3b8827b554983571c1b43`
with 3,018 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, candidate, evidence-claim, and
pre-acceptance receipt exact-SHA CI boundaries are green. The project owner
accepted this bounded result at `2026-08-23T17:25:54Z`. The owner-accepted
closeout claim and excluded closeout receipt also passed exact-SHA CI, so I1
is effectively done only for IPAddress scalar create/PUT/PATCH presence,
operation-specific generated API contracts, matching REST/gRPC semantics, Vue
form serialization/validation, and their focused tests. The pinned external
differential remained unavailable, so REST T2 and corresponding gRPC T3 are
unearned. I1 closes no 13-resource parent, compatibility tier, profile state,
or traceability consumer. `CW1-V2-02-I2` has an exact tested candidate at
`87863efd38fe71dfa05c818b860b37b7e94d67b4`, source digest
`source-v2:sha256:c7c1b86c2bcd768bb719149a54dddbceccf2b5b2e4087dd4b79eec20bef5a37c`
with 3,022 entries. Its exact-name, affected/race, real-PostgreSQL, complete
L4, generated-contract, Vue, pinned repository, independent candidate,
evidence-claim, and pre-acceptance receipt exact-SHA CI boundaries are green.
The project owner accepted this bounded result at `2026-08-24T04:46:51Z`.
The owner-accepted closeout claim and excluded closeout receipt both passed
their exact-SHA repository CI boundaries, so I2 is effectively done only for
Site `name`, `slug`, `status`, `facility`, `description`, and `comments`
create/PUT/PATCH presence, generated request/response contracts, matching
REST/gRPC semantics, PostgreSQL durability, Vue dirty-field
serialization/validation, and the eight named focused tests. The external
differential was unavailable before execution because Docker rejected its
temporary source bind, so REST T2 and corresponding gRPC T3 remain unearned.
I2 closes no Site uniqueness, deletion, list/query, full CRUD,
compatibility-tier, parent, profile, or traceability consumer boundary.
`CW1-V2-02-I3` has an exact tested candidate at
`651d33bc3fb2c8e663b6b14320af405b8501471f`, source digest
`source-v2:sha256:09499a6618569d2dae224edfb339ac82585bf0248d20e5a4d5ff23d19221fe6f`
with 3,029 entries. Its exact-name, affected/race, real-PostgreSQL, complete
L4, generated-contract, Vue, pinned repository, independent exact-candidate,
evidence-claim, and pre-acceptance receipt exact-SHA CI boundaries are green.
The project owner's acceptance of this bounded result was recorded at
`2026-08-24T10:28:34Z`. Its owner-accepted closeout claim and excluded
closeout receipt both passed exact-SHA repository CI, so I3 is effectively
done only for Manufacturer `name`, `slug`, and
`description` create/PUT/PATCH presence, operation-specific generated API
contracts, matching REST/gRPC semantics, PostgreSQL durability, Vue
dirty-field serialization/validation, and the eight named focused tests. No
retained pinned differential accompanies this result, so REST T2 and
corresponding gRPC T3 remain unearned. I3 closes no Manufacturer uniqueness,
deletion, list/query, full CRUD, compatibility-tier, parent, profile, or
traceability-consumer boundary. `CW1-V2-02-I4` has exact tested candidate
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe` at source digest
`source-v2:sha256:68db7a9835545d66ef9a651b9c4a000c91f3d834b5503a1b89e4a122275c3bc9`
with 3,036 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-24T18:51:56Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I4 is effectively `done`. I4 owns only
RackRole `name`, `slug`, `color`, and `description` create/PUT/PATCH presence,
defaults, validation, generated API contracts, matching REST/gRPC semantics,
PostgreSQL durability, Vue serialization/validation, and eight fixed tests.
No retained differential accompanies it, so T2/T3 remain unearned. I4 closes
no RackRole uniqueness, deletion, list/query behavior, full CRUD, another
resource, tier, parent, profile, or traceability-consumer boundary.
`CW1-V2-02-I5` has exact tested candidate
`89507d95d2743de7f97d64ca14cc43f6b834770b` at source digest
`source-v2:sha256:a8325eaae703aa801ed587deefae7e8d08d9c9e0189c80ff7569da95c36d6f90`
with 3,043 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-25T03:56:48Z`. Its
owner-accepted closeout claim and excluded receipt both passed exact-SHA CI,
so I5 is effectively `done`. I5 owns only RackType common writable-field
create/PUT/PATCH presence, including the numeric-ID Manufacturer envelope,
operation-specific generated contracts, matching REST/gRPC semantics,
PostgreSQL durability, Vue dirty-field handling, and its focused tests. It
excludes RackType uniqueness, Rack propagation semantics, deletion,
list/query behavior, alternate nested Manufacturer inputs, every
tier/consumer, the parent, and any later child. `CW1-V2-02-I6` has exact tested
candidate `dddd7adbda72f5dd760202c4862ce23b17cdf180` at source digest
`source-v2:sha256:be2168180bfbbb10406772d57e75210d32e5084051c4b765d4cf2954d958e621`
with 3,050 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-25T08:13:40Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I6 is effectively `done`. I6 owns only DeviceRole `parent`, `name`,
`slug`,
`color`, `vm_role`, `description`, and `comments` create/PUT/PATCH presence,
operation-specific generated contracts, matching REST/gRPC semantics,
PostgreSQL durability, Vue dirty-field handling, and eight fixed tests. It
does not close DeviceRole hierarchy, uniqueness, deletion, list/query, full
CRUD, another resource, a tier/consumer, the parent, the profile, or any later
child. `CW1-V2-02-I7` has exact tested candidate
`e2ad1acc33b84f20f24418d89b3b881b897b7ed3` at source digest
`source-v2:sha256:8930fa5fbe487df4c225a301c674db096664cccaf3f0c53df93c51e59edaeba7`
with 3,057 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-25T14:12:13Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I7 is effectively `done`. I7 owns only DeviceType `manufacturer`, `model`,
`slug`, `part_number`, `u_height`, `exclude_from_utilization`,
`is_full_depth`, `airflow`, `description`, and `comments` common
create/PUT/PATCH presence, operation-specific generated API contracts,
matching REST/gRPC semantics, PostgreSQL durability, Vue form
serialization/validation, and eight fixed tests. It does not close uniqueness,
height transitions, deletion/cascades, list/query behavior, full CRUD, another
resource, a tier/consumer, the parent, the profile, or any later child.
`CW1-V2-02-I8` has exact tested candidate
`b216d4c217cf863a8760494fd6499e54899ef368` at source digest
`source-v2:sha256:da2b93d51dbdc2dfbcb6e348e2f5b23f42439b5973169f683ddc39de285c5048`
with 3,064 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-26T02:27:06Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I8 is effectively `done`. I8 owns only InterfaceTemplate `device_type`,
`name`, `label`, `type`, `enabled`, `mgmt_only`, and `description` common
create/PUT/PATCH presence, operation-specific generated API contracts,
matching REST/gRPC semantics, PostgreSQL durability, Vue form
serialization/validation, and eight fixed tests. It exercises existing
DeviceType-owner containment but does not close owner immutability,
uniqueness, Device instantiation/snapshot/rollback, non-retroactivity, bridge
behavior, deletion, list/query behavior, full CRUD, another resource, a
tier/consumer, the parent, the profile, or any later child.
`CW1-V2-02-I9` has exact tested candidate
`9c257b04b7cf798199c5aa4b7ae076cebbbbdff1` at source digest
`source-v2:sha256:343a1767534de69acad81e7fdfbf8bd23cf1e7a22450c00df70b038c4c54c152`
with 3,071 entries. Its focused, race, real-PostgreSQL, complete L4,
generated-contract, Vue, repository, candidate-CI, evidence-claim, and
pre-acceptance receipt boundaries are green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-26T11:01:23Z`. I9 is
conditional `done`; that state becomes effective only after the exact
owner-accepted closeout claim passes repository CI, and its excluded receipt
must then pass exact-SHA CI. I9 owns only Rack `site`, `name`, `facility_id`,
`rack_type`, `status`, `role`, `serial`, `asset_tag`, `form_factor`, `width`,
`u_height`, `starting_unit`, `desc_units`, `airflow`, `description`, and
`comments` common create/PUT/PATCH presence, direct-save RackType copy
precedence, operation-specific generated contracts, matching REST/gRPC
semantics, PostgreSQL durability, Vue dirty-field handling, and eight fixed
tests. It excludes uniqueness, mounted-device/placement rules, RackType-update
propagation, Device site propagation, deletion, list/query, full CRUD, another
resource, every tier/consumer, the parent, the profile, and later children. Do
not begin or merge another Capability Profile while V0 is
red or before `CW1-V6-03`. Once those gates are
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
