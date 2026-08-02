---
status: accepted
---

# Target core REST compatibility with an equal gRPC interface

NetBox Go will be drop-in compatible for the declared in-scope core REST contract of the checked-in source snapshot at commit `fbb948d30e79ce657fac62994a22aca72c1770a9` (`v4.4.6-7-gfbb948d30`). gRPC is an added first-class interface with semantic parity over the same application and domain implementation; supported operations must have equivalent validation, authorization, side effects, transaction boundaries, and errors, while transport-specific message shapes may differ. We rejected both a frontend-only compatibility target and independent REST/gRPC business implementations because either would permit behavioral drift and make the replacement unsafe to embed in a larger system.

## Consequences

- REST compatibility is measured against upstream NetBox behavior; gRPC parity is measured against the same domain scenarios.
- Transport handlers perform translation and protocol concerns only. They do not query or mutate persistence directly and do not own domain rules.
- A capability is not complete until both interfaces exercise the shared behavior and their contract tests pass.
- Capabilities absent from, or intentionally incompatible with, upstream REST may be exposed through REST and/or gRPC only as explicitly documented extensions. They are tested separately and never counted as baseline compatibility; when both interfaces expose one, they must retain semantic parity through the shared core.
- The existing generated, table-oriented RPCs are unstable scaffolding and are not a public compatibility commitment. They may coexist temporarily during replacement but can be removed without a compatibility adapter.
- New gRPC contracts use versioned packages such as `netbox.dcim.v1` and `netbox.ipam.v1`. Once one of those contracts is declared published, compatible evolution is additive and a breaking change requires a new API version.
- The deployed application must not depend on Python, Django, or the upstream source tree. The pinned source is a development reference and compatibility oracle only.
- GraphQL is not a supported public interface. The supported interface set is NetBox-compatible HTTPS REST plus the parity gRPC API.
