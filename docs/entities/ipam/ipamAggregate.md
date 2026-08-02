# Aggregate

> Module: `ipam` | Table: `ipam_aggregate` | Python class: `Aggregate` | File: `ipam/models/ip.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/aggregates/`

## Implementation Status

- [ ] Go model (`internal/model/ipamAggregate.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition
- [ ] Proto generated code
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
- [ ] Response envelope compatible
- [ ] Bulk operations
- [ ] Filtering support
- [ ] Pagination support
- [ ] RBAC / permissions
- [ ] API integration tests
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `rir` | [RIR](./ipamRir.md) | `PROTECT` | No | `aggregates` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `aggregates` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `prefix` | IPNetworkField | Required (custom field type) |
| `date_added` | DateField | null=True |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `tags`

## Constraints

- UniqueConstraint: `prefix`, `rir` (from model clean)

## Notes

- **Python source:** `ipam/models/ip.py`
- `IPNetworkField` is a custom field storing both network address and prefix length

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `rir` set, prefix valid.
2. Save.

### UPDATE
Standard flow.

### DELETE
Change log + event.
