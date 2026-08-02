# Role

> Module: `ipam` | Table: `ipam_role` | Python class: `Role` | File: `ipam/models/ip.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/ipam/roles/`

## Implementation Status

- [ ] Go model (`internal/model/ipamRole.go`)
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
| `weight` | PositiveIntegerField | default=1000 |

## Inherited Fields

- From **OrganizationalModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `tags`

## Referenced By

- [Prefix](./ipamPrefix.md) via `role` (FK)
- [VLAN](./ipamVlan.md) via `role` (FK)

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
1. **SET_NULL:** Prefixes (`prefix.role=null`).
