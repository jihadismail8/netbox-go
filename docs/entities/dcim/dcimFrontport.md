# FrontPort

> Module: `dcim` | Table: `dcim_frontport` | Python class: `FrontPort` | File: `dcim/models/device_components.py`

**Inheritance:** `ModularComponentModel <- CabledObjectModel <- TrackingModelMixin`

**REST URL:** `/api/dcim/front-ports/`

## Implementation Status

- [ ] Go model (`internal/model/dcimFrontport.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimFrontport.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimFrontport.go`)
- [ ] DAO unit tests (`internal/dao/dcimFrontport_test.go`)
- [ ] Cache layer (`internal/cache/dcimFrontport.go`)
- [ ] Cache unit tests (`internal/cache/dcimFrontport_test.go`)
- [ ] Service layer (`internal/service/dcimFrontport.go`)
- [ ] Service unit tests (`internal/service/dcimFrontport_test.go`)
- [ ] Handler layer (`internal/handler/dcimFrontport.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/front-ports/`)
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

### Foreign Keys (inherited + own)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `frontports` |
| `_site` | [Site](./dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `+` (cached) |
| `module` | [Module](./dcimModule.md) | `CASCADE` | Yes | `frontports` |
| `cable` | [Cable](./dcimCable.md) | `SET_NULL` | Yes | `+` (none) |
| `rear_port` | [RearPort](./dcimRearport.md) | `CASCADE` | No | `frontports` |

### Regular Fields (defined directly)

| Field | Type | Notes |
|-------|------|-------|
| `type` | CharField(50) | choices=PortTypeChoices; default=8p8c |
| `rear_port_position` | PositiveSmallIntegerField | default=1; min=1, max=64 |

### Fields Inherited from ComponentModel

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort" |
| `label` | CharField(64) | blank=True |
| `description` | CharField(200) | blank=True |

### Fields Inherited from CabledObjectModel

| Field | Type | Notes |
|-------|------|-------|
| `cable_end` | CharField(1) | choices=CableEndChoices; null=True |
| `mark_connected` | BooleanField | default=False |

## Inherited Fields (from base classes)

- From **NetBoxModel** (via ComponentModel): `id`, `created`, `last_updated`, `custom_field_data`, `tags` (M2M to Tag)
- From **TrackingModelMixin**: `_is_signal_during_save`

## Constraints

- UniqueConstraint: `(device, name)` (from ComponentModel)
- UniqueConstraint: `(rear_port, rear_port_position)` — each position on a rear port used once

## Dependencies

- [Device](./dcimDevice.md)
- [Module](./dcimModule.md) (optional)
- [RearPort](./dcimRearport.md)
- [Cable](./dcimCable.md) (optional)

## Referenced By

- [CableTermination](./dcimCabletermination.md) via `termination` (GenericFK)
- [InventoryItem](./dcimInventoryitem.md) via `component` (GenericFK)

## Notes

- **Python source:** `dcim/models/device_components.py`
- **Go model file:** `internal/model/dcimFrontport.go`
- **Proto file:** `api/netbox_go/v1/dcimFrontport.proto`
- **Pass-through port** — does NOT inherit PathEndpoint
- `rear_port_position` maps this front port to a specific position on the parent rear port
- Validation: `rear_port` must belong to the same device
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
