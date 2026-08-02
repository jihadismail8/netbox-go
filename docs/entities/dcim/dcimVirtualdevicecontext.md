# VirtualDeviceContext

> Module: `dcim` | Table: `dcim_virtualdevicecontext` | Python class: `VirtualDeviceContext` | File: `dcim/models/devices.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/dcim/virtual-device-contexts/`

## Implementation Status

- [ ] Go model (`internal/model/dcimVirtualdevicecontext.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimVirtualdevicecontext.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimVirtualdevicecontext.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimVirtualdevicecontext.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimVirtualdevicecontext.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimVirtualdevicecontext.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
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

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `vdcs` |
| `primary_ip4` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` (none) |
| `primary_ip6` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` (none) |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `vdcs` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `status` | CharField(50) | choices=VirtualDeviceContextStatusChoices; default=active |
| `identifier` | PositiveSmallIntegerField | null=True; unique within device |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Constraints

- UniqueConstraint: `(device, name)`

## Referenced By

- [Interface](./dcimInterface.md) via `vdcs` (M2M, related_name=`interfaces`)

## Notes

- **Python source:** `dcim/models/devices.py`
- Represents a VRF-like virtual routing/forwarding context on a physical device
- Interfaces can be assigned to VDCs via M2M relationship

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `device` set, name unique within device.
2. Save.

### UPDATE
Standard flow.

### DELETE
Change log + event.
