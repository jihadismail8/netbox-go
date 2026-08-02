# Rack Type

> Module: `dcim` | Table: `dcim_racktype` | Python class: `RackType` | File: `dcim/models/racks.py`

**Inheritance:** `ContactsMixin <- WeightMixin <- PrimaryModel`

**REST URL:** `/api/dcim/rack-types/`

## Implementation Status

- [ ] Go model (`internal/model/dcimRacktype.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimRacktype.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimRacktype.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimRacktype.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimRacktype.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimRacktype.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/rack-types/`)
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
| `manufacturer` | [Manufacturer](./dcimManufacturer.md) | `PROTECT` | Yes | `rack_types` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `model` | CharField(100) | Required |
| `slug` | CharField(100) | Required |
| `form_factor` | ChoiceField | |
| `width` | ChoiceField | |
| `u_height` | PositiveSmallIntegerField | Default 42 |
| `starting_unit` | PositiveSmallIntegerField | Default 1 |
| `desc_units` | BooleanField | Default False |
| `outer_width` | PositiveSmallIntegerField | |
| `outer_depth` | PositiveSmallIntegerField | |
| `outer_unit` | ChoiceField | |
| `mounting_depth` | PositiveSmallIntegerField | |
| `weight` | DecimalField | |
| `weight_unit` | ChoiceField | |
| `description` | CharField(200) | |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **WeightMixin**: `weight`, `weight_unit`, `_abs_weight`
- From **ContactsMixin**: Contact assignments

## Dependencies

- [Manufacturer](./dcimManufacturer.md)

## Referenced By

- [Rack](./dcimRack.md) via `rack_type` (FK, on_delete=PROTECT, related_name=`racks`)

## Notes

- **Python source:** `dcim/models/racks.py`
- **Go model file:** `internal/model/dcimRacktype.go`
- **Proto file:** `api/netbox_go/v1/dcimRacktype.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Racks referencing this type prevent deletion.
