# Testing

Verified structurally: 2026-08-03

A test file or harness is not a passed gate. Results must come from the current
revision, use the required external boundary, and be retained according to the
[evidence policy](evidence/README.md). The latest full post-hardening external
runs are pending, so this document does not claim T2, T3, or T4.

## Fast repository gate

From the repository root:

```bash
make check
```

This runs:

1. backend format, vet, lint, race-unit, non-mutating coverage non-regression,
   build, frozen legacy-client compile, frozen proto compile, and canonical Buf
   checks;
2. pinned frontend install, formatting, lint, application/test typechecks,
   coverage tests, and production build;
3. deterministic current-surface inventory drift checks;
4. Capability Profile, inventory, protobuf descriptor, generated contract,
   OpenAPI, and Markdown link validation.

The obsolete Vue-registry API catalogue and DRF registry generators have been
retired. `generated-check` now validates only the authoritative contract
inventory. Frozen live-client tests still require an unmanaged service and are
compile-checked under the owned exclusion in
[`legacy-exclusions.yml`](../netbox-backend/quality/legacy-exclusions.yml);
they are not parity evidence.

Backend `coverage-check` and its backward-compatible `cover` alias write the
atomic profile only beneath `/tmp`, validate it, and remove it. The reviewed
[`coverage-baseline.tsv`](../netbox-backend/quality/coverage-baseline.tsv)
records 54 measured module packages, the one owned generated live-client
exclusion, and the current statement baseline of 36,013/290,425 (12.4001%).
The prior post-recovery floor was 35,987/292,320 (12.3108%). Package-set drift,
an unowned exclusion, or a lower exact ratio fails the gate; lowering the
baseline is not a permitted repair. Durable V0 status is recorded only in the
[evidence ledger](evidence/README.md).

The [2026-08-03 artifact](evidence/2026-08-03-post-cleanup-v0.md) is historical:
it used the superseded mode-blind v1 digest. The
[source-v2 artifact](evidence/2026-08-17-core-workflow-v1-source-v2-v0.md)
retains the exact entry revision after human review. Every later source change
must refresh that continuous gate. Accepted bounded `CW1-V1-02-I1` evidence
separately proves token lookup classification, strict `last_used` ordering, affected local
REST/gRPC mapping, and actual-update behavior on an owned disposable
PostgreSQL schema. It does not complete the parent credential matrix or promote
T2/T3. Accepted bounded `CW1-V1-02-I2` evidence adds the exact baseline REST Token
grammar and details, typed credential-state causes, strict unary gRPC Bearer
grammar, an explicit 28-read/57-write method classification over 85 protected
RPCs, malformed-input zero-lookup containment, and cross-transport
principal/cause/effect parity. It remains direct-peer-only and does not close
session, CSRF, throttle, trusted-proxy, simultaneous-credential, or streaming
rows. Accepted bounded `CW1-V1-02-I3` evidence adds typed browser-session outcomes,
valid-session-first REST arbitration, exact double-submit CSRF pairing,
transactional login/logout, deterministic active-session CSRF recovery, fixed
session-cookie shape, and real-PostgreSQL rollback/advisory-lock/concurrency
proof. It remains a classified browser-identity extension and does not close
password-session, throttle, trusted-proxy, streaming, aggregate all-transport,
REST T2, gRPC T3, or browser T4 work. GitHub CI currently runs the V0 command
only. `CW1-V1-02-I4` has an exact tested candidate at
`c4b1ce1f00cb255b684fb9d795e4e5c7a578907f`, source digest
`source-v2:sha256:3f37417bb791ac6bc97ac4a0d23c5f928062feecf81a0f8a4fb9e57445d53670`
over 3,013 entries. Its rebased red-first, focused, PostgreSQL,
affected-package, race, backend, root, feature-candidate CI, and main exact-SHA
CI boundaries passed. The current nine-path claim-only transition records
bounded `evidence` only if the exact attestation revision passes repository CI;
its digest-excluded receipt and project-owner review then remain. The owned
PostgreSQL, deployment, differential, gRPC-promotion, and
browser gates remain manual until connected jobs retain durable artifacts.

`CW1-V2-02-I1` has an exact tested candidate at
`7acba402f0de2bd59e5b342a6f05df268bc9120b`, source digest
`source-v2:sha256:ed330b0a5bbeafd70b7b16a4ce4d1052fa9a385313a3b8827b554983571c1b43`
over 3,018 entries. Its eight named domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the affected race,
generated-contract, Vue adapter/form, complete L4, pinned repository, and
independent exact-candidate CI boundaries. Its evidence claim and
pre-acceptance receipt also passed exact-SHA repository CI. The project owner
accepted only this bounded result at `2026-08-23T17:25:54Z`; the
owner-accepted closeout claim and excluded closeout receipt also passed
exact-SHA CI, so I1 is effectively `done`. The
pinned REST oracle remained unavailable. The local transport/parity and Vue
results do not substitute for that oracle or promote T2, T3, or T4.

