# MACAddress

> Module: `dcim` | Table: `dcim_macaddress` | Python class: `MACAddress` | File: `dcim/models/devices.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/dcim/mac-addresses/`

## Implementation Status

- [ ] Go model (`internal/model/dcimMacaddress.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimMacaddress.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimMacaddress.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimMacaddress.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimMacaddress.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimMacaddress.go`)
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
| `mac_address` | MACAddressField | Required (custom field type) |

### Generic FK (assignment)

| Field | Related Model | Via |
|-------|---------------|-----|
| `assigned_object` | (polymorphic: Interface, VMInterface) | `assigned_object_type` (ContentType) + `assigned_object_id` |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`

## Referenced By

- [Interface](./dcimInterface.md) via `primary_mac_address` (reverse O2O, SET_NULL)

## Notes

- **Python source:** `dcim/models/devices.py`
- `MACAddressField` is a custom Django field storing 48-bit MAC addresses
- Assigned to interfaces via GenericFK (supports both dcim.Interface and virtualization.VMInterface)

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
Standard flow. Assigned to Interface or VMInterface.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
