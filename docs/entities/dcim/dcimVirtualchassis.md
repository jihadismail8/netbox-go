# Virtual Chassis

> Module: `dcim` | Table: `dcim_virtualchassis` | Python class: `VirtualChassis` | File: `dcim/models/devices.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/virtual-chassis/`

## Implementation Status

- [ ] Go model (`internal/model/dcimVirtualchassis.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimVirtualchassis.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimVirtualchassis.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimVirtualchassis.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimVirtualchassis.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimVirtualchassis.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/virtual-chassis/`)
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
| `name` | CharField(64) | |
| `domain` | CharField(30) | |
| `description` | CharField(200) | |
| `master` | OneToOneField to Device (on_delete=SET_NULL) | Reverse: `virtual_chassis` |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments

## Dependencies

No external model dependencies (`master` is a reverse O2O from Device).

## Referenced By

- [Device](./dcimDevice.md) via `virtual_chassis` (FK, on_delete=SET_NULL, related_name=`members`)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimVirtualchassis.go`
- **Proto file:** `api/netbox_go/v1/dcimVirtualchassis.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Devices (`device.virtual_chassis=null`, `vc_position=null`).
