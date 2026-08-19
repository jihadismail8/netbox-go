# Core Workflow Implementation Plan

- Status: active execution plan
- Updated: 2026-08-03

This plan implements the first named [Capability Profile](../CONTEXT.md) for the standalone NetBox Go rewrite. It is ordered by dependency and exit evidence rather than time estimates. A listed file or phase is planned work, not implementation evidence; [Project status](STATUS.md) remains authoritative for what exists today.

The governing documents are [Architecture](ARCHITECTURE.md), [Compatibility](COMPATIBILITY.md), [Coding standards](CODING_STANDARDS.md), [Testing](TESTING.md), and the accepted [ADRs](adr/README.md).

The detailed executor contract, current typed-cutover recovery steps, stable
rule IDs, repeatable future-profile factory, and production-readiness sequence
are in the
[whole-project execution playbook](IMPLEMENTATION_EXECUTION_PLAYBOOK.md).
This document remains the authoritative behavioral plan for
`core-workflow-v1`; the playbook does not broaden its scope.

## Execution status

The phase sections below remain the required design and exit gates. This table
records present implementation without converting partial work into a passed
gate; [Project status](STATUS.md) contains the detailed evidence ledger.

| Phase                               | Present foundation                                                                                                                                                                                                                                                                  | Missing implementation and evidence before exit                                                                                                         |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0 — unsafe-runtime containment      | Legacy business surfaces, public configuration output, demo auth, and undeclared Vue routes are not registered. The canonical router has no dedicated `GET /ping` or `/ping` SPA bypass; frozen runtime-disabled ADR 0004 wiring remains source only.                               | Keep containment assertions green and retain them through the exact current V0.                                                                         |
| 1 — V0 quality baseline             | Deterministic backend/frontend/contract/documentation gates and an owned legacy-test exclusion exist. Coverage is a non-mutating part of `make check`; current post-wrapper-cleanup V0 is retained.                                                                                 | Keep V0 green on every behavioral, contract, test, security, or toolchain change.                                                                       |
| 2 — profile and canonical contracts | The 13-resource profile, resource/scenario metadata, OpenAPI, versioned protobufs, generated docs, and inventories exist. The contract remains pre-publication and T1.                                                                                                              | Add rule-to-test/evidence traceability; do not publish or promote without the required evidence.                                                        |
| 3 — PostgreSQL/bootstrap            | Missing-table-only bootstrap, a 198-entry persistence registry, 13 typed profile tables, typed object changes, focused constraints/concurrency tests, and a Compose harness exist.                                                                                                  | Implement dependency-aware HTTP/gRPC readiness, dependency loss/recovery, and missing locking cases; retain PostgreSQL/deployment artifacts.            |
| 4 — shared Site skeleton            | Shared transaction/authorization/change abstractions and typed per-resource application services exist. REST, gRPC, and parity fixtures use the same service; generic workflow packages are retired and prohibited.                                                                 | Trace and retain negative dual-adapter/PostgreSQL behavior for every declared rule.                                                                     |
| 5 — identity/RBAC                   | Users, groups, memberships, direct/group/global/object grants, sessions, tokens, REST/gRPC adapters, bearer auth, administrator CLI, retained trusted-origin CORS, accepted token lookup/touch and token-transport evidence, and bounded browser-session/CSRF evidence are present. | Review I3, then complete password-session, Origin/Referer/masking, throttle/proxy/streaming, RBAC, CLI, administration, and secret evidence.            |
| 6 — compatibility/parity harness    | Pinned-SHA/config refusal, a strict comparator, explicit normalizers, state checks, a sensitivity test, and broad resource/parity suites exist.                                                                                                                                     | Add every missing presence, permission, conflict, invariant, rollback, and side-effect scenario; then retain T2 and corresponding T3 reports.           |
| 7 — DCIM chain                      | Ten DCIM resources have typed shared services, persistence, REST/gRPC/Vue adapters, focused tests, and no remaining first-profile legacy stacks.                                                                                                                                    | Complete traceability and missing hierarchy, propagation, placement, uniqueness, template, permission, rollback, PostgreSQL, oracle, and browser cases. |
| 8 — IPAM/assignment                 | VRF, Prefix, IPAddress, assign/unassign, typed persistence/Vue adapters, normalization, concurrency tests, and focused rollback/parity cases exist.                                                                                                                                 | Complete traceability and IPv4/IPv6, hierarchy, presence, uniqueness, invalid-pair, reassignment, rollback, oracle, gRPC, and browser matrices.         |
| 9 — Vue workflows                   | Profile-only routes, typed DTO/form/filter adapters, and a real-Chrome/CDP clean-deployment harness exist.                                                                                                                                                                          | Add session refresh, edit/list/filter, late template rollback, explicit null, reassignment rollback, conflict, not-found, and exact-state cases.        |
| 10 — retirement/sign-off            | The 13 displaced legacy stacks were removed early; ADR 0005 separately removed 118 dormant handler/router wrapper pairs without changing inventory or tier.                                                                                                                         | Retain V0–V5 together, freeze the protobuf baseline, obtain V6 sign-off, and enforce ADR 0004/CP-13 for later retirement.                               |

This plan closes only the first profile. Project-wide accepted scope is 153
resource/action entries: 13 are T1 and 140 remain T0. After V6, the playbook
still contains 19 candidate profiles, 10 module-closeout passes, and seven
production-readiness programs. See [Project status](STATUS.md) for the exact
module and evidence ledger.

## Intended outcome

From a clean disposable PostgreSQL database, an administrator can use the Go-owned identity flow and Vue application to complete both workflows:

```text
Manufacturer -> DeviceType -> InterfaceTemplate
                                      |
Site -> Rack -> Device ---------------+-> Interface

VRF -> Prefix -> IPAddress -> assign to Interface -> unassign
```

Baseline-declared operations and scenarios will be available through exact NetBox-compatible HTTPS REST and a versioned gRPC API. Profile-declared security extensions will have documented REST/gRPC semantic parity but are not counted as baseline-compatible. Both adapters invoke the same application use cases and produce the same authorization, validation, transaction, change-log, and durable-state outcomes; cookie/CSRF mechanics remain transport-specific.

Passing this profile does not make all of DCIM or IPAM complete.

## Resolved implementation choices

| Topic                  | Recommended decision used by this plan                                                                                                    |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Delivery unit          | A named Core Workflow Profile with an explicit included and deferred surface                                                              |
| REST operations        | Collection `GET`/single-object `POST` and detail `GET`/`PUT`/`PATCH`/`DELETE`, limited to profile-declared fields, filters, and scenarios |
| gRPC operations        | Typed `List`, `Get`, `Create`, `Replace`, `Update`, and `Delete` RPCs grouped into module services                                        |
| Partial updates        | REST preserves omitted/null/value states; gRPC uses presence-aware fields and `FieldMask`                                                 |
| IP assignment          | REST uses baseline `PATCH` fields; gRPC exposes explicit assign/unassign RPCs; both reach one atomic use case                             |
| Device components      | Include DeviceType-backed `InterfaceTemplate` and instantiate Interfaces atomically during Device creation                                |
| Convenience operations | Defer bulk operations, rack elevation, and automatic available-prefix/address allocation                                                  |
| gRPC compatibility     | Freeze and retire the generated table RPCs; only new versioned contracts become public commitments                                        |
| Backend layout         | Layer-first `domain`, `application`, `adapters`, and `platform` packages                                                                  |
| Persistence            | Private PostgreSQL adapter rows; GORM is not the domain model                                                                             |
| Schema lifecycle       | On disposable databases, create absent tables only; never alter/backfill an existing table shape                                          |
| Security               | Fail closed, one Principal/RBAC path, session cookies for Vue, API tokens for REST, bearer metadata for gRPC                              |
| Caching                | No cache in the first shared path; add one only after a measured need and never for correctness                                           |
| Generated source       | Immutable transitional output; change owned definitions/generators and regenerate                                                         |
| Quality                | Establish a green self-contained V0 gate before feature work; legacy exclusions are explicit and expiring                                 |

## Core Workflow Profile

The committed machine-readable profile described in Phase 2 becomes authoritative for exact defaults, nullability, nested shapes, errors, filters, permissions, and side effects. The tables below fix the intended boundary that the profile must encode.

Stable behavioral rules in this section carry semantic `plan.*` labels. These
labels are traceability identifiers, not implementation or evidence claims.
Adding, removing, splitting, or changing an accepted behavioral rule requires
updating its machine-readable traceability row in the same increment. Resource
fields, filters, choices, relationships, and operations remain enumerated by
the linked profile/resource metadata rather than being copied into rule IDs.
The traceability validator independently pins reviewed row semantics and
applicability plus each reference set's exact capability ownership, pinned
source, upstream tests, and dimension-specific test inventory. Retained
verification must cite a canonical result artifact whose non-attestation
raw-byte hash is committed by the owned matrix. The complete artifact must be
valid UTF-8, its marker line must be ASCII, and no decoding or newline
normalization occurs before hashing. Its versioned strict RFC 8259 JSON
attestation is bound to the active profile, Compatibility Baseline,
classification and tier or extension axes, covered proof dimensions, a full
reachable Git `tested_revision`, versioned tested/current source digests, and
successful result. An exact claim digest covers the complete verification set
and evidence references plus every sorted consumer's full row, assessment,
applicability, and proof assignment/object; those claims cannot be rebound
behind an unchanged artifact.

