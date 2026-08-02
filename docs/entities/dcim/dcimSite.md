# Site

> Module: `dcim` | Table: `dcim_site` | Python class: `Site` | File: `dcim/models/sites.py`

**Inheritance:** `ContactsMixin <- ImageAttachmentsMixin <- PrimaryModel`

**REST URL:** `/api/dcim/sites/`

## Implementation Status

- [ ] Go model (`internal/model/dcimSite.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimSite.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimSite.go`)
- [ ] DAO unit tests (`internal/dao/dcimSite_test.go`)
- [ ] Cache layer (`internal/cache/dcimSite.go`)
- [ ] Cache unit tests (`internal/cache/dcimSite_test.go`)
- [ ] Service layer (`internal/service/dcimSite.go`)
- [ ] Service unit tests (`internal/service/dcimSite_test.go`)
- [ ] Handler layer (`internal/handler/dcimSite.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/sites/`)
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

### Foreign Keys (3)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `region` | [Region](./dcimRegion.md) | `SET_NULL` | Yes | `sites` |
| `group` | [SiteGroup](./dcimSitegroup.md) | `SET_NULL` | Yes | `sites` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `sites` |

### ManyToMany Fields (1)

| Field | Related Model | through | related_name |
|-------|---------------|---------|--------------|
| `asns` | [ASN](./../ipam/ipamAsn.md) | `(auto: dcim_site_asns)` | `sites` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | CharField(100) | Required, unique |
| `status` | CharField(50) | Required |
| `facility` | CharField(50) | |
| `time_zone` | TimeZoneField | |
| `description` | CharField(200) | |
| `physical_address` | CharField(200) | |
| `shipping_address` | CharField(200) | |
| `latitude` | DecimalField | |
| `longitude` | DecimalField | |
| `comments` | TextField | |

## Inherited Fields (from base classes)

- From **PrimaryModel**: `id`, `created`, `last_updated`, `custom_field_data`, `description`, `comments`, `tags` (M2M to Tag)
- From **ContactsMixin**: Contact assignments (reverse relation from `tenancy.ContactAssignment`)
- From **ImageAttachmentsMixin**: Image attachments (reverse relation from `extras.ImageAttachment`)

## Dependencies (depends on 3 models)

These models must exist before this one can function:

- [Region](./dcimRegion.md)
- [SiteGroup](./dcimSitegroup.md)
- [Tenant](./../tenancy/tenancyTenant.md)

## Referenced By (23 models)

These models have FK or M2M pointing to this model:

- [Location](./dcimLocation.md) via `site` (FK, related_name=`locations`)
- [Rack](./dcimRack.md) via `site` (FK, related_name=`racks`)
- [RackReservation](./dcimRackreservation.md) via `site` (FK)
- [Device](./dcimDevice.md) via `site` (FK, related_name=`devices`)
- [ModuleBay](./dcimModulebay.md) via `_site` (FK, cached)
- [ConsolePort](./dcimConsoleport.md) via `_site` (FK, cached)
- [ConsoleServerPort](./dcimConsoleserverport.md) via `_site` (FK, cached)
- [PowerPort](./dcimPowerport.md) via `_site` (FK, cached)
- [PowerOutlet](./dcimPoweroutlet.md) via `_site` (FK, cached)
- [Interface](./dcimInterface.md) via `_site` (FK, cached)
- [FrontPort](./dcimFrontport.md) via `_site` (FK, cached)
- [RearPort](./dcimRearport.md) via `_site` (FK, cached)
- [DeviceBay](./dcimDevicebay.md) via `_site` (FK, cached)
- [InventoryItem](./dcimInventoryitem.md) via `_site` (FK, cached)
- [PowerPanel](./dcimPowerpanel.md) via `site` (FK, related_name=`powerpanels`)
- [CableTermination](./dcimCabletermination.md) via `_site` (FK, cached)
- [CircuitTermination](./../circuits/circuitsCircuittermination.md) via `_site` (FK, cached)
- [VirtualMachine](./../virtualization/virtualizationVirtualmachine.md) via `site` (FK, related_name=`virtual_machines`)
- [Cluster](./../virtualization/virtualizationCluster.md) via `site` (FK, related_name=`clusters`)
- [Prefix](./../ipam/ipamPrefix.md) via `scope` (GenericFK)
- [VLAN](./../ipam/ipamVlan.md) via `site` (FK, related_name=`vlans`)
- [ConfigContext](./../extras/extrasConfigcontext.md) via `sites` (M2M, related_name=`+`)
- [WirelessLAN](./../wireless/wirelessWirelesslan.md) via `scope` (GenericFK)

## Notes

- **Python source:** `dcim/models/sites.py`
- **Go model file:** `internal/model/dcimSite.go`
- **Proto file:** `api/netbox_go/v1/dcimSite.proto`
## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), C (Cache Source)

### CREATE
1. `clean()` validates status, tenant FK.
2. Save to DB.
3. Change log + event queued.

### UPDATE
1. Snapshot pre-change state.
2. `clean()` re-validates.
3. Save.
4. Change log + event queued.

### DELETE
1. **CASCADE:** Locations, Racks, PowerPanels, CablePaths with origin here.
2. **SET_NULL:** Components (`_site=null`), CircuitTermination (`_site=null`), CableTermination (`_site=null`).
3. **GenericFK scope:** Prefix.scope, VLAN.site, WirelessLAN.scope become null.
4. Change log + event.

### Interdependencies
- **Counter fields:** `rack_count`, `device_count`, `prefix_count`, `vlan_count`, `circuit_count`, `virtual_machine_count`.
- **Cache source for:** ComponentModels (`_site`), CableTermination (`_site`), CircuitTermination (`_site`).
