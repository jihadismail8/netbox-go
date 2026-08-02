# IPAddress

> Module: `ipam` | Table: `ipam_ipaddress` | Python class: `IPAddress` | File: `ipam/models/ip.py`

**Inheritance:** `PrimaryModel <- CachedScopeMixin`

**REST URL:** `/api/ipam/ip-addresses/`

## Implementation Status

- [ ] Go model (`internal/model/ipamIpaddress.go`)
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
| `vrf` | [VRF](./ipamVrf.md) | `CASCADE` | Yes | `ip_addresses` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `ip_addresses` |
| `status` | CharField(50) | — | — | choices=IPAddressStatusChoices; default=active |
| `role` | CharField(50) | — | — | choices=IPAddressRoleChoices; null=True |
| `nat_inside` | [IPAddress](./ipamIpaddress.md) (self) | `SET_NULL` | Yes | `nat_outside` |
| `_site` | [Site](./../dcim/dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_region` | [Region](./../dcim/dcimRegion.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./../dcim/dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_site_group` | [SiteGroup](./../dcim/dcimSitegroup.md) | `SET_NULL` | Yes | `+` (cached) |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `address` | IPAddressField | Required (IP + mask) |
| `dns_name` | CharField(255) | blank=True |

### Generic FK (assignment)

| Field | Related Model | Via |
|-------|---------------|-----|
| `assigned_object` | (polymorphic: Interface, VMInterface, FHRPGroup) | `assigned_object_type` + `assigned_object_id` |

## Referenced By

- [Device](./../dcim/dcimDevice.md) via `primary_ip4`/`primary_ip6`/`oob_ip` (O2O)
- [VirtualMachine](./../virtualization/virtualizationVirtualmachine.md) via `primary_ip4`/`primary_ip6` (O2O)
- [IPAddress](./ipamIpaddress.md) via `nat_inside` (self-ref FK)
- [Interface](./../dcim/dcimInterface.md) via `assigned_object` (GenericFK)

## Notes

- **Python source:** `ipam/models/ip.py`
- `nat_inside` creates NAT mapping between addresses
- `assigned_object` is GenericFK to Interface, VMInterface, or FHRPGroup

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** E (IPAM)

### CREATE
1. `clean()` validates: no duplicate (address, vrf, tenant) if `role!='anycast'`; assigned VRF matches interface VRF.
2. `save()` sets `assigned_object` GFK.
3. Save.
4. **Counter increment:** VRF, Tenant, Interface `ip_address_count`.
5. Change log + event.

### UPDATE
1. Snapshot.
2. If `assigned_object` changed: old interface counter decremented, new incremented.
3. Save.
4. Change log + event.

### DELETE
1. **Primary IP cleanup:** If this is a Device's `primary_ip4`/`primary_ip6`, that FK is SET_NULL.
2. **Counter decrement:** VRF, Tenant, Interface.
3. Change log + event.

### Interdependencies
- **Assignment:** `assigned_object` GenericFK to Interface or VMInterface.
- **Primary IP:** Referenced by Device.primary_ip4/primary_ip6 (SET_NULL on delete).
- **Counter source for:** VRF, Tenant, Interface.