`CW1-V2-02-I2` has an exact tested candidate at
`87863efd38fe71dfa05c818b860b37b7e94d67b4`, source digest
`source-v2:sha256:c7c1b86c2bcd768bb719149a54dddbceccf2b5b2e4087dd4b79eec20bef5a37c`
over 3,022 entries. Its eight named Site domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 17-test Vue adapter/form, complete L4, pinned repository,
and independent exact-candidate CI boundaries. Its evidence claim and
pre-acceptance receipt also passed exact-SHA repository CI. The project owner
accepted only this bounded result at `2026-08-24T04:46:51Z`; the
owner-accepted closeout claim and excluded closeout receipt both passed
exact-SHA repository CI, so I2 is effectively `done`. Docker
rejected the differential harness's temporary source bind before oracle
execution. The local transport/parity and Vue results do not substitute for
that oracle or promote T2, T3, or T4.

`CW1-V2-02-I3` has an exact tested candidate at
`651d33bc3fb2c8e663b6b14320af405b8501471f`, source digest
`source-v2:sha256:09499a6618569d2dae224edfb339ac82585bf0248d20e5a4d5ff23d19221fe6f`
over 3,029 entries. Its eight named Manufacturer domain, application, REST,
gRPC, real-PostgreSQL, and parity tests passed, as did the exact
affected/race, generated-contract, 23-test Vue adapter/form, complete L4,
pinned repository, and independent exact-candidate CI boundaries. Its evidence
claim and pre-acceptance receipt also passed exact-SHA repository CI. The
project owner's acceptance of only this bounded result was recorded at
`2026-08-24T10:28:34Z`; its owner-accepted closeout claim and excluded
closeout receipt both passed exact-SHA repository CI, so I3 is effectively
`done`. No retained pinned differential accompanies this result. The local
transport/parity and Vue results do not substitute for that oracle or promote
T2, T3, or T4.

`CW1-V2-02-I4` has exact tested candidate
`f1ef3d5e21b66a8e2f77bd380c09c81a8ef5dbfe` at source digest
`source-v2:sha256:68db7a9835545d66ef9a651b9c4a000c91f3d834b5503a1b89e4a122275c3bc9`
over 3,036 entries. Its exact-name, affected/race, real-PostgreSQL, complete
L4, generated-contract, Vue, backend/frontend/root, candidate-CI,
evidence-claim, and pre-acceptance receipt boundaries are green. The project
owner accepted only this bounded result at `2026-08-24T18:51:56Z`. Its
owner-accepted closeout claim and excluded receipt passed exact-SHA repository
CI, so I4 is effectively `done`. The matrix remains limited to RackRole
`name`, `slug`, `color`, and
`description` create/PUT/PATCH omitted/null/blank/concrete states across eight
fixed domain, application, REST, gRPC, real-PostgreSQL, and parity tests,
generated OpenAPI, and Vue adapter/form checks. No differential was run, so
T2/T3 remain unearned; the Vue boundary earns no T4.

`CW1-V2-02-I5` has exact tested candidate
`89507d95d2743de7f97d64ca14cc43f6b834770b` at source digest
`source-v2:sha256:a8325eaae703aa801ed587deefae7e8d08d9c9e0189c80ff7569da95c36d6f90`
over 3,043 entries. Its eight named RackType domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 35-test Vue adapter/form, complete L4, backend/frontend/
root, candidate-CI, evidence-claim, pre-acceptance receipt,
owner-accepted closeout claim, and excluded receipt boundaries. The project
owner accepted only this bounded result at `2026-08-25T03:56:48Z`, so I5 is
effectively `done`. The matrix is limited to the
ten declared RackType writable fields across create/PUT/PATCH presence,
numeric-ID Manufacturer input, generated OpenAPI, shared REST/gRPC semantics,
PostgreSQL durability, and Vue dirty-field handling. No differential was run,
so T2/T3 remain unearned; the Vue boundary earns no T4. RackType uniqueness,
propagation, deletion, list/query behavior, alternate Manufacturer inputs,
full CRUD, the parent, profile promotion, and every later child remain open.

