# Implementation handoffs

This directory contains bounded execution packets for implementation agents.
The canonical goal state and dependency graph remain in the
[roadmap](../ROADMAP.md); these packets refine one ready goal into exact code,
test, evidence, and review boundaries. They do not prove that the described
implementation exists.

## Rules

1. One packet owns one reviewable outcome and one roadmap goal or an explicitly
   named portion of it.
2. The executor records the starting source digest and existing dirty paths
   before editing. Existing user work is never reset, regenerated, or rewritten
   wholesale.
3. The packet's permitted-file list is closed. A needed file outside it is a
   stop condition until the packet is reviewed and amended.
4. The packet's forbidden scope, authoritative inputs, wire behavior, scenario
   matrix, command ladder, and exit evidence are requirements, not suggestions.
5. The executor fixes a failure at the lowest responsible layer. Tests,
   comparators, normalizers, security controls, coverage floors, inventories,
   and generation checks are never weakened to obtain a pass.
6. Implementation evidence and compatibility-tier promotion are separate.
   Passing a bounded increment closes only the named goal; it does not imply
   T2, T3, T4, profile completion, module completion, or production readiness.
7. A claim-changing source edit uses the reviewed two-digest attestation
   protocol in the
   [execution playbook](../IMPLEMENTATION_EXECUTION_PLAYBOOK.md#two-digest-claim-attestation).
8. Completion reports contain no password, password/credential hash, cookie,
   CSRF value, bearer/API token, DSN, raw credential-bearing header, or
   complete configuration. Source and content digests required for evidence
   remain valid.

## Packets

| Goal           | State         | Intended executor | Packet                                                  |
| -------------- | ------------- | ----------------- | ------------------------------------------------------- |
| `CW1-V1-01`    | `done`        | Codex GPT-5.6 Sol | [Trusted-origin CORS](CW1-V1-01.md)                     |
| `CW1-V1-02-I1` | `done`        | Codex GPT-5.6 Sol | [Token credential foundation](CW1-V1-02.md)             |
| `CW1-V1-02-I2` | `done`        | Codex GPT-5.6 Sol | [Token transport and unary RPC safety](CW1-V1-02-I2.md) |
| `CW1-V1-02-I3` | `done`        | Codex GPT-5.6 Sol | [Browser session and CSRF lifecycle](CW1-V1-02-I3.md)   |
| `CW1-V1-02-I4` | `evidence`    | Codex GPT-5.6 Sol | [Password change and session rotation](CW1-V1-02-I4.md) |
| `CW1-V2-01`    | `done`        | Codex GPT-5.6 Sol | [Machine-readable traceability](CW1-V2-01.md)           |
| `CW1-V2-02-I1` | `done`        | Codex GPT-5.6 Sol | [IPAddress scalar write presence](CW1-V2-02-I1.md)      |
| `CW1-V2-02-I2` | `done`        | Codex GPT-5.6 Sol | [Site scalar write presence](CW1-V2-02-I2.md)           |
| `CW1-V2-02-I3` | `done`        | Codex GPT-5.6 Sol | [Manufacturer scalar write presence](CW1-V2-02-I3.md)   |
| `CW1-V2-02-I4` | `done`        | Codex GPT-5.6 Sol | [RackRole scalar write presence](CW1-V2-02-I4.md)       |
| `CW1-V2-02-I5` | `done`        | Codex GPT-5.6 Sol | [RackType scalar write presence](CW1-V2-02-I5.md)       |
| `CW1-V2-02-I6` | `done`        | Codex GPT-5.6 Sol | [DeviceRole scalar write presence](CW1-V2-02-I6.md)     |
| `CW1-V2-02-I7` | `done`        | Codex GPT-5.6 Sol | [DeviceType scalar write presence](CW1-V2-02-I7.md)     |
| `CW1-V2-02-I8` | `done`        | Codex GPT-5.6 Sol | [InterfaceTemplate scalar write presence](CW1-V2-02-I8.md) |
| `CW1-V2-02-I9` | `done`        | Codex GPT-5.6 Sol | [Rack scalar write presence](CW1-V2-02-I9.md)              |
| `CW1-V2-02-I10`| `done`        | Codex             | [Device scalar write presence](CW1-V2-02-I10.md)           |
| `CW1-V2-02-I11`| `done`        | Codex             | [Standing-authorized Interface scalar presence](CW1-V2-02-I11.md) |
| `CW1-V2-02-I12`| `done`        | Codex             | [Standing-authorized bounded VRF scalar presence](CW1-V2-02-I12.md); exact closeout/final receipt and integrity green |
| `CW1-V2-02-I13`| `packet`      | Codex             | [Prefix scalar write presence](CW1-V2-02-I13.md); independent reviews/standing packet permission recorded, local and exact CI pending |
| `CW1-V3-01`    | `done`        | Codex GPT-5.6 Sol | [Dependency-aware readiness](CW1-V3-01.md)                 |
| `CW1-V3-02`    | `done`        | Codex             | [Bootstrap and schema contract](CW1-V3-02.md)              |

`in-progress` means the packet owns only its explicitly active bounded
increment; no implementation evidence or parent-goal completion is implied.

`packet` means implementation remains blocked until the exact independently
reviewed packet claim passes repository CI.

`blocked` means the packet is reviewed but a named hard entry gate remains.

`evidence` means the bounded candidate has current retained results but still
awaits any human acceptance required by its packet. `done` means that bounded
goal or increment has the required accepted evidence. A child increment
entering either state never closes its parent automatically.

`done` for `CW1-V2-01` records accepted structural authority only. Its pending
and contradicted behavior remains open, it retains no V0 artifact, and it does
not promote a compatibility tier.

`CW1-V1-02-I4` has an exact tested candidate at
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f` with green accepted-I3 red-first,
local, real-PostgreSQL, feature-CI, main-CI, claim-CI, and receipt-CI
boundaries. Its `evidence` state is effective; project-owner review remains
before `done`. No parent, tier, or profile claim changes.

`CW1-V2-02-I1` has an exact tested candidate at
`7acba402f0de2bd59e5b342a6f05df268bc9120b` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, pinned repository, and
independent candidate-CI boundaries. Its evidence claim, pre-acceptance
receipt, owner-accepted closeout claim, and excluded closeout receipt all
passed exact-SHA CI. The project owner accepted only this bounded result at
`2026-08-23T17:25:54Z`; I1 is now effectively `done`. The external
differential was unavailable, and no parent, tier, profile, or
traceability-consumer claim changes.

`CW1-V2-02-I2` has an exact tested candidate at
`87863efd38fe71dfa05c818b860b37b7e94d67b4` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, pinned repository, and
independent candidate-CI boundaries. Its evidence claim and pre-acceptance
receipt exact-SHA CI are also green. The project owner accepted only this
bounded result at `2026-08-24T04:46:51Z`. Its owner-accepted closeout claim
and excluded closeout receipt both passed exact-SHA repository CI, so I2 is
effectively `done`. Docker rejected the external differential's
temporary source bind before oracle execution. Site uniqueness, deletion,
list/query behavior, full CRUD, and the 13-resource parent remain open. I2
changes no compatibility-tier or traceability-consumer boundary.

`CW1-V2-02-I3` has an exact tested candidate at
`651d33bc3fb2c8e663b6b14320af405b8501471f` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, pinned repository, and
independent candidate-CI boundaries. Its evidence claim and pre-acceptance
receipt exact-SHA repository CI are also green. The project owner's acceptance
of only this bounded result was recorded at `2026-08-24T10:28:34Z`. Its
owner-accepted closeout claim and excluded closeout receipt both passed
exact-SHA repository CI, so I3 is effectively `done`. No retained pinned
differential accompanies this bounded result. Manufacturer uniqueness,
deletion, list/query behavior, full CRUD, the 13-resource parent, all tiers,
and every traceability consumer remain open.

`CW1-V2-02-I4` has exact tested candidate
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, repository,
candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-24T18:51:56Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA
repository CI, so I4 is effectively `done`. I4 owns only
RackRole `name`, `slug`, `color`, and `description` create/PUT/PATCH presence,
defaults, validation, generated request/response contracts, matching
REST/gRPC semantics, PostgreSQL durability, Vue serialization/validation, and
eight fixed tests. No differential was run. RackRole uniqueness, deletion,
list/query behavior, full CRUD, the parent, all tiers, and every traceability
consumer remain open.

`CW1-V2-02-I5` has exact tested candidate
`89507d95d2743de7f97d64ca14cc43f6b834770b` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, repository,
candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-25T03:56:48Z`.
Its owner-accepted closeout claim and excluded receipt both passed exact-SHA
repository CI, so I5 is effectively `done`. I5 owns only the ten
declared RackType write fields across create/PUT/PATCH operation metadata,
typed domain/application behavior, REST/gRPC semantics, PostgreSQL durability,
Vue dirty-field serialization, and eight fixed tests. No differential was
run. RackType uniqueness, propagation, deletion, list/query behavior,
relationship-dictionary input, full CRUD, all tiers, every consumer, the
parent, and every later child remain open.

`CW1-V2-02-I6` has exact tested candidate
`dddd7adbda72f5dd760202c4862ce23b17cdf180` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, repository,
candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-25T08:13:40Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA
repository CI, so I6 is effectively `done`. I6 owns only
DeviceRole `parent`, `name`, `slug`, `color`, `vm_role`, `description`, and
`comments` create/PUT/PATCH presence across operation-specific contracts,
shared typed semantics, PostgreSQL durability, Vue dirty-field handling, and
eight fixed tests. No differential was run. DeviceRole hierarchy, uniqueness,
deletion, alternate parent input, list/query behavior, full CRUD, all tiers,
every consumer, the parent, profile promotion, and every later child remain
open.

`CW1-V2-02-I7` has exact tested candidate
`e2ad1acc33b84f20f24418d89b3b881b897b7ed3` with green focused, race,
real-PostgreSQL, complete L4, generated-contract, Vue, repository,
candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-25T14:12:13Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA
repository CI, so I7 is effectively `done`. I7 owns only the ten
declared DeviceType fields across create/PUT/PATCH operation metadata, typed
domain/application behavior, REST/gRPC semantics, PostgreSQL durability, Vue
dirty-field serialization, and eight fixed tests. No differential was run.
DeviceType uniqueness, positioned-Device height transitions,
deletion/cascades, alternate Manufacturer input, list/query behavior, full
CRUD, all tiers, every consumer, the parent, profile promotion, and every later
child remain open.

`CW1-V2-02-I8` has exact tested candidate
`b216d4c217cf863a8760494fd6499e54899ef368` and green candidate,
evidence-claim, and pre-acceptance receipt exact-SHA CI boundaries. The
project owner accepted only this bounded result at `2026-08-26T02:27:06Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA
repository CI, so I8 is effectively `done`. I8 owns only the
seven declared InterfaceTemplate writable fields across create/PUT/PATCH
operation metadata, typed domain/application behavior, REST/gRPC semantics,
PostgreSQL durability, Vue dirty-field serialization, and eight fixed tests.
No differential was run. Existing DeviceType-owner containment is exercised
without closing owner immutability; uniqueness, ModuleType ownership, bridge
behavior, Device instantiation/snapshot/rollback, non-retroactivity,
deletion, list/query behavior, full CRUD, all tiers, every consumer, the
parent, profile promotion, and every later child remain open.

`CW1-V2-02-I9` has exact tested candidate
`9c257b04b7cf798199c5aa4b7ae076cebbbbdff1` and green candidate,
evidence-claim, and pre-acceptance receipt exact-SHA CI boundaries. The
project owner accepted only this bounded result at `2026-08-26T11:01:23Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA CI, so
bounded I9 is effectively `done`. I9 owns only the 16 declared Rack writable
fields across create/PUT/PATCH operation metadata, typed domain/application
behavior, REST/gRPC semantics, PostgreSQL durability, Vue dirty-field
serialization, direct-save RackType copy precedence, and eight fixed tests.
No differential was run. Rack uniqueness, mounted-device and placement rules,
RackType-update propagation, Device site propagation, deletion, list/query
behavior, full CRUD, all tiers, every consumer, the parent, profile promotion,
and every later child remain open.

`CW1-V2-02-I10` has exact tested candidate
`c74821a0473a52873ea2bf26e781bf0fa5e33ebb` at source digest
`source-v2:sha256:9937f5371b3a50525f7fad0e0bfc41231857af1b043e764aa2abdd419b3e3640`
with 3,086 entries and green contract, focused, race, real-PostgreSQL,
complete-L4, Vue, backend, frontend, repository, exact committed replay,
candidate-CI, evidence-claim, and pre-closeout receipt gates. The project
owner accepted only the bounded evidence at receipt
`dde91785d34c4d76cbae40b99a191ad76437e9fd`, recorded at
`2026-10-12T14:07:30Z`; the independently reviewed owner-closeout scope
amendment `707c698c4dfc2952f0891935f3e6ba2e12f05e3c` passed exact-SHA CI.
Owner-accepted closeout claim
`9ae1bc3b46c0e8a5111b67d331355ee9e426abfd` and excluded receipt
`189633c42ab5aa9f9b2a7a271dbc5be48a4470f3` passed exact-SHA repository CI,
so bounded I10 is effectively `done`.
It owns only the 13 declared Device writable fields across create/PUT/PATCH
operation metadata, create-only
DeviceType-airflow inheritance, typed shared behavior, representable
REST/gRPC intent, PostgreSQL durability, Vue dirty-field handling, and eight
fixed tests. No differential was retained, so T2/T3 remain unearned and the
Vue unit/type boundary earns no T4. Device create's InterfaceTemplate
snapshot-order contradiction remains assigned to `CW1-V2-04`; uniqueness,
placement, template transactions,
list/get/delete, full CRUD, all tiers, every consumer, the parent, and profile
promotion remain open.

`CW1-V2-02-I11` is effectively standing-authorized implementation `done` only for
Interface common write presence in
`device`, `name`, `label`, `type`, `enabled`, `mgmt_only`, `mtu`, `speed`,
`duplex`, and `description`. It implements signed-int64 numeric intent,
operation-specific contracts, shared typed REST/gRPC semantics,
PostgreSQL durability/rollback tests, Vue dirty-field handling, and eight
fixed regressions. Tested candidate `8f4c6d6d3d8a981519a2bd8800428152b9b92d78`
has digest
`source-v2:sha256:e9e8520e5d2699f90e2e2d59a722112b12f5a84b2388da5d053ecb57fc4e4711`
with 3,093 entries. Focused/race, real PostgreSQL including six late-error
rollbacks, complete L4, generated/Vue, repository, committed replay, and
candidate CI passed. Initial claim `4874aa52c816660fb8077e0a798181ff2b7d83b4`
and excluded receipt `d455eb7943f5d169e0a7490ae6e645e124711710` passed
independent integrity, pinned local, and exact-SHA CI gates. Existing standing
acceptance was applied by the executor at `2026-10-12T22:13:41Z`, not a new
human decision or inspection. Eight-path closeout
`3c05346ffde251bf51de76a3187bc2367463704e` and final two-path receipt
`cb5d3f52a78b2baaab3503255181206ecbaeac81` passed independent, pinned local,
and exact-SHA CI gates. Final receipt run `38321116994`, job `115049181063`,
passed all 12 steps on unchanged source; frozen-byte/full-manifest checks
passed. No artifact-specific
human review or owner-immutability, tier, parent, profile, or consumer
completion is claimed.

`CW1-V2-02-I12` is effectively standing-authorized bounded implementation
`done` only for its five declared VRF writable fields. Reviewed packet
`009ab93e86114d36b85550d75e5c18d0cd09b8c9`
passed pinned local and exact packet CI gates before behavior editing.
Compiling semantic red proof preceded fixes. The corrected 21-path local
implementation covers only VRF-owned scalar validation,
operation contracts, shared representable API intent, PostgreSQL durability/
rollback, and Vue dirty-field/server-error presentation tests. Its local
focused/race, real PostgreSQL/complete L4, contracts/Vue, backend/frontend,
and repository checks passed. Exact candidate
`1cdfbbf6f39cbe494a367285fb2f2c5240523a0b` subsequently passed assembly,
unchanged committed focused/race, real PostgreSQL/complete L4 and repository
replay, and exact candidate CI. The
[bounded artifact](../evidence/2026-10-12-core-workflow-v2-vrf-scalar-presence-i12-v0.md)
and [evidence preparation](CW1-V2-02-I12.md#bounded-evidence-claim-preparation)
record those actual boundaries and verified disposable-database cleanup.
Evidence claim `14e9267e931debd38eb921357a39dda6b11cccd9` and pre-acceptance
receipt `ed8af4b653ef9f78110f7c4f2b80d590d57c99b2` passed their independent
integrity, pinned local and exact-SHA CI gates. The executor applied the owner's
existing standing instruction at `2026-10-13T04:27:07Z`, not a new human decision
or artifact inspection. Eight-path closeout
`bf3c4755c3e32beccf5053ca59b66c53b389d865` and final two-path excluded receipt
`4d8ea929b2c56e02b7e5ff021fb88d14090b77d0` passed their required independent
integrity, frozen-file pinned local and exact-SHA CI gates. Final run
`38354974081` / job `115155618986` passed all 12 reported steps, captured at
`2026-10-13T06:27:51.125Z` and independently audited at the separate time
`2026-10-13T06:30:30Z`. The [verified final leaf](CW1-V2-02-I12.md#verified-final-leaf-and-bounded-implementation-done)
consumes that boundary without a new acceptance or human inspection.
Shared Prefix/accepted IPAddress helpers and RD
list/query parsing remain untouched. No uniqueness/downstream policy,
deletion/list/query, tier, consumer, parent/profile, or rewrite completion
follows.

`CW1-V2-02-I13` is the authored Prefix scalar-presence `packet`, owned by
`ipam-core-workflow` / Codex, for `prefix`, `vrf`, `status`, `is_pool`,
`mark_utilized`, `description`, and `comments`. Its eight fixed regressions and
API/PostgreSQL/Vue work are planned, not implementation evidence. Independent
technical/companion reviews passed; existing standing routine packet permission
was applied as recorded in the packet, not implementation acceptance. Pinned
local and exact packet CI remain required before behavior editing. After effective
I13, review a bounded Prefix network-value packet under existing `CW1-V2-05`;
do not invent I14 or close the common-mutation parent, any tier or profile.

`CW1-V3-01` has exact tested candidate
`58241d7f87b642e6b7eca279592ad3759a0f5a06` at source digest
`source-v2:sha256:5cb32b753a4ac62544b2db24720b70c89904b0ae24731627e236dcd6c68564e9`
with 3,075 entries. Its focused, affected, race, complete-L4
real-PostgreSQL, coverage, backend, repository, candidate-CI, evidence-claim,
and pre-closeout receipt boundaries are green. The project owner accepted
only this bounded result at `2026-08-27T21:48:42Z`. Its owner-accepted
closeout claim `711d9e0e5ce34529fffeed840dc4812824c09f64` and excluded receipt
`e2f1863b7d925c195d6e759185de85a6497cb5e0` passed exact-SHA repository CI,
so bounded V3-01 is effectively `done`. It owns only process-only HTTP
liveness, request-time PostgreSQL-aware HTTP/gRPC readiness, empty-service
gRPC Health Check, fail-closed named-service/Watch behavior, shared
constructor injection, seven fixed named tests, and auxiliary constructor
coverage. V3-03 through V3-05, real deployment loss/recovery, streaming
monitoring, locking/rollback/reconnect/restart, every tier and consumer, the V3
parent, profile, rewrite completion, and production readiness remain open.

The independently reviewed V3-02 packet owns only bootstrap/schema proof.
Exact candidate `0cc4fe2365b7402929f65288eb776a425159d85b` at
`source-v2:sha256:ecc65583666214d2a470ff02806fc684b642c14cb5abc0df33ccad6c8b5188d6`
with 3,079 entries passed its registry-order regression plus six
PostgreSQL-backed scenarios, race, complete L4, backend, repository, and
candidate-CI, evidence-claim, and pre-closeout receipt boundaries without
changing production source. The project owner accepted only this bounded
result at `2026-10-12T04:01:48Z`. Owner-accepted closeout claim
`60efc848d0e52d9e446307d3a730cf76128161ca` and excluded receipt
`ced8f3bb31536e92f8a87be827e925cec306afaa` passed exact-SHA repository CI,
so bounded V3-02 is effectively `done`. It owns only deterministic startup ordering,
missing-table-only bootstrap/non-repair proof, the private schema catalog,
content-type natural-key uniqueness, and their test fixtures. It closes no V3
parent, compatibility tier, consumer, profile, deployment, rewrite, production
migration, or production-readiness boundary.

This bounded V3-02 closeout does not close V3 or make V3-03 ready because
V1-03 and the corresponding V2-03 through V2-07 dependencies remain
incomplete.
