# IPRange

> Module: `ipam` | Table: `ipam_iprange` | Python class: `IPRange` | File: `ipam/models/ip.py`

**Inheritance:** `PrimaryModel <- CachedScopeMixin`

**REST URL:** `/api/ipam/ip-ranges/`

## Implementation Status

- [ ] Go model (`internal/model/ipamIprange.go`)
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
| `vrf` | [VRF](./ipamVrf.md) | `CASCADE` | Yes | `ip_ranges` |
| `role` | [Role](./ipamRole.md) | `SET_NULL` | Yes | `ip_ranges` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `ip_ranges` |
| `_site` | [Site](./../dcim/dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_region` | [Region](./../dcim/dcimRegion.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./../dcim/dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_site_group` | [SiteGroup](./../dcim/dcimSitegroup.md) | `SET_NULL` | Yes | `+` (cached) |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `start_address` | IPAddressField | Required (IP range start) |
| `end_address` | IPAddressField | Required (IP range end) |
| `status` | CharField(50) | choices=IPRangeStatusChoices; default=active |
| `mark_utilized` | BooleanField | default=False |

## Notes

- **Python source:** `ipam/models/ip.py`
- Uses `CachedScopeMixin` for polymorphic site/region/location assignment

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** E (IPAM)

### CREATE
1. `clean()` validates start < end address.
2. `save()` sets `vrf`.
3. Save.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