The validator reconstructs tested owned source from the reachable commit. A
same-digest claim must equal that committed tree exactly. When source digests
differ, the owned verification set also commits the raw-byte SHA-256 of a
complete tested-source manifest which must match the Git tree entry for entry.
Versioned `source-v2` entries include file executable mode, byte size, content
hash, and closed relative symlinks to independently hashed owned content. The
validator compares full tested/current path and kind sets, derives every mode
or content change, and requires the attestation's claim manifest to match that
derived claim-only diff exactly. Added, deleted, type-changed, or symlink-
retargeted entries require a new evidence run. These checks establish source
and claim integrity; they do not prove a command ran or its result/environment
was truthful. Durable runner provenance and human evidence review remain
mandatory, including semantic review of every mechanically allowed claim-only
change. A row's primary `capability` is its sole tier owner;
`affected_capabilities` records dependencies and observable effects without
granting or capping a secondary capability's tier. A source file, test file,
evidence ledger, unrelated historical pass, or self-asserted digest is never
retained boundary evidence by itself.

- **Rule `plan.profile.surface.manifest-exact`.** The supported runtime surface equals the validated profile plus explicitly classified system and identity entries.
- **Rule `plan.profile.surface.deferred-rejected-absent`.** A deferred route, RPC, field, filter, relationship, or UI capability is rejected or absent rather than silently accepted.
- **Rule `plan.profile.transport.shared-application-use-case`.** REST and corresponding gRPC operations invoke the same typed application use case.
- **Rule `plan.profile.transport.central-error-mapping`.** Transport adapters map the shared stable error taxonomy without exposing internal causes.
- **Rule `plan.profile.oracle.pin-sha-configuration`.** Differential evidence is valid only against the exact pinned source SHA and effective oracle configuration. The oracle source is materialized by replacement- and graft-disabled direct committed-tree/blob reads which bypass repository-local export attributes; live untracked, ignored, modified, Git-metadata-substituted, or attribute-filtered content is never mounted.
- **Rule `plan.profile.mutation.single-transaction`.** One public mutation owns one application-level transaction including authorization, validation, persistence, derived state, object changes, and required durable event intent.
- **Rule `plan.profile.authorization.authenticate-default`.** Every Managed Object read and mutation authenticates by default unless the profile explicitly classifies a public system/credential endpoint.
- **Rule `plan.profile.authorization.permission-parity`.** REST and gRPC resolve the same Principal and view/add/change/delete/object-visibility decisions.
- **Rule `plan.profile.evidence.no-tier-without-boundary`.** A compatibility tier changes only after its required current external boundary is retained and reviewed.

### Included resources and paths

**Rule `plan.common.resource.six-single-object-operations`.** Each included resource exposes the declared `list`, `get`, `create`, `replace`, `update`, and `delete` operations as independently traceable operations. The resource remains the tier-owning capability.

| Module | Resource          | Baseline REST path               |
| ------ | ----------------- | -------------------------------- |
| DCIM   | Site              | `/api/dcim/sites/`               |
| DCIM   | Manufacturer      | `/api/dcim/manufacturers/`       |
| DCIM   | RackRole          | `/api/dcim/rack-roles/`          |
| DCIM   | RackType          | `/api/dcim/rack-types/`          |
| DCIM   | Rack              | `/api/dcim/racks/`               |
| DCIM   | DeviceRole        | `/api/dcim/device-roles/`        |
| DCIM   | DeviceType        | `/api/dcim/device-types/`        |
| DCIM   | InterfaceTemplate | `/api/dcim/interface-templates/` |
| DCIM   | Device            | `/api/dcim/devices/`             |
| DCIM   | Interface         | `/api/dcim/interfaces/`          |
| IPAM   | VRF               | `/api/ipam/vrfs/`                |
| IPAM   | Prefix            | `/api/ipam/prefixes/`            |
| IPAM   | IPAddress         | `/api/ipam/ip-addresses/`        |

**Rule `plan.common.response.standard-envelope`.** All resources include the applicable baseline response envelope (`id`, URLs, display fields, `created`, and `last_updated`).

**Rule `plan.common.response.nested-relationships`.** Responses include the nested/brief representations needed by included relationships.

**Rule `plan.common.response.safe-pagination`.** Collection responses use safe pagination.

**Rule `plan.common.response.object-visibility`.** Read responses apply object visibility.

**Rule `plan.common.response.read-only-counts`.** Responses include only the relevant declared read-only counts.

**Rule `plan.common.response.reject-read-only-writes`.** Response fields are never accepted as writable merely because they are returned.

**Rule `plan.common.collection.single-object-post`.** Collection `POST` accepts exactly one JSON object.

**Rule `plan.common.collection.reject-bulk`.** Array payloads and every other bulk mutation shape are deferred.

The response-only projection is fixed as follows; all other serializer counters are deferred:

| Resource     | Included response-only fields                                                                                 |
| ------------ | ------------------------------------------------------------------------------------------------------------- |
| Site         | `device_count`, `prefix_count`, `rack_count` (`prefix_count` remains zero while scoped Prefixes are deferred) |
| Manufacturer | `devicetype_count`                                                                                            |
| RackRole     | `rack_count`                                                                                                  |
| Rack         | `device_count`                                                                                                |
| DeviceRole   | `device_count`, `_depth`                                                                                      |
| DeviceType   | `device_count`, `interface_template_count`                                                                    |
| Device       | `interface_count`                                                                                             |
| Interface    | `count_ipaddresses`                                                                                           |
| VRF          | `ipaddress_count`, `prefix_count`                                                                             |
| Prefix       | `family`, `children`, `_depth`                                                                                |
| IPAddress    | `family`                                                                                                      |

### Identity and system contract

**Rule `plan.identity.interfaces.explicit-contract`.** Identity is a prerequisite contract, not an implied middleware detail. The profile declares:

| Interface                             | Included operations                                                                                                                                                                                                                 |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Browser/identity REST extension       | `GET /api/auth/session/`, `GET /api/auth/csrf/`, `POST /api/auth/login/`, `POST /api/auth/logout/`, `POST /api/auth/password/change/`, plus authenticated `GET`/`POST /api/auth/tokens/` and `DELETE /api/auth/tokens/{id}/`        |
| NetBox-compatible REST authentication | `Authorization: Token <key>` for automation, including write-enabled, expiry, allowed-IP restrictions, and throttled `last_used` persistence declared by the profile                                                                |
| gRPC `IdentityService`                | `GetCurrentUser`, `ListAPITokens`, `CreateAPIToken`, `RevokeAPIToken`, `ChangePassword` using bearer metadata where authentication is required                                                                                      |
| Administrator CLI                     | One-time bootstrap, protected `reset-password`, authenticated `create-user`, and authenticated `grant-permission`; passwords stay on protected stdin and no REST/RPC bootstrap, recovery, or anonymous provisioning endpoint exists |

The identity contract above decomposes into these independently traceable
security assertions:

- **Rule `plan.identity.session.csrf-bootstrap-cookie-body`.** CSRF bootstrap returns the documented readable SameSite cookie and matching response-body value.
- **Rule `plan.identity.session.current-principal`.** The authenticated session endpoint returns the current Principal without exposing reusable credential material.
- **Rule `plan.identity.session.login-rotation`.** Login rotates any presented session identifier and does not accept caller-chosen session material.
- **Rule `plan.identity.session.hardened-cookie-flags`.** Browser session cookies use the documented hardened flags.
- **Rule `plan.identity.session.mutation-requires-csrf`.** Cookie-authenticated mutations require the matching CSRF cookie/header pair.
- **Rule `plan.identity.session.logout-revokes`.** Logout revokes the current browser session.
- **Rule `plan.identity.session.expiry-fails-closed`.** Expired browser sessions fail closed at the authentication boundary.
- **Rule `plan.identity.token.owner-only-management`.** A principal may list and revoke only its own API tokens.
- **Rule `plan.identity.token.rest-token-scheme`.** REST automation accepts only the documented `Authorization: Token` credential path.
- **Rule `plan.identity.token.grpc-bearer`.** gRPC accepts token credentials only through bearer metadata.
- **Rule `plan.identity.token.fail-closed-conditions`.** Unknown, expired, revoked, write-disabled, disallowed-IP, and inactive-user token conditions fail closed as applicable.
- **Rule `plan.identity.token.create-throttled`.** Token creation uses the declared bounded throttle without logging submitted credential material.
- **Rule `plan.identity.token.revoke-immediate`.** A revoked token is rejected on its next use.
- **Rule `plan.identity.password.change-invalidates-sessions`.** Password change verifies the current credential and invalidates sessions according to the declared policy.
- **Rule `plan.identity.rbac.direct-group-global-object`.** Effective authorization combines direct, group, global-model, and object-scoped grants.
- **Rule `plan.identity.rbac.visibility-before-pagination`.** Object visibility is applied before count, ordering, and pagination.
- **Rule `plan.identity.rbac.authorization-before-write`.** Mutation authorization completes before persistence begins.
- **Rule `plan.identity.admin.bootstrap-empty-store-once`.** Administrator bootstrap succeeds only against an empty identity store and refuses later calls.
- **Rule `plan.identity.admin.reset-local-recovery-only`.** Password reset is a protected local administrator recovery boundary, not a network endpoint.
- **Rule `plan.identity.admin.create-user-active-superuser`.** Local non-superuser creation requires an existing active superuser.
- **Rule `plan.identity.admin.grant-global-model-permission`.** Local permission administration grants only the declared global model permission by username.
- **Rule `plan.identity.admin.password-protected-stdin`.** Every administrator password is read from protected stdin and stays out of argv, output, logs, and evidence.
- **Rule `plan.identity.admin.no-network-provisioning`.** No anonymous HTTP or gRPC bootstrap, reset, user-creation, permission-grant, or token-provisioning surface exists.

