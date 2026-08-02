# Location

> Module: `dcim` | Table: `dcim_location` | Python class: `Location` | File: `dcim/models/sites.py`

**Inheritance:** `ContactsMixin <- ImageAttachmentsMixin <- NestedGroupModel`

**REST URL:** `/api/dcim/locations/`

## Implementation Status

- [ ] Go model (`internal/model/dcimLocation.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimLocation.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimLocation.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimLocation.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimLocation.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimLocation.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/locations/`)
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

### Foreign Keys (3)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `site` | [Site](./dcimSite.md) | `PROTECT` | No | `locations` |
| `parent` | self (Location) | `CASCADE` | Yes | `children` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `locations` |

**Note:** `parent` is a `TreeForeignKey`.

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `slug` | CharField(100) | Required |
| `status` | CharField(50) | Required |
| `facility` | CharField(50) | |
| `description` | CharField(200) | |

## Inherited Fields

- From **NestedGroupModel**: `id`, `created`, `last_updated`, `custom_field_data`, `name`, `slug`, `description`, `tags` (M2M)
- From **ContactsMixin**: Contact assignments
- From **ImageAttachmentsMixin**: Image attachments

## Dependencies

- [Site](./dcimSite.md)
- [Tenant](./../tenancy/tenancyTenant.md)

## Referenced By

- [Rack](./dcimRack.md) via `location` (FK, related_name=`racks`)
- [Device](./dcimDevice.md) via `location` (FK, related_name=`devices`)
- [PowerPanel](./dcimPowerpanel.md) via `location` (FK, related_name=`powerpanels`)
- [ConfigContext](./../extras/extrasConfigcontext.md) via `locations` (M2M)
- Many component models via `_location` cached FK

## Notes

- **Python source:** `dcim/models/sites.py`
- **Go model file:** `internal/model/dcimLocation.go`
- **Proto file:** `api/netbox_go/v1/dcimLocation.proto`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates `site` is set, parent location belongs to same site.
2. MPTT tree fields computed.
3. Save.

### UPDATE
1. Snapshot. If `parent` or `site` changed: MPTT re-balanced.
3. Save.

### DELETE
1. **CASCADE:** Sub-locations, Racks.
2. **SET_NULL:** Components (`_location=null`), CableTermination (`_location=null`).
3. MPTT re-balanced.
