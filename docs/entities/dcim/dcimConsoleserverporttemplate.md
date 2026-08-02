# Consoleserverporttemplate

> Module: `dcim` | Table: `dcim_consoleserverporttemplate` | Python class: `Consoleserverporttemplate` | File: `dcim/models/device_component_templates.py`

**Inheritance:** `ComponentTemplateModel <- ChangeLoggedModel <- TrackingModelMixin`

**REST URL:** `/api/dcim/console-server-port-templates/`

## Implementation Status

- [ ] Go model (`internal/model/dcimConsoleserverporttemplate.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimConsoleserverporttemplate.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimConsoleserverporttemplate.go`)
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
- [ ] Vue.js views

## Django Model Fields (from Python source)

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device_type` | [DeviceType](./dcimDevicetype.md) | `CASCADE` | No | `consoleserverporttemplates` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `type` | CharField(50) | choices=ConsolePortTypeChoices; null=True |
| `speed` | PositiveIntegerField | choices=ConsolePortSpeedChoices; null=True |

### Fields Inherited from ComponentTemplateModel

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort" |
| `label` | CharField(64) | blank=True |
| `description` | CharField(200) | blank=True |

## Inherited Fields

- From **ChangeLoggedModel**: `id`, `created`, `last_updated`, `custom_field_data`

## Constraints

- UniqueConstraint: `(device_type, name)`

## Referenced By

- [DeviceType](./dcimDevicetype.md) counter cache `console_server_port_template_count`

## Notes

- **Python source:** `dcim/models/device_component_templates.py`
- Template models define the component blueprint for auto-creation when a new Device/Module is created

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
1. `clean()` validates `device_type` is set.
2. Save. DeviceType counter incremented.

### UPDATE
Standard flow.

### DELETE
1. DeviceType counter decremented.
2. Change log + event.