`CW1-V2-02-I6` has exact tested candidate
`dddd7adbda72f5dd760202c4862ce23b17cdf180` at source digest
`source-v2:sha256:be2168180bfbbb10406772d57e75210d32e5084051c4b765d4cf2954d958e621`
over 3,050 entries. Its eight named DeviceRole domain, application, REST,
gRPC, real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 41-test Vue adapter/form, complete L4, backend/frontend/
root, candidate-CI, evidence-claim, and pre-acceptance receipt boundaries. The
project owner accepted only this bounded result at `2026-08-25T08:13:40Z`.
Its owner-accepted closeout claim and excluded receipt passed exact-SHA
repository CI, so I6 is effectively `done`. The matrix is limited to
DeviceRole `parent`, `name`, `slug`, `color`, `vm_role`, `description`, and
`comments` create/PUT/PATCH presence, numeric-ID parent input, generated
OpenAPI, shared REST/gRPC semantics, PostgreSQL durability, and Vue dirty-field
handling. No differential was run, so T2/T3 remain unearned; the Vue boundary
earns no T4. DeviceRole hierarchy, uniqueness, deletion, alternate parent
inputs, list/query behavior, full CRUD, the parent, profile promotion, and
every later child remain open.

`CW1-V2-02-I7` has exact tested candidate
`e2ad1acc33b84f20f24418d89b3b881b897b7ed3` at source digest
`source-v2:sha256:8930fa5fbe487df4c225a301c674db096664cccaf3f0c53df93c51e59edaeba7`
with 3,057 entries. Its eight named DeviceType domain, application, REST,
gRPC, real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, 47-test focused Vue adapter/form, complete L4,
backend/frontend/root, candidate-CI, evidence-claim, and pre-acceptance receipt
boundaries. The project owner accepted only this bounded result at
`2026-08-25T14:12:13Z`. Its owner-accepted closeout claim and excluded receipt
passed exact-SHA repository CI, so I7 is effectively `done`. The matrix is
limited to DeviceType `manufacturer`, `model`, `slug`,
`part_number`, `u_height`, `exclude_from_utilization`, `is_full_depth`,
`airflow`, `description`, and `comments` create/PUT/PATCH presence,
numeric-ID Manufacturer input, generated OpenAPI, shared REST/gRPC semantics,
PostgreSQL durability, and Vue dirty-field handling. No differential was run,
so T2/T3 remain unearned; the Vue boundary earns no T4. DeviceType uniqueness,
positioned-Device height transitions, deletion/cascades, list/query behavior,
full CRUD, the parent, profile promotion, and every later child remain open.

`CW1-V2-02-I8` has exact tested candidate
`b216d4c217cf863a8760494fd6499e54899ef368` at source digest
`source-v2:sha256:da2b93d51dbdc2dfbcb6e348e2f5b23f42439b5973169f683ddc39de285c5048`
with 3,064 entries. Its eight named InterfaceTemplate domain, application,
REST, gRPC, real-PostgreSQL, and parity tests passed, as did the exact
affected/race, generated-contract, focused Vue adapter/form, complete L4,
backend/frontend/root, candidate-CI, evidence-claim, and pre-acceptance
receipt boundaries. The project owner accepted only this bounded result at
`2026-08-26T02:27:06Z`. Its owner-accepted closeout claim and excluded receipt
passed exact-SHA repository CI, so I8 is effectively `done`. The matrix is
limited to InterfaceTemplate `device_type`, `name`,
`label`, `type`, `enabled`, `mgmt_only`, and `description` create/PUT/PATCH
presence, generated OpenAPI, shared REST/gRPC semantics, PostgreSQL
durability, and Vue dirty-field handling. No differential was run, so T2/T3
remain unearned; the Vue boundary earns no T4. Existing DeviceType-owner
containment is exercised without closing owner immutability. Uniqueness,
ModuleType ownership, bridge behavior, Device instantiation/snapshot/rollback,
non-retroactivity, deletion, list/query, full CRUD, the parent, profile
promotion, every tier, and every later child remain open.

`CW1-V2-02-I9` has exact tested candidate
`9c257b04b7cf798199c5aa4b7ae076cebbbbdff1` at source digest
`source-v2:sha256:343a1767534de69acad81e7fdfbf8bd23cf1e7a22450c00df70b038c4c54c152`
with 3,071 entries. Its eight named Rack domain, application, REST, gRPC,
real-PostgreSQL, and parity tests passed, as did the exact affected/race,
generated-contract, focused Vue adapter/form, complete L4,
backend/frontend/root, candidate-CI, evidence-claim, and pre-acceptance receipt
boundaries. The project owner accepted only this bounded result at
`2026-08-26T11:01:23Z`. Its owner-accepted closeout claim and excluded receipt
passed exact-SHA repository CI, so bounded I9 is effectively `done`. The
matrix is limited to Rack `site`, `name`, `facility_id`,
`rack_type`, `status`, `role`, `serial`, `asset_tag`, `form_factor`, `width`,
`u_height`, `starting_unit`, `desc_units`, `airflow`, `description`, and
`comments` create/PUT/PATCH presence, generated OpenAPI, shared REST/gRPC
semantics, direct-save RackType copy precedence, PostgreSQL durability, and
Vue dirty-field handling. No differential was run, so T2/T3 remain unearned;
the Vue boundary earns no T4. Rack uniqueness, mounted-device and placement
rules, RackType-update propagation, Device site propagation, deletion,
list/query, full CRUD, the parent, profile promotion, every tier, and every
later child remain open.

