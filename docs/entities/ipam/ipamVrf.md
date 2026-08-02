# VRF

> Module: `ipam` | Table: `ipam_vrf` | Python class: `VRF` | File: `ipam/models/vrfs.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/vrfs/`

## Implementation Status

- [ ] Go model (`internal/model/ipamVrf.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/ipamVrf.proto`)
- [ ] Proto generated code
- [ ] DAO layer
- [ ] DAO unit tests
- [ ] Cache layer
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
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `vrfs` |

### ManyToMany Fields

| Field | Related Model | through | related_name |
|-------|---------------|---------|--------------|
| `import_targets` | [RouteTarget](./ipamRoutetarget.md) | `(auto)` | `importing_vrfs` |
| `export_targets` | [RouteTarget](./ipamRoutetarget.md) | `(auto)` | `exporting_vrfs` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required; db_collation="natural_sort" |
| `rd` | CharField(VRF_RD_MAX_LENGTH) | unique=True; null=True (route distinguisher) |
| `enforce_unique` | BooleanField | default=True |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Referenced By

- [Prefix](./ipamPrefix.md) via `vrf` (FK)
- [IPRange](./ipamIprange.md) via `vrf` (FK)
- [IPAddress](./ipamIpaddress.md) via `vrf` (FK)
- [Interface](./../dcim/dcimInterface.md) via `vrf` (FK)
- [VMInterface](./../virtualization/virtualizationVminterface.md) via `vrf` (FK)

## Notes

- **Python source:** `ipam/models/vrfs.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Prefixes (`prefix.vrf=null`), IPAddresses (`ipaddress.vrf=null`).
2. Change log + event.

### Interdependencies
- **Counter fields:** `prefix_count`, `ipaddress_count`.
