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

| Goal           | State       | Intended executor | Packet                                                  |
| -------------- | ----------- | ----------------- | ------------------------------------------------------- |
| `CW1-V1-01`    | `done`      | Codex GPT-5.6 Sol | [Trusted-origin CORS](CW1-V1-01.md)                     |
| `CW1-V1-02-I1` | `done`      | Codex GPT-5.6 Sol | [Token credential foundation](CW1-V1-02.md)             |
| `CW1-V1-02-I2` | `done`      | Codex GPT-5.6 Sol | [Token transport and unary RPC safety](CW1-V1-02-I2.md) |
| `CW1-V1-02-I3` | `done`      | Codex GPT-5.6 Sol | [Browser session and CSRF lifecycle](CW1-V1-02-I3.md)   |
| `CW1-V1-02-I4` | `evidence`  | Codex GPT-5.6 Sol | [Password change and session rotation](CW1-V1-02-I4.md) |
| `CW1-V2-01`    | `done`      | Codex GPT-5.6 Sol | [Machine-readable traceability](CW1-V2-01.md)           |
| `CW1-V2-02-I1` | `evidence`* | Codex GPT-5.6 Sol | [IPAddress scalar write presence](CW1-V2-02-I1.md)      |

`in-progress` means the packet owns only its explicitly active bounded
increment; no implementation evidence or parent-goal completion is implied.

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
independent candidate-CI boundaries. Its `evidence` row is conditional: it
becomes effective only after the exact claim-only revision carrying this
transition passes repository CI. The digest-excluded receipt, unavailable
external differential, and project-owner review then remain. No parent, tier,
or profile claim changes.
