# SiteGroup

> Module: `dcim` | Table: `dcim_sitegroup` | Python class: `SiteGroup` | File: `dcim/models/sites.py`

**Inheritance:** `ContactsMixin <- NestedGroupModel`

**REST URL:** `/api/dcim/site-groups/`

## Implementation Status

- [ ] Go model (`internal/model/dcimSitegroup.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimSitegroup.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimSitegroup.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimSitegroup.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimSitegroup.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimSitegroup.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/site-groups/`)
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
| `parent` | self (SiteGroup) | `CASCADE` | Yes | `children` |

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

- [Site](./dcimSite.md) via `group` (FK, related_name=`sites`)
- [Location](./dcimLocation.md) via `_site_group` (FK, cached)
- [CircuitTermination](./../circuits/circuitsCircuittermination.md) via `_site_group` (FK, cached)
- [Prefix](./../ipam/ipamPrefix.md) via `scope` (GenericFK)
- [ConfigContext](./../extras/extrasConfigcontext.md) via `site_groups` (M2M)

## Notes

- **Python source:** `dcim/models/sites.py`
- **Go model file:** `internal/model/dcimSitegroup.go`
- **Proto file:** `api/netbox_go/v1/dcimSitegroup.proto`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates parent group exists.
2. MPTT tree fields computed.
3. Save.

### UPDATE
1. Snapshot. If `parent` changed: MPTT re-balanced.
3. Save.

### DELETE
1. **CASCADE:** Sub-groups.
2. **SET_NULL:** Sites (`site.group=null`).
3. MPTT re-balanced.
