# NetBox Go

NetBox Go is an early-stage, standalone rewrite of NetBox using a Go backend and a Vue 3 frontend. The compatibility target is the core behavior of the pinned post-4.4.6 NetBox source submodule at commit `fbb948d30e79ce657fac62994a22aca72c1770a9` (`git describe --tags`: `v4.4.6-7-gfbb948d30`).

The supported runtime is intentionally smaller than NetBox: it exposes the
pre-publication `core-workflow-v1` profile (13 DCIM/IPAM resources plus Go-owned
identity), while the broad generated scaffold is disabled and transitional. It
is **not yet a drop-in replacement**; [current status](docs/STATUS.md) records
evidence by compatibility tier and does not infer completion from generated
files.

As of the 2026-08-03 audit, 13 of 153 accepted baseline resource/action
entries are T1 (8.50%), 140 remain T0, and none has retained T2 REST, T3 gRPC,
or T4 browser evidence. The first profile still has implementation gaps before
those external gates; this is not simply a backlog of test executions.

This remains a development build. The supported REST and gRPC routes fail closed, do not expose generic user rows, and no longer register the generated direct-GORM business surfaces, but production hardening and full NetBox coverage are not complete. Bind development ports to loopback, use disposable data, and do not treat deferred resources as supported.

## Product boundary

- HTTPS REST is the NetBox-compatible interface used by automation and the Vue application.
- gRPC is a first-class service-to-service interface with semantic parity over the same application and domain logic.
- The delivered application must not depend on Python, Django, or the checked-in upstream source to build, migrate, start, or run.
- Vue targets operator-workflow parity, not a pixel-perfect copy of the upstream UI.
- GraphQL and Python plugin, custom-script, and report execution are outside the compatibility target.

The canonical REST and gRPC adapters share one application layer for identity, authorization, validation, transactions, change logging, DCIM, and IPAM. Their presence establishes a T1 implementation; baseline differential compatibility, complete browser evidence, and later NetBox modules still require their own gates before higher tiers are claimed.

## Read this first

| Document                                                        | Purpose                                                                           |
| --------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| [Documentation index](docs/README.md)                           | Canonical, generated, reference, and historical documents                         |
| [Current status](docs/STATUS.md)                                | Evidence-backed audit of what works and what does not                             |
| [Compatibility contract](docs/COMPATIBILITY.md)                 | Meaning of REST compatibility, gRPC parity, and “complete”                        |
| [Architecture](docs/ARCHITECTURE.md)                            | Audited current structure and normative target architecture                       |
| [Coding standards](docs/CODING_STANDARDS.md)                    | Required Go, REST, gRPC, database, frontend, generation, and testing practices    |
| [Contributing](CONTRIBUTING.md)                                 | Change workflow and review checklist                                              |
| [Implementation plan](docs/IMPLEMENTATION_PLAN.md)              | First Capability Profile, file-level execution order, and exit evidence           |
| [Execution playbook](docs/IMPLEMENTATION_EXECUTION_PLAYBOOK.md) | Agent-ready recovery sequence, stable rules, later profiles, and production gates |
| [Roadmap](docs/ROADMAP.md)                                      | Stable goals, dependencies, checklists, and outcome-gated implementation sequence |
| [Testing](docs/TESTING.md)                                      | Current test results and required compatibility strategy                          |
| [Project language](CONTEXT.md)                                  | Agreed terms and scope boundaries                                                 |
| [ADRs](docs/adr/README.md)                                      | Accepted architectural decisions                                                  |

## Repository layout

```text
netbox-backend/    Go backend, REST adapters, gRPC services, and PostgreSQL access
netbox-frontend/   Vue 3 application
netbox/            Pinned upstream oracle submodule; test-only and never a runtime dependency
docs/              Canonical plans, decisions, audits, and generated inventories
scripts/           Repository-level documentation tooling
```

