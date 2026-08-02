# ModuleTypeProfile

> Module: `dcim` | Table: `dcim_moduletypeprofile` | Python class: `ModuleTypeProfile` | File: `dcim/models/modules.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/dcim/module-type-profiles/`

## Implementation Status

- [ ] Go model (`internal/model/dcimModuletypeprofile.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimModuletypeprofile.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimModuletypeprofile.go`)
- [ ] DAO unit tests
- [ ] Cache layer
- [ ] Service layer
- [ ] Handler layer
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

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Referenced By

- [ModuleType](./dcimModuletype.md) via `profile` (FK, SET_NULL)

## Notes

- **Python source:** `dcim/models/modules.py`
- Defines attributes/profiles for module types

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow. No side effects.

### UPDATE
Standard flow. No side effects.

### DELETE
No downstream effects.
