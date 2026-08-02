# API documentation

The legacy 125-resource Vue/protobuf catalogue was retired because it mixed
disabled scaffolding, deferred bulk operations, and unsupported routes with the
public contract.

Use these authoritative sources instead:

- [Core workflow REST contract](../contracts/core-workflow-v1.md)
- [Canonical gRPC v1 inventory](../contracts/grpc-v1.md)
- [Compatibility contract](../COMPATIBILITY.md)
- [Current evidence ledger](../STATUS.md)
- [Canonical OpenAPI source](../../netbox-backend/api/openapi/netbox-go-v1.yaml)

The versioned machine-readable capability profile and inventories live under
[`contracts/netbox/v4.4.6-post7/`](../../contracts/netbox/v4.4.6-post7/). The
checked-in Python NetBox tree is a development oracle only; it is not a runtime
or build dependency.
