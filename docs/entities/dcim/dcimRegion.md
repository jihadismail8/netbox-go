# Region

> Module: `dcim` | Table: `dcim_region` | Python class: `Region` | File: `dcim/models/sites.py`

**Inheritance:** `ContactsMixin <- NestedGroupModel`

**REST URL:** `/api/dcim/regions/`

## Implementation Status

- [ ] Go model (`internal/model/dcimRegion.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimRegion.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimRegion.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimRegion.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimRegion.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimRegion.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/regions/`)
- [ ] Response envelope compatible
- [ ] Bulk operations
- [ ] Filtering support
- [ ] Pagination support
- [ ] RBAC / permissions
- [ ] API integration tests
- [ ] Vue.js list view
- [ ] Vue.js detail view
- [ ] Vue.js create/edit form
- [ ] Vue.js delete confirmation
- [ ] E2E test

## Django Model Fields (from Python source)

### Foreign Keys (1)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `parent` | self (Region) | `CASCADE` | Yes | `children` |

**Note:** `parent` is a `TreeForeignKey` (self-referential, supports tree hierarchy).

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | CharField(100) | Required, unique |
| `description` | CharField(200) | |

## Inherited Fields

- From **NestedGroupModel**: `id`, `created`, `last_updated`, `custom_field_data`, `name`, `slug`, `description`, `tags` (M2M), `parent` (self-ref TreeForeignKey)
- From **ContactsMixin**: Contact assignments (reverse from `tenancy.ContactAssignment`)

## Dependencies

No external model dependencies (self-referential only).

## Referenced By

- [Site](./dcimSite.md) via `region` (FK, related_name=`sites`)
- [Location](./dcimLocation.md) via `_region` (FK, cached)
- [Device](./dcimDevice.md) via `_region` (FK, cached)
- [Rack](./dcimRack.md) via `_region` (FK, cached)
- [PowerPanel](./dcimPowerpanel.md) via `_region` (FK, cached)
- [CircuitTermination](./../circuits/circuitsCircuittermination.md) via `_region` (FK, cached)
- [CableTermination](./dcimCabletermination.md) via `_region` (FK, cached)
- [Prefix](./../ipam/ipamPrefix.md) via `scope` (GenericFK)
- [ConfigContext](./../extras/extrasConfigcontext.md) via `regions` (M2M)
- [Cluster](./../virtualization/virtualizationCluster.md) via `_region` (FK, cached)

## Notes

- **Python source:** `dcim/models/sites.py`
- **Go model file:** `internal/model/dcimRegion.go`
- **Proto file:** `api/netbox_go/v1/dcimRegion.proto`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates parent region exists.
2. MPTT tree fields (`lft`, `rgt`, `tree_id`, `level`) computed.
3. Save.

### UPDATE
1. Snapshot.
2. If `parent` changed: MPTT tree re-balanced (all descendants reindexed).
3. Save.

### DELETE
1. **CASCADE:** Sub-regions.
2. **SET_NULL:** Sites (`site.region=null`).
3. MPTT tree re-balanced.
4. Change log + event.

### Interdependencies
- MPTT hierarchy via `parent` FK. Tree in `lft`, `rgt`, `tree_id`, `level`.
- Referenced by Site (`region`), Prefix/VLAN/Cluster (cached `_region`).
