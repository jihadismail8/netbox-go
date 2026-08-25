# Evidence ledger

This directory records durable verification results for compatibility tiers and
release gates. Source files and test harnesses are linked from
[Testing](../TESTING.md); they are not pass evidence by themselves.

## Current ledger

The [2026-08-17 source-v2 V0](2026-08-17-core-workflow-v1-source-v2-v0.md),
[trusted-origin CORS evidence](2026-08-03-core-workflow-v1-cors-v0.md), and
[bounded token-authentication foundation](2026-08-17-core-workflow-v1-token-auth-i1-v0.md)
passed their exact committed local boundaries and linked independent GitHub
runs. The
[bounded token-transport and unary-safety result](2026-08-17-core-workflow-v1-token-transport-i2-v0.md)
also passed its exact committed local boundary, independent candidate CI, and
evidence-receipt CI. The
[bounded browser-session and CSRF result](2026-08-17-core-workflow-v1-browser-session-i3-v0.md)
passed its exact committed local, real-PostgreSQL, independent candidate CI,
and evidence-receipt CI boundaries. The exact
[password-change and session-rotation result](2026-08-19-core-workflow-v1-password-session-i4-v0.md)
passed its accepted-I3 red-first, focused, race, real-PostgreSQL, complete L4,
pinned repository, feature-candidate CI, and main exact-SHA CI boundaries. Its
claim-only revision
`d2a19d8e61838785be939c77e0e1a35ac95a075e` also passed its pinned local gate
and exact-SHA repository CI, making bounded I4 `evidence`. This digest-excluded
receipt also passed exact-SHA CI; project-owner review remains.
The
[bounded IPAddress scalar write-presence result](2026-08-19-core-workflow-v2-ipaddress-scalar-presence-i1-v0.md)
passed its focused, race, real-PostgreSQL, complete L4, generated-contract,
Vue, pinned repository, and independent exact-candidate CI boundaries. Claim
revision `05f24e211b4b446d3337941ba8811b7ffdd68ae6` also passed its pinned
local gate and exact-SHA repository CI, making bounded I1 `evidence`. Its
digest-excluded receipt `0d2638182f141858963ab9b66a837244772c3972`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at `2026-08-23T17:25:54Z`. Owner-accepted closeout revision
`6e3568fd8eded6280a1957e2b30b6c6a2541a0c2` passed its pinned local gate and
exact-SHA repository CI, making only bounded I1 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI. The unavailable external differential remains pending.
The
[bounded Site scalar write-presence result](2026-08-23-core-workflow-v2-site-scalar-presence-i2-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries. Claim revision
`65d37902b4111e1f091ff8c9a75548966514082b` also passed its pinned local gate
and exact-SHA repository CI, making bounded I2 `evidence`. Its
digest-excluded receipt `eee2b835d9052566d40b1c95af356157029c277b`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at `2026-08-24T04:46:51Z`. Owner-accepted closeout revision
`dd2550d3e721d27efc525838d3a9bc0b1bf329f9` passed its pinned local gate and
exact-SHA repository CI, making only bounded I2 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI. Docker rejected the differential harness's temporary source bind
before oracle execution, so no T2/T3 claim is made.
The
[bounded Manufacturer scalar write-presence result](2026-08-24-core-workflow-v2-manufacturer-scalar-presence-i3-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries. Claim revision
`2771102e04671492a351177547a1395cd9512524` also passed its pinned local gate
and exact-SHA repository CI, making bounded I3 `evidence`. Its
digest-excluded receipt `414a2c22758ea9b92b98780f97a69f4416b472ce`
also passed exact-SHA repository CI. The project owner's acceptance of only
this bounded result was recorded at `2026-08-24T10:28:34Z`. Owner-accepted
closeout revision `9274f8dbe331af14e5f3f7953f1ae8a74c100e15` passed its
pinned local gate and exact-SHA repository CI, making only bounded I3 `done`.
This two-path excluded receipt preserves that result without changing source
and requires its own exact-SHA CI. No retained pinned differential accompanies
this bounded result, so no T2/T3 claim is made.
The
[bounded RackRole scalar write-presence result](2026-08-24-core-workflow-v2-rack-role-scalar-presence-i4-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe`. Its two-path,
source-digest-excluded claim revision
`7277f09afa22d5b17dde5c12e8607ef7c7ebb33f` passed its pinned local gate and
exact-SHA repository CI, making only bounded I4 `evidence`. Its excluded
receipt `9dd7727dbe66463bafe4e7adb17c90835ae91c56` also passed exact-SHA
repository CI. The project owner accepted only this bounded result at
`2026-08-24T18:51:56Z`. Owner-accepted closeout revision
`0ff9cf8481e6b703cffcd07b6f562ca04119b07b` passed its pinned local gate and
exact-SHA repository CI, making only bounded I4 `done`. Its two-path excluded
receipt `9f1ef5697f0e1457ee2e7574977444b8f2356bbe` also passed exact-SHA
repository CI and preserves that result without changing source. No retained
pinned differential accompanies this bounded result, so no T2/T3 claim is
made.
The
[bounded RackType scalar write-presence result](2026-08-24-core-workflow-v2-rack-type-scalar-presence-i5-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`89507d95d2743de7f97d64ca14cc43f6b834770b`. Its current two-path,
source-digest-excluded claim revision
`25d6f10bab78b9b92c1682fe9b951cc4f2286ea8` passed its pinned local gate and
exact-SHA repository CI, making only bounded I5 `evidence`. This receipt
`bbd8d014735188cb1b621cb1053f58850db2a805` also passed exact-SHA
repository CI. The project owner accepted only this bounded result at
`2026-08-25T03:56:48Z`. Owner-accepted closeout revision
`492b95b9494e6c3bc8bf1f7e7fe3280b20ac928c` passed its pinned local gate and
exact-SHA repository CI, making only bounded I5 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI. No retained pinned differential accompanies this bounded
result, so no T2/T3 claim is made.
The
[bounded DeviceRole scalar write-presence result](2026-08-25-core-workflow-v2-device-role-scalar-presence-i6-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`dddd7adbda72f5dd760202c4862ce23b17cdf180`. Its current two-path,
source-digest-excluded claim revision
`ad897478df134dcc71a08f068651de0b832ab913` passed its pinned local gate and
exact-SHA repository CI, making only bounded I6 `evidence`. This receipt
`f2dfb6ef1a95dda6b5fa4ebabdff447ac28521c4` also passed exact-SHA
repository CI. The project owner accepted only this bounded result at
`2026-08-25T08:13:40Z`. Owner-accepted closeout revision
`54ce96884281a94829ebbef26dd64f0e6c112178` passed its pinned local gate and
exact-SHA repository CI, making only bounded I6 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI. No retained pinned differential accompanies this bounded
result, so no T2/T3 claim is made.
The
[bounded DeviceType scalar write-presence result](2026-08-25-core-workflow-v2-device-type-scalar-presence-i7-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`e2ad1acc33b84f20f24418d89b3b881b897b7ed3`. Its current two-path,
source-digest-excluded claim revision
`c6ffdec22b5f47b529feba7fc4bb864b95a7ac13` passed its pinned local gate and
exact-SHA repository CI, making only bounded I7 `evidence`. This
source-excluded receipt `2ac51b6fce930d24949343b56a8a4338abaec9ce`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at `2026-08-25T14:12:13Z`. Its `done` state is conditional on
the exact owner-accepted closeout claim passing repository CI; a two-path
excluded receipt must then preserve that result and pass its own exact-SHA CI.
No retained pinned differential accompanies this bounded result, so no T2/T3
claim is made.
The project owner reviewed and retained the entry V0, CORS, bounded token I1,
bounded token I2, bounded browser-session I3, bounded IPAddress I1, and bounded
Site I2, Manufacturer I3, RackRole I4, RackType I5, DeviceRole I6, and
DeviceType I7 results.
The
[2026-08-03 post-cleanup V0](2026-08-03-post-cleanup-v0.md) used the superseded
mode-blind v1 source digest and is now historical. The 2026-08-01
repository V0 and recovery-scoped PostgreSQL replay also remain historical in
the
[Core Workflow recovery artifact](2026-08-01-core-workflow-v1-v0.md). The
Profile remains T1 and pre-publication; V1-V6 remain open. `CW1-G00` passed for
the exact I7 tested source; `CW1-V1-01`, `CW1-V1-02-I1`, `CW1-V1-02-I2`, and
`CW1-V1-02-I3` are done. The parent credential/session matrix remains open.
`CW1-V2-02-I1` has an effective owner-accepted `done` closeout only for
IPAddress scalar
create/PUT/PATCH presence. The current excluded receipt preserves that fact and
is evaluated by exact-SHA CI without a self-referential documentation commit.
`CW1-V2-02-I2` has an effective owner-accepted `done` closeout at exact
revision `dd2550d3e721d27efc525838d3a9bc0b1bf329f9`, only for Site scalar
create/PUT/PATCH presence. This excluded receipt preserves that exact result
without changing source and must pass its own exact-SHA CI.
`CW1-V2-02-I3` has an effective owner-accepted `done` closeout at exact
revision `9274f8dbe331af14e5f3f7953f1ae8a74c100e15`, only for Manufacturer
`name`, `slug`, and `description` create/PUT/PATCH presence. This excluded
receipt preserves that exact result without changing source and must pass its
own exact-SHA CI. I3 promotes no compatibility-tier, profile, parent, or
traceability-consumer boundary.
`CW1-V2-02-I4` has an effective owner-accepted `done` closeout at exact
revision `0ff9cf8481e6b703cffcd07b6f562ca04119b07b`, only for RackRole
`name`, `slug`, `color`, and `description` create/PUT/PATCH presence. Excluded
receipt `9f1ef5697f0e1457ee2e7574977444b8f2356bbe` passed exact-SHA repository
CI and preserves that exact result without changing source. I4 promotes no
compatibility tier, profile, parent, or traceability-consumer boundary.
`CW1-V2-02-I5` has an effective owner-accepted `done` closeout at exact
revision `492b95b9494e6c3bc8bf1f7e7fe3280b20ac928c`, only for RackType common
scalar create/PUT/PATCH presence. This excluded receipt preserves that exact
result without changing source and must pass its own exact-SHA CI. I5 promotes
no compatibility tier, profile state, parent, later child, or traceability
consumer. The external differential, REST T2, corresponding gRPC T3, browser
T4, and 13-resource `CW1-V2-02` parent stay open.
`CW1-V2-02-I6` has an effective owner-accepted `done` closeout at exact
revision `54ce96884281a94829ebbef26dd64f0e6c112178`, only for DeviceRole
common scalar create/PUT/PATCH presence. This excluded receipt preserves that
exact result without changing source and must pass its own exact-SHA CI. I6
promotes no compatibility tier, profile state, parent, later child, or
traceability consumer. I2 and I3 promote no
compatibility-tier or traceability-consumer
boundary.
`CW1-V2-02-I7` has owner-accepted conditional `done`, only for DeviceType
common scalar create/PUT/PATCH presence. Its evidence claim and
pre-acceptance receipt passed exact-SHA repository CI. The closeout becomes
effective only after the exact owner-accepted claim revision passes repository
CI; its excluded receipt then remains. I7 promotes no compatibility tier,
profile state, parent, later child, or traceability consumer. Feature work must
re-establish V0 after its next owned-source change.

| Evidence                 | Required command or boundary                                  | Current state                                                    |
| ------------------------ | ------------------------------------------------------------- | ---------------------------------------------------------------- |
| Repository quality       | Revised `make check` including non-mutating backend coverage  | V1-I4 and V2-I1/I2/I3/I4 retained; V2-I5/I6 done; V2-I7 owner-accepted closeout claim conditional |
| Strict REST differential | `make compatibility-test`                                     | Harness present; current result pending                          |
| gRPC semantic parity     | `go test ./test/parity -count=1` plus corresponding T2 report | Tests present; current result pending                            |
| Real PostgreSQL          | DSN-enabled bootstrap/schema/concurrency/identity suites      | I1, I3, V1-I4, V2-I1/I2/I3/I4/I5/I6, and bounded V2-I7 passed; V1/V3 remain open |
| Standalone deployment    | `make deployment-smoke`                                       | CORS-scoped exact-commit result retained; V3 still open          |
| Browser workflow         | `make browser-e2e`                                            | Harness present; current result pending                          |

GitHub CI currently runs V0 only. Real PostgreSQL, deployment, differential
oracle, gRPC promotion, and browser gates remain operator-run until durable,
connected CI jobs retain their credential-free artifacts.

## Required artifact shape

A result may be cited from [Project status](../STATUS.md) only when it records:

- the full reachable tested Git revision and its exact versioned
  `make source-digest` value; retained source-v2 claims are unavailable when
  repository Git metadata is unavailable, and the oracle remains identified
  separately by its pinned Git SHA; all authoritative historical-object reads
  disable local Git replacement refs and grafts;
- command, toolchain, and relevant non-secret configuration;
- start/end time and pass/fail status;
- the pinned oracle SHA/configuration for differential runs;
- per-capability scenario totals and failures where a tier is affected;
- durable-state and side-effect checks required by the scenario; and
- a credential-free artifact location retained by CI or committed as a concise
  summary.

Raw logs must not contain passwords, cookies, CSRF values, bearer/API tokens,
hashes, DSNs, or complete configuration objects. Temporary paths under `/tmp`
are useful diagnostics but are not durable evidence after the environment is
gone.

## Promotion rule

1. Retain the successful current artifact.
2. Link it from this ledger and the affected Capability Profile/report.
3. Review that the comparator used only committed normalizers.
4. Bind evidence to only the exact rows exercised. Promote a tier-owning
   resource only after every applicable row it owns closes that tier boundary.
5. Keep the contract pre-publication until every first-profile exit condition
   is satisfied together.

An earlier successful run cannot be reused after comparator, behavioral
contract, implementation, fixture, or security changes without a fresh
execution. Claim-only tier/link/status metadata uses the audited exception
below.

## Claim-only attestation

Updating tier/evidence links and Status after a successful run changes
`make source-digest`, even though the tested behavior did not change. Use the
two-digest protocol in
[the execution playbook](../IMPLEMENTATION_EXECUTION_PLAYBOOK.md#two-digest-claim-attestation):

1. retain gates against an unchanged `tested_digest`;
2. allow only reviewed evidence summaries/links, exact tier/contract-state
   metadata, metadata-only generated docs, and Status claim changes;
3. record a file/content-hash diff proving no behavior, schema, scenario,
   fixture, comparator, normalizer, security rule, or toolchain changed;
4. compute the final `attestation_digest` and run `make check`; and
5. retain a mapping from the final digest to the tested digest and artifacts.

Anything outside that claim-only boundary invalidates the affected evidence and
requires a fresh external run.

Retained claims use `source-v2:sha256:<hex>` and
`netbox-go-evidence-v2`. The tested revision must be a commit reachable from
the claimed `HEAD`; its complete owned-source manifest is reconstructed from
the Git tree. File mode (`100644`/`100755`), byte size/hash, and closed relative
symlinks are part of that manifest. Two-digest claims retain the tested
manifest and must match it to the Git tree entry for entry.

Result payload SHA-256 commitments cover exact raw non-marker bytes before any
decoding; the whole artifact must be valid UTF-8 and marker lines must be ASCII.
The marker's claim digest covers the complete verification/evidence references
and each exact consumer row's full assessment, applicability, and proof
assignment/object. These checks prove structural integrity and prevent claim
rebinding. They cannot prove a command ran or its output/environment was
truthful, so durable runner provenance and designated human evidence acceptance
remain mandatory.