**Rule `plan.identity.browser.session-csrf`.** Browser session creation/revocation and CSRF are transport-specific credential mechanics.

**Rule `plan.identity.grpc.no-cookie-equivalent`.** Those browser mechanics do not require a cookie-shaped gRPC equivalent.

**Rule `plan.identity.admin.defer-network-management`.** Network-exposed user/group/permission administration is deferred to a later identity profile.

**Rule `plan.identity.cli.protected-administration`.** The protected local CLI may create a non-superuser and grant one global model permission by username for development, deployment, and recovery workflows.

**Rule `plan.identity.cli.defer-group-object-grants`.** Group administration and object-scoped CLI grants remain deferred.

**Rule `plan.identity.authorization.shared-rbac`.** Shared RBAC enforcement is not deferred.

**Rule `plan.identity.token.authenticated-extension`.** The authenticated `/api/auth/tokens/` management surface is an intentional security extension paired with `IdentityService`.

**Rule `plan.identity.token.one-time-secret`.** A server-generated secret is returned once at creation and never by list/get.

**Rule `plan.identity.token.defer-baseline-crud`.** The pinned baseline's `/api/users/tokens/` serializer hides a generated key under the default `ALLOW_TOKEN_RETRIEVAL=false`; baseline token CRUD remains deferred.

**Rule `plan.identity.token.reject-anonymous-provision`.** The baseline `/api/users/tokens/provision/` action accepts a username and password anonymously, and this profile explicitly rejects that action.

**Rule `plan.identity.token.extension-not-t2`.** The divergence is recorded in the manifest and is never normalized into a T2 compatibility pass.

**Rule `plan.identity.token.credential-use-baseline`.** Only use of a resulting token as a REST credential is claimed baseline-compatible.

**Rule `plan.identity.token.last-used-rate`.** For a recognized token key, the baseline persists `last_used` at most once per minute.

**Rule `plan.identity.token.last-used-before-restrictions`.** The recognized-key update occurs before expiry, inactive-user, or allowed-IP rejection.

**Rule `plan.identity.token.unknown-key-no-write`.** An unknown key produces no write.

**Rule `plan.identity.token.maintenance-off`.** The profile pins maintenance mode off.

**Rule `plan.identity.token.shared-authentication-path`.** The shared identity path, rather than transport middleware, preserves token ordering semantics.

### Writable field boundary

**Rule `plan.common.fields.presence-states`.** When the contract distinguishes them, commands preserve absent, explicit null, zero, empty, and concrete values without collapsing one state into another.

**Rule `plan.common.fields.operation-specific-nullability`.** Create, PUT, and PATCH enforce the pinned field-specific required, blank, and null semantics for that operation rather than inferring writability or nullability from a response schema.

| Resource          | Included writable fields                                                                                                                                                                     | Explicitly deferred fields/relationships                                                                                                     |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Site              | `name`, `slug`, `status`, `facility`, `description`, `comments`                                                                                                                              | region, site group, tenant, ASNs, time zone, addresses, coordinates, tags, custom fields                                                     |
| Manufacturer      | `name`, `slug`, `description`                                                                                                                                                                | tags, custom fields, unrelated inventory/platform behavior                                                                                   |
| RackRole          | `name`, `slug`, `color`, `description`                                                                                                                                                       | tags, custom fields                                                                                                                          |
| RackType          | `manufacturer`, `model`, `slug`, `form_factor`, `width`, `u_height`, `starting_unit`, `desc_units`, `description`, `comments`                                                                | weight, outer dimensions/unit, mounting depth, tags, custom fields                                                                           |
| Rack              | `site`, `name`, `facility_id`, `rack_type`, `status`, `role`, `serial`, `asset_tag`, `form_factor`, `width`, `u_height`, `starting_unit`, `desc_units`, `airflow`, `description`, `comments` | location, tenant, reservations, weight/outer dimensions, tags, custom fields                                                                 |
| DeviceRole        | `parent`, `name`, `slug`, `color`, `vm_role`, `description`, `comments`                                                                                                                      | config template, tags, custom fields, virtualization behavior beyond the scalar `vm_role` choice                                             |
| DeviceType        | `manufacturer`, `model`, `slug`, `part_number`, `u_height`, `exclude_from_utilization`, `is_full_depth`, `airflow`, `description`, `comments`                                                | default platform, subdevice role, images, weight profile, other component templates, tags, custom fields                                     |
| InterfaceTemplate | `device_type`, `name`, `label`, `type`, `enabled`, `mgmt_only`, `description`                                                                                                                | module type, bridge, PoE, RF/wireless fields, all non-interface template families                                                            |
| Device            | `device_type`, `role`, `name`, `site`, `rack`, `position`, `face`, `status`, `serial`, `asset_tag`, `airflow`, `description`, `comments`                                                     | tenant, platform, location, cluster, virtual chassis/child devices, primary/OOB IP designation, config context/template, tags, custom fields |
| Interface         | `device`, `name`, `label`, `type`, `enabled`, `mgmt_only`, `mtu`, `speed`, `duplex`, `description`                                                                                           | module/VDC ownership, parent/bridge/LAG, VRF binding, MAC management, VLAN/QinQ, cable, wireless, PoE/RF, tags, custom fields                |
| VRF               | `name`, `rd`, `enforce_unique`, `description`, `comments`                                                                                                                                    | tenant, import/export route targets, tags, custom fields                                                                                     |
| Prefix            | `prefix`, `vrf`, `status`, `is_pool`, `mark_utilized`, `description`, `comments`                                                                                                             | scope, tenant, VLAN, role, automatic allocation, tags, custom fields                                                                         |
| IPAddress         | `address`, `vrf`, `status`, `role`, `dns_name`, `description`, `comments`, Interface assignment fields                                                                                       | tenant, NAT, primary/OOB designation, VM/FHRP assignment, tags, custom fields                                                                |

**Rule `plan.common.fields.reject-undeclared`.** An omitted field is deferred, not silently ignored. REST rejects or reports it according to the committed profile rather than writing arbitrary columns.

**Rule `plan.common.fields.promote-vertically`.** Promoting a deferred field requires its domain rules, both adapters, PostgreSQL behavior, and tests in one reviewed increment.

### Required list behavior

**Rule `plan.common.list.id-filter`.** Every collection supports verified baseline behavior for `id`.

**Rule `plan.common.list.search`.** Every collection supports search (`q`) where the baseline provides it.

**Rule `plan.common.list.bounded-pagination`.** Every collection supports bounded `limit` and `offset` behavior.

**Rule `plan.common.list.ordering-allowlist`.** Every collection uses an explicit ordering allowlist.

The first profile also includes these filter groups:

| Resource          | Required filters                                                                                                |
| ----------------- | --------------------------------------------------------------------------------------------------------------- |
| Site              | name, slug, status                                                                                              |
| Manufacturer      | name, slug                                                                                                      |
| RackRole          | name, slug                                                                                                      |
| RackType          | manufacturer ID/slug, model, slug                                                                               |
| Rack              | site ID/slug, name, status, role ID/slug, rack-type ID/slug                                                     |
| DeviceRole        | name, slug                                                                                                      |
| DeviceType        | manufacturer ID/slug, model, slug                                                                               |
| InterfaceTemplate | device-type ID, name, type, enabled, management-only                                                            |
| Device            | site ID/slug, rack ID, device-type ID/slug, role ID/slug, name, status                                          |
| Interface         | device ID/name, name, type, enabled, management-only                                                            |
| VRF               | name, route distinguisher, enforce-unique                                                                       |
| Prefix            | VRF ID/route distinguisher, prefix, family, status, `within`, `within_include`, `contains`                      |
| IPAddress         | VRF ID/route distinguisher, address, family, parent prefix, status, assigned state, Interface/device assignment |

**Rule `plan.common.list.oracle-filter-semantics`.** Filter spellings and edge semantics are copied from the pinned oracle into the profile and differential scenarios.

**Rule `plan.common.list.grpc-typed-fields`.** gRPC list requests use typed fields, never generic table columns/operators.

### Assignment semantics

REST preserves the baseline contract: assignment occurs through `PATCH /api/ipam/ip-addresses/{id}/` with `assigned_object_type` and `assigned_object_id`. The adapter applies field presence to the current assignment and validates the resulting pair:

