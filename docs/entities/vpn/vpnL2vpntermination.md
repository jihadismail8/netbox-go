# L2VPNTermination

> Module: `vpn` | Table: `vpn_l2vpntermination` | Python class: `L2VPNTermination` | File: `vpn/models/l2vpn.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/vpn/l2vpn-terminations/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `l2vpn` | [L2VPN](./vpnL2vpn.md) | `CASCADE` | No | `terminations` |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `assigned_object` | (polymorphic: Interface, VMInterface, VLAN) | `assigned_object_type` + `assigned_object_id` |

## Notes

- **Python source:** `vpn/models/l2vpn.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates `l2vpn` set, assigned object (Interface/VMInterface).
2. Save.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
