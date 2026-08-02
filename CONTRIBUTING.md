# Contributing to NetBox Go

NetBox Go is replacing behavior, not translating files one for one. Every change must preserve the standalone Go/Vue boundary and move supported operations toward one shared application implementation used by REST and gRPC.

## Read before changing code

1. [Project language](CONTEXT.md)
2. [Architecture](docs/ARCHITECTURE.md)
3. [Compatibility contract](docs/COMPATIBILITY.md)
4. [Coding standards](docs/CODING_STANDARDS.md)
5. [Core Workflow implementation plan](docs/IMPLEMENTATION_PLAN.md)
6. [Whole-project execution playbook](docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md)
7. [Current project status](docs/STATUS.md)
8. [Testing and evidence gates](docs/TESTING.md)
9. [Accepted decisions](docs/adr/README.md)

When documents disagree, accepted ADRs govern hard-to-reverse decisions; the
Capability Profile governs declared surface; Compatibility governs proof and
tier promotion; the pinned upstream source/oracle governs baseline behavior;
Architecture and Coding Standards govern implementation boundaries; and Status
governs current claims. Generated code is not architectural precedent.

## Change workflow

For a capability change:

1. Identify the exact baseline operation and its positive, negative, permission, and rollback scenarios.
2. Record the capability and initial evidence tier in the machine-readable Capability Profile.
3. Add or update transport-neutral domain/application tests before implementing behavior.
4. Implement the use case and its consumer-owned repository ports.
5. Implement the PostgreSQL adapter and prove constraints, transaction rollback, and change logging on real PostgreSQL.
6. Connect both REST and gRPC adapters to the same use case and prove equivalent outcomes.
7. Add or update typed Vue DTO mapping and component evidence in the same increment when the capability is operator-facing; keep it non-T4 until the integrated browser workflow passes.
8. Regenerate owned outputs, run the required gates, and update status only with links to passing evidence.

The repository remains bounded by first-profile verification. Check the current
retained V0 result in [Testing](docs/TESTING.md) and
[Evidence](docs/evidence/README.md) before starting or merging another
capability, then continue V1-V6 in the file-specific
[execution playbook](docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md). Do not begin a
later profile before first-profile V6 sign-off.

## Generated files

- Never hand-edit generated output, including existing generated files that lack a reliable marker.
- Change the owned source definition or generator, then regenerate.
- Do not add behavior to the current Sponge-generated models, DAOs, caches, services, handlers, routers, or table-oriented protobuf contracts.
- Retire a displaced generated path only after its replacement satisfies the capability completion gate in ADR 0004.
- A generation change must use pinned tools, have deterministic sorted inputs, fail on missing inputs, and reproduce a clean worktree.
- Keep generated source needed for builds and review; do not commit binaries, coverage, frontend build output, editor state, or local secrets.

See [ADR 0004](docs/adr/0004-generated-scaffolding-is-immutable-and-transitional.md) for the governing decision.

## Review checklist

- [ ] The change follows the dependency direction and contains no transport or GORM types in domain/application contracts.
- [ ] REST and gRPC call the same use case; neither calls the other transport or persistence directly.
- [ ] Authentication, authorization, validation, transaction handling, and change logging are covered.
- [ ] Inputs preserve absent, explicit-null, zero, and concrete values where the contract distinguishes them.
- [ ] Database behavior is verified on PostgreSQL when constraints, locking, arrays, JSONB, CIDR, or transactions matter.
- [ ] Tests own their dependencies and do not require an unmanaged live service.
- [ ] Generated outputs are reproducible and contain no handwritten edits.
- [ ] No credential material, hash, DSN, or full configuration is logged or retained in evidence. Responses emit only the documented hardened session cookie, CSRF bootstrap cookie/body value, and one-time token-creation secret; reusable secrets are never returned afterward.
- [ ] Public behavior and evidence status are documented accurately; scaffolding is not reported as compatibility.
- [ ] Any exception names the rule, exact path or symbol, reason, owner, and removal milestone.
