# ModuleType

> Module: `dcim` | Table: `dcim_moduletype` | Python class: `ModuleType` | File: `dcim/models/modules.py`

**Inheritance:** `ImageAttachmentsMixin <- PrimaryModel <- WeightMixin`

**REST URL:** `/api/dcim/module-types/`

## Implementation Status

- [ ] Go model (`internal/model/dcimModuletype.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimModuletype.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimModuletype.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimModuletype.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimModuletype.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimModuletype.go`)
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
| `manufacturer` | [Manufacturer](./dcimManufacturer.md) | `PROTECT` | No | `module_types` |
| `profile` | [ModuleTypeProfile](./dcimModuletypeprofile.md) | `SET_NULL` | Yes | `module_types` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `model` | CharField(100) | Required |
| `part_number` | CharField(50) | blank=True |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`
- From **WeightMixin**: `weight` (DecimalField), `weight_unit` (CharField), `_abs_weight` (cached, in grams)
- From **ImageAttachmentsMixin**: Image attachments (reverse from extras.ImageAttachment)

## Constraints

- UniqueConstraint: `(manufacturer, model)`

## Referenced By

- [Module](./dcimModule.md) via `module_type` (FK, CASCADE)
- Component templates (InterfaceTemplate, etc.) via `module_type` (FK)

## Notes

- **Python source:** `dcim/models/modules.py`
- Similar to DeviceType but for hot-swappable modules
- Has component templates that define what ports/components a module provides

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Modules referencing this type prevent deletion.
