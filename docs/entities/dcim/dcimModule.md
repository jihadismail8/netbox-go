# Module

> Module: `dcim` | Table: `dcim_module` | Python class: `Module` | File: `dcim/models/modules.py`

**Inheritance:** `ConfigContextModel <- PrimaryModel`

**REST URL:** `/api/dcim/modules/`

## Implementation Status

- [ ] Go model (`internal/model/dcimModule.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimModule.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimModule.go`)
- [ ] DAO unit tests
- [ ] Cache layer (`internal/cache/dcimModule.go`)
- [ ] Cache unit tests
- [ ] Service layer (`internal/service/dcimModule.go`)
- [ ] Service unit tests
- [ ] Handler layer (`internal/handler/dcimModule.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention (`/api/dcim/modules/`)
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
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `modules` |
| `module_bay` | [ModuleBay](./dcimModulebay.md) | `CASCADE` | Yes | `installed_module` |
| `module_type` | [ModuleType](./dcimModuletype.md) | `CASCADE` | No | `instances` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `status` | CharField(50) | choices=ModuleStatusChoices; default=active |
| `serial` | CharField(50) | blank=True |
| `asset_tag` | CharField(50) | unique=True; null=True |

### Counter Cache Fields

| Field | Tracked Model | Tracked FK |
|-------|---------------|------------|
| `console_port_template_count` | ConsolePortTemplate | `module_type` |
| `console_server_port_template_count` | ConsoleServerPortTemplate | `module_type` |
| `power_port_template_count` | PowerPortTemplate | `module_type` |
| `power_outlet_template_count` | PowerOutletTemplate | `module_type` |
| `interface_template_count` | InterfaceTemplate | `module_type` |
| `front_port_template_count` | FrontPortTemplate | `module_type` |
| `rear_port_template_count` | RearPortTemplate | `module_type` |
| `module_bay_template_count` | ModuleBayTemplate | `module_type` |

## Inherited Fields

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags`
- From **ConfigContextModel**: `local_context_data` (JSONField, null=True)

## Constraints

- UniqueConstraint: `(device, module_bay)` — one module per bay

## Dependencies

- [Device](./dcimDevice.md)
- [ModuleBay](./dcimModulebay.md) (optional)
- [ModuleType](./dcimModuletype.md)

## Notes

- **Python source:** `dcim/models/modules.py`
- On creation, auto-instantiates component templates from ModuleType
- On deletion, cleans up child components that were created from the module type

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `module_bay` belongs to same device.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. Components replicated from module are disassociated.
2. Change log + event.