- **Rule `plan.ipam.assignment.both-null-unassign`.** Both explicitly `null`: unassign.
- **Rule `plan.ipam.assignment.both-omitted-preserve`.** Both omitted: preserve the current assignment.
- **Rule `plan.ipam.assignment.unassigned-requires-pair`.** Assigning an unassigned IPAddress requires a complete non-null type/ID pair.
- **Rule `plan.ipam.assignment.partial-reassignment`.** For an already assigned IPAddress, an ID-only patch retains the existing content type and may reassign it to another Interface; a type-only patch retains the current ID if the resulting Interface target is valid.
- **Rule `plan.ipam.assignment.invalid-pairs-reject`.** One null and one non-null value, a missing target, or a non-Interface type is a validation error.
- **Rule `plan.ipam.assignment.atomic-update`.** Assignment plus other IPAddress changes is one atomic application update, not two commits.

**Rule `plan.ipam.assignment.grpc-actions`.** gRPC exposes `AssignIPAddress` and `UnassignIPAddress`.

**Rule `plan.ipam.assignment.grpc-update-presence`.** `UpdateIPAddress` may express the same transition with presence-aware assignment fields.

**Rule `plan.ipam.assignment.shared-transport-use-case`.** REST PATCH and all gRPC assignment forms invoke the same application operation.

**Rule `plan.ipam.assignment.interface-only-target`.** The first profile permits only `dcim.interface` targets.

### InterfaceTemplate semantics

`CreateDevice` owns instantiation; there is no public “instantiate interfaces” action.

1. **Rule `plan.dcim.device-create.authorize-load`.** Authorize and load the DeviceType and related objects.
2. **Rule `plan.dcim.device-create.validate-placement`.** Validate Device state and rack placement.
3. **Rule `plan.dcim.device-create.snapshot-templates`.** Snapshot the DeviceType's InterfaceTemplates.
4. **Rule `plan.dcim.device-create.persist-device`.** Persist the Device.
5. **Rule `plan.dcim.device-create.copy-template-fields`.** Create one Interface per template, copying `name`, `label`, `type`, `enabled`, and `mgmt_only`.
6. **Rule `plan.dcim.device-create.record-changes`.** Record required Device and Interface changes.
7. **Rule `plan.dcim.device-create.atomic-commit`.** Commit everything, or roll everything back.

**Rule `plan.dcim.device-create.omit-template-description`.** The baseline does not copy the template description to the Interface.

**Rule `plan.dcim.interface-template.non-retroactive`.** Template changes are not retroactive.

**Rule `plan.dcim.interface.independently-editable`.** Existing Interfaces remain independently editable.

**Rule `plan.dcim.device-create.defer-other-template-families`.** Other template families remain deferred even though the full baseline instantiates them.

**Rule `plan.dcim.interface-template.bridge-deferred`.** Bridge-template creation and post-instantiation bridge resolution are named deferred scenarios, not silently ignored behavior.

**Rule `plan.dcim.device-create.bridge-free-scope`.** The Device-create compatibility claim is limited to DeviceTypes whose InterfaceTemplates have no bridge.

### Deferred operations

- **Rule `plan.deferred.bulk-mutations`.** Bulk create, edit, rename, import, and delete.
- **Rule `plan.deferred.rack-elevation`.** Rack elevation JSON and SVG actions and the upstream rack elevation UI.
- **Rule `plan.deferred.automatic-ipam-allocation`.** Prefix `available-prefixes` and `available-ips`, next-available address selection, and automatic allocation RPCs.
- **Rule `plan.deferred.baseline-token-management`.** Baseline `/api/users/tokens/` CRUD and anonymous `/api/users/tokens/provision/`; the secure authenticated identity extension above is used instead.
- **Rule `plan.deferred.physical-cabling`.** Cable tracing/connections and physical cabling.
- **Rule `plan.deferred.generic-query-operators`.** Arbitrary lookup operators and generic SQL-shaped filters.
- **Rule `plan.deferred.unbounded-results`.** Unbounded result sets.
- **Rule `plan.deferred.graphql`.** GraphQL.
- **Rule `plan.deferred.python-runtime`.** Python scripts/reports and in-process Python plugins.

**Rule `plan.dcim.rack.manual-position`.** Manual rack position input remains included, with full validation and conflict handling.

**Rule `plan.ipam.prefix.manual-input`.** Manual prefix input remains included, with full validation and conflict handling.

**Rule `plan.ipam.ip-address.manual-input`.** Manual IP address input remains included, with full validation and conflict handling.

### Baseline invariants that must not be simplified

These rules are easy to “clean up” into incompatible behavior and therefore receive explicit scenarios before implementation:

- **Rule `plan.ipam.prefix.network-derived-containment`.** Prefix containment is derived from network values within a VRF.
- **Rule `plan.ipam.ip-address.no-prefix-id`.** IPAddress has no `prefix_id`.
- **Rule `plan.ipam.ip-address.no-containing-prefix-required`.** IPAddress may exist without a containing Prefix.
- **Rule `plan.ipam.ip-address.preserve-host-mask`.** IPAddress preserves its host address plus mask.
- **Rule `plan.ipam.prefix.reject-host-bits-with-suggestion`.** Prefix input with host bits is rejected with the canonical network suggestion.
- **Rule `plan.ipam.prefix.store-canonical-network`.** Accepted Prefixes are stored canonically.
- **Rule `plan.ipam.interface-vrf.deferred`.** Interface VRF is deferred.
- **Rule `plan.ipam.assignment.no-interface-vrf-equality`.** IPAddress assignment must not invent a VRF-equality rule; when Interface VRF is promoted later, it remains independent from IPAddress VRF as in the baseline.
- **Rule `plan.ipam.uniqueness.global-default`.** The oracle profile fixes `ENFORCE_GLOBAL_UNIQUE=true`, matching the pinned baseline default.
- **Rule `plan.ipam.uniqueness.vrf-enforce-unique`.** VRF `enforce_unique` affects Prefix and IPAddress uniqueness.
- **Rule `plan.ipam.ip-address.uniqueness-host-not-mask`.** IPAddress uniqueness compares hosts rather than masks.
- **Rule `plan.ipam.ip-address.duplicate-role-exceptions`.** IPAddress uniqueness preserves the baseline role exceptions.
- **Rule `plan.ipam.prefix.reject-slash-zero`.** Prefix rejects `/0`.
- **Rule `plan.ipam.ip-address.reject-slash-zero`.** IPAddress rejects `/0`.
- **Rule `plan.ipam.assignment.reject-network-broadcast`.** IP assignment rejects ordinary network and broadcast addresses.
- **Rule `plan.ipam.assignment.edge-prefix-exceptions`.** IP assignment preserves the baseline IPv4 `/31`/`/32` and IPv6 `/127`/`/128` exceptions.
- **Rule `plan.ipam.ip-address.slaac-ipv6-only`.** SLAAC is IPv6-only.
- **Rule `plan.ipam.ip-address.dns-lowercase`.** DNS names are normalized to lowercase.
- **Rule `plan.dcim.site.name-global`.** Site name is globally unique.
- **Rule `plan.dcim.site.slug-global`.** Site slug is globally unique.
- **Rule `plan.dcim.manufacturer.name-global`.** Manufacturer name is globally unique.
- **Rule `plan.dcim.manufacturer.slug-global`.** Manufacturer slug is globally unique.
- **Rule `plan.dcim.rack-role.name-global`.** RackRole name is globally unique.
- **Rule `plan.dcim.rack-role.slug-global`.** RackRole slug is globally unique.
- **Rule `plan.dcim.rack-type.slug-global`.** RackType slug is globally unique.
- **Rule `plan.dcim.rack-type.manufacturer-model-unique`.** RackType `(manufacturer, model)` is unique.
- **Rule `plan.dcim.device-type.manufacturer-model-unique`.** DeviceType model is unique within its Manufacturer.
- **Rule `plan.dcim.device-type.manufacturer-slug-unique`.** DeviceType slug is unique within its Manufacturer.
- **Rule `plan.ipam.vrf.rd-global-unique`.** A non-null VRF route distinguisher is globally unique.
- **Rule `plan.ipam.vrf.name-repeatable`.** VRF name is not globally unique.
- **Rule `plan.dcim.rack.name-location-scoped`.** Rack name uniqueness is scoped to Location, not Site.
- **Rule `plan.dcim.rack.facility-location-scoped`.** Rack facility-ID uniqueness is scoped to Location, not Site.
- **Rule `plan.dcim.rack.null-location-duplicates`.** Because Location is deferred and null in this profile, duplicate Rack names or facility IDs in one Site must not be rejected unless the pinned baseline rejects the exact scenario.
- **Rule `plan.dcim.rack.direct-save-full-validation`.** Assigning or directly updating a Rack runs Rack validation.
- **Rule `plan.dcim.rack.direct-save-mounted-device-protection`.** Direct Rack changes protect mounted Devices.
- **Rule `plan.dcim.rack-type.propagates-rack-physical-fields`.** RackType updates propagate owned physical attributes by saving each Rack.
- **Rule `plan.dcim.rack-type.propagation-skips-rack-full-validation`.** RackType propagation does not run the same full Rack validation; preserve that behavior provisionally and lock it with a differential oracle scenario rather than silently strengthening REST compatibility.
- **Rule `plan.dcim.device-role.reject-self-descendant-cycle`.** DeviceRole hierarchy rejects self and descendant cycles.
- **Rule `plan.dcim.device-role.name-sibling-root-unique`.** DeviceRole name is unique among siblings and separately at the top level.
- **Rule `plan.dcim.device-role.slug-sibling-root-unique`.** DeviceRole slug is unique among siblings and separately at the top level.
- **Rule `plan.dcim.device-role.delete-descendant-cascade`.** DeviceRole parent deletion cascades through descendants.
- **Rule `plan.dcim.device-role.delete-device-protection`.** A Device `PROTECT` reference blocks DeviceRole deletion.
- **Rule `plan.dcim.device-type.height-half-unit`.** DeviceType `u_height` is zero or a multiple of 0.5.
- **Rule `plan.dcim.device-type.height-increase-fit`.** Increasing DeviceType height must fit every positioned instance.
- **Rule `plan.dcim.device-type.height-zero-position-protection`.** Changing DeviceType height to zero is rejected while any instance remains positioned.
- **Rule `plan.ipam.prefix.is-pool-contract`.** Prefix `is_pool` retains its baseline default, validation, serialization, and effect on included projections.
- **Rule `plan.ipam.prefix.mark-utilized-contract`.** Prefix `mark_utilized` retains its baseline default, validation, serialization, and effect on included projections.
- **Rule `plan.ipam.prefix.flags-no-auto-allocation`.** Utilization flags do not promote automatic allocation actions.
- **Rule `plan.dcim.device.name-nullable`.** Device name is nullable.
- **Rule `plan.dcim.device.name-case-insensitive-site-unique`.** In this tenantless profile, a non-null Device name is case-insensitively unique per Site.
- **Rule `plan.dcim.device.rack-site-consistency`.** Rack must belong to Device Site.
- **Rule `plan.dcim.device.face-requires-rack`.** Device face requires a Rack.
- **Rule `plan.dcim.device.position-requires-rack`.** Device position requires a Rack.
- **Rule `plan.dcim.device.position-requires-face`.** Device position requires a face.
- **Rule `plan.dcim.device.position-range`.** Device position is between 1.0 and 100.5.
- **Rule `plan.dcim.device.position-half-unit`.** Device position uses 0.5-unit increments.
- **Rule `plan.dcim.device.position-rack-bounds`.** Device placement must fit Rack bounds.
- **Rule `plan.dcim.device.position-occupancy`.** Device placement must not overlap occupied Rack space.
- **Rule `plan.dcim.device.zero-u-unpositioned`.** A zero-U DeviceType cannot be positioned.
- **Rule `plan.dcim.device.airflow-create-inheritance`.** Device inherits DeviceType airflow when blank at creation.
- **Rule `plan.dcim.device-create.interfaces-and-changes-atomic`.** Device creation, instantiated Interfaces, and required changes are atomic.
- **Rule `plan.dcim.interface-template.edits-do-not-modify-interfaces`.** Later InterfaceTemplate edits do not modify existing Interfaces.
- **Rule `plan.dcim.interface.name-device-unique`.** Interface name is unique per Device.
- **Rule `plan.dcim.interface.owner-immutable`.** An Interface cannot move to another Device.
- **Rule `plan.dcim.interface-template.name-device-type-unique`.** InterfaceTemplate name is unique per DeviceType.
- **Rule `plan.dcim.interface-template.owner-immutable`.** An InterfaceTemplate cannot move to another DeviceType.
- **Rule `plan.dcim.delete.rack-protected`.** Rack deletion is protected while any Device references it.
- **Rule `plan.dcim.delete.site-protected`.** Referenced Site deletion is protected.
- **Rule `plan.dcim.delete.rack-type-protected`.** Referenced RackType deletion is protected.
- **Rule `plan.dcim.delete.rack-role-protected`.** Referenced RackRole deletion is protected.
- **Rule `plan.dcim.delete.manufacturer-protected`.** Referenced Manufacturer deletion is protected.
- **Rule `plan.dcim.delete.device-type-protected`.** Referenced DeviceType deletion is protected.
- **Rule `plan.dcim.delete.device-role-protected`.** Referenced DeviceRole deletion is protected.
- **Rule `plan.ipam.delete.vrf-protected`.** VRF deletion is protected by included Prefix and IPAddress references.
- **Rule `plan.ipam.interface-vrf.future-set-null`.** A future Interface VRF reference uses `SET_NULL`.
- **Rule `plan.dcim.delete.device-cascades-interfaces`.** Device deletion cascades its Interfaces.
- **Rule `plan.dcim.delete.device-type-cascades-templates`.** DeviceType deletion cascades its templates.
- **Rule `plan.dcim.delete.interface-cascades-ip-addresses`.** The baseline Interface `GenericRelation` cascades assigned IPAddresses when an Interface is deleted.
- **Rule `plan.dcim.delete.interface-cascade-change-effects`.** Preserve and prove the resulting IPAddress and object-change effects with an oracle scenario.

## Target source layout

```text
contracts/netbox/v4.4.6-post7/
  baseline.yaml
  oracle-profile.yaml
  inventory/{baseline-rest,current-rest,current-grpc,current-vue}.yaml
  profiles/core-workflow-v1.yaml
  resources/{identity,dcim,ipam}.yaml
  scenarios/
  fixtures/
  normalizers.yaml
  schema/profile.schema.json

netbox-backend/
  api/proto/netbox/{types,identity,dcim,ipam}/v1/
  api/openapi/netbox-go-v1.yaml
  gen/go/netbox/{types,identity,dcim,ipam}/v1/
  internal/domain/{shared,identity,dcim,ipam}/
  internal/application/{transaction,presence,identity,authz,changelog,dcim,ipam}/
  internal/adapters/postgres/{bootstrap,identity,changelog,dcim,ipam}/
  internal/adapters/postgres/{dcim,ipam}/row/
  internal/adapters/rest/netbox/{httpapi,identity,dcim,ipam}/
  internal/adapters/grpc/{statusmap,identity,dcim,ipam}/
  internal/platform/{config,logging,server}/
  internal/testkit/
  test/{integration,contract,parity}/

netbox-frontend/src/
  api/{http,errors,session}.ts
  features/identity/
  features/dcim/
  features/ipam/

scripts/
  validate_contract_inventory.mjs
  validate_capability_profile.mjs
  generate_contract_docs.mjs
  check_markdown_links.mjs
```

Handwritten protobuf sources live under `api/proto`; generated Go is isolated under `gen/go`. The legacy `api/netbox_go/v1` tree is frozen until retired. The capability manifest may validate routes and generate inventory documentation, but it never generates business handlers, repositories, or domain rules.

## Execution phases

### Phase 0 — contain the unsafe runtime

Purpose: eliminate accidental exposure before building new features.

Primary changes:

- `netbox-backend/internal/handler/auth.go`: reject absent/invalid credentials by default; the actual broad bypass is its no-Authorization-header branch.
- `netbox-backend/internal/routers/routers.go`: disable the entire direct-GORM generic Managed Object registry and bespoke token-management routes by default, remove the public `/config` response, and restrict `/codes`, `/metrics`, and any profiling surface to a non-production authenticated/admin listener or disable them.
- `netbox-backend/internal/adapters/rest/netbox/router/router.go`: remove the
  redundant public `/ping` route and SPA bypass; health and readiness are the
  only public process probes.
- `netbox-backend/internal/server/grpc.go`: do not register generated Managed Object services until shared bearer authentication exists; remove hard-coded credentials.
- `netbox-backend/internal/server/grpc.go`: remove the auxiliary public `/config` response and restrict `/codes`, metrics, and profiling to a non-production authenticated/admin listener or disable that listener.
- `netbox-backend/cmd/netbox_go/initial/initApp.go`: stop logging complete configuration values.
- `netbox-frontend/src/router/index.ts`, `netbox-frontend/src/stores/auth.ts`, and `netbox-frontend/src/pages/auth/LoginPage.vue`: remove demo auto-login, dummy tokens, and credential storage.

Only health, readiness, login, and CSRF endpoints may be public where their protocol requires it. Administrator bootstrap is CLI-only and no provisioning endpoint is public. It is acceptable for unfinished functionality to become unavailable; it is not acceptable for it to remain anonymously mutable.

The root composition/security files named above are hand-owned transitional
wiring under ADR 0004. Per-resource generated registries, services, and other
retained outputs remain immutable. ADR 0005 later authorized removal, not
editing, of exactly 118 untouched, runtime-dormant Sponge handler delegates and
their 118 matching top-level router wrappers. It retained `auth.go`, canonical
adapter handlers, hand-owned router code, the separate direct-GORM DRF
registry, and every lower generated stack. If ownership discovery identifies a
retained named file as reproducible output, make the containment change in its
owned source/template and regenerate instead of patching the output.

Exit evidence:

- Anonymous REST reads and mutations fail closed, and no direct-GORM business route is registered in the default runtime.
- Generated gRPC business services are not registered and therefore are unavailable to unauthenticated callers.
- `/config` is absent on both listeners; `/codes`, metrics, and profiling are either disabled or unreachable from the public listener without the documented admin authentication.
- The canonical runtime route inventory contains no dedicated `GET /ping`,
  and `/ping` is not an SPA bypass; health/readiness remain narrow public
  probes. Frozen runtime-disabled ADR 0004 source is not a canonical route.
- User responses cannot expose password hashes or privilege internals.
- Opening Vue cannot silently create an administrator session.

### Phase 1 — establish verification checkpoint V0

Purpose: make one default, non-mutating, self-contained quality gate green before feature work.

Primary changes:

- Add a root `Makefile` orchestrating backend, protobuf, frontend, generated-output, and documentation checks.
- Replace mutating `ci-lint` behavior in `netbox-backend/Makefile` with `fmt`, `fmt-check`, `vet`, `lint`, `unit`, `build`, and `check` targets.
- Add a non-mutating backend coverage target that writes its temporary profile
  outside the source tree, establishes a reviewed baseline, enforces
  non-regression, and is included by `check`.
- Add a pinned `netbox-backend/.golangci.yml`; initially exclude only identified legacy generated paths, never new handwritten packages.
- Classify current generated live-client tests through their owned generator/template. Until regenerated, use a visibly named compile-only legacy target and an explicit package exclusion; do not hand-fix hundreds of outputs or report them as behavioral tests.
- Add `scripts/check_markdown_links.mjs` and a non-mutating documentation check
  for V0. The transitional `generate_api_docs.mjs` was later retired when the
  canonical contract generator took ownership.
- Pin the current Node 24 LTS line (initial toolchain pin `24.18.0`) and its npm version in repository/package metadata, pin `@types/node` to the matching 24 major, preserve `package-lock.json`, add the matching Vitest coverage provider, and add `typecheck`, `typecheck:test`, and `check` scripts.
- Add `tsconfig.test.json`, enable production `no-explicit-any`, replace direct `v-html` with one tested sanitized component, and close current format/lint failures.
- Compile frozen legacy protobufs without invoking `netbox-backend/scripts/protoc.sh`.
- Add CI only after the same commands pass locally.

Target default gate:

```text
root make check
  -> backend format/vet/lint/race-unit/coverage/build/legacy-compile
  -> legacy proto descriptor compile
  -> frontend format/lint/typecheck/test/coverage/build
  -> deterministic generated-doc and link checks
```

PostgreSQL, differential, parity, and browser suites remain separate owned jobs.

The minimum command contract is:

```bash
make check
```

The root target runs `make -C netbox-backend check` and, from `netbox-frontend/`, the pinned `npm ci`, `npm run format:check`, `npm run lint`, `npm run typecheck`, `npm run typecheck:test`, `npm run test:coverage`, and `npm run build` sequence. CI invokes the same root target rather than maintaining a second command list.

Exit evidence: checkpoint V0 in [Testing](TESTING.md) is green; no warning, allow-fail job, unmanaged live server, or unowned exclusion remains.

### Phase 2 — commit the profile and canonical API contracts

Purpose: make scope and both public interfaces machine-verifiable before broad implementation.

Create:

```text
contracts/netbox/v4.4.6-post7/{baseline.yaml,oracle-profile.yaml,normalizers.yaml}
contracts/netbox/v4.4.6-post7/inventory/{baseline-rest,current-rest,current-grpc,current-vue}.yaml
contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml
contracts/netbox/v4.4.6-post7/resources/{identity,dcim,ipam}.yaml
contracts/netbox/v4.4.6-post7/scenarios/*.yaml
contracts/netbox/v4.4.6-post7/schema/profile.schema.json

netbox-backend/api/proto/netbox/types/v1/{pagination,errors}.proto
netbox-backend/api/proto/netbox/identity/v1/identity_service.proto
netbox-backend/api/proto/netbox/dcim/v1/{resources,dcim_service}.proto
netbox-backend/api/proto/netbox/ipam/v1/{resources,ipam_service}.proto
netbox-backend/api/openapi/netbox-go-v1.yaml
netbox-backend/{buf.yaml,buf.gen.yaml,buf.lock}
scripts/{validate_contract_inventory,validate_capability_profile,generate_contract_docs}.mjs
```

The versioned inventory classifies every baseline route/custom action and every
current REST, gRPC, and Vue entry as `in_profile`, `extension`, `deferred`, or
`out_of_scope`, with an owner. `baseline-rest.yaml` is the immutable catalogue
view and therefore records entries at T0; current promoted tiers come from the
Capability Profile, current-interface inventories, Status, and retained
evidence. Extensions use `tier: not_applicable` plus separate `contract`,
`parity`, and `security` verification statuses, so an extension can never
appear to have earned T2 against the oracle. First-profile resources carry the
detailed methods, fields, filters, permissions, mappings, scenarios, and
evidence links. This reconciles 102 frozen REST resources, 176 frozen generated
gRPC services, 13 canonical REST/Vue resources, three canonical gRPC services,
and eight identity REST extension operations without pretending that disabled
legacy artifacts are implemented capabilities.

Those counts are unaffected by ADR 0005's later source-hygiene deletion of
118 dormant Sponge handler/route wrapper pairs. The 102 frozen REST entries
belong to the separate retained DRF registry, and the 176 generated gRPC
services remain frozen with their model, DAO, cache, service, and protobuf
source. Wrapper-file removal is not inventory or compatibility evidence.

Use three capability services: `IdentityService`, `DCIMService`, and `IPAMService`. Do not generate a service per table. Use `int64` IDs, typed pagination, protobuf timestamps, enum zero values named `*_UNSPECIFIED`, field presence plus `FieldMask`, and reserved names/numbers on removal. Omit `google.api.http` annotations because REST is a separate adapter. The `netbox.types.v1` package contains pagination messages and protobuf wire details mapped from canonical application errors; it must not define a second error taxonomy.

`scripts/generate_contract_docs.mjs` owns deterministic generation of `netbox-backend/api/openapi/netbox-go-v1.yaml`, capability inventory pages, and protobuf reference documentation. Its REST inputs are the validated Capability Profile plus hand-owned REST schema metadata in the versioned `resources/*.yaml`; it never infers the contract from Vue or generated handlers. Its protobuf input is the canonical Buf image produced after formatting, linting, and generation. Phase 2 installs the OpenAPI/router conformance test and validates the spec structurally. Each capability promoted in Phases 4, 7, and 8 must make its portion of that test pass, and Phase 10 requires full-profile conformance. The Go server serves the document at `/api/schema/`, and the Vue API documentation view consumes that endpoint instead of the stale proto-derived Swagger bundle.

The new v1 contract remains explicitly pre-publication while the first slice is being proven. Freeze it as the supported v1 contract when its T3 scenarios pass; after that, compatible evolution is additive and breaking changes require a new package version.

Exit evidence:

- Schema validation rejects an undeclared route, RPC, field, filter, or deferred operation.
- The baseline/profile files pin the oracle SHA and configuration inputs that Phase 6 must assert at runtime.
- Pinned protobuf formatting, lint, generation, and regeneration-diff checks pass.
- Every baseline route/action and current REST/gRPC/Vue entry has a high-level classification and owner; the profile maps every detailed first-slice capability to planned REST, gRPC, application, persistence, Vue, and test owners.

Required commands after the tooling is added:

From the repository root:

```bash
git status --porcelain=v1 --untracked-files=all
node scripts/validate_contract_inventory.mjs contracts/netbox/v4.4.6-post7/inventory
node scripts/validate_capability_profile.mjs contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml
```

From `netbox-backend/`:

```bash
buf format --diff --exit-code
buf lint
buf generate
buf build -o /tmp/netbox-go-core-workflow-v1.bin
go test ./gen/go/...
```

Then, from the repository root:

```bash
node scripts/generate_contract_docs.mjs contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml /tmp/netbox-go-core-workflow-v1.bin
node scripts/check_markdown_links.mjs
```

Then, from the repository root, generation must be idempotent:

```bash
git status --porcelain=v1 --untracked-files=all
```

That command must produce no output in a clean checkout both before generation and after all validation/generation commands. This whole-worktree check catches unrelated tracked mutations and untracked generated files; a path-limited diff is insufficient.

After v1 is frozen, add `buf breaking` against the accepted main-branch descriptor/module baseline. Pin and bootstrap Buf, protoc plugins, and documentation generators; do not depend on whichever binaries happen to be on a workstation `PATH`.

### Phase 3 — PostgreSQL test and missing-table bootstrap foundation

Purpose: establish real database evidence and the accepted disposable schema lifecycle.

Create:

```text
netbox-backend/internal/testkit/{postgres,clock,principal,fixtures}.go
netbox-backend/internal/adapters/postgres/bootstrap/{registry,bootstrap}.go
netbox-backend/internal/adapters/postgres/bootstrap/bootstrap_integration_test.go
netbox-backend/internal/adapters/postgres/unit_of_work.go
tests/deployment/compose_smoke.sh
```

Use the implemented topologically ordered registry to preserve this policy for
each registered row type:

1. Check `Migrator().HasTable`.
2. Call GORM `AutoMigrate` for that registered row type only when its table is absent, allowing it to create the table, constraints, and indexes from scratch.
3. Never invoke `AutoMigrate` or any shape-altering migration behavior for an existing table; do not inspect or repair its columns.
4. Fail startup on bootstrap or content-type failure.

Legacy tables can remain in the registry transitionally. First-profile entries switch to authoritative private adapter rows as they are implemented. The content-type table definition includes a unique constraint on `(app_label, model)` when that table is created; this phase does not retrofit the constraint onto an existing table.

The root `docker-compose.yml` now uses a named PostgreSQL volume, lets the
pinned image create `POSTGRES_DB`, and has no initialization-SQL mount. Preserve
that ownership. The smoke test creates fresh Compose resources, waits on
PostgreSQL and application health/readiness, proves startup schema bootstrap,
and tears everything down.

Exit evidence:

- Empty PostgreSQL bootstraps successfully.
- Re-running bootstrap is idempotent.
- A missing table is created in a non-empty database.
- A deliberately mismatched existing table is unchanged.
- There is no backfill, destructive alteration, or drift-repair code path.
- Clean Compose startup passes from fresh volumes and no directory is mounted as an initialization SQL file.

### Phase 4 — build a Site walking skeleton through the shared core

Purpose: prove the dependency direction, unit of work, and dual-adapter path on one resource before repeating it.

Create:

```text
netbox-backend/
  internal/domain/shared/{id,slug,errors,timestamp}.go
  internal/domain/identity/principal.go
  internal/domain/dcim/{choices,site}.go
  internal/application/transaction/unit_of_work.go
  internal/application/changelog/{change,recorder}.go
  internal/application/authz/{actions,authorizer}.go
  internal/application/dcim/{site_commands,site_queries,site_ports,site_service}.go
  internal/adapters/postgres/dcim/row/rows.go
  internal/adapters/postgres/dcim/site_repository.go
  internal/adapters/postgres/changelog/{row,recorder}.go
  internal/adapters/rest/netbox/httpapi/{errors,pagination}.go
  internal/adapters/rest/netbox/dcim/{site_dto,site_mapper,site_handlers}.go
  internal/adapters/grpc/statusmap/errors.go
  internal/adapters/grpc/dcim/{server,site_mapper}.go
```

Implement Site list/get/create/replace/update/delete, explicit `Principal`, a mandatory fail-closed `Authorizer` port, field-presence semantics, transactional change logging, and central error mapping. Phase 4 uses deterministic test authorizers while Phase 5 supplies the identity-backed evaluator; no use case may omit the authorization dependency. Add an import-boundary test that rejects Gin, protobuf/generated API types, GORM/PostgreSQL drivers, global database/configuration packages, and adapter/platform implementations from domain/application packages. Register dependencies explicitly from platform/server composition; do not use `init()`.

The walking skeleton was kept non-public until the Phase 5 credential path was
wired. Its current publication remains T1 until the remaining Phase 4 and 5
exit evidence passes.

Exit evidence:

- REST and in-process gRPC tests prove the same application use case is invoked.
- Domain validation, PostgreSQL rollback, and object-change tests pass.
- A denied representative write fails identically through REST and gRPC before persistence, proving authorization cannot be bypassed.
- Neither adapter imports GORM or performs a second persistence operation.
- A deliberate failure after the Site write rolls back both Site and change record.

### Phase 5 — implement unified identity and authorization

Purpose: make every later compatibility scenario run under the intended security model.

Create:

```text
netbox-backend/
  internal/domain/identity/{user,group,permission,token,session}.go
  internal/application/identity/{service,ports,password,session,token}.go
  internal/application/authz/{constraints,service}.go
  internal/adapters/postgres/identity/{rows,repository}.go
  internal/adapters/rest/netbox/identity/{login,logout,session,token,csrf,middleware}.go
  internal/adapters/grpc/identity/{interceptor,server}.go
  cmd/netbox_go_admin/main.go

netbox-frontend/
  src/api/{http,errors,session}.ts
  src/features/identity/{api,types,store}.ts
  src/pages/auth/LoginPage.vue
  src/router/authGuard.ts
```

Recommendations fixed by this plan:

- The administrator bootstrap is an explicit one-time CLI. Its
  `reset-password` subcommand updates an existing administrator without
  exposing a public recovery endpoint. `create-user` creates only a local
  non-superuser, and `grant-permission` grants a global model permission by
  username; both authenticate an existing active superuser. Every password is
  read from protected stdin/TTY, never a command argument or log, and unsafe
  targets or repeated initialization are rejected.
- Password hashes use a reviewed, versioned, self-describing policy with upgrade support; hash parameters are pinned and tested before implementation is accepted.
- The secure REST extension and `IdentityService` disclose a server-generated API token secret only at creation. Tokens use a one-way verification representation, are scoped and revocable, and are never listed in reusable form. Baseline anonymous token provisioning remains disabled and explicitly deferred.
- Browser sessions use `HttpOnly`, `Secure` in production, `SameSite` cookies plus CSRF protection; JavaScript stores no credential. Session identifiers rotate on authentication and relevant privilege/password changes without permitting fixation.
- Browser CORS uses an explicit origin allowlist with credentials; wildcard credentialed origins are prohibited. Login and token-creation failures are throttled and audited without logging submitted credentials.
- Authentication resolves credentials to one Principal. Shared application authorization owns view/add/change/delete and object constraints.
- List visibility is constrained before count, ordering, and pagination.

Exit evidence:

- Clean database -> bootstrap administrator -> REST session login/logout -> secure API-token creation -> gRPC bearer succeeds without Python.
- Vue identity store, login page, and router-guard tests prove cookie/CSRF behavior and contain no credential persistence; full browser E2E remains Phase 9 evidence.
- The CLI can reset that administrator's password from protected stdin, invalidates sessions according to the declared policy, and leaves no secret in arguments or logs.
- Missing and invalid credentials fail on both transports.
- The same Principal receives matching view/add/change/delete allow and deny results through REST and gRPC.
- At least one object constraint has both allow and deny cases and changes list visibility before pagination.
- Shared scenario tests prove `GetCurrentUser`, token list/create/revoke, and password change through both the identity REST extension and `IdentityService`, with equivalent principals, authorization, committed state, error reasons, and one-time secret handling.
- Baseline token-authentication tests prove unknown-key, expiry, inactive-user, allowed-IP, write-enabled, and revocation cases plus the at-most-once-per-minute `last_used` ordering on both accepted and recognized-but-rejected credentials.
- CSRF, CORS, login and token-creation throttling, expiry/revocation, and secret non-disclosure tests pass.
- Session fixation, rotation, password-change invalidation, and CLI password-reset tests pass.

### Phase 6 — establish differential REST and gRPC parity harnesses

Purpose: prevent implementation volume from outrunning compatibility evidence.

Create:

```text
netbox-backend/test/contract/{rest,grpc}/
netbox-backend/test/parity/
netbox-backend/test/report/
netbox-backend/internal/testkit/{http_server,grpc_server,oracle,state_snapshot}.go
tests/compatibility/{compose.yaml,oracle,fixtures}/
```

The compatibility job starts the pinned oracle and standalone Go implementation with isolated databases, asserts the SHA/config profile, applies versioned fixtures, compares REST status/shapes/errors/order/pagination/durable state, and replays equivalent intent through in-process gRPC. Normalizers are explicit and reviewed.

Exit evidence:

- Positive Site CRUD, field validation, permission denial, and rolled-back mutation scenarios pass.
- The harness refuses to run when the oracle SHA or committed configuration profile differs from the pinned baseline.
- A deliberately introduced REST or gRPC divergence fails the suite.
- Python/Django is confined to the development oracle job and is absent from build, startup, migration, and deployment.
- Reports distinguish baseline T0–T4 evidence from extension verification status and never infer either from generated files.

### Phase 7 — implement the DCIM chain in end-to-end increments

Implement resources in this dependency order:

1. Site walking-skeleton completion.
2. Manufacturer.
3. DeviceRole and RackRole.
4. RackType, depending on Manufacturer.
5. DeviceType, depending on Manufacturer.
6. InterfaceTemplate, depending on DeviceType.
7. Rack, depending on Site and optionally RackRole/RackType.
8. Device and Interface as one coordinated increment, depending on Site, DeviceType, DeviceRole, and optionally Rack. Implement the Interface domain/application/persistence core before `CreateDevice`, then publish both resource surfaces only when atomic template instantiation is proven.

For each resource, add the same backend artifacts and its typed Vue boundary:

```text
netbox-backend/internal/
  domain/dcim/<resource>.go
  application/dcim/<resource>_{commands,queries,ports,service}.go
  adapters/postgres/dcim/<resource>_repository.go
  adapters/rest/netbox/dcim/<resource>_{dto,mapper,handlers}.go
  adapters/grpc/dcim/<resource>_mapper.go

netbox-frontend/src/features/dcim/
  api/{dto,mappers,dcimApi}.ts
  resources/<resource>.ts
  components/<resource>*.vue
  pages/<resource>*.vue
```