`CW1-V2-02-I10` has exact tested candidate
`c74821a0473a52873ea2bf26e781bf0fa5e33ebb` at source digest
`source-v2:sha256:9937f5371b3a50525f7fad0e0bfc41231857af1b043e764aa2abdd419b3e3640`
with 3,086 entries. Its eight fixed Device tests
exist exactly once and pass across scalar normalization, command presence,
pinned PUT omission defaults, failed-validation non-effects, REST, gRPC,
real-PostgreSQL durability, and REST/gRPC parity. The local candidate also
passes 254 capability/OpenAPI adversarial tests, 96 traceability tests, the
293-row profile check, 33-path/86-operation OpenAPI validation, affected and
race suites, complete L4, 65 focused Vue tests and both typechecks, backend
coverage, the 185-test frontend gate, and the repository gate. Exact committed
replay, candidate-CI, evidence-claim, and pre-closeout receipt CI are green.
The disposable PostgreSQL role/database were removed and verified absent; no credential is
retained. The project owner accepted only this bounded evidence at receipt
`dde91785d34c4d76cbae40b99a191ad76437e9fd`, recorded at
`2026-10-12T14:07:30Z`; the independently reviewed owner-closeout scope
amendment `707c698c4dfc2952f0891935f3e6ba2e12f05e3c` passed exact-SHA
repository CI. Owner-accepted closeout claim
`9ae1bc3b46c0e8a5111b67d331355ee9e426abfd` and excluded receipt
`189633c42ab5aa9f9b2a7a271dbc5be48a4470f3` passed exact-SHA repository CI,
so bounded I10 is effectively `done`. The
[retained I10 artifact](evidence/2026-10-12-core-workflow-v2-device-scalar-presence-i10-v0.md)
records the bounded tests and their command results. No differential was
run, so T2/T3/T4 remain unearned. Device create's template-snapshot
contradiction, placement, uniqueness, list/get/delete, every tier, and the
parent remain open.

The [`CW1-V2-02-I11` increment](increments/CW1-V2-02-I11.md) has eight
fixed Interface regressions across domain, command/application presence,
validation non-effects, REST, gRPC, real-PostgreSQL durability/rollback, and
representable transport parity, plus operation-specific contract and Vue
dirty-field checks. They cover only the ten declared writable fields,
including signed-int64 numeric intent. Independently reviewed packet
`840d889346e846b46efc98bb79305f0dcfb01957` passed its pinned local gate and
exact-SHA CI run `38295273529` before behavior editing. Semantic red proofs
preceded fixes; the eight named tests, focused/race, real-PostgreSQL raw/reload/
snapshot and six late-error rollback cases, complete L4, generated/Vue, and
complete repository gates passed. Tested candidate
`8f4c6d6d3d8a981519a2bd8800428152b9b92d78`, digest
`source-v2:sha256:e9e8520e5d2699f90e2e2d59a722112b12f5a84b2388da5d053ecb57fc4e4711`
(3,093 entries), also passed unchanged committed replay and exact-candidate
CI. Initial claim `4874aa52c816660fb8077e0a798181ff2b7d83b4` and excluded
receipt `d455eb7943f5d169e0a7490ae6e645e124711710` passed independent
integrity, pinned local, and exact-SHA CI gates. Existing standing acceptance
was applied by the executor at `2026-10-12T22:13:41Z`, without claiming a
new human decision or inspection. Standing-authorized implementation `done`
is effective after eight-path closeout `3c05346ffde251bf51de76a3187bc2367463704e`
and final receipt `cb5d3f52a78b2baaab3503255181206ecbaeac81` passed independent,
pinned local, and exact-SHA CI gates. Final receipt run `38321116994`, job
`115049181063`, passed all 12 steps; frozen bytes and the complete closeout
source manifest were verified unchanged.
The pinned oracle image is not cached;
no differential T2 or corresponding T3 is retained and Vue checks earn no T4.
No owner-immutability, parent, profile, or consumer completion follows.

