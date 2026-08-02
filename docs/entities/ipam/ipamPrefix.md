# Prefix

> Module: `ipam` | Table: `ipam_prefix` | Python class: `Prefix` | File: `ipam/models/ip.py`

**Inheritance:** `PrimaryModel <- CachedScopeMixin`

**REST URL:** `/api/ipam/prefixes/`

## Implementation Status

- [ ] Go model (`internal/model/ipamPrefix.go`)
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
| `vrf` | [VRF](./ipamVrf.md) | `CASCADE` | Yes | `prefixes` |
| `rir` | [RIR](./ipamRir.md) | `PROTECT` | Yes | `prefixes` |
| `vlan` | [VLAN](./ipamVlan.md) | `PROTECT` | Yes | `prefixes` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `prefixes` |
| `role` | [Role](./ipamRole.md) | `SET_NULL` | Yes | `prefixes` |
| `_site` | [Site](./../dcim/dcimSite.md) | `SET_NULL` | Yes | `+` (cached, from CachedScopeMixin) |
| `_region` | [Region](./../dcim/dcimRegion.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./../dcim/dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_site_group` | [SiteGroup](./../dcim/dcimSitegroup.md) | `SET_NULL` | Yes | `+` (cached) |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `prefix` | IPNetworkField | Required (custom field type) |
| `status` | CharField(50) | choices=PrefixStatusChoices; default=active |
| `is_pool` | BooleanField | default=False |
| `mark_utilized` | BooleanField | default=False |

### CachedScopeMixin Fields

| Field | Type | Notes |
|-------|------|-------|
| `scope_type` | FK to ContentType | GenericFK type |
| `scope_id` | PositiveBigIntegerField | GenericFK ID |
| `scope` | GenericFK | Polymorphic (Site, Region, Location, SiteGroup, etc.) |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `tags`

## Referenced By

- [IPAddress](./ipamIpaddress.md) via `prefix` (FK)
- [IPRange](./ipamIprange.md) via `prefix` (FK)

## Notes

- **Python source:** `ipam/models/ip.py`
- Uses `CachedScopeMixin` for polymorphic site/region/location assignment
- `_site`, `_region`, `_location`, `_site_group` are cached from `scope` for efficient filtering

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** E (IPAM Hierarchy), F (Scoped)

### CREATE
1. `clean()` validates: status valid; scope type allowed; no duplicate (prefix, vrf, scope) if `is_pool=False`.
2. `save()` computes `vrf`, `tenant` overrides.
3. Save.
4. **Signal `cache_prefix_hierarchy()`:** Recomputes `_depth`/`_children` for all prefixes in VRF+scope.
5. **Counter increment:** VRF, Site, Tenant counters.
6. Change log + event.

### UPDATE
1. Snapshot. `clean()` re-validates.
3. Save.
4. **If prefix (network) changed:** All `_depth`/`_children` recalculated for entire hierarchy.
5. **If scope changed:** `_region`, `_site_group` re-cached.
6. Change log + event.

### DELETE
1. **Counter decrement:** VRF, Site, Tenant.
2. **Signal `cache_prefix_hierarchy()`:** Sibling/parent `_children` and child `_depth` recalculated.
3. IPAddresses are NOT deleted (no direct FK).
4. Change log + event.

### Interdependencies
- **Hierarchy fields:** `_depth` (parent count), `_children` (direct child count) — recalculated on any prefix CRUD.
- **Scope cache:** `_region`, `_site_group` from `scope` GenericFK.
- **Counter source for:** VRF, Site, Tenant.
