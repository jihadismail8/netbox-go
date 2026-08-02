# RackReservation

> Module: `dcim` | Table: `dcim_rackreservation` | Python class: `RackReservation` | File: `dcim/models/racks.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/dcim/rack-reservations/`

## Implementation Status

- [ ] Go model (`internal/model/dcimRackreservation.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimRackreservation.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimRackreservation.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimRackreservation.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimRackreservation.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimRackreservation.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
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

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `rack` | [Rack](./dcimRack.md) | `CASCADE` | No | `reservations` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `rackreservations` |
| `user` | User (settings.AUTH_USER_MODEL) | `PROTECT` | No | — |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `units` | ArrayField | base_field=PositiveSmallIntegerField; list of unit numbers |
| `status` | CharField(50) | choices=RackReservationStatusChoices; default=active |
| `description` | CharField(200) | Required |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `tags`

## Notes

- **Python source:** `dcim/models/racks.py`
- `units` is a PostgreSQL ArrayField storing rack unit numbers
- Validation: units must exist within the rack, no duplicate reservations

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
Standard flow. No side effects.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
