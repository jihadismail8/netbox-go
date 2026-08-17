# Business-logic discovery reference

> **Partial upstream analysis, not an accepted specification or implementation
> guide.** These notes were derived from the pinned post-4.4.6 source snapshot
> at commit `fbb948d30e79ce657fac62994a22aca72c1770a9`. They contain known gaps
> and stale claims. The pinned source, accepted Capability Profile, and
> executable compatibility evidence remain authoritative.

## Contents and known coverage

| Document                        | Discovery value                                                           | Known limitation                                                                          |
| ------------------------------- | ------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| [DCIM models](dcim-models.md)   | Candidate DCIM fields, relationships, validation, and computed behavior   | Not reconciled to every current first-profile rule or later DCIM resource                 |
| [IPAM models](ipam-models.md)   | Candidate IPAM fields, network rules, hierarchy, and utilization behavior | Not a complete source-linked rule matrix                                                  |
| [API patterns](api-patterns.md) | Historical serializer/action/bulk observations                            | Does not reconcile to all 23 authoritative actions; bulk is deferred in the first profile |
| [FilterSets](filtersets.md)     | Partial record of upstream lookup syntax and selected filters             | Covers only part of the baseline and must not drive a generic SQL parser                  |

The separate [`entities/`](../entities/) inventory has 104 derived entity
pages, while the baseline REST catalogue has 132 resources. These sets are not
one-to-one. Entity checkboxes, generated paths, and status labels must never be
used as progress evidence.

## Required use

For each Capability Profile:

1. Start from the authoritative baseline inventory and select a coherent
   operator or automation outcome.
2. Trace every candidate field, relationship, filter, action, validation,
   permission, transaction, error, deletion rule, and durable side effect to
   the pinned source and upstream tests.
3. Record the accepted subset in the machine-readable profile, resource
   metadata, and scenario metadata before implementation.
4. Maintain traceability from accepted rule to source reference, Go test,
   differential REST case, corresponding gRPC case, applicable browser case,
   and retained evidence.
5. Keep every omitted behavior explicitly deferred or excluded. A derived
   note never silently broadens a profile.

Before first-profile V2, the relationship/delete claims for all 13 resources
must be reconciled directly to the pinned source. The current derived material
has known `CASCADE`/`SET_NULL`/`PROTECT` errors and stale references to retired
generated paths.

## Translation rules

- Python inheritance describes upstream behavior; it does not prescribe Go
  struct embedding or one Go type per Django class.
- Domain packages own pure Managed Object invariants and typed policies. They
  do not import GORM, Gin, protobuf, SQL drivers, or database handles.
- Application services own authorization, cross-object orchestration,
  transaction boundaries, object-change intent, and required durable effects.
- PostgreSQL adapters own private row types, GORM mappings, constraints,
  locking, and explicit parameterized queries.
- REST and gRPC adapters translate exact transport contracts into the same
  application use cases. They do not make independent business decisions.
- Vue implements presentation and operator workflow over REST; it does not
  duplicate backend authorization or validation.
- `BeforeSave`/`AfterSave` hooks, generic post-hooks, or generic CRUD event
  pipelines must not become hidden business-logic centers.
- IDs and presence semantics use the accepted typed contracts. Do not infer
  `uint`, nullability, or zero-value behavior from old generated structs.

## Filtering rule

Each profile declares an exact allowlist of accepted filter and ordering keys.
The REST adapter rejects undeclared keys, parses declared values into a typed
list query, and passes them to explicit repository predicates. Shared syntax
helpers may be introduced only behind the allowlist; an arbitrary query key
must never become a column name, GORM clause, or SQL fragment.

## Authority order

When this directory disagrees with another source, use:

1. accepted ADRs and [project language](../../CONTEXT.md);
2. the machine-readable Capability Profile for declared scope;
3. [Compatibility](../COMPATIBILITY.md) for proof and tier claims;
4. the pinned upstream source and strict oracle for baseline behavior;
5. [Architecture](../ARCHITECTURE.md) and
   [Coding standards](../CODING_STANDARDS.md) for implementation boundaries;
6. [Status](../STATUS.md) for current claims.

Do not implement directly from this reference directory without completing
the source reconciliation and accepted-profile steps above.
