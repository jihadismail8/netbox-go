# Rack

> Module: `dcim` | Table: `dcim_rack` | Python class: `Rack` | File: `dcim/models/racks.py`

**Inheritance:** `ContactsMixin <- ImageAttachmentsMixin <- RackBase(WeightMixin, PrimaryModel)`

**REST URL:** `/api/dcim/racks/`

## Implementation Status

- [ ] Go model (`internal/model/dcimRack.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimRack.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimRack.go`)
- [ ] DAO unit tests (`internal/dao/dcimRack_test.go`)
- [ ] Cache layer (`internal/cache/dcimRack.go`)
- [ ] Cache unit tests (`internal/cache/dcimRack_test.go`)
- [ ] Service layer (`internal/service/dcimRack.go`)
- [ ] Service unit tests (`internal/service/dcimRack_test.go`)
- [ ] Handler layer (`internal/handler/dcimRack.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/racks/`)
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

### Foreign Keys (5)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `rack_type` | [RackType](./dcimRacktype.md) | `PROTECT` | Yes | `racks` |
| `site` | [Site](./dcimSite.md) | `PROTECT` | No | `racks` |
| `location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `racks` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `racks` |
| `role` | [RackRole](./dcimRackrole.md) | `PROTECT` | Yes | `racks` |

### Regular Fields (defined directly on Rack)

| Field | Type | Notes |
|-------|------|-------|
| `form_factor` | CharField(50) | choices=RackFormFactorChoices; null=True |
| `name` | CharField(100) | db_collation="natural_sort" |
| `facility_id` | CharField(50) | null=True |
| `status` | CharField(50) | choices=RackStatusChoices; default=active |
| `serial` | CharField(50) | blank=True |
| `asset_tag` | CharField(50) | unique=True; null=True |
| `airflow` | CharField(50) | choices=RackAirflowChoices; null=True |

### Fields Inherited from RackBase

| Field | Type | Notes |
|-------|------|-------|
| `width` | PositiveSmallIntegerField | choices=RackWidthChoices; default=19in |
| `u_height` | PositiveSmallIntegerField | default=42; min=1, max=100 |
| `starting_unit` | PositiveSmallIntegerField | default=1 |
| `desc_units` | BooleanField | default=False |
| `outer_width` | PositiveSmallIntegerField | null=True |
| `outer_height` | PositiveSmallIntegerField | null=True |
| `outer_depth` | PositiveSmallIntegerField | null=True |
| `outer_unit` | CharField(50) | choices=RackDimensionUnitChoices; null=True |
| `mounting_depth` | PositiveSmallIntegerField | null=True |
| `max_weight` | PositiveIntegerField | null=True |
| `_abs_max_weight` | PositiveBigIntegerField | null=True (cached, in grams) |

## Inherited Fields (from base classes)

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M to Tag)
- From **WeightMixin**: `weight` (DecimalField), `weight_unit` (CharField), `_abs_weight` (DecimalField, cached in grams)
- From **ContactsMixin**: Contact assignments (reverse relation from `tenancy.ContactAssignment`)
- From **ImageAttachmentsMixin**: Image attachments (reverse relation from `extras.ImageAttachment`)

## Generic Relations

- `vlan_groups` → `ipam.VLANGroup` (via `scope_type`/`scope_id` GenericFK)

## Constraints

- UniqueConstraint: `(location, name)`
- UniqueConstraint: `(location, facility_id)`

## Dependencies (depends on 5 models)

These models must exist before this one can function:

- [Site](./dcimSite.md)
- [Location](./dcimLocation.md) (optional)
- [Tenant](./../tenancy/tenancyTenant.md) (optional)
- [RackRole](./dcimRackrole.md) (optional)
- [RackType](./dcimRacktype.md) (optional)

## Referenced By (6 models)

These models have FK or M2M pointing to this model:

- [Device](./dcimDevice.md) via `rack` (FK, related_name=`devices`)
- [RackReservation](./dcimRackreservation.md) via `rack` (FK, related_name=`reservations`, CASCADE)
- [PowerFeed](./dcimPowerfeed.md) via `rack` (FK, related_name=`powerfeeds`)
- [VLANGroup](./../ipam/ipamVlangroup.md) via `scope` (GenericFK)
- [Module](./dcimModule.md) via `_rack` (FK, cached, via device)
- [CableTermination](./dcimCabletermination.md) via `_rack` (FK, cached)

## Notes

- **Python source:** `dcim/models/racks.py`
- **Go model file:** `internal/model/dcimRack.go`
- **Proto file:** `api/netbox_go/v1/dcimRack.proto`
- When a `rack_type` is assigned, physical attributes (`RACKTYPE_FIELDS`) are copied via `copy_racktype_attrs()` on save
- `RACKTYPE_FIELDS` = `form_factor, width, u_height, starting_unit, desc_units, outer_width, outer_height, outer_depth, outer_unit, mounting_depth, weight, weight_unit, max_weight`
- `_abs_max_weight` is a computed cached field stored in grams for database ordering
- Complex business logic: `get_available_units()`, `get_utilization()`, `get_power_utilization()`, `get_elevation_svg()` must be ported
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), C (Cache Source)

### CREATE
1. `clean()` validates unique (site, location, name) and (site, location, facility_id).
2. Save.

### UPDATE
1. Snapshot. `clean()` re-validates uniqueness.
3. Save.

### DELETE
1. **CASCADE:** RackReservations.
2. **SET_NULL:** Devices (`device.rack=null`), Components (`_rack=null`), CableTermination (`_rack=null`).
3. **PROTECT:** PowerFeed prevents deletion.
4. Change log + event.

### Interdependencies
- **Counter fields:** `device_count`.
- **Cache source for:** ComponentModels (`_rack`), CableTermination (`_rack`).