The [`CW1-V2-02-I12` increment](increments/CW1-V2-02-I12.md) has eight fixed
VRF regressions across domain normalization, command/application presence,
validation non-effects, REST, gRPC, real-PostgreSQL durability and six
delegate-then-error rollbacks, and representable parity. Operation-specific
contract and Vue dirty-field/error presentation tests accompany them.
Optional PUT tests seed false and concrete optional state; NUL plus length
errors must aggregate without hiding RD failures. Shared Prefix/accepted
IPAddress helper bodies and RD list/query parsing remain unchanged.
Independently reviewed packet `009ab93e86114d36b85550d75e5c18d0cd09b8c9`
passed pinned local and exact packet CI gates before behavior editing. Four
compiling core regressions failed behaviorally on unchanged production before
fixes. A protobuf mutex-copy vet failure in the initial parity fixture was
corrected without weakening vet; earlier snapshot passes are not the final
implementation boundary. The corrected 21-path snapshot passed exact-name,
focused/race, real-PostgreSQL durability and complete L4, 404 capability/OpenAPI
and 96 traceability tests, 84 focused Vue tests, both typechecks, backend,
204-test frontend, and pinned repository checks. Those are historical
pre-governance local results. Exact candidate
`1cdfbbf6f39cbe494a367285fb2f2c5240523a0b` subsequently passed assembly,
unchanged committed focused/race, real-PostgreSQL durability/complete L4 and
repository replay, and exact candidate CI. Its repository replay includes
the 204-test frontend suite, both typechecks, 404 capability/OpenAPI and 96
traceability tests, and the unchanged strict quality policies. The
[bounded artifact](evidence/2026-10-12-core-workflow-v2-vrf-scalar-presence-i12-v0.md)
and [evidence preparation](increments/CW1-V2-02-I12.md#bounded-evidence-claim-preparation)
bind the prior packet, red proofs, actual candidate receipts and full manifest.
The preceding eight-path claim `14e9267e931debd38eb921357a39dda6b11cccd9`
and pre-acceptance receipt `ed8af4b653ef9f78110f7c4f2b80d590d57c99b2` passed
their independent integrity, pinned local and exact-SHA CI gates. The executor
applied existing standing acceptance at `2026-10-13T04:27:07Z`, not fresh human
inspection. Eight-path closeout `bf3c4755c3e32beccf5053ca59b66c53b389d865`
and final two-path excluded receipt `4d8ea929b2c56e02b7e5ff021fb88d14090b77d0`
passed their independent integrity, frozen-file pinned local and exact-SHA CI
gates. Final run `38354974081` / job `115155618986`, attempt 1, passed all 12
reported steps; actual raw capture was `2026-10-13T06:27:51.125Z`, followed by
the separate independent integrity audit at `2026-10-13T06:30:30Z`. Only bounded
I12 implementation `done` is effective; its original standing application is
not reapplied or relabelled as human inspection.
Actual I12 toolchain inventory and strict
gate results are retained, not borrowed from I11. The read-only
oracle preflight found the NetBox image unavailable; no differential ran,
same-core parity earns no corresponding T3, and Vue checks earn no T4. The
23 existing npm advisories (7 moderate, 14 high, 2 critical) remain unchanged.
No tier, consumer, parent/profile, or browser outcome is earned.

The [`CW1-V2-02-I13` increment](increments/CW1-V2-02-I13.md) implements exactly eight
Prefix regressions: `TestPrefixScalarNormalizationContract`,
`TestPrefixScalarCommandPresenceMatrix`,
`TestReplacePrefixPreservesOmittedOptionalState`,
`TestPrefixScalarValidationLeavesStateUnchanged`,
`TestPrefixRESTScalarPresenceMatrix`, `TestPrefixGRPCScalarPresenceMatrix`,
`TestPostgresPrefixScalarPresenceDurability`, and
`TestPrefixScalarPresenceRESTGRPCParity`. The implemented matrix covers
seven-field intent, optional PUT preservation, exact Prefix-owned status and
numeric-VRF diagnostics, mixed-validation non-effects, real PostgreSQL
durability/six delegate-then-error rollbacks and representable same-core parity.
Operation-metadata/OpenAPI and actual DynamicForm dirty-field/server-error
checks accompany them. Reviewed packet
`44a1f2018e7cda1c85273b714b96667f02577555` passed its pinned local and exact
CI gates before behavior editing. Five core/gRPC regressions compiled and
failed semantically on unchanged production; contract red preserved 404 existing
passes and Vue red preserved 84. Subsequent v1 parity compilation failure was
corrected by three response `VrfId` references; v2 staticcheck QF1003 by two
equivalent tagged switches in the new REST test. Those are correction history,
not the genuine compiling red proof or final-source passes.

Exact tested candidate `1dc5a472032abda9c9b41a855ca7df8f8d4e2bb7`, source
`source-v2:sha256:9266635d9d19148844271972e45514a224dd4b5e4d4a843ebba536ffaa68bfab`
with 3,103 entries, passed all 14 local and 14 unchanged committed replay gates.
The eight fixed names each exist exactly once. The replay includes 539
capability/OpenAPI and 96 traceability tests, 219 Vue tests in 29 files, both
typechecks, generated checks and 187 Markdown files with valid links. Backend
coverage was 38,637/291,842 (13.2390%), above the unchanged 37,876/291,613
(12.9884%) floor, with 55 measured packages and one reviewed exclusion.
Real PostgreSQL executed 64 named subcases, including six actual delegated
transaction-local row/change observations followed by sentinel errors and outer
rollback, plus raw-row/reload/identity/timestamp and exact seven-field snapshots.
The focused PostgreSQL matrix and complete L4 passed; the owned role/database
was removed and verified `0|0`, with its ephemeral credential removed.

Two original v3 standalone typechecks failed while parallel `npm ci` removed
dependencies (exits 127 and 2); immutable logs remain failures. Fresh
`implementation-v3-vue-typecheck-after-install` and
`implementation-v3-vue-test-typecheck-after-install` passed after installation
on unchanged final source. Eight actual tool outputs and two failed ENOENT
plugin-path probes are retained separately, without pin or policy changes.
Exact candidate run `38377586847` / job `115229749443`, attempt 1, passed
all 12 reported steps; capture was `2026-10-13T09:58:17.032Z`, root audit
`10:00:42.271Z` and independent full integrity `10:02:31Z`.
The [bounded artifact](evidence/2026-10-13-core-workflow-v2-prefix-scalar-presence-i13-v0.md)
retains initial claim `d161e251b0f838c57ed0a057a4493fbb02a7c40b` and excluded
pre-acceptance receipt `77197e854725bba6e625e5fcfb05468aa43caf27`; both passed
independent integrity, two frozen pinned local gates and exact-SHA CI. The
receipt's source is
`source-v2:sha256:74f1cb1101a64ed15f4b8ea5f544c015d8bee4b78a6dcd8acb0e0fab667bf08c`,
3,103 entries, distinct from the tested candidate. Receipt run `38398462515` /
job `115299600865`, attempt 1, passed all 12 steps; actual capture
`2026-10-13T12:57:20.760Z`, root review `13:00:08Z` and independent review
`13:01:22.591Z` are separate. Its local gates preserve the historical candidate
records and report 188 Markdown files, not a replacement for the candidate's
187-file result. Existing standing acceptance was applied by the executor at
`2026-10-13T13:02:43Z`; bounded implementation `done` remains conditional,
not effective, until separate eight-path closeout and final two-path receipt
pass their own independent integrity, two frozen pinned local and exact-SHA CI
gates. No fresh human inspection or formal sign-off follows.
The unavailable oracle earns no T2 or corresponding T3;
same-core parity and Vue units/types earn no T4. The 23 existing npm advisories
(7 moderate, 14 high, 2 critical) remain. Existing network/query code and tests
are unchanged; no canonicalization, containment, `/0`, uniqueness, tier,
consumer, parent/profile, fresh human inspection or formal sign-off follows.

`CW1-V3-01` has exact tested candidate
`58241d7f87b642e6b7eca279592ad3759a0f5a06` at source digest
`source-v2:sha256:5cb32b753a4ac62544b2db24720b70c89904b0ae24731627e236dcd6c68564e9`
with 3,075 entries. Its focused, affected, race, complete-L4
real-PostgreSQL, coverage, backend, repository, candidate-CI, evidence-claim,
and pre-closeout receipt boundaries are green. The project owner accepted
only this bounded result at `2026-08-27T21:48:42Z`; its owner-accepted
closeout claim `711d9e0e5ce34529fffeed840dc4812824c09f64` and excluded receipt
`e2f1863b7d925c195d6e759185de85a6497cb5e0` passed exact-SHA repository CI,
so bounded V3-01 is effectively `done`. Seven
fixed named tests plus auxiliary constructor coverage prove `/health` remains
200 without a ping, `/ready` maps current dependency state to 200/503,
empty-service gRPC Health Check maps it to SERVING/NOT_SERVING, named services
and Watch fail closed, and missing composition cannot silently serve ready.
Real deployment loss/recovery remains V3-04; no T2/T3/T4, profile, V3-parent,
deployment, rewrite, or production claim is made.

## Owned external gates

### Differential REST oracle

```bash
make compatibility-comparator-test
make compatibility-test
```

[`tests/compatibility/`](../tests/compatibility/) owns the pinned NetBox oracle,
isolated oracle/Go databases, deterministic fixtures, strict comparator, and
artifact directory. It refuses a dirty or mismatched oracle SHA and mismatched
effective configuration.

The comparator permits only the normalizers declared in
[`normalizers.yaml`](../contracts/netbox/v4.4.6-post7/normalizers.yaml). It must
fail on status, validation-reason, authorization, path/trailing-slash,
query-order, choice-envelope/label, numeric JSON type, missing/extra-field,
committed-state, or side-effect divergence. Generated IDs are bound to named
scenario objects; an unbound ID is a failure. A successful current report is
required for T2; the harness's existence is only T1 evidence. The current
driver has broad CRUD/filter/ordering coverage but does not yet cover every
declared permission, presence, conflict, invariant, rollback, and side-effect
case. Build the scenario/rule traceability matrix and add the uncovered cases
before treating `make compatibility-test` as a complete first-profile T2 gate.

### gRPC semantic parity

The backend parity packages cover canonical resource lifecycle, error mapping,
authorization, Device-instantiation rollback, and IPAddress assign/unassign
through shared REST-visible state:

```bash
cd netbox-backend
go test ./test/parity -count=1
```

These tests support T3 only after the corresponding REST capabilities have
earned T2 and the current full parity run is retained. A separate transport
shape does not permit separate business behavior.

### Real PostgreSQL

Provide a disposable PostgreSQL DSN; each suite creates and removes its own
schema:

```bash
cd netbox-backend
NETBOX_TEST_POSTGRES_DSN='postgres://...' \
  go test \
    ./internal/database \
    ./internal/adapters/postgres/bootstrap \
    ./internal/adapters/postgres/dcim \
    ./internal/adapters/postgres/dcim/row \
    ./internal/adapters/postgres/ipam \
    ./internal/adapters/postgres/ipam/row \
    ./internal/adapters/postgres/changelog \
    ./internal/application/dcim \
    ./internal/application/ipam \
    ./internal/application/identity \
    ./cmd/netbox_go_admin \
    -count=1
```

The checked-in suites exercise missing-table-only bootstrap, the 198-entry
startup registry, the ten DCIM rows owned by
[`dcim/row/rows.go`](../netbox-backend/internal/adapters/postgres/dcim/row/rows.go),
the three IPAM rows owned by
[`ipam/row/rows.go`](../netbox-backend/internal/adapters/postgres/ipam/row/rows.go),
the change row owned by
[`changelog/row.go`](../netbox-backend/internal/adapters/postgres/changelog/row.go),
constraints/indexes, concurrent duplicate and conditional-IP handling,
assignment/audit serialization, one-time administrator bootstrap,
authenticated non-superuser creation, and authenticated model-permission
grants. SQLite tests may support orchestration, but cannot prove PostgreSQL
constraints, locking, or concurrency.

Current retained real-PostgreSQL/deployment runs remain required for V3-03
through V3-05 and the V3 parent. The bounded V3-02 retained result does not
prove those later boundaries.

The independently reviewed
[CW1-V3-02 packet](increments/CW1-V3-02.md) fixes the bootstrap/schema proof
boundary. Exact candidate `0cc4fe2365b7402929f65288eb776a425159d85b`
passed the registry-order regression plus six PostgreSQL-backed scenarios,
the affected race suite, complete L4, backend, repository, candidate-CI,
evidence-claim, and pre-closeout receipt gates at source digest
`source-v2:sha256:ecc65583666214d2a470ff02806fc684b642c14cb5abc0df33ccad6c8b5188d6`.
The project owner accepted only this bounded result at
`2026-10-12T04:01:48Z`. Owner-accepted closeout claim
`60efc848d0e52d9e446307d3a730cf76128161ca` and excluded receipt
`ced8f3bb31536e92f8a87be827e925cec306afaa` passed exact-SHA repository CI,
so bounded V3-02 is effectively `done`. It changes no production source, and neither it nor the existing broad PostgreSQL
suites close V3-03 through V3-05 or the V3 parent.

### Deployment smoke

```bash
make deployment-smoke
```

[`compose_smoke.sh`](../tests/deployment/compose_smoke.sh) owns a uniquely named
Compose project and volume, starts a clean PostgreSQL and standalone Go/Vue
application, verifies health and schema/content-type bootstrap, restarts the
application, checks idempotence, and tears everything down. It does not use a
Django migration or initialization-SQL mount.

The current smoke checks `/health` but does not exercise actual PostgreSQL
loss and recovery. The runtime `/ready` and empty-service gRPC Health Check now
map request-time dependency state under the bounded deterministic V3-01 tests;
V3-04 must extend the real deployment run with PostgreSQL loss/recovery before
its exit condition can pass.

### Real-browser workflow

```bash
make browser-e2e
```

[`tests/browser/`](../tests/browser/) drives an installed headless Chrome over
the disposable deployment using Chrome DevTools Protocol. It retains
credential-free diagnostics on failure. The harness is the required boundary
for T4; component tests or a typed adapter alone cannot earn workflow parity.

A successful current run must cover the declared DCIM/IPAM workflow and the
applicable authentication, validation, real limited-user permission denial,
rollback, assignment/unassignment, and delete-warning outcomes before T4 is
considered.

The current browser driver covers the main creation chains and a subset of
negative cases. Session refresh, edit/list/filter, late template rollback,
explicit-null update, reassignment/assignment rollback, conflict, not-found,
and exact change-state outcomes remain to be added and retained.

## Implemented test layers

| Layer                      | Checked-in boundary                                    | What it can prove                                | What it cannot prove alone                              |
| -------------------------- | ------------------------------------------------------ | ------------------------------------------------ | ------------------------------------------------------- |
| Domain unit                | Table-driven Go tests without I/O                      | Invariants and stable reasons                    | SQL, transport, or oracle behavior                      |
| Application integration    | Typed resource/identity services and transaction tests | Orchestration, authorization, rollback intent    | External PostgreSQL semantics unless a real DSN is used |
| REST contract              | In-process canonical router and OpenAPI equality       | Runtime surface and local mappings               | Baseline compatibility                                  |
| gRPC contract/parity       | In-process canonical services over shared state        | RPC mapping and same-core outcomes               | T3 before REST reaches T2                               |
| PostgreSQL integration     | Disposable schemas on a real server                    | Tables, constraints, indexes, locks, concurrency | Oracle or browser parity                                |
| Differential compatibility | Pinned NetBox and Go systems with isolated databases   | Exact-in-profile REST compatibility              | Vue workflow parity                                     |
| Vue component              | Typed adapters/components with controlled API input    | Hydration, payloads, filters, UI behavior        | Real session/backend/database workflow                  |
| Browser E2E                | Built Vue + Go + PostgreSQL in real Chrome             | Operator Workflow Parity                         | Undeclared capabilities                                 |
| Deployment smoke           | Fresh Compose project and restart                      | Standalone startup/idempotence                   | Production upgrade/operations                           |

## Verification checkpoints

| Checkpoint               | Implementation state                                                                       | Exit condition                                                                                                            | Current status                                           |
| ------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| V0 repository baseline   | Complete deterministic gate, including non-mutating backend coverage                       | `make check` retained green on each exact relevant source digest                                                          | Entry source-v2 retained; active-source refresh required |
| V1 identity/RBAC         | Persisted users/groups/grants, sessions/tokens, CLI and focused tests                      | Pass I4's exact claim CI, retain its receipt, obtain owner review, then complete the remaining identity matrix            | I4 evidence claim conditional; parent open               |
| V2 profile behavior      | Shared 13-resource service, broad tests, bounded I1-I12 done; I13 tested candidate green | Gate I13 eight-path closeout/final two-path receipt through independent integrity, two pinned local and exact-SHA CI gates; trace remaining invariant, rollback and external-tier boundaries | I1-I12 bounded done; I13 known claim/receipt green, executor standing applied `2026-10-13T13:02:43Z`; in-progress, bounded done conditional/not effective; parent/profile open |
| V3 PostgreSQL/deployment | Typed tables, bootstrap/concurrency tests, Compose harness, and owner-accepted bounded V3-01/V3-02 evidence with green closeout receipts | Complete V3-03 through V3-05 and retain deployment loss-recovery artifacts | V3-01/V3-02 done; V3 parent open |
| V4 REST/gRPC             | Strict comparator/orchestrator and broad resource/parity suites                            | Complete T2 scenario report, then equivalent T3 report                                                                    | T1; implementation gap                                   |
| V5 Vue workflow          | Typed adapters and real-browser harness                                                    | Both workflows and every required negative/state case retained green                                                      | T1; implementation gap                                   |
| V6 sign-off              | First-profile legacy stacks retired                                                        | V0-V5 green together with linked evidence                                                                                 | Not earned                                               |

## Merge rules

- Tests own their process, database, clock, randomness, ports, and fixtures.
- Do not depend on an already-running developer service, execution order, or
  arbitrary sleeps as synchronization.
- Every defect fix adds the lowest-layer regression that proves it.
- A tier changes only when its required boundary passes and the evidence is
  retained. Compilation, route enumeration, generation, SQLite rollback, and
  one happy path cannot substitute for a stronger boundary.
- Coverage is a non-regression gate. Security and compatibility-critical code
  cannot be excluded to improve the number.
- External artifacts must contain no credentials, cookies, token secrets,
  password material, or DSNs.
