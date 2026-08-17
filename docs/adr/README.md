# Architecture Decisions

These records contain accepted, cross-cutting decisions for the rewrite. Current implementation may not yet satisfy a decision; use [Project status](../STATUS.md) for that distinction.

| Decision                                                                                                           | Status   | Summary                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------ | -------- | ------------------------------------------------------------------------------------------------------------------- |
| [0001 — Core REST compatibility and equal gRPC interface](0001-netbox-compatibility-and-interface-parity.md)       | Accepted | Pin one source snapshot, preserve its in-scope REST contract, and expose semantic gRPC parity through a shared core |
| [0002 — Out-of-process extension services](0002-replace-python-plugins-with-extension-services.md)                 | Accepted | Exclude the Python plugin/script/report runtime and integrate extensions through supported contracts                |
| [0003 — Unified authentication and authorization](0003-unified-authentication-and-authorization.md)                | Accepted | Resolve browser sessions, REST tokens, and gRPC bearer credentials to one principal and RBAC path                   |
| [0004 — Immutable transitional generated scaffolding](0004-generated-scaffolding-is-immutable-and-transitional.md) | Accepted | Keep business logic out of generated files and replace the independent generated paths by vertical slice            |
| [0005 — Retire dormant Sponge HTTP wrappers](0005-retire-dormant-sponge-http-wrappers.md)                          | Accepted | Remove the exact dormant handler/router wrapper pairs while retaining the frozen lower legacy stacks                |

Add an ADR only for a durable decision that constrains multiple parts of the system. Keep reversible development-phase tactics, such as disposable database resets, in [Architecture](../ARCHITECTURE.md) or the [Roadmap](../ROADMAP.md).
