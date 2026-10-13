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
bounded result at `2026-08-25T14:12:13Z`. Owner-accepted closeout revision
`aa03e18239d5a1655e8254a148ae46f77c87c6d3` passed its pinned local gate and
exact-SHA repository CI, making only bounded I7 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI.
No retained pinned differential accompanies this bounded result, so no T2/T3
claim is made.
The
[bounded InterfaceTemplate scalar write-presence result](2026-08-25-core-workflow-v2-interface-template-scalar-presence-i8-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`b216d4c217cf863a8760494fd6499e54899ef368`. Its current two-path,
source-digest-excluded claim revision
`537a40ccf55e07a4455ea04e7d654a671f16ccfe` passed its pinned local gates and
exact-SHA repository CI, making only bounded I8 `evidence`. This
source-excluded receipt `354400176f0b122de6d654e12988048afff78e47`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at `2026-08-26T02:27:06Z`. Owner-accepted closeout revision
`9f69a1e31e012074a36d55281fbc5153b9e500fd` passed its pinned local gates and
exact-SHA repository CI, making only bounded I8 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI.
No retained pinned differential accompanies this bounded result, so no T2/T3
claim is made.
The
[bounded Rack scalar write-presence result](2026-08-26-core-workflow-v2-rack-scalar-presence-i9-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`9c257b04b7cf798199c5aa4b7ae076cebbbbdff1`. Its current two-path,
source-digest-excluded claim revision
`f24511f2dada4ecac2773614c2d237afbfd4275e` passed its pinned local gates and
exact-SHA repository CI, making only bounded I9 `evidence`. This
source-excluded receipt `199a1e0ad7edd4afd987b573c91a879485d97ffd`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at `2026-08-26T11:01:23Z`. Owner-accepted closeout revision
`8d59ba3629267d3fe851bf2f4fe09f0bca540b45` passed its pinned local gates and
exact-SHA repository CI, making only bounded I9 `done`. This two-path excluded
receipt preserves that result without changing source and requires its own
exact-SHA CI.
No retained pinned differential accompanies this bounded result, so no T2/T3
claim is made.
The
[bounded Device scalar write-presence result](2026-10-12-core-workflow-v2-device-scalar-presence-i10-v0.md)
passed its exact-name, affected/race, real-PostgreSQL, complete L4,
generated-contract, Vue, pinned repository, and independent exact-candidate
CI boundaries on exact revision
`c74821a0473a52873ea2bf26e781bf0fa5e33ebb`. Its current two-path,
source-v2-excluded claim revision
`33b595a5d5c7a1cac5444213cc7531a5382b7415` passed its pinned local gates and
exact-SHA repository CI, making only bounded I10 `evidence`. This
source-v2-excluded receipt `dde91785d34c4d76cbae40b99a191ad76437e9fd`
also passed exact-SHA repository CI. The project owner accepted only this
bounded result at that exact receipt, recorded at `2026-10-12T14:07:30Z`.
The independently reviewed one-file closeout-scope amendment
`707c698c4dfc2952f0891935f3e6ba2e12f05e3c` passed its pinned local gates
and exact-SHA repository CI. Owner-accepted eight-path closeout revision
`9ae1bc3b46c0e8a5111b67d331355ee9e426abfd` passed its pinned local gates
and exact-SHA repository CI, making only bounded I10 `done`. This two-path
excluded receipt preserves that result without changing source and requires
its own pinned local gates and exact-SHA CI. No retained pinned
differential accompanies this bounded result, so no T2/T3 claim is made.
The
[bounded Interface scalar write-presence result](2026-10-12-core-workflow-v2-interface-scalar-presence-i11-v0.md)
records exact candidate `8f4c6d6d3d8a981519a2bd8800428152b9b92d78` at
`source-v2:sha256:e9e8520e5d2699f90e2e2d59a722112b12f5a84b2388da5d053ecb57fc4e4711`,
3,093 entries. Its focused/race, real-PostgreSQL durability and six late-error
rollbacks, complete L4, generated/Vue, repository, unchanged committed replay,
and exact-candidate CI boundaries passed. Initial two-path excluded claim
`4874aa52c816660fb8077e0a798181ff2b7d83b4` passed both independent artifact
reviews, its assembly local gate, and its final no-later-edit `make check`
(`2026-10-12T20:57:35.063Z`–`21:04:03.428Z`, exit 0, log SHA-256
`e772eca17e9947a0fc45f182994b33ca6d346b3e4c30b90a49a76052f9e5a959`).
Its exact-head repository run
[38309027966](https://github.com/jihadismail8/netbox-go/actions/runs/38309027966)
and job
[115009960122](https://github.com/jihadismail8/netbox-go/actions/runs/38309027966/job/115009960122)
passed with all 12 reported steps successful; the artifact retains exact
times, raw JSON commitments, and the frozen claim mapping. The pre-closeout
two-path excluded receipt `d455eb7943f5d169e0a7490ae6e645e124711710` passed
independent integrity review and its assembly
`make check` (`2026-10-12T21:32:26.974Z`–`21:38:49.608Z`, exit 0, unchanged
source, log SHA-256
`ccc79153402d35cc712683cf17ecd5afe2f0a32a8564596f8da45fe74f581164`).
Its final no-later-edit pinned `make check` passed
`2026-10-12T21:42:38.915Z`–`2026-10-12T21:47:42.696Z`, exit 0, unchanged
source, log SHA-256
`f6e4f716fa4a351a82b40946d07a439fca7cfa402382152e3edc549da3529d48`.
Exact receipt run
[38313218478](https://github.com/jihadismail8/netbox-go/actions/runs/38313218478)
and job
[115023630424](https://github.com/jihadismail8/netbox-go/actions/runs/38313218478/job/115023630424)
passed, all 12 reported steps successful; the artifact retains exact times,
raw JSON commitments, and the complete frozen receipt mapping provenance.
The executor applied existing standing acceptance at `2026-10-12T22:13:41Z`,
not a new human decision or inspection. Exact eight-path closeout
`3c05346ffde251bf51de76a3187bc2367463704e` claims standing-authorized
implementation `done` only for I11. The complete tested manifest and
six-governance-path source diff are retained in the artifact.
Two independent closeout integrity reviews passed with no findings; pinned
assembly `make check` passed `2026-10-12T22:17:46.715Z`–
`2026-10-12T22:24:06.706Z`, exit 0, unchanged closeout source, log SHA-256
`3eb5148f705d9f9fbc8d0cfa1424ab6e6d71d219be6927f234f03ba9959de0d7`.
The closeout's final frozen-file pinned `make check` passed
`2026-10-12T22:27:52.390Z`–`2026-10-12T22:33:09.773Z`, exit 0, unchanged
source, log SHA-256
`6c6368921e29c31ab32eb95158e414c685cb013da182c687912cbb0165d1f0bf`.
Its committed-byte audit passed; the complete eight-path mapping and both
actual local-gate objects are retained in the reachable commit message.
Exact closeout run
[38317550055](https://github.com/jihadismail8/netbox-go/actions/runs/38317550055)
and job
[115037685537](https://github.com/jihadismail8/netbox-go/actions/runs/38317550055/job/115037685537)
passed, all 12 reported steps successful; the artifact retains exact times
and raw JSON commitments. This final two-path excluded receipt preserves
source `source-v2:sha256:57687459b6e0286bc43cf791395dfe3b933649c7422e270fa91b00dcbf8b7e98`
and 3,093 entries. Two independent receipt integrity reviews passed; pinned
assembly `make check` passed `2026-10-12T22:58:47.764Z`–
`2026-10-12T23:03:50.545Z`, exit 0, source unchanged, log SHA-256
`cb53920aac81f4db36b815fb89641f5a16dfbced913a4c2472750cd945db8d44`.
At that final-receipt assembly boundary, its final frozen-file local replay
and own exact-SHA CI were pending, so I11 was not yet effectively done and
I12 could not begin. Those gates subsequently passed at exact leaf receipt
`cb5d3f52a78b2baaab3503255181206ecbaeac81`: run
[38321116994](https://github.com/jihadismail8/netbox-go/actions/runs/38321116994)
and job
[115049181063](https://github.com/jihadismail8/netbox-go/actions/runs/38321116994/job/115049181063),
attempt 1, all 12 reported steps successful. The
[verified I12 entry](../increments/CW1-V2-02-I12.md#verified-i11-exit-not-a-new-human-acceptance)
retains the exact final local/CI/frozen-byte provenance without another I11
receipt chain. Only bounded standing-authorized I11 implementation `done`
became effective; its executor application time remains `2026-10-12T22:13:41Z`.
No new human artifact inspection or formal retained-claim authentication,
tier/consumer/parent/profile promotion, or production completion is claimed.
The pinned oracle image is unavailable, so T2/T3 remain unearned; Vue checks
earn no T4. The unchanged pinned dependency replay still reports 23 npm
advisories (seven moderate, fourteen high, two critical).
The
[bounded VRF scalar write-presence result](2026-10-12-core-workflow-v2-vrf-scalar-presence-i12-v0.md)
records exact tested candidate `1cdfbbf6f39cbe494a367285fb2f2c5240523a0b`,
source digest
`source-v2:sha256:fe70ecdbe799c99c93c3c8f9702905996c624205d7a72fa9559a6f8bef1d6c11`,
3,098 entries, full manifest SHA-256
`ebd789e2a9883286f543f640cdb748e776dd70c3282298d0435a0c7befb9c323`.
It binds the reviewed packet, genuine red proofs and correction history,
candidate assembly and unchanged committed focused/race, real-PostgreSQL
durability/complete-L4 and repository replay. Exact candidate run
[38333544757](https://github.com/jihadismail8/netbox-go/actions/runs/38333544757)
and job
[115088455846](https://github.com/jihadismail8/netbox-go/actions/runs/38333544757/job/115088455846)
completed success at the exact SHA, attempt 1, all 12 reported steps successful;
actual raw capture was `2026-10-13T02:00:12.149Z`. Raw/source/frozen-byte
verification and post-CI source CLI checks passed before this claim preparation.
Disposable cleanup ended `2026-10-13T01:40:49.180Z`, with `0|0` role/database
counts and the ephemeral credential removed. Actual I12 tool versions and the
complete tested manifest/governance-only source diff are retained by the artifact.
The preceding eight-path claim changed exactly six authorized source-governance
files plus that new source-v2-excluded artifact and this ledger. At its assembly
boundary, its independent review, pinned local and exact-SHA CI gates, and the
later receipt's gates, were pending. Those claim gates subsequently passed at
`14e9267e931debd38eb921357a39dda6b11cccd9`, single parent
`1cdfbbf6f39cbe494a367285fb2f2c5240523a0b`, final mapping SHA-256
`48cb35b9facdb761b9874b997cd5d54bc1136c9cb1fd675df800041ae3430bdf`.
Both corrected frozen-file pinned `make check` gates passed, exit 0, unchanged
claim source: assembly `2026-10-13T02:41:29.179Z`–`02:47:54.956Z` and final
`2026-10-13T02:48:34.742Z`–`02:55:00.772Z`. The artifact preserves the initial
failed gate and corrected-runner diagnostic as historical, not complete passes.
Exact claim run
[38339600527](https://github.com/jihadismail8/netbox-go/actions/runs/38339600527)
and job
[115107381745](https://github.com/jihadismail8/netbox-go/actions/runs/38339600527/job/115107381745)
completed success at that exact SHA, attempt 1, all 12 reported steps successful.
Run chronology was `2026-10-13T03:00:17Z`–`03:19:59Z`, job
`03:00:20Z`–`03:19:58Z`, repository step `03:01:22Z`–`03:19:55Z`.
Actual raw capture was `2026-10-13T03:22:00.711Z`, distinct from the separate
source/frozen-byte audit time. The artifact retains the complete final claim
mapping/freeze, both actual local-gate objects, normal exact-SHA push and
terminal CI receipt with raw commitments.
Pre-acceptance receipt `ed8af4b653ef9f78110f7c4f2b80d590d57c99b2` changed
only the two excluded evidence paths, preserving source
`source-v2:sha256:7f5042b01b55ad87e4b2b34419599fe10dc12f63ff85995629d7bc910b12b531`,
3,098 entries, full manifest SHA-256
`b92c249789a68a9e6cec23bcc198a9fa2d6930dc07b0bc0a5dc3af3ef5aa6bae`,
and the complete tested manifest/six-governance diff verbatim. At that receipt's
assembly its own review, two frozen-file pinned local gates and exact-SHA CI
were pending. Those gates subsequently passed with full two-path mapping
SHA-256 `ca7a80c90da93bfe3570a3f189e0d1f78ffe480b1aa815f188f0dd227f63653f`.
Receipt assembly `make check` passed
`2026-10-13T03:43:19.731Z`–`03:49:42.762Z`; final frozen-file gate passed
`2026-10-13T03:53:36.708Z`–`04:00:03.132Z`, both exit 0, unchanged source.
The reachable receipt commit retains its complete mapping and both actual
gate objects. Normal exact-SHA push passed
`2026-10-13T04:07:51.079Z`–`04:07:54.331Z`, with the recorded remote head
matching that receipt. Exact receipt run
[38344969001](https://github.com/jihadismail8/netbox-go/actions/runs/38344969001)
and job
[115124130023](https://github.com/jihadismail8/netbox-go/actions/runs/38344969001/job/115124130023)
completed success at that exact SHA, attempt 1, all 12 reported steps successful.
Run start–terminal update was `2026-10-13T04:07:55Z`–`04:23:03Z`; job
start–completion was `04:07:57Z`–`04:23:03Z`; repository step
`04:08:58Z`–`04:22:59Z`. Actual raw capture was `2026-10-13T04:24:28.686Z`,
distinct from the separate source/frozen-byte audit time. The artifact retains
the complete actual receipt proof and fixed raw commitments.
At `2026-10-13T04:27:07Z` the executor applied the owner's existing standing
instruction to that exact known gated bounded implementation result, not a
fresh human decision or artifact inspection. The artifact retains the immutable
standing-application record and its scope/exclusions. At that eight-path
closeout assembly, its own gates and the final excluded receipt's gates were
pending. The preceding standing-authorized closeout subsequently passed at
`bf3c4755c3e32beccf5053ca59b66c53b389d865`, single parent
`ed8af4b653ef9f78110f7c4f2b80d590d57c99b2`, full eight-path mapping SHA-256
`368c4326644d1b9e9149d10dc4a90e80d556a21931d282b0955b085bb25255c4`.
Both unchanged frozen-file pinned `make check` gates passed, exit 0: assembly
`2026-10-13T04:48:58.311Z`–`04:55:25.808Z` and final
`2026-10-13T04:55:55.824Z`–`05:02:25.561Z`. The actual reachable closeout
commit retains the complete mapping and both actual gate objects. Normal
exact-SHA push passed `2026-10-13T05:07:31.209Z`–`05:07:34.425Z`, with the
recorded remote head matching that closeout. Exact closeout run
[38349814434](https://github.com/jihadismail8/netbox-go/actions/runs/38349814434)
and job
[115139396635](https://github.com/jihadismail8/netbox-go/actions/runs/38349814434/job/115139396635)
completed success at that exact SHA, attempt 1, all 12 reported steps successful.
Run start–terminal update was `2026-10-13T05:07:36Z`–`05:27:11Z`; job
start–completion was `05:07:38Z`–`05:27:10Z`; repository step
`05:08:36Z`–`05:27:06Z`. Actual raw capture was `2026-10-13T05:28:49.645Z`,
distinct from the separate committed/source/frozen-byte audit time. The artifact
retains the complete actual closeout commit/freeze, both local-gate objects,
normal push and terminal CI receipt with fixed raw commitments.
This current final receipt changes only the artifact and this ledger,
preserving source
`source-v2:sha256:d8beb4bc21f52a991ab4e5b3eaa3deaa1bc5051686347a1991fe566116d89ea1`,
3,098 entries, complete manifest SHA-256
`88fdd755bd3b9eb47ce1fc53a3937630d138445ee9a88ffcbdf2111225ea007e`,
all preceding proof payloads, full tested manifest and original/closeout source
diffs verbatim. Its own independent integrity review, two frozen-file pinned
local gates and exact-SHA CI remain pending. I12 stays `in-progress`:
implementation `done` becomes effective only after this final leaf's actual
gates and captured integrity pass. The existing executor application remains
`2026-10-13T04:27:07Z`; no fresh acceptance application or source-governance
done promotion is made. No own final-receipt SHA/hash/CI or extra self-referential
receipt chain is predeclared.
No formal marker, fresh human inspection, designated retained-claim sign-off,
tier/consumer/parent/profile promotion or I13 start is claimed. The unavailable
pinned oracle leaves T2/T3 unearned; Vue checks earn no T4. Existing npm
advisories remain 23 (7 moderate, 14 high, 2 critical).

The planned
[bounded Prefix scalar write-presence artifact](2026-10-13-core-workflow-v2-prefix-scalar-presence-i13-v0.md)
is being prepared only for `CW1-V2-02-I13`'s seven declared writable fields.
The preceding I12 final-receipt pending statements describe its historical
assembly boundary; the [verified I13 entry](../increments/CW1-V2-02-I13.md#verified-i12-exit-not-a-new-human-acceptance)
retains the actual green final leaf and separate integrity audit, without
reapplying its `2026-10-13T04:27:07Z` standing acceptance or inventing inspection.
Reviewed packet `44a1f2018e7cda1c85273b714b96667f02577555` passed pinned
local and exact packet CI before compiling semantic red proof. Exact tested
candidate `1dc5a472032abda9c9b41a855ca7df8f8d4e2bb7`, single parent that
packet, has source
`source-v2:sha256:9266635d9d19148844271972e45514a224dd4b5e4d4a843ebba536ffaa68bfab`,
3,103 entries, complete tested manifest 754,272 bytes / SHA-256
`99a09218b8450b58fc7fa936719a42cca7a930f76b93f1544ab89310f3832978`,
and complete 21-path mapping SHA-256
`396ae79f3b32eb8d2dc6d3b450b67a983115988abcbad6253e994438067bd89b`.
All 14 local and 14 unchanged committed replay gates passed, including the
eight fixed tests, contracts/Vue, affected/race, real-PostgreSQL durability/
six late rollbacks, complete L4, backend/frontend and repository checks.
Committed repository replay ran `2026-10-13T09:26:42.305Z`–`09:33:09.312Z`,
exit 0, unchanged source, raw log SHA-256
`edac96f71711c30eb0ab366448e973ef3d4916f76555a69355afe83e0ca6cf23`.
The compiling red manifest and both original raw patches, historical v1
compile/v2 lint failures, two original v3 dependency-install/typecheck failures
and fresh successful after-install commands remain distinct actual records,
never relabelled as green. Actual tools include eight successful version outputs
and two failed ENOENT probes. Disposable `netbox_i13_20261013` cleanup ran
`2026-10-13T09:26:14.617Z`–`09:26:15.668Z`, `1|1` to `0|0`, zero active
connections, ephemeral credential removed.

Normal exact-SHA push ran `2026-10-13T09:36:52.678Z`–`09:36:56.000Z`,
with boundary verification `09:36:59.518Z`. Exact candidate run
[38377586847](https://github.com/jihadismail8/netbox-go/actions/runs/38377586847)
and job
[115229749443](https://github.com/jihadismail8/netbox-go/actions/runs/38377586847/job/115229749443)
completed success at that SHA, attempt 1, all 12 reported steps successful.
Run start–terminal update was `2026-10-13T09:36:57Z`–`09:56:37Z`; job
start–completion `09:37:00Z`–`09:56:36Z`; repository step
`09:37:53Z`–`09:56:33Z`. Actual raw capture was
`2026-10-13T09:58:17.032Z`, raw run SHA-256
`866ded328458e7e5796a34203240e8a38d330e71af24f1c7c299d1d99ec7fc86`
and job SHA-256
`be486d347f7fc9157e94fe0c5754cec45a10cd0b8706f4a131e5fbda58c840d3`.
Root audit passed at `2026-10-13T10:00:42.271Z`; independent complete captured
CI/source/Git/mapping/message/push/replay integrity passed separately at
`2026-10-13T10:02:31Z`, not at the borrowed capture time.

This initial conditional claim is `in-progress`, not effective `evidence` or
`done`. Its assembly changes only six authorized source-governance paths plus
the planned new source-v2-excluded artifact and this ledger. The artifact must
retain the complete tested/red manifests, both raw red patches, complete actual
histories/receipts and derived six-governance source diff. Its own independent
reviews, two frozen-file pinned local gates and exact claim CI, then the later
two-path excluded receipt's gates, remain pending. Only after those known gates
and integrity may the executor honestly apply existing standing implementation
acceptance; no I13 application has occurred. Separate eight-path closeout and
final two-path receipt gates remain before bounded `done` or successor work.
No own future claim digest/SHA/CI, marker, fresh human inspection, formal
retained-claim/row/consumer/tier/profile sign-off, parent/profile or production
promotion is declared. The oracle image is unavailable; T2/T3/T4 stay unearned
and the 23 existing npm advisories remain. Network/containment/uniqueness and
every other excluded boundary stay open.

The
[bounded dependency-aware readiness result](2026-08-27-core-workflow-v3-dependency-readiness-v0.md)
passed its exact-name, affected, race, complete L4 PostgreSQL-regression,
coverage, pinned repository, and independent exact-candidate CI boundaries on
exact revision `58241d7f87b642e6b7eca279592ad3759a0f5a06`. Its two-path,
source-v2-excluded claim revision
`e3b416d37e9dfb3f220122414b0bea658d993b3a` passed its pinned local gates and
exact-SHA repository CI, making only bounded V3-01 `evidence`. Its excluded
receipt `0c4d39fafa03d7f9088d1258d5339e06dfd00e78` also passed exact-SHA
repository CI. The project owner accepted only this bounded result at
`2026-08-27T21:48:42Z`. Owner-accepted closeout revision
`711d9e0e5ce34529fffeed840dc4812824c09f64` passed its pinned local gates and
exact-SHA repository CI, making only bounded V3-01 `done`. This two-path
excluded receipt preserves that result without changing source and requires
its own exact-SHA CI. Actual deployment dependency loss/recovery remains
`CW1-V3-04`, so V3 and production readiness stay open.
The
[bounded bootstrap/schema result](2026-10-11-core-workflow-v3-bootstrap-schema-v0.md)
passed its registry-order regression plus six PostgreSQL-backed scenarios,
race, complete L4, backend, repository, and independent exact-candidate CI
boundaries on exact revision
`0cc4fe2365b7402929f65288eb776a425159d85b`. This ten-path claim records only
candidate-state governance and source-v2-excluded evidence. Exact claim
revision `979a1558bbba2993ebb4f0d512a94fb0036c6b8b` passed its pinned local
gates and exact-SHA repository CI, making only bounded V3-02 `evidence`. This
two-path, source-v2-excluded receipt
`452f80dc257f02d1d0b03b524a208c84a1d66abc` passed exact-SHA repository CI
without changing source. The project owner accepted only this bounded result
at `2026-10-12T04:01:48Z`. Owner-accepted closeout revision
`60efc848d0e52d9e446307d3a730cf76128161ca` passed its pinned local gates and
exact-SHA repository CI, making only bounded V3-02 `done`. This two-path
excluded receipt preserves that result without changing source and requires
its own exact-SHA CI. It promotes no tier, consumer, V3 parent, profile,
deployment, rewrite, production migration, or production-readiness boundary.
The project owner reviewed and retained the entry V0, CORS, bounded token I1,
bounded token I2, bounded browser-session I3, bounded IPAddress I1, and bounded
Site I2, Manufacturer I3, RackRole I4, RackType I5, DeviceRole I6, DeviceType
I7, InterfaceTemplate I8, Rack I9, and Device I10 results.
The
[2026-08-03 post-cleanup V0](2026-08-03-post-cleanup-v0.md) used the superseded
mode-blind v1 source digest and is now historical. The 2026-08-01
repository V0 and recovery-scoped PostgreSQL replay also remain historical in
the
[Core Workflow recovery artifact](2026-08-01-core-workflow-v1-v0.md). The
Profile remains T1 and pre-publication; V1-V6 remain open. `CW1-G00` passed for
the exact I10 tested source; `CW1-V1-01`, `CW1-V1-02-I1`, `CW1-V1-02-I2`, and
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
`CW1-V2-02-I7` has an effective owner-accepted `done` closeout at exact
revision `aa03e18239d5a1655e8254a148ae46f77c87c6d3`, only for DeviceType
common scalar create/PUT/PATCH presence. This excluded receipt preserves that
exact result without changing source and must pass its own exact-SHA CI. I7
promotes no compatibility tier, profile state, parent, later child, or
traceability consumer. Feature work must re-establish V0 after its next
owned-source change.
`CW1-V2-02-I8` has an effective owner-accepted `done` closeout at exact
revision `9f69a1e31e012074a36d55281fbc5153b9e500fd`, only for InterfaceTemplate
common scalar create/PUT/PATCH presence. This excluded receipt preserves that
exact result without changing source and must pass its own exact-SHA CI. I8
promotes no compatibility tier, profile state, parent, later child, or
traceability consumer.
`CW1-V2-02-I9` has an effective owner-accepted `done` closeout at exact
revision `8d59ba3629267d3fe851bf2f4fe09f0bca540b45`, only for Rack common scalar
create/PUT/PATCH presence. This excluded receipt preserves that exact result
without changing source and must pass its own exact-SHA CI. I9 promotes no
compatibility tier, profile state, parent, later child, or traceability
consumer.
`CW1-V2-02-I10` has an effective owner-accepted `done` closeout at exact
revision `9ae1bc3b46c0e8a5111b67d331355ee9e426abfd`, only for Device common
scalar and relationship create/PUT/PATCH presence. Its exact evidence
claim `33b595a5d5c7a1cac5444213cc7531a5382b7415`, excluded receipt
`dde91785d34c4d76cbae40b99a191ad76437e9fd`, and reviewed one-file scope
amendment `707c698c4dfc2952f0891935f3e6ba2e12f05e3c` passed exact-SHA
CI. Owner acceptance is bound to that receipt at `2026-10-12T14:07:30Z`.
The eight-path closeout passed independent validation, its assembly and final
no-later-edit local gates, and exact-SHA CI. This excluded receipt preserves
that result without changing source. Its assembly local gate passed; the
final no-later-edit replay and its own exact-SHA CI remain. I10 promotes no
compatibility tier, profile state,
parent, later child, or traceability consumer.
`CW1-V3-01` has an effective owner-accepted `done` closeout at exact revision
`711d9e0e5ce34529fffeed840dc4812824c09f64`, only for its bounded
operational-probe slice. This excluded receipt preserves that exact result
without changing source and must pass its own exact-SHA CI. It promotes no
compatibility tier, profile state, V3 parent, traceability consumer,
deployment, rewrite, or production boundary.
`CW1-V3-02` has an effective owner-accepted `done` closeout at exact revision
`60efc848d0e52d9e446307d3a730cf76128161ca`, only for its bounded
bootstrap/schema proof. This excluded receipt preserves that exact result
without changing source and must pass its own exact-SHA CI. It promotes no
compatibility tier, profile state, V3 parent, traceability consumer,
deployment, migration, rewrite, or production boundary.

| Evidence                 | Required command or boundary                                  | Current state                                                    |
| ------------------------ | ------------------------------------------------------------- | ---------------------------------------------------------------- |
| Repository quality       | Revised `make check` including non-mutating backend coverage  | Prior bounded results retained; I11/I12 bounded done at verified final leaves; I13 tested candidate/replay/exact CI green, own conditional claim/receipt gates pending |
| Strict REST differential | `make compatibility-test`                                     | Harness present; current result pending                          |
| gRPC semantic parity     | `go test ./test/parity -count=1` plus corresponding T2 report | Tests present; current result pending                            |
| Real PostgreSQL          | DSN-enabled bootstrap/schema/concurrency/identity suites      | Prior bounded suites and I11/I12/I13 committed durability/complete-L4 passed; I13 own conditional claim/receipt gates pending; V3-03 through V3-05 open |
| Standalone deployment    | `make deployment-smoke`                                       | CORS-scoped exact-commit result retained; V3-04 remains open     |
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

## I13 excluded pre-acceptance receipt preparation

The preceding I13 initial-claim pending statements describe its historical assembly boundary. Its actual known green eight-file claim, complete frozen mapping, both pinned local gates, normal publication, terminal exact-SHA CI and distinct post-capture integrity reviews are retained in the [bounded Prefix scalar artifact](2026-10-13-core-workflow-v2-prefix-scalar-presence-i13-v0.md#actual-initial-claim-boundary-and-excluded-pre-acceptance-receipt).

This two-file excluded receipt preserves the complete original artifact and native source at source-v2:sha256:74f1cb1101a64ed15f4b8ea5f544c015d8bee4b78a6dcd8acb0e0fab667bf08c, 3,103 entries. Its own independent integrity review, two frozen pinned local gates and exact-SHA CI remain pending. I13 stays in progress; standing implementation acceptance, separate closeout and effective done have not been applied. No fresh human inspection, formal sign-off, tier/consumer/parent/profile or production promotion follows. It predicts no own future receipt SHA/hash/CI or acceptance time.

## I13 standing-authorized bounded closeout preparation

The earlier I13 table cells report historical assembly boundaries; this appendix
supersedes their status without altering the retained table bytes.

The preceding I13 initial-claim and pre-acceptance pending statements describe
their historical assembly boundaries. Actual initial claim
`d161e251b0f838c57ed0a057a4493fbb02a7c40b` and excluded pre-acceptance
receipt `77197e854725bba6e625e5fcfb05468aa43caf27` passed independent
integrity, two frozen pinned local and exact-SHA CI gates. The accepted receipt
preserves source
`source-v2:sha256:74f1cb1101a64ed15f4b8ea5f544c015d8bee4b78a6dcd8acb0e0fab667bf08c`,
3,103 entries/full manifest SHA-256
`b50630c0c1ec6ea4d7de9e6e9f7a27d2c36fcc56f8934f5d53de685e57a2cdca`;
the historical tested candidate remains `1dc5a472032abda9c9b41a855ca7df8f8d4e2bb7`
at tested digest
`source-v2:sha256:9266635d9d19148844271972e45514a224dd4b5e4d4a843ebba536ffaa68bfab`.
The [bounded Prefix artifact](2026-10-13-core-workflow-v2-prefix-scalar-presence-i13-v0.md)
preserves the complete prior manifests, raw patches, canonical proofs and
operational failures; their historical results are not replaced.

Receipt run `38398462515` / job `115299600865`, attempt 1, passed all 12
reported steps. Actual raw capture was `2026-10-13T12:57:20.760Z`, distinct
from root review `2026-10-13T13:00:08Z` and nonauthor independent review
`2026-10-13T13:01:22.591Z`. The canonical integrity record SHA-256 is
`226ecf41739afd88871fdae6b5928d8e61d34bab36473d4db4751eee3db71e2a`.
At `2026-10-13T13:02:43Z`, after both actual reviews, the executor applied
existing standing authorization, immutable record SHA-256
`e39451540dd367f3b65544ba4990723f37f119b58bbda4234a2bb7ec53b25eed`.
This accepts only Prefix's seven-field scalar create/PUT/PATCH presence,
operation contracts, shared representable API semantics, PostgreSQL
durability/rollback and Vue dirty-field/error tests; it is not new human
inspection, formal retained-claim authentication or a designated sign-off.

Only bounded implementation `done` is conditional. It is not effective until
the separately reviewed eight-path closeout and final two-path excluded receipt
each pass their own independent integrity, two frozen pinned local and exact-SHA
CI gates. Those future boundaries remain pending; no own closeout hash/SHA,
source digest, CI or acceptance is predicted. No successor, full CRUD,
network/containment/uniqueness, tier/consumer/parent/profile, rewrite, deployment
or production boundary closes. The oracle remains unavailable, T2/T3/T4 and
formal human sign-offs unearned, disposable database/credential absent and
23 existing npm advisories unchanged.
