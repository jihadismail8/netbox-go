# RIR

> Module: `ipam` | Table: `ipam_rir` | Python class: `RIR` | File: `ipam/models/ip.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/ipam/rirs/`

## Implementation Status

- [ ] Go model (`internal/model/ipamRir.go`)
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

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |
| `is_private` | BooleanField | default=False |

## Inherited Fields

- From **OrganizationalModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `tags`

## Referenced By

- [Aggregate](./ipamAggregate.md) via `rir` (FK)
- [ASN](./ipamAsn.md) via `rir` (FK)
- [ASNRange](./ipamAsnrange.md) via `rir` (FK)

## Notes

- **Python source:** `ipam/models/ip.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Aggregates (`aggregate.rir=null`).
