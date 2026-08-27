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
| `CW1-V3-01`    | `in-progress` | Codex GPT-5.6 Sol | [Dependency-aware readiness](CW1-V3-01.md)                 |

`in-progress` means the packet owns only its explicitly active bounded
increment; no implementation evidence or parent-goal completion is implied.

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

`CW1-V3-01` is the active packet for process-only HTTP liveness and
PostgreSQL-aware HTTP/gRPC readiness. It owns request-time `PingContext`, exact
public `/health` and `/ready` state mapping, empty-service gRPC Health Check,
fail-closed named-service/Watch behavior, shared constructor injection, seven
fixed named tests, and auxiliary constructor coverage. Its independently
reviewed exact packet and amendment gates are green, and the implementation
candidate has green local gates; exact committed replay, candidate CI, and
evidence remain. V3-02 through V3-05, deployment loss/recovery, streaming monitoring, tiers,
the profile, rewrite completion, and production readiness remain open.
