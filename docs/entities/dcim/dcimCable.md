# Cable

> Module: `dcim` | Table: `dcim_cable` | Python class: `Cable` | File: `dcim/models/cables.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/dcim/cables/`

## Implementation Status

- [ ] Go model (`internal/model/dcimCable.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimCable.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimCable.go`)
- [ ] DAO unit tests (`internal/dao/dcimCable_test.go`)
- [ ] Cache layer (`internal/cache/dcimCable.go`)
- [ ] Cache unit tests (`internal/cache/dcimCable_test.go`)
- [ ] Service layer (`internal/service/dcimCable.go`)
- [ ] Service unit tests (`internal/service/dcimCable_test.go`)
- [ ] Handler layer (`internal/handler/dcimCable.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/cables/`)
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

### Foreign Keys (1)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `cables` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `type` | CharField(50) | choices=CableTypeChoices; null=True |
| `status` | CharField(50) | choices=LinkStatusChoices; default=connected |
| `label` | CharField(100) | blank=True |
| `color` | ColorField | blank=True |
| `length` | DecimalField(8,2) | null=True |
| `length_unit` | CharField(50) | choices=CableLengthUnitChoices; null=True |
| `_abs_length` | DecimalField(10,4) | null=True (cached, in meters) |

## Inherited Fields (from base classes)

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M to Tag)

## Termination Architecture

Cables use a **many-to-many termination model** via `CableTermination`. Each Cable has A-side and B-side terminations that can connect to multiple endpoints (polymorphic via GenericFK).

## Dependencies (depends on 1 model)

- [Tenant](./../tenancy/tenancyTenant.md) (optional)

## Referenced By (8+ models)

- [CableTermination](./dcimCabletermination.md) via `cable` (FK, CASCADE)
- [ConsolePort](./dcimConsoleport.md) via `cable` (FK, SET_NULL)
- [ConsoleServerPort](./dcimConsoleserverport.md) via `cable` (FK, SET_NULL)
- [PowerPort](./dcimPowerport.md) via `cable` (FK, SET_NULL)
- [PowerOutlet](./dcimPoweroutlet.md) via `cable` (FK, SET_NULL)
- [Interface](./dcimInterface.md) via `cable` (FK, SET_NULL)
- [FrontPort](./dcimFrontport.md) via `cable` (FK, SET_NULL)
- [RearPort](./dcimRearport.md) via `cable` (FK, SET_NULL)
- [PowerFeed](./dcimPowerfeed.md) via `cable` (FK, SET_NULL)

## Notes

- **Python source:** `dcim/models/cables.py`
- **Go model file:** `internal/model/dcimCable.go`
- **Proto file:** `api/netbox_go/v1/dcimCable.proto`
- Cable uses a **polymorphic termination model** via CableTermination with GenericFK
- `_abs_length` is cached field storing normalized length in meters
- `save()` triggers `trace_paths` signal for cable path recalculation

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected)

### CREATE
1. `clean()` validates: termination A and B are different; neither is already connected.
2. Save.
3. **Signal `update_connected_endpoints()`:** Sets `cable` FK on both terminations.
4. **Signal `update_cablepaths()`:** Recomputes CablePath graph.
5. Change log + event.

### UPDATE
1. Snapshot.
2. If terminations changed: old endpoints disconnected, new connected.
3. `update_cablepaths()` recomputes.
4. Change log + event.

### DELETE
1. **`pre_delete`:** `nullify_connected_endpoints()` — sets `cable=null` on both terminations.
2. **`post_delete`:** `update_cablepaths()` — deletes/recomputes paths.
3. Change log + event.
