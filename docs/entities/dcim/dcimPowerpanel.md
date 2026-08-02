# PowerPanel

> Module: `dcim` | Table: `dcim_powerpanel` | Python class: `PowerPanel` | File: `dcim/models/power.py`

**Inheritance:** `ContactsMixin <- ImageAttachmentsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/power-panels/`

## Implementation Status

- [ ] Go model (`internal/model/dcimPowerpanel.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimPowerpanel.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimPowerpanel.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimPowerpanel.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimPowerpanel.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimPowerpanel.go`)
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
| `site` | [Site](./dcimSite.md) | `PROTECT` | No | `powerpanels` |
| `location` | [Location](./dcimLocation.md) | `PROTECT` | Yes | `powerpanels` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`
- From **ContactsMixin**: Contact assignments
- From **ImageAttachmentsMixin**: Image attachments

## Constraints

- UniqueConstraint: `(site, name)`

## Referenced By

- [PowerFeed](./dcimPowerfeed.md) via `power_panel` (FK, related_name=`powerfeeds`)

## Notes

- **Python source:** `dcim/models/power.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `site` (and optional `location`) set.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** PowerFeeds.
2. Change log + event.
