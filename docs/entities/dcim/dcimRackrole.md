# Rack Role

> Module: `dcim` | Table: `dcim_rackrole` | Python class: `RackRole` | File: `dcim/models/racks.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/rack-roles/`

## Implementation Status

- [ ] Go model (`internal/model/dcimRackrole.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimRackrole.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimRackrole.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimRackrole.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimRackrole.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimRackrole.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/rack-roles/`)
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
| `description` | CharField(200) | |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments

## Dependencies

No external model dependencies.

## Referenced By

- [Rack](./dcimRack.md) via `role` (FK, on_delete=PROTECT, related_name=`racks`)

## Notes

- **Python source:** `dcim/models/racks.py`
- **Go model file:** `internal/model/dcimRackrole.go`
- **Proto file:** `api/netbox_go/v1/dcimRackrole.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Racks (`rack.role=null`).
