# Differential compatibility job

This directory is the only place where Python/Django is permitted at test
time. The job mounts the checked-in NetBox oracle source, refuses a source or
configuration mismatch, creates disposable databases, and compares the
profile-declared REST projection with the standalone Go implementation.

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
asserts the oracle's effective security and uniqueness settings from inside
the running Django process.

The differential driver covers all 13 resources in `core-workflow-v1`. It
checks authentication, required-field and network validation, create/get/
patch/replace/delete, relationship and computed-field projections, canonical
network values, template-created interfaces, pagination, and every filter and
ordering field declared for all 13 resources (including relationship,
containment, and assignment filters). Every writable field is compared on
mutations; GET/list/PATCH/PUT must expose every
declared writable and response-only field. The exact POST-only annotated
counter omissions are pinned explicitly, including a check that Go does not
emit them. Only generated IDs, configured URL origins, and timestamps are
normalized as declared in `normalizers.yaml`. Choice envelopes, numeric JSON
types, paths, queries, trailing slashes, validation reasons, and field
presence stay exact. Each generated identifier is paired to an explicit
scenario symbol (including template-created objects); an unbound identifier is
a hard failure rather than a generic placeholder.

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
