# Differential compatibility job

This directory is the only place where Python/Django is permitted at test
time. The job materializes the exact regular-file blobs and executable modes
from the pinned NetBox commit into a temporary tree and mounts that tree as the
oracle source. Every authoritative Git operation disables local replacement
refs and grafts, and direct tree/blob reads bypass repository-local export
attributes. It never executes untracked, ignored, modified,
replacement-substituted, graft-authorized, or attribute-filtered live-checkout
content. It refuses a source or configuration mismatch, creates disposable
databases, and compares the profile-declared REST projection with the standalone
Go implementation.

Run from the repository root:

```bash
make compatibility-test
```

The job uses only pinned/local images and always removes its containers and
volumes. It never contributes code, migrations, or runtime dependencies to the
Go/Vue application.

The gate refuses to start unless the original checkout is clean at commit
`fbb948d30e79ce657fac62994a22aca72c1770a9`, the release reports `4.4.6`,
and the capability profile names the `netbox-v4.4.6-post7` baseline. It then
materializes that exact Git tree and asserts the oracle's effective security
and uniqueness settings from inside the running Django process. Unsupported
Git modes or object types fail closed.

The same production materializer has a fast adversarial regression which
plants untracked Python, ignored bytecode, tracked dirt, a wrong trusted commit,
a malicious Git replacement ref, a forged graft, and repository-local export
attributes:

```bash
make compatibility-oracle-source-test
```

This check requires Git, but no Docker daemon, network, or Python runtime.

The differential driver exercises all 13 resources in `core-workflow-v1` with
broad authentication, CRUD, projection, canonical-network, template,
pagination, filter, and ordering coverage. For the payloads it exercises, the
comparator remains strict about declared writable/response fields, annotated
counter omissions, choice envelopes, numeric JSON types, paths, queries,
trailing slashes, validation reasons, field presence, generated identifiers,
committed state, and side effects.

This is not yet the complete first-profile T2 matrix. The 2026-08-03 coverage
audit found missing permission, presence, conflict, invariant, rollback, and
durable-effect scenarios. The profile now has machine-readable row/test/
evidence traceability, but uncovered rows remain explicitly pending. Implement
and exercise every required case before treating a successful run as a per-
capability T2 report.

Failure diagnostics are retained in the artifact directory printed by the
job. They contain redacted exchanges, the effective oracle configuration,
Compose logs, and standalone-server logs. Set `NETBOX_COMPAT_ARTIFACT_DIR` to
choose the location.

The fast comparator sensitivity test is also available independently:

```bash
make compatibility-comparator-test
```

It supplies a deliberate semantic divergence and fails unless the comparator
rejects it at the expected field path.
