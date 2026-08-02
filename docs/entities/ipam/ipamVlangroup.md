# VLANGroup

> Module: `ipam` | Table: `ipam_vlangroup` | Python class: `VLANGroup` | File: `ipam/models/vlans.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/vlan-groups/`

## Implementation Status

- [ ] Go model (`internal/model/ipamVlangroup.go`)
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
| `name` | CharField(100) | Required |
| `slug` | SlugField(100) | Required |
| `min_vid` | PositiveIntegerField | default=VLAN_VID_MIN; min=1, max=4094 |
| `max_vid` | PositiveIntegerField | default=VLAN_VID_MAX; min=1, max=4094 |

### Scope (GenericFK)

| Field | Related Model | Via |
|-------|---------------|-----|
| `scope` | (polymorphic: Site, Region, Location, SiteGroup, Cluster, etc.) | `scope_type` + `scope_id` |

## Referenced By

- [VLAN](./ipamVlan.md) via `group` (FK)

## Notes

- **Python source:** `ipam/models/vlans.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source)

### CREATE
1. `clean()` validates scope type if scope set.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** Child VLANGroups.
2. **SET_NULL:** VLANs (`vlan.group=null`).

### Interdependencies
- **Counter fields:** `vlan_count`.
