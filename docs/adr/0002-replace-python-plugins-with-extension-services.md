---
status: accepted
---

# Replace Python plugins with out-of-process extension services

NetBox Go will not execute or emulate NetBox's Python plugins. The standalone Go runtime will expose stable REST, gRPC, webhook, and event contracts through which independently deployed extension services can integrate. We accept that “drop-in-compatible” therefore covers the pinned NetBox core rather than its Python plugin runtime; embedding Python would violate the standalone constraint, while inventing a second in-process plugin ABI would add substantial coupling before the core rewrite is compatible.

## Consequences

- Python plugin installation, imports, templates, models, hooks, custom scripts, and reports are explicit non-goals.
- A plugin-dependent deployment needs a separately planned replacement for each required plugin.
- Script/report automation must be redesigned as an Extension Service or external job; the core runtime will not execute user-supplied Python.
- Extension contracts must enforce the same authentication, authorization, and compatibility guarantees as other external consumers.
- An extension cannot bypass the shared application/domain layer by writing directly to the database.
