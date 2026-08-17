# Derived NetBox entity reference

> **Partial generated discovery material, not implementation documentation or
> status.** The 104 entity pages in this directory were derived from the pinned
> post-4.4.6 Python source, but they do not cover the 132-resource baseline
> one-to-one and contain known stale paths and relationship errors.

## Coverage

| Module         | Derived entity pages |
| -------------- | -------------------: |
| Circuits       |                    6 |
| Core           |                    5 |
| DCIM           |                   45 |
| Extras         |                    8 |
| IPAM           |                   18 |
| Tenancy        |                    6 |
| Users          |                    3 |
| Virtualization |                    5 |
| VPN            |                    5 |
| Wireless       |                    3 |
| **Total**      |              **104** |

The baseline inventory separately contains 132 resources and 23 custom
actions. Some resources have no page, some pages are not public REST resources,
and actions are not represented consistently. Never derive a completion ratio
or public scope from this directory.

## What a page may help discover

- candidate upstream fields and inheritance;
- candidate relationships and constraints;
- likely dependencies and reverse references; and
- source files that deserve direct review.

Each item must be rechecked against the exact pinned source and upstream tests.
Known inaccuracies include contradictory `CASCADE`, `SET_NULL`, and `PROTECT`
claims, invented relationships, incorrect nullability, and references to
retired generated Go paths. The embedded implementation checklists are stale;
zero checked boxes or many unchecked boxes say nothing about current Go/Vue
status.

## Required workflow

1. Start from the authoritative baseline inventory and selected Capability
   Profile, not from this directory's file list.
2. Verify field, relationship, deletion, validation, permission, and side-effect
   behavior against pinned source and oracle.
3. Record accepted behavior in resource/scenario metadata and the traceability
   matrix.
4. Implement through the typed domain/application/adapters architecture.
5. Use [Status](../STATUS.md) and retained [Evidence](../evidence/README.md) for
   completion claims.

See [Business-logic discovery](../business-logic/README.md),
[Compatibility](../COMPATIBILITY.md), and
[Coding standards](../CODING_STANDARDS.md) for the governing boundaries.
