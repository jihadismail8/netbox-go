# PowerFeed

> Module: `dcim` | Table: `dcim_powerfeed` | Python class: `PowerFeed` | File: `dcim/models/power.py`

**Inheritance:** `PrimaryModel <- CabledObjectModel <- PathEndpoint`

**REST URL:** `/api/dcim/power-feeds/`

## Implementation Status

- [ ] Go model (`internal/model/dcimPowerfeed.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimPowerfeed.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimPowerfeed.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimPowerfeed.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimPowerfeed.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimPowerfeed.go`)
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
| `power_panel` | [PowerPanel](./dcimPowerpanel.md) | `CASCADE` | No | `powerfeeds` |
| `rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `powerfeeds` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `powerfeeds` |
| `cable` | [Cable](./dcimCable.md) | `SET_NULL` | Yes | `+` (none) |
| `_path` | [CablePath](./dcimCablepath.md) | `SET_NULL` | Yes | — |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `status` | CharField(50) | choices=PowerFeedStatusChoices; default=active |
| `type` | CharField(50) | choices=PowerFeedTypeChoices; default=primary |
| `supply` | CharField(50) | choices=PowerFeedSupplyChoices; default=AC |
| `phase` | CharField(50) | choices=PowerFeedPhaseChoices; default=single-phase |
| `voltage` | PositiveSmallIntegerField | default=120 |
| `amperage` | PositiveSmallIntegerField | default=15 |
| `max_utilization` | PositiveSmallIntegerField | default=80; percentage (0-100) |
| `available_power` | PositiveIntegerField | Calculated: volts * amps * (max_util/100) |

### Fields Inherited from CabledObjectModel

| Field | Type | Notes |
|-------|------|-------|
| `cable_end` | CharField(1) | choices=CableEndChoices; null=True |
| `mark_connected` | BooleanField | default=False |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Constraints

- UniqueConstraint: `(power_panel, name)`

## Dependencies

- [PowerPanel](./dcimPowerpanel.md)
- [Rack](./dcimRack.md) (optional)
- [Cable](./dcimCable.md) (optional)

## Notes

- **Python source:** `dcim/models/power.py`
- `available_power` is a computed property: voltage * amperage * (max_utilization / 100)

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected)

### CREATE
1. `clean()` validates `power_panel` set, unique (power_panel, name).
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **Cable disconnect:** If connected, cable termination removed, path recomputed.
2. Change log + event.