Critical shared files include `netbox-backend/internal/domain/dcim/rack_unit.go`, `netbox-backend/internal/domain/dcim/rack_placement.go`, `netbox-backend/internal/application/dcim/create_device.go`, and `netbox-backend/internal/application/dcim/instantiate_interface_templates.go`.

Required behavior includes baseline field/default/null rules, the baseline's exact (sometimes surprising) uniqueness scopes, RackType inheritance/propagation, included rack-unit geometry, half-U placement, face/overlap/0U protection, Device site/rack consistency, Interface name uniqueness and immovability, and atomic InterfaceTemplate instantiation.

Exit evidence per promoted resource:

- Table-driven domain positive/negative tests.
- Real-PostgreSQL constraint, concurrency where applicable, and rollback tests.
- Authenticated REST differential scenarios declared for that resource in the profile reach T2.
- Equivalent profile-declared gRPC scenarios reach T3 through the same use case.
- Change-log and list-visibility assertions pass.
- Typed Vue DTO mapping, form hydration/serialization, deferred-control absence, and component tests pass for the resource.

Additional Device exit evidence: failure during the third instantiated Interface proves the Device and all Interfaces/change records roll back.

Phase 7 increments include component-level Vue evidence but remain explicitly non-T4 until Phase 9 registers the manifest-backed routes and proves the complete workflows in a browser.

### Phase 8 — implement IPAM and Interface assignment

Implement in this order:

1. VRF.
2. Prefix.
3. IPAddress.
4. Assign/unassign IPAddress to a DCIM Interface.

Create the parallel domain/application/PostgreSQL/REST/gRPC artifacts under the `ipam` packages plus `application/ipam/assignment.go`. Add the matching typed DTO/mappers, resource definitions, components, pages, and component tests under `netbox-frontend/src/features/ipam/` in the same increments.

Required behavior includes unique nullable route distinguishers, Prefix host-bit rejection plus canonical storage, `/0` rejection, VRF/global uniqueness policy, containment/depth/child counts, IP host-plus-mask preservation, IPv4/IPv6 and SLAAC rules, network/broadcast restrictions with narrow prefix exceptions, lowercase DNS, safe Interface-only content types, and concurrency control for uniqueness/assignment.

Prefix containment is derived from network values; do not add an artificial `prefix_id` to IPAddress.

Exit evidence:

- IPv4 and IPv6 domain suites pass.
- PostgreSQL concurrent duplicate and assignment scenarios pass.
- Prefix hierarchy remains correct after create, update, and delete.
- REST reaches T2 and gRPC reaches T3 for every IPAM/assignment operation, field, and scenario declared by the profile.
- A failed assignment leaves IPAddress, Interface projection, and change log unchanged.
- Typed Vue mapping and component tests cover Prefix/IP normalization plus assignment/unassignment payloads without constructing content-type fields in pages.

Phase 8 increments carry component-level Vue evidence under the same rule: the profile remains incomplete until the operator workflows reach T4 in Phase 9.

### Phase 9 — integrate and prove the Vue workflows

Integrate and prove the typed feature boundaries created with Phases 7 and 8:

```text
netbox-frontend/src/features/core/{resources,adapters,api}.ts
netbox-frontend/src/pages/core/
tests/browser/{run.sh,browser_e2e.mjs,README.md}
tests/deployment/compose_smoke.sh
```

Only API infrastructure imports Axios. Typed feature adapters own exact
snake-case DTOs, form hydration, mutation payloads, and filter conversion for
all 13 resources. Pages/components never construct content-type assignment
payloads or endpoint strings directly.

Extend the Phase 5 identity API/store and auth guard rather than creating a second frontend authentication path.

The browser workflow supports creating the support resources,
InterfaceTemplates, Site/Rack/Device, viewing instantiated Interfaces, creating
VRF/Prefix/IPAddress, assignment, and unassignment. Asynchronous edit hydration
and nested relationship normalization pass through the typed adapters. The
active route tree contains no rack-elevation tab, unsupported component tabs,
bulk controls/routes, automatic allocation, GraphQL, Scripts, or Reports.

The registry is now bounded to the 13 profile resources and backed by typed
adapters. Broad legacy bulk/special routes are not part of the supported route
tree. The Capability Profile and generated inventory, not the Vue registry,
remain authoritative for public scope.

Exit evidence:

- Both clean-database workflows reach T4.
- Browser tests cover validation, conflict, permission denial, explicit-null update, delete protection, assignment/unassignment, and rollback.
- Interface deletion presents an explicit warning that assigned IPAddresses will also be deleted; browser/oracle scenarios prove cancel and confirm behavior plus the resulting IP/change records.
- No credential appears in `localStorage` or logs.
- Vue uses REST only and contains no independent authorization/domain rule.
- An automated manifest/runtime test fails if a dual-interface profile capability is missing from REST or gRPC, or if Vue exposes an undeclared/deferred route. Only schema-approved transport-scope entries, such as cookie/CSRF mechanics and the administrator CLI, may be exceptions.

The required browser gate is `make browser-e2e`. The deployment smoke harness
owns the frontend/backend/PostgreSQL processes and an ephemeral administrator;
the Node driver controls an installed headless Chrome through Chrome DevTools
Protocol. It retains credential-free screenshots, sanitized DOM, Chrome logs,
and request method/URL/status summaries on failure.

### Phase 10 — retire displaced legacy paths and sign off

Runtime registration is now explicit and profile-bounded. Generic REST and
Sponge gRPC registration are disabled in the default composition and excluded
from compatibility evidence. Do not restore a development switch unless it is
named, warning-emitting, isolated, and documented in the inventory.

For the first profile, the displaced artifacts have already been removed from:

```text
api/netbox_go/v1/
internal/{model,dao,cache,service,handler,routers}/
```

The obsolete DRF and legacy API-document generators were retired with those
outputs. The interface inventories separately retain 102 deferred REST
configurations and 176 deferred gRPC services alongside canonical surfaces.
The startup persistence registry is a different 198-table inventory: 176
legacy rows, 8 identity rows, 13 profile rows, and 1 object-change row. Apply
the same owned-source rule to later profiles; do not hand-edit generated
output.

Separately, ADR 0005 removed exactly 118 untouched, runtime-dormant Sponge
per-resource handler delegates and their 118 matching top-level route wrappers
as source hygiene. That bounded cleanup did not remove or alter the separate
102 direct-GORM DRF configurations, the 176 generated gRPC services, or their
frozen models, DAOs, caches, services, and protobufs, which remain available
for controlled reuse or reference but stay runtime-disabled and unpublished.
It earns no tier or capability-completion claim. Every future stack retirement
still requires capability completion under ADR 0004 and CP-13.

The displaced frontend routes/pages and the legacy Swagger/OpenAPI bundle are
retired. Canonical OpenAPI and contract documentation are generated from the
validated profile and protobuf descriptor by `generate_contract_docs.mjs`.

Exit evidence:

- Runtime REST inventory equals the profile plus named
  health/readiness/identity/schema endpoints and contains no dedicated
  `GET /ping`; ordinary SPA history fallback is not a diagnostic route.
- gRPC reflection lists only canonical supported services and health.
- No first-profile code imports legacy model/DAO/cache/service packages.
- V0–V5 remain green after physical retirement.
- [Project status](STATUS.md) links the evidence for each tier and does not call whole DCIM/IPAM modules complete.

## Recommended pull-request sequence

Each change is reviewable and must leave its applicable gates green:

1. Unsafe-runtime containment.
2. V0 quality baseline and pinned tooling.
3. Capability profile schema/data and canonical pre-publication protos.
4. PostgreSQL testkit and missing-table-only bootstrap.
5. Site walking skeleton and dual-adapter architecture proof.
6. Identity, administrator CLI, sessions/tokens, and shared RBAC.
7. Differential/parity harness.
8. Site profile promotion with its typed Vue boundary.
9. Manufacturer, DeviceRole, RackRole, and RackType with their typed Vue boundaries.
10. DeviceType, InterfaceTemplate, Rack, Device, and Interface with their typed Vue boundaries.
11. VRF and Prefix promotion.
12. IPAddress assignment/unassignment.
13. Vue route cutover, integrated workflows, and real-Chrome/CDP evidence.
14. Legacy retirement and first-profile sign-off.

Do not combine all resources into a single branch, and do not merge a REST-only or gRPC-only business implementation for later reconciliation.

## First-profile completion gate

The Core Workflow Profile closes only when:

- the default V0 quality gate and real-PostgreSQL bootstrap suite pass;
- all baseline-declared REST operations, fields, filters, and scenarios in the profile are T2 against the pinned oracle;
- all corresponding baseline gRPC capabilities are T3 through the same application path;
- profile-declared identity extensions pass their standalone REST/gRPC contract, shared-core parity, and security suites while remaining labelled extensions rather than T2;
- identity/RBAC, transaction, rollback, and object-change evidence exists;
- both Vue workflows are T4;
- deferred fields/actions are visible in the manifest and absent from the supported UI;
- first-profile legacy paths are disabled and then retired; and
- status and API documentation are regenerated from evidence without claiming full DCIM/IPAM completion.

The next profile is chosen from the visible deferred ledger after this gate, not by whichever generated resource happens to be easiest to expose.
