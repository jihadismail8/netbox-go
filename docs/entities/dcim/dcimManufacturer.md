# Manufacturer

> Module: `dcim` | Table: `dcim_manufacturer` | Python class: `Manufacturer` | File: `dcim/models/devices.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/manufacturers/`

## Implementation Status

- [ ] Go model (`internal/model/dcimManufacturer.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimManufacturer.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimManufacturer.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimManufacturer.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimManufacturer.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimManufacturer.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/manufacturers/`)
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
| `description` | CharField(200) | |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments (reverse from `tenancy.ContactAssignment`)

## Dependencies

No external model dependencies.

## Referenced By

- [DeviceType](./dcimDevicetype.md) via `manufacturer` (FK, on_delete=PROTECT)
- [ModuleType](./dcimModuletype.md) via `manufacturer` (FK, on_delete=PROTECT)
- [Platform](./dcimPlatform.md) via `manufacturer` (FK, on_delete=PROTECT)
- [RackType](./dcimRacktype.md) via `manufacturer` (FK, on_delete=PROTECT)
- [InventoryItemTemplate](./dcimInventoryitemtemplate.md) via `manufacturer` (FK, on_delete=PROTECT)
- [InventoryItem](./dcimInventoryitem.md) via `manufacturer` (FK, on_delete=PROTECT)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimManufacturer.go`
- **Proto file:** `api/netbox_go/v1/dcimManufacturer.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** DeviceTypes and Platforms prevent deletion.
2. Change log + event.

### Interdependencies
- **Counter fields:** `device_type_count`.
