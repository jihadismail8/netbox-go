# Device

> Module: `dcim` | Table: `dcim_device` | Python class: `Device` | File: `dcim/models/devices.py`

**Inheritance:** `ContactsMixin <- ImageAttachmentsMixin <- RenderConfigMixin <- ConfigContextModel <- TrackingModelMixin <- PrimaryModel`

**REST URL:** `/api/dcim/devices/`

## Implementation Status

- [ ] Go model (`internal/model/dcimDevice.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimDevice.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimDevice.go`)
- [ ] DAO unit tests (`internal/dao/dcimDevice_test.go`)
- [ ] Cache layer (`internal/cache/dcimDevice.go`)
- [ ] Cache unit tests (`internal/cache/dcimDevice_test.go`)
- [ ] Service layer (`internal/service/dcimDevice.go`)
- [ ] Service unit tests (`internal/service/dcimDevice_test.go`)
- [ ] Handler layer (`internal/handler/dcimDevice.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/devices/`)
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

### Foreign Keys (8)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device_type` | [DeviceType](./dcimDevicetype.md) | `PROTECT` | No | `instances` |
| `role` | [DeviceRole](./dcimDevicerole.md) | `PROTECT` | No | `devices` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `devices` |
| `platform` | [Platform](./dcimPlatform.md) | `SET_NULL` | Yes | `devices` |
| `site` | [Site](./dcimSite.md) | `PROTECT` | No | `devices` |
| `location` | [Location](./dcimLocation.md) | `PROTECT` | Yes | `devices` |
| `rack` | [Rack](./dcimRack.md) | `PROTECT` | Yes | `devices` |
| `cluster` | [Cluster](./../virtualization/virtualizationCluster.md) | `SET_NULL` | Yes | `devices` |
| `virtual_chassis` | [VirtualChassis](./dcimVirtualchassis.md) | `SET_NULL` | Yes | `members` |

### OneToOne Fields (3)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `primary_ip4` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` (none) |
| `primary_ip6` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` (none) |
| `oob_ip` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` (none) |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort"; null=True |
| `serial` | CharField(50) | blank=True |
| `asset_tag` | CharField(50) | unique=True; null=True |
| `position` | DecimalField(4,1) | null=True; min=1 |
| `face` | CharField(50) | choices=DeviceFaceChoices; null=True |
| `status` | CharField(50) | choices=DeviceStatusChoices; default=active |
| `airflow` | CharField(50) | choices=DeviceAirflowChoices; null=True |
| `vc_position` | PositiveIntegerField | null=True |
| `vc_priority` | PositiveSmallIntegerField | null=True; max=255 |
| `latitude` | DecimalField(8,6) | null=True |
| `longitude` | DecimalField(9,6) | null=True |

### Counter Cache Fields (10)

| Field | Tracked Model | Tracked FK |
|-------|---------------|------------|
| `console_port_count` | ConsolePort | `device` |
| `console_server_port_count` | ConsoleServerPort | `device` |
| `power_port_count` | PowerPort | `device` |
| `power_outlet_count` | PowerOutlet | `device` |
| `interface_count` | Interface | `device` |
| `front_port_count` | FrontPort | `device` |
| `rear_port_count` | RearPort | `device` |
| `device_bay_count` | DeviceBay | `device` |
| `module_bay_count` | ModuleBay | `device` |
| `inventory_item_count` | InventoryItem | `device` |

## Inherited Fields (from base classes)

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M to Tag)
- From **RenderConfigMixin**: `config_template` (FK to `extras.ConfigTemplate`, PROTECT, null=True)
- From **ConfigContextModel**: `local_context_data` (JSONField, null=True), `local_context_data_owner_content_type`/`local_context_data_owner_object_id` (GenericFK)
- From **ContactsMixin**: Contact assignments (reverse relation from `tenancy.ContactAssignment`)
- From **ImageAttachmentsMixin**: Image attachments (reverse relation from `extras.ImageAttachment`)
- From **TrackingModelMixin**: `_is_signal_during_save` (internal tracking)

## Generic Relations

- `services` → `ipam.Service` (via `parent_object_type`/`parent_object_id` GenericFK)

## Constraints

- UniqueConstraint: `(Lower(name), site, tenant)`
- UniqueConstraint: `(Lower(name), site)` when `tenant IS NULL`
- UniqueConstraint: `(rack, position, face)`
- UniqueConstraint: `(virtual_chassis, vc_position)`

