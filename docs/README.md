# Documentation guide

This index separates accepted decisions, current evidence, executable
contracts, and historical analysis. The compatibility baseline is the
checked-in NetBox source at commit
`fbb948d30e79ce657fac62994a22aca72c1770a9`
(`v4.4.6-7-gfbb948d30`), not the official `v4.4.6` tag.

## Canonical documents

| Document                                                   | Authority                                                                                              |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| [Project language](../CONTEXT.md)                          | Ubiquitous language and compatibility boundary                                                         |
| [Architecture](ARCHITECTURE.md)                            | Current architecture and normative dependency direction                                                |
| [Compatibility](COMPATIBILITY.md)                          | REST exactness, gRPC parity, evidence tiers, definition of complete                                    |
| [Coding standards](CODING_STANDARDS.md)                    | Normative Go, REST, gRPC, PostgreSQL, Vue, generation, and test rules                                  |
| [Implementation plan](IMPLEMENTATION_PLAN.md)              | First Capability Profile, implementation sequence, and exit conditions                                 |
| [Execution playbook](IMPLEMENTATION_EXECUTION_PLAYBOOK.md) | Executor rules, stuck-work recovery, repeatable profile factory, expansion queue, and production gates |
| [Project status](STATUS.md)                                | Dated, conservative implementation and evidence ledger                                                 |
| [Testing](TESTING.md)                                      | Commands, boundary of each suite, and verification checkpoints                                         |
| [Evidence](evidence/README.md)                             | Durable artifact policy and current result ledger                                                      |
| [Roadmap](ROADMAP.md)                                      | Ordered delivery gates after the first profile                                                         |
| [ADRs](adr/README.md)                                      | Accepted hard-to-reverse decisions                                                                     |

When documents disagree, accepted ADRs govern hard-to-reverse decisions; the
machine-readable Capability Profile governs declared surface; Compatibility
governs proof and tier promotion; the pinned source/oracle governs baseline
behavior; Architecture and Coding Standards govern implementation boundaries;
and Status governs current claims. Generated-file counts never override those
documents.

## Machine-readable contracts and inventories

- [`contracts/netbox/v4.4.6-post7/`](../contracts/netbox/v4.4.6-post7/)
  pins the oracle, committed normalizers, resource metadata, scenarios, and the
  pre-publication `core-workflow-v1` profile.
- [`inventory/`](../contracts/netbox/v4.4.6-post7/inventory/) is the
  authoritative current-surface inventory: baseline REST 155; current REST 123
  (102 frozen + 13 canonical + 8 identity); current gRPC 179 (176 frozen + 3
  canonical); Vue 13.
- [`contracts/`](contracts/) is generated from the validated profile and
  canonical protobuf descriptor. It documents a contract; it does not prove
  that the oracle, PostgreSQL, gRPC, or browser gates passed.
- [`apis/`](apis/) now points to the canonical contract and status documents.
  The obsolete per-model Vue-registry catalogue and legacy Swagger/OpenAPI
  bundle were retired.

Regenerate and validate the authoritative artifacts from the repository root:

```bash
make generated-check
make contracts-check
make docs-check
```

Update the Capability Profile/resource metadata or canonical protobuf sources,
not generated outputs, when contract content changes.

## Reference analysis

- [Business logic](business-logic/README.md) records upstream behavior
  discovered during analysis. Recheck it against the pinned source and strict
  oracle before treating it as current behavior.
- [Entity inventory](entities/README.md) and [CRUD paradigm](CRUD_PARADIGM.md)
  are analysis aids, not implementation status or public scope.

## Historical planning

[REWRITE_PLAN.md](../REWRITE_PLAN.md) is now a short redirect to the canonical
plan. The following files remain historical reference and their estimates,
counts, checklists, and completion language are superseded:

- [API implementation plan](API_IMPLEMENTATION_PLAN.md)
- [Frontend analysis](FRONTEND_ANALYSIS.md)
- [Frontend plan](FRONTEND_PLAN.md)
- [Frontend completion plan](FRONTEND_COMPLETION_PLAN.md)
- [Test plan](TEST_PLAN.md)

## Documentation rules

1. Use the canonical terms in [CONTEXT.md](../CONTEXT.md): Managed Object,
   Compatibility Baseline, Standalone Runtime, Interface Parity, Workflow
   Parity, Capability Profile, and Extension Service.
2. Do not call a capability complete above its retained evidence tier.
3. Keep implementation structure separate from executed evidence.
4. Keep current observations separate from normative target decisions.
5. Link claims to code, contracts, tests, or retained artifacts.
6. Never convert artifact counts into progress percentages.
7. Do not cite temporary `/tmp` output as durable evidence.
