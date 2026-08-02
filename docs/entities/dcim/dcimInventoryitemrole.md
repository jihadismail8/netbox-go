# InventoryItemRole

> Module: `dcim` | Table: `dcim_inventoryitemrole` | Python class: `InventoryItemRole` | File: `dcim/models/device_components.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/dcim/inventory-item-roles/`

## Implementation Status

- [ ] Go model (`internal/model/dcimInventoryitemrole.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimInventoryitemrole.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimInventoryitemrole.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimInventoryitemrole.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimInventoryitemrole.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimInventoryitemrole.go`)
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
| `slug` | CharField(100) | Required, unique |
| `color` | ColorField | default=grey |

## Inherited Fields

- From **OrganizationalModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `tags`

## Referenced By

- [InventoryItem](./dcimInventoryitem.md) via `role` (FK, SET_NULL)

## Notes

- **Python source:** `dcim/models/device_components.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