## Dependencies (depends on 9 models)

These models must exist before this one can function:

- [DeviceType](./dcimDevicetype.md)
- [DeviceRole](./dcimDevicerole.md)
- [Tenant](./../tenancy/tenancyTenant.md) (optional)
- [Platform](./dcimPlatform.md) (optional)
- [Site](./dcimSite.md)
- [Location](./dcimLocation.md) (optional)
- [Rack](./dcimRack.md) (optional)
- [Cluster](./../virtualization/virtualizationCluster.md) (optional)
- [VirtualChassis](./dcimVirtualchassis.md) (optional)

## Referenced By (16+ models)

These models have FK or M2M pointing to this model:

- [ConsolePort](./dcimConsoleport.md) via `device` (FK)
- [ConsoleServerPort](./dcimConsoleserverport.md) via `device` (FK)
- [PowerPort](./dcimPowerport.md) via `device` (FK)
- [PowerOutlet](./dcimPoweroutlet.md) via `device` (FK)
- [Interface](./dcimInterface.md) via `device` (FK)
- [FrontPort](./dcimFrontport.md) via `device` (FK)
- [RearPort](./dcimRearport.md) via `device` (FK)
- [DeviceBay](./dcimDevicebay.md) via `device` (FK) and via `installed_device` (OneToOne)
- [ModuleBay](./dcimModulebay.md) via `device` (FK)
- [InventoryItem](./dcimInventoryitem.md) via `device` (FK)
- [Module](./dcimModule.md) via `device` (FK)
- [MACAddress](./dcimMacaddress.md) via `device` (M2M via interface)
- [IPAddress](./../ipam/ipamIpaddress.md) via `primary_ip4`/`primary_ip6`/`oob_ip` (reverse O2O)
- [Service](./../ipam/ipamService.md) via `parent_object` (GenericFK)
- [VirtualDeviceContext](./dcimVirtualdevicecontext.md) via `device` (FK)
- [RearPort](./dcimRearport.md) via `device` (FK)

## Notes

- **Python source:** `dcim/models/devices.py`
- **Go model file:** `internal/model/dcimDevice.go`
- **Proto file:** `api/netbox_go/v1/dcimDevice.proto`
- On new Device creation, `_instantiate_components()` auto-creates all component objects (console/power/interface/front/rear ports, device/module bays) based on the DeviceType templates
- `save()` inherits airflow from DeviceType if not set, and `default_platform` if platform not set
- `save()` inherits location from Rack if not set
- Uses `ConfigContextModelQuerySet` as manager for config context resolution
- Complex validation: rack space, primary IP ownership, platform/manufacturer consistency, cluster/site consistency, virtual chassis assignment
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), C (Cache Source), D (Cable-Connected)

### CREATE
1. `clean()` validates:
   - Unique (site, tenant, serial) if serial is set.
   - `vc_position` unique within virtual chassis if set.
2. Save.
3. Change log + event.

### UPDATE
1. Snapshot.
2. `clean()` re-validates.
3. Save.
4. **CRITICAL:** If `site`/`location`/`rack` changed — signal `update_component_caches()` refreshes ALL child components' `_site`, `_location`, `_rack`.
5. If `primary_ip4`/`primary_ip6` changed: validates IP assigned to this device.
6. Change log + event.

### DELETE
1. **CASCADE (all components):** ConsolePort, ConsoleServerPort, PowerPort, PowerOutlet, Interface, FrontPort, RearPort, DeviceBay, ModuleBay, InventoryItem.
2. **Cable cleanup:** Cables connected to components are disconnected (termination FK nulled).
3. **CablePath recomputation:** All paths through this device deleted/recomputed.
4. Site `device_count` decremented.
5. Change log + event.

### Interdependencies
- **Counter fields:** `console_port_count`, `console_server_port_count`, `power_port_count`, `power_outlet_count`, `interface_count`, `front_port_count`, `rear_port_count`, `device_bay_count`, `module_bay_count`, `inventory_item_count`.
- **Cache source for:** ALL ComponentModels (`_site`, `_location`, `_rack`).
- **Virtual chassis:** `vc_position` must be unique within chassis.
- **Primary IP:** `primary_ip4`/`primary_ip6` must be assigned to an interface on this device.
