# RouteTarget

> Module: `ipam` | Table: `ipam_routetarget` | Python class: `RouteTarget` | File: `ipam/models/vrfs.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/route-targets/`

## Implementation Status

- [ ] Go model (`internal/model/ipamRoutetarget.go`)
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
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `route_targets` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(ROUTE_TARGET_MAX_LENGTH) | Required, unique |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Referenced By

- [VRF](./ipamVrf.md) via `import_targets`/`export_targets` (M2M)

## Notes

- **Python source:** `ipam/models/vrfs.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow. No side effects.

### UPDATE
Standard flow. No side effects.

### DELETE
No downstream effects.
