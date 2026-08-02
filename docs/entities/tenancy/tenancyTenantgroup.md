# TenantGroup

> Module: `tenancy` | Table: `tenancy_tenantgroup` | Python class: `TenantGroup` | File: `tenancy/models/tenants.py`

**Inheritance:** `NestedGroupModel`

**REST URL:** `/api/tenancy/tenant-groups/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyTenantgroup.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `parent` | [TenantGroup](./tenancyTenantgroup.md) (self) | `SET_NULL` | Yes | `children` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |
| `description` | CharField(200) | blank=True |

## Referenced By

- [Tenant](./tenancyTenant.md) via `group` (FK)

## Notes

- **Python source:** `tenancy/models/tenants.py`
- Uses tree structure (parent self-ref FK) for hierarchical grouping

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates parent group exists.
2. MPTT tree computed.
3. Save.

### UPDATE
1. If `parent` changed: MPTT re-balanced.
2. Save.

### DELETE
1. **CASCADE:** Sub-groups.
2. **SET_NULL:** Tenants (`tenant.group=null`).
3. MPTT re-balanced.