Clone with the pinned compatibility oracle, or initialize it after a normal
clone:

```bash
git clone --recurse-submodules https://github.com/jihadismail8/netbox-go.git

# Existing clone:
git submodule update --init
```

The compatibility harness refuses to run if the submodule is dirty or differs
from the pinned commit.

## Development checks

Use the repository-owned deterministic gate from the root. It pins the Go, Node/npm, lint, protobuf, generated-contract, frontend, and documentation checks used by CI:

```bash
make check
```

PostgreSQL, deployment, differential-oracle, and real-browser suites are separately owned integration gates; see [Testing](docs/TESTING.md). Do not use raw test counts as a compatibility metric.

To start local development, provide PostgreSQL and review [`netbox_go.yml`](netbox-backend/configs/netbox_go.yml) before running the backend. Environment variables such as `NETBOX_DATABASE_DSN`, `NETBOX_HTTP_PORT`, and `NETBOX_GRPC_PORT` override checked-in development values. When a browser frontend runs on a different origin, set the backend's `NETBOX_CORS_ALLOWED_ORIGINS` to a comma-separated list of exact HTTP(S) origins. It defaults to empty, which grants no cross-origin access; wildcard and regular-expression origins are rejected.

```bash
cd netbox-backend
make run Config=configs/netbox_go.yml

cd ../netbox-frontend
npm ci
npm run dev
```

Development databases are disposable. Startup walks a topologically ordered registry and calls GORM `AutoMigrate` only for a missing table. Existing tables are never altered, inspected for missing columns, backfilled, or repaired. Drop and recreate the development database after schema changes; versioned upgrades and backfills are intentionally deferred until production hardening.

Create the first administrator locally through protected stdin; no anonymous
provisioning endpoint exists:

```bash
cd netbox-backend
go run ./cmd/netbox_go_admin bootstrap --config configs/netbox_go.yml --username admin
```

Additional local identity administration is deliberately CLI-only. Both
commands authenticate an existing active superuser and read every password
from protected stdin; passwords are never accepted as flags or positional
arguments:

```bash
go run ./cmd/netbox_go_admin create-user \
  --config configs/netbox_go.yml \
  --actor-username admin \
  --username operator \
  --email operator@example.invalid

go run ./cmd/netbox_go_admin grant-permission \
  --config configs/netbox_go.yml \
  --actor-username admin \
  --username operator \
  --permission dcim.view_site
```

The current CLI can create a non-superuser and grant one global
`view`/`add`/`change`/`delete` model permission. Broader user, group, and
object-permission administration remains deferred.

## Container development

The root Compose stack lets PostgreSQL create the configured database and stores it in the `postgres_data` named volume. It mounts no initialization SQL: the standalone Go process owns development schema bootstrap. Published ports bind to `127.0.0.1` by default and can be changed with `NETBOX_POSTGRES_PORT`, `NETBOX_HTTP_PORT`, and `NETBOX_GRPC_PORT`. Redis and an auxiliary diagnostics listener are not part of the supported stack.

```bash
docker compose up --build --wait
docker compose down
```

Development data survives an ordinary `docker compose down`. To exercise the agreed disposable-database workflow after a schema change, remove the named volume explicitly:

```bash
docker compose down --volumes
```

The separately owned deployment check builds a fresh stack under a unique Compose project, verifies the PostgreSQL volume and application-owned schema bootstrap, restarts the application to exercise idempotence, and always removes its containers and volumes:

```bash
make deployment-smoke
```

See verification checkpoint V3 in [docs/TESTING.md](docs/TESTING.md) for the exact assertions.

## License

NetBox Go is licensed under the [Apache License 2.0](LICENSE). The upstream
oracle submodule remains separately covered by the
[NetBox license](netbox/LICENSE.txt). Third-party source retains its upstream
license and attribution requirements recorded in
[Third-party notices](THIRD_PARTY_NOTICES.md).
