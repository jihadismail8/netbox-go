# Interface

> Module: `dcim` | Table: `dcim_interface` | Python class: `Interface` | File: `dcim/models/device_components.py`

**Inheritance:** `InterfaceValidationMixin <- ModularComponentModel <- BaseInterface <- CabledObjectModel <- PathEndpoint <- TrackingModelMixin`

**REST URL:** `/api/dcim/interfaces/`

## Implementation Status

- [ ] Go model (`internal/model/dcimInterface.go`)
- [ ] GORM mapping verified (column names, types, constraints)
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimInterface.proto`)
- [ ] Proto generated code (`.pb.go`, `_grpc_pb.go`, `.pb.validate.go`)
- [ ] DAO layer (`internal/dao/dcimInterface.go`)
- [ ] DAO unit tests (`internal/dao/dcimInterface_test.go`)
- [ ] Cache layer (`internal/cache/dcimInterface.go`)
- [ ] Cache unit tests (`internal/cache/dcimInterface_test.go`)
- [ ] Service layer (`internal/service/dcimInterface.go`)
- [ ] Service unit tests (`internal/service/dcimInterface_test.go`)
- [ ] Handler layer (`internal/handler/dcimInterface.go`)
- [ ] HTTP routes registered
- [ ] Error codes defined (`internal/ecode/`)
- [ ] REST URL matches NetBox convention (`/api/dcim/interfaces/`)
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

### Foreign Keys (15)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `device` | [Device](./dcimDevice.md) | `CASCADE` | No | `interfaces` |
| `_site` | [Site](./dcimSite.md) | `SET_NULL` | Yes | `+` (cached) |
| `_location` | [Location](./dcimLocation.md) | `SET_NULL` | Yes | `+` (cached) |
| `_rack` | [Rack](./dcimRack.md) | `SET_NULL` | Yes | `+` (cached) |
| `module` | [Module](./dcimModule.md) | `CASCADE` | Yes | `interfaces` |
| `parent` | [Interface](./dcimInterface.md) (self) | `RESTRICT` | Yes | `child_interfaces` |
| `bridge` | [Interface](./dcimInterface.md) (self) | `SET_NULL` | Yes | `bridge_interfaces` |
| `lag` | [Interface](./dcimInterface.md) (self) | `SET_NULL` | Yes | `member_interfaces` |
| `untagged_vlan` | [VLAN](./../ipam/ipamVlan.md) | `SET_NULL` | Yes | `interfaces_as_untagged` |
| `qinq_svlan` | [VLAN](./../ipam/ipamVlan.md) | `SET_NULL` | Yes | `interfaces_svlan` |
| `vlan_translation_policy` | [VLANTranslationPolicy](./../ipam/ipamVlantranslationpolicy.md) | `PROTECT` | Yes | — |
| `wireless_link` | [WirelessLink](./../wireless/wirelessWirelesslink.md) | `SET_NULL` | Yes | `+` (none) |
| `vrf` | [VRF](./../ipam/ipamVrf.md) | `SET_NULL` | Yes | `interfaces` |
| `cable` | [Cable](./dcimCable.md) | `SET_NULL` | Yes | `+` (none) |
| `_path` | [CablePath](./dcimCablepath.md) | `SET_NULL` | Yes | — |

### OneToOne Fields (1)

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `primary_mac_address` | [MACAddress](./dcimMacaddress.md) | `SET_NULL` | Yes | `+` (none) |

### ManyToMany Fields (3)

| Field | Related Model | through | related_name |
|-------|---------------|---------|--------------|
| `tagged_vlans` | [VLAN](./../ipam/ipamVlan.md) | `(auto)` | `interfaces_as_tagged` |
| `vdcs` | [VirtualDeviceContext](./dcimVirtualdevicecontext.md) | `(auto)` | `interfaces` |
| `wireless_lans` | [WirelessLAN](./../wireless/wirelessWirelesslan.md) | `(auto)` | `interfaces` |

### Regular Fields (defined on Interface directly)

| Field | Type | Notes |
|-------|------|-------|
| `_name` | NaturalOrderingField(100) | Internal natural ordering field |
| `type` | CharField(50) | choices=InterfaceTypeChoices; required |
| `mgmt_only` | BooleanField | default=False |
| `speed` | PositiveIntegerField | null=True (Kbps) |
| `duplex` | CharField(50) | choices=InterfaceDuplexChoices; null=True |
| `wwn` | WWNField | null=True (64-bit World Wide Name) |
| `rf_role` | CharField(30) | choices=WirelessRoleChoices; null=True |
| `rf_channel` | CharField(50) | choices=WirelessChannelChoices; null=True |
| `rf_channel_frequency` | DecimalField(7,2) | null=True (MHz) |
| `rf_channel_width` | DecimalField(7,3) | null=True (MHz) |
| `tx_power` | SmallIntegerField | null=True; min=-40, max=127 (dBm) |
| `poe_mode` | CharField(50) | choices=InterfacePoEModeChoices; null=True |
| `poe_type` | CharField(50) | choices=InterfacePoETypeChoices; null=True |

### Fields Inherited from BaseInterface

| Field | Type | Notes |
|-------|------|-------|
| `enabled` | BooleanField | default=True |
| `mtu` | PositiveIntegerField | null=True; min=1, max=65536 |
| `mode` | CharField(50) | choices=InterfaceModeChoices; null=True |

### Fields Inherited from CabledObjectModel

| Field | Type | Notes |
|-------|------|-------|
| `cable_end` | CharField(1) | choices=CableEndChoices; null=True |
| `mark_connected` | BooleanField | default=False |

## Generic Relations (6)

| Field | Related Model | Via |
|-------|---------------|-----|
| `ip_addresses` | `ipam.IPAddress` | `assigned_object_type`/`assigned_object_id` |
| `mac_addresses` | `dcim.MACAddress` | `assigned_object_type`/`assigned_object_id` |
| `fhrp_group_assignments` | `ipam.FHRPGroupAssignment` | `interface_type`/`interface_id` |
| `tunnel_terminations` | `vpn.TunnelTermination` | `termination_type`/`termination_id` |
| `l2vpn_terminations` | `vpn.L2VPNTermination` | `assigned_object_type`/`assigned_object_id` |
| `inventory_items` | `dcim.InventoryItem` | `component_type`/`component_id` |

## Constraints

- UniqueConstraint: `(device, name)` (from ComponentModel)

## Notes

- **Most complex model in NetBox** — 15 FKs, 1 O2O, 3 M2M, 6 GenericRelations
- `_site`, `_location`, `_rack` are denormalized cached fields copied from parent Device on save
- `_path` is set/cleared by `dcim.signals` in response to cable path changes
- Complex validation: parent/bridge/LAG membership, wireless channel attributes, VLAN mode constraints
- `save()` auto-populates `rf_channel_frequency` and `rf_channel_width` from selected `rf_channel`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected), C (Cache Consumer)

### CREATE
1. `clean()` validates `device` is set. Components **cannot** be moved to a different device after creation.
2. Denormalized cache (`_site`, `_location`, `_rack`) populated from parent Device.
3. Save.
4. Parent Device counter incremented (e.g., `Device.interface_count`).
5. Change log + event.

### UPDATE
1. Snapshot.
2. `clean()` validates device hasn't changed (immutable).
3. Save.
4. Change log + event.

### DELETE
1. **Cable disconnect:** If cable connected, cable's termination reference removed. CablePath recomputed.
2. Parent Device counter decremented.
3. Change log + event.

### Interdependencies
- **Cache consumer:** `_site`, `_location`, `_rack` sourced from parent Device.
- **Cable connectivity:** Each component can have a `cable` FK. Path tracing originates here.
- **Parent counter:** Device's `<component>_count` tracks children.
