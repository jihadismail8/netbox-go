# DeviceBay

> Module: `dcim` | Table: `dcim_devicebay` | Python class: `DeviceBay` | File: `dcim/models/device_components.py`

**Inheritance:** `ComponentModel <- TrackingModelMixin`

**REST URL:** `/api/dcim/device-bays/`

## Implementation Status

- [ ] Go model (`internal/model/dcimDevicebay.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimDevicebay.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimDevicebay.go`)
- [ ] DAO unit tests (`internal/dao/dcimDevicebay_test.go`)
- [ ] Cache layer (`internal/cache/dcimDevicebay.go`)
- [ ] Cache unit tests (`internal/cache/dcimDevicebay_test.go`)
- [ ] Service layer (`internal/service/dcimDevicebay.go`)
- [ ] Service unit tests (`internal/service/dcimDevicebay_test.go`)
- [ ] Handler layer (`internal/handler/dcimDevicebay.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/device-bays/`)
- [ ] Response envelope compatible
- [ ] Bulk operations (create/update/delete)
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
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `devicebays` |
| `_site` | [Site](./dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `+` (cached) |

### OneToOne Fields

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `installed_device` | [Device](./dcimDevice.md) | `SET_NULL` | Yes | `parent_bay` |

### Fields Inherited from ComponentModel

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort" |
| `label` | CharField(64) | blank=True |
| `description` | CharField(200) | blank=True |

## Inherited Fields (from base classes)

- From **NetBoxModel** (via ComponentModel): `id`, `created`, `last_updated`, `custom_field_data`, `tags` (M2M to Tag)
- From **TrackingModelMixin**: `_is_signal_during_save`

## Constraints

- UniqueConstraint: `(device, name)` (from ComponentModel)

## Dependencies

- [Device](./dcimDevice.md) (parent device)

## Referenced By

- [Device](./dcimDevice.md) via `parent_bay` (reverse O2O from `installed_device`)

## Notes

- **Python source:** `dcim/models/device_components.py`
- **Go model file:** `internal/model/dcimDevicebay.go`
- **Proto file:** `api/netbox_go/v1/dcimDevicebay.proto`
- Does NOT inherit ModularComponentModel, CabledObjectModel, or PathEndpoint
- Device bays are only applicable to parent device types (subdevice_role=parent)
- `installed_device` is the child device housed within this bay
- Validation: installed device's parent must be the bay's device
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected), C (Cache Consumer)

### CREATE
1. `clean()` validates `device` is set. Components **cannot** be moved to a different device after creation.
2. Denormalized cache (`_site`, `_location`, `_rack`) populated from parent Device.
3. Save.
4. Parent Device counter incremented (e.g., `Device.interface_count`).
5. Change log + event.

### UPDATE
1. Snapshot.
2. `clean()` validates device hasn't changed (immutable).
3. Save.
4. Change log + event.

### DELETE
1. **Cable disconnect:** If cable connected, cable's termination reference removed. CablePath recomputed.
2. Parent Device counter decremented.
3. Change log + event.

### Interdependencies
- **Cache consumer:** `_site`, `_location`, `_rack` sourced from parent Device.
- **Cable connectivity:** Each component can have a `cable` FK. Path tracing originates here.
- **Parent counter:** Device's `<component>_count` tracks children.
