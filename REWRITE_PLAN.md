# NetBox Go rewrite plan

This file is retained as the stable entry point for links created during the
early rewrite. The original estimate-driven plan is superseded by the
evidence-gated documentation below.

## Objective

Build a [Standalone Runtime](CONTEXT.md) with a Go backend and Vue frontend for
the pinned NetBox Compatibility Baseline. Python, Django, and the checked-in
upstream source are development references only; they are not build, migration,
startup, or deployment dependencies.

The application has two first-class interfaces:

- HTTPS REST with exact compatibility inside a reviewed Capability Profile;
- versioned gRPC with semantic parity over the same application/domain path.

The Vue application uses REST and targets Workflow Parity, not pixel parity.

## Current plan and status

- [Core Workflow implementation plan](docs/IMPLEMENTATION_PLAN.md) defines the
  13-resource first profile and its exit conditions.
- [Whole-project execution playbook](docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md)
  gives implementation agents the current recovery sequence, stable rule IDs,
  repeatable profile factory, full expansion queue, and production gates.
- [Project status](docs/STATUS.md) records what is implemented and what remains
  unproved.
- [Compatibility](docs/COMPATIBILITY.md) defines T0–T4 and the meaning of
  complete.
- [Testing](docs/TESTING.md) defines the executable gates and their evidence
  boundaries.
- [Evidence ledger](docs/evidence/README.md) records durable results.
- [Roadmap](docs/ROADMAP.md) orders later profiles and production hardening.
- [Architecture](docs/ARCHITECTURE.md),
  [Coding standards](docs/CODING_STANDARDS.md), and the
  [ADRs](docs/adr/README.md) govern implementation choices.

## Current first-profile boundary

```text
Manufacturer -> DeviceType -> InterfaceTemplate
                                      |
Site -> Rack -> Device ---------------+-> Interface

VRF -> Prefix -> IPAddress -> assign to Interface -> unassign
```

RackRole and DeviceRole are supporting Managed Objects. The profile has typed
per-table PostgreSQL persistence, persisted groups/RBAC, typed Vue adapters,
canonical REST/gRPC adapters, and physical retirement of the 13 displaced
legacy stacks.

The profile remains T1 and pre-publication because the latest strict oracle,
real-PostgreSQL, deployment, gRPC parity, and browser executions are not yet
retained as durable evidence. No whole-module or production-readiness claim is
made.

## Development database decision

Development and test databases are disposable. Startup may use GORM
`AutoMigrate` only to create a missing table. It must not alter, inspect for
missing columns, repair, or backfill an existing table. Schema shape changes
are tested by dropping and recreating the database; production upgrade
migrations are a later release concern.

The current startup registry contains 198 tables: 176 deferred legacy rows, 8
Go-owned identity rows, 13 typed first-profile rows, and 1 typed object-change
row.

Do not use historical file counts, week estimates, generic CRUD checklists, or
the presence of generated code as completion evidence.
