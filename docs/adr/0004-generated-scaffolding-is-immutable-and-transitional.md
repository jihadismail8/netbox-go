---
status: accepted
---

# Treat generated scaffolding as immutable, transitional code

The current Sponge-generated per-resource models, DAOs, caches, services, handlers, routers, per-table protobuf contracts, and other reproducible outputs are transitional scaffolding and must not receive handwritten business logic. New behavior belongs in handwritten domain, application, persistence-adapter, and transport-adapter code; new public `.proto` files are handwritten contract sources, while their generated Go files are immutable outputs. This avoids losing fixes during regeneration and prevents the current independent REST and gRPC paths from becoming permanent architecture.

## Consequences

- Change a generated artifact only by changing its owned source definition or generator and regenerating it.
- New generated files must identify their generator and source and carry the standard `Code generated ... DO NOT EDIT.` marker when the file format supports comments.
- Generation must be deterministic, use pinned tools, and be checked in CI by regenerating and failing on an unexpected diff.
- Generated source required to build or review the application may be committed; binaries, coverage, frontend build output, and other disposable build artifacts are not source.
- Existing generated paths may remain temporarily while a vertical slice is replaced, but they are not extension points or evidence of compatibility.
- Retire each displaced REST route, gRPC service, DAO, model, and cache path only after its replacement satisfies the capability completion gate.
- Transitional process composition and custom security files (`internal/handler/auth.go`, `internal/routers/routers.go`, `internal/server/grpc.go`, and `cmd/netbox_go/initial/initApp.go`) are hand-owned wiring, not per-resource generated outputs. They may receive narrow containment changes before replacement. Generated per-resource files and registries remain immutable; if later ownership discovery identifies any named composition file as reproducible output, its source/template must be changed and regenerated instead.

## Narrow wrapper-retirement amendment

[ADR 0005](0005-retire-dormant-sponge-http-wrappers.md) accepts one bounded
exception to the capability-completion ordering above: the 118
template-identical, runtime-dormant per-resource HTTP handler wrappers and their
118 matching Sponge router wrappers may be removed early. Their canonical
handlers, authentication wiring, lower generated stacks, protobufs, and the
separate frozen DRF registry are not part of that exception. The removal earns
no compatibility or completion claim, must be protected against regeneration,
and does not relax any other consequence of this decision.
