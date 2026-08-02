# VLAN

> Module: `ipam` | Table: `ipam_vlan` | Python class: `VLAN` | File: `ipam/models/vlans.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/vlans/`

## Implementation Status

- [ ] Go model (`internal/model/ipamVlan.go`)
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
| `site` | [Site](./../dcim/dcimSite.md) | `PROTECT` | Yes | `vlans` |
| `group` | [VLANGroup](./ipamVlangroup.md) | `PROTECT` | Yes | `vlans` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `vlans` |
| `role` | [Role](./ipamRole.md) | `SET_NULL` | Yes | `vlans` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `vid` | PositiveIntegerField | Required; min=1, max=4094 |
| `name` | CharField(64) | Required |
| `status` | CharField(50) | choices=VLANStatusChoices; default=active |

## Referenced By

- [Prefix](./ipamPrefix.md) via `vlan` (FK)
- [Interface](./../dcim/dcimInterface.md) via `untagged_vlan`, `tagged_vlans`, `qinq_svlan` (FK/M2M)

## Notes

- **Python source:** `ipam/models/vlans.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), F (Scoped)

### CREATE
1. `clean()` validates unique (group, vid) if group set.
2. Save.
3. Counter increment: VLANGroup, Site, Tenant.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Interfaces (`untagged_vlan`/`tagged_vlans`), Prefixes (`vlan`) prevent deletion.
2. Counter decrement: VLANGroup, Site, Tenant.
3. Change log + event.
