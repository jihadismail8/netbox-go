# Platform

> Module: `dcim` | Table: `dcim_platform` | Python class: `Platform` | File: `dcim/models/devices.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/platforms/`

## Implementation Status

- [ ] Go model (`internal/model/dcimPlatform.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimPlatform.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimPlatform.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimPlatform.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimPlatform.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimPlatform.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/platforms/`)
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
| `manufacturer` | [Manufacturer](./dcimManufacturer.md) | `PROTECT` | Yes | `platforms` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | CharField(100) | Required, unique |
| `python_class` | CharField(200) | |
| `description` | CharField(200) | |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments

## Dependencies

- [Manufacturer](./dcimManufacturer.md)

## Referenced By

- [Device](./dcimDevice.md) via `platform` (FK, on_delete=SET_NULL)
- [VirtualMachine](./../virtualization/virtualizationVirtualmachine.md) via `platform` (FK, on_delete=SET_NULL)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimPlatform.go`
- **Proto file:** `api/netbox_go/v1/dcimPlatform.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Devices and VirtualMachines prevent deletion.
