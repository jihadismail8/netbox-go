# InventoryItem

> Module: `dcim` | Table: `dcim_inventoryitem` | Python class: `InventoryItem` | File: `dcim/models/device_components.py`

**Inheritance:** `MPTTModel <- NetBoxModel <- TrackingModelMixin`

**REST URL:** `/api/dcim/inventory-items/`

## Implementation Status

- [ ] Go model (`internal/model/dcimInventoryitem.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimInventoryitem.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimInventoryitem.go`)
- [ ] DAO unit tests (`internal/dao/dcimInventoryitem_test.go`)
- [ ] Cache layer (`internal/cache/dcimInventoryitem.go`)
- [ ] Cache unit tests (`internal/cache/dcimInventoryitem_test.go`)
- [ ] Service layer (`internal/service/dcimInventoryitem.go`)
- [ ] Service unit tests (`internal/service/dcimInventoryitem_test.go`)
- [ ] Handler layer (`internal/handler/dcimInventoryitem.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/inventory-items/`)
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

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `inventoryitems` |
| `_site` | [Site](./dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `+` (cached) |
| `parent` | [InventoryItem](./dcimInventoryitem.md) (self, TreeForeignKey) | `CASCADE` | Yes | `children` |
| `role` | [InventoryItemRole](./dcimInventoryitemrole.md) | `SET_NULL` | Yes | `inventory_items` |
| `manufacturer` | [Manufacturer](./dcimManufacturer.md) | `SET_NULL` | Yes | `inventory_items` |

### Generic FK (component assignment)

| Field | Related Model | Via |
|-------|---------------|-----|
| `component` | (polymorphic) | `component_type` (ContentType) + `component_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | db_collation="natural_sort" |
| `label` | CharField(64) | blank=True |
| `description` | CharField(200) | blank=True |
| `part_id` | CharField(50) | blank=True; manufacturer part ID |
| `serial` | CharField(50) | blank=True |
| `asset_tag` | CharField(50) | unique=True; null=True |
| `discovered` | BooleanField | default=False |
| `status` | CharField(50) | choices=InventoryItemStatusChoices; default=active |

### MPTT Tree Fields

| Field | Type | Notes |
|-------|------|-------|
| `lft` | PositiveIntegerField | MPTT left edge |
| `rght` | PositiveIntegerField | MPTT right edge |
| `tree_id` | PositiveIntegerField | MPTT tree ID |
| `level` | PositiveIntegerField | MPTT nesting level |

## Constraints

- UniqueConstraint: `(device, name)` for items with parent=None
- UniqueConstraint: `(parent, name)` for child items

## Notes

- Uses **MPTT** for hierarchical inventory trees
- Unlike other components, InventoryItem CAN be moved to a different device
- `component` is a GenericFK pointing to any device component

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** C (Cache Consumer)

### CREATE
1. `clean()` validates `device` is set. InventoryItem **can** be moved between devices.
2. Cache populated from Device.
3. Save. Device counter incremented.

### UPDATE
1. Snapshot.
2. Save.
3. If `device` changed: `_site`/`_location`/`_rack` re-cached from new device.

### DELETE
1. Device counter decremented.
2. Change log + event.
