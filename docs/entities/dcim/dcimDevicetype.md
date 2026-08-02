# DeviceType

> Module: `dcim` | Table: `dcim_devicetype` | Python class: `DeviceType` | File: `dcim/models/devices.py`

**Inheritance:** `ImageAttachmentsMixin <- PrimaryModel <- WeightMixin`

**REST URL:** `/api/dcim/device-types/`

## Implementation Status

- [ ] Go model (`internal/model/dcimDevicetype.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimDevicetype.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimDevicetype.go`)
- [ ] DAO unit tests (`internal/dao/dcimDevicetype_test.go`)
- [ ] Cache layer (`internal/cache/dcimDevicetype.go`)
- [ ] Cache unit tests (`internal/cache/dcimDevicetype_test.go`)
- [ ] Service layer (`internal/service/dcimDevicetype.go`)
- [ ] Service unit tests (`internal/service/dcimDevicetype_test.go`)
- [ ] Handler layer (`internal/handler/dcimDevicetype.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/device-types/`)
- [ ] Response envelope compatible
- [ ] Bulk operations (create/update/delete)
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

### Foreign Keys (2)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `manufacturer` | [Manufacturer](./dcimManufacturer.md) | `PROTECT` | No | `device_types` |
| `default_platform` | [Platform](./dcimPlatform.md) | `SET_NULL` | Yes | `+` (none) |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `model` | CharField(100) | Required; unique with manufacturer |
| `slug` | SlugField(100) | Required; unique with manufacturer |
| `part_number` | CharField(50) | Optional |
| `u_height` | DecimalField(4,1) | Default 1.0; must be increments of 0.5 |
| `exclude_from_utilization` | BooleanField | Default False |
| `is_full_depth` | BooleanField | Default True |
| `subdevice_role` | CharField(50) | choices=SubdeviceRoleChoices; null=True |
| `airflow` | CharField(50) | choices=DeviceAirflowChoices; null=True |
| `front_image` | ImageField | upload_to='devicetype-images'; blank=True |
| `rear_image` | ImageField | upload_to='devicetype-images'; blank=True |

### Counter Cache Fields (10)

| Field | Tracked Model | Tracked FK |
|-------|---------------|------------|
| `console_port_template_count` | ConsolePortTemplate | `device_type` |
| `console_server_port_template_count` | ConsoleServerPortTemplate | `device_type` |
| `power_port_template_count` | PowerPortTemplate | `device_type` |
| `power_outlet_template_count` | PowerOutletTemplate | `device_type` |
| `interface_template_count` | InterfaceTemplate | `device_type` |
| `front_port_template_count` | FrontPortTemplate | `device_type` |
| `rear_port_template_count` | RearPortTemplate | `device_type` |
| `device_bay_template_count` | DeviceBayTemplate | `device_type` |
| `module_bay_template_count` | ModuleBayTemplate | `device_type` |
| `inventory_item_template_count` | InventoryItemTemplate | `device_type` |

## Inherited Fields (from base classes)

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M to Tag)
- From **WeightMixin**: `weight` (DecimalField), `weight_unit` (CharField), `_abs_weight` (DecimalField, cached)
- From **ImageAttachmentsMixin**: Image attachments (reverse relation from `extras.ImageAttachment`)

## Constraints

- UniqueConstraint: `(manufacturer, model)`
- UniqueConstraint: `(manufacturer, slug)`

## Dependencies (depends on 2 models)

These models must exist before this one can function:

- [Manufacturer](./dcimManufacturer.md)
- [Platform](./dcimPlatform.md) (optional, for `default_platform`)

## Referenced By (11 models)

These models have FK or M2M pointing to this model:

- [Device](./dcimDevice.md) via `device_type` (FK, related_name=`instances`)
- [ConsolePortTemplate](./dcimConsoleporttemplate.md) via `device_type` (FK)
- [ConsoleServerPortTemplate](./dcimConsoleserverporttemplate.md) via `device_type` (FK)
- [PowerPortTemplate](./dcimPowerporttemplate.md) via `device_type` (FK)
- [PowerOutletTemplate](./dcimPoweroutlettemplate.md) via `device_type` (FK)
- [InterfaceTemplate](./dcimInterfacetemplate.md) via `device_type` (FK)
- [FrontPortTemplate](./dcimFrontporttemplate.md) via `device_type` (FK)
- [RearPortTemplate](./dcimRearporttemplate.md) via `device_type` (FK)
- [DeviceBayTemplate](./dcimDevicebaytemplate.md) via `device_type` (FK)
- [ModuleBayTemplate](./dcimModulebaytemplate.md) via `device_type` (FK)
- [InventoryItemTemplate](./dcimInventoryitemtemplate.md) via `device_type` (FK)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimDevicetype.go`
- **Proto file:** `api/netbox_go/v1/dcimDevicetype.proto`
- Component templates are auto-created on new Device creation via `_instantiate_components()`
- `front_image` and `rear_image` are stored as file uploads — need file storage strategy in Go
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source)

### CREATE
1. `clean()` validates manufacturer, front/rear image presence.
2. Save.

### UPDATE
1. Snapshot. `clean()` re-validates.
3. Save.

### DELETE
1. **PROTECT:** Devices prevent deletion.
2. **CASCADE:** Component templates.
3. Manufacturer counter decremented.

### Interdependencies
- **Counter fields:** `console_port_template_count`, `console_server_port_template_count`, `power_port_template_count`, `power_outlet_template_count`, `interface_template_count`, `front_port_template_count`, `rear_port_template_count`, `device_bay_template_count`, `module_bay_template_count`, `inventory_item_template_count`.
