# Device Role

> Module: `dcim` | Table: `dcim_devicerole` | Python class: `DeviceRole` | File: `dcim/models/devices.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/device-roles/`

## Implementation Status

- [ ] Go model (`internal/model/dcimDevicerole.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimDevicerole.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimDevicerole.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimDevicerole.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimDevicerole.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimDevicerole.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/device-roles/`)
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

### Foreign Keys (0)

No direct foreign keys.

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | CharField(100) | Required, unique |
| `color` | ColorField | Required |
| `vm_role` | BooleanField | Default True |
| `config_template` | FK to ConfigTemplate (on_delete=PROTECT) |
| `description` | CharField(200) | |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments

## Dependencies

- [ConfigTemplate](./../extras/extrasConfigtemplate.md) (optional, via `config_template` FK)

## Referenced By

- [Device](./dcimDevice.md) via `role` (FK, on_delete=PROTECT, related_name=`devices`)
- [VirtualMachine](./../virtualization/virtualizationVirtualmachine.md) via `role` (FK, on_delete=PROTECT, related_name=`virtual_machines`)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimDevicerole.go`
- **Proto file:** `api/netbox_go/v1/dcimDevicerole.proto`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Devices and VirtualMachines prevent deletion.
