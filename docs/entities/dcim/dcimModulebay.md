# ModuleBay

> Module: `dcim` | Table: `dcim_modulebay` | Python class: `ModuleBay` | File: `dcim/models/device_components.py`

**Inheritance:** `ComponentModel <- TrackingModelMixin`

**REST URL:** `/api/dcim/module-bays/`

## Implementation Status

- [ ] Go model (`internal/model/dcimModulebay.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimModulebay.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimModulebay.go`)
- [ ] DAO unit tests (`internal/dao/dcimModulebay_test.go`)
- [ ] Cache layer (`internal/cache/dcimModulebay.go`)
- [ ] Cache unit tests (`internal/cache/dcimModulebay_test.go`)
- [ ] Service layer (`internal/service/dcimModulebay.go`)
- [ ] Service unit tests (`internal/service/dcimModulebay_test.go`)
- [ ] Handler layer (`internal/handler/dcimModulebay.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/module-bays/`)
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
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `modulebays` |
| `_site` | [Site](./dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `+` (cached) |

### Fields Inherited from ComponentModel

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort" |
| `label` | CharField(64) | blank=True |
| `description` | CharField(200) | blank=True |

### Position Field (defined directly)

| Field | Type | Notes |
|-------|------|-------|
| `position` | CharField(30) | blank=True; optional slot identifier |

## Constraints

- UniqueConstraint: `(device, name)` (from ComponentModel)
- UniqueConstraint: `(device, position)` when position is set

## Dependencies

- [Device](./dcimDevice.md) (parent device)

## Referenced By

- [Module](./dcimModule.md) via `parent_module_bay` (reverse O2O)

## Notes

- **Python source:** `dcim/models/device_components.py`
- Does NOT inherit ModularComponentModel, CabledObjectModel, or PathEndpoint
- Module bays house Module instances (hot-swappable hardware modules)

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** C (Cache Consumer)

### CREATE
1. `clean()` validates `device` set, `position` unique.
2. Cache populated from Device.
3. Save. Device counter incremented.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Module (`module.module_bay=null`).
2. Device counter decremented.
