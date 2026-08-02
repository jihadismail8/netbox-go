# VMInterface

> Module: `virtualization` | Table: `virtualization_vminterface` | Python class: `VMInterface` | File: `virtualization/models/virtualmachines.py`

**Inheritance:** `BaseInterface <- PrimaryModel`

**REST URL:** `/api/virtualization/interfaces/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `virtual_machine` | [VirtualMachine](./virtualizationVirtualmachine.md) | `CASCADE` | No | `interfaces` |
| `parent` | [VMInterface](./virtualizationVminterface.md) (self) | `RESTRICT` | Yes | `child_interfaces` |
| `bridge` | [VMInterface](./virtualizationVminterface.md) (self) | `SET_NULL` | Yes | `bridge_interfaces` |
| `untagged_vlan` | [VLAN](./../ipam/ipamVlan.md) | `SET_NULL` | Yes | `vminterfaces_as_untagged` |
| `qinq_svlan` | [VLAN](./../ipam/ipamVlan.md) | `SET_NULL` | Yes | `vminterfaces_svlan` |
| `vrf` | [VRF](./../ipam/ipamVrf.md) | `SET_NULL` | Yes | `vminterfaces` |

### ManyToMany Fields

| Field | Related Model | related_name |
|-------|---------------|--------------|
| `tagged_vlans` | [VLAN](./../ipam/ipamVlan.md) | `vminterfaces_as_tagged` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | Required |
| `enabled` | BooleanField | default=True |
| `mac_address` | MACAddressField | null=True |
| `mtu` | PositiveIntegerField | null=True |
| `mode` | CharField(50) | choices=InterfaceModeChoices; null=True |
| `description` | CharField(200) | blank=True |

## Notes

- **Python source:** `virtualization/models/virtualmachines.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected), C (Cache Consumer)

### CREATE
1. `clean()` validates `virtual_machine` set, unique (virtual_machine, name).
2. Cache (`_cluster`) populated from parent VM.
3. Save.
4. Change log + event.

### UPDATE
1. Snapshot.
2. Save.
3. Change log + event.

### DELETE
1. **Cable disconnect:** If connected, cable termination removed, path recomputed.
2. Change log + event.

### Interdependencies
- **Cache consumer:** `_cluster` from parent VirtualMachine.
- **Cable connectivity:** Can have `cable` FK for path tracing.
- **IP addresses:** Interface `ip_address_count` counter tracks assigned IPs.
