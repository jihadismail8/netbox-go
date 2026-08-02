# TunnelTermination

> Module: `vpn` | Table: `vpn_tunneltermination` | Python class: `TunnelTermination` | File: `vpn/models/tunnels.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/vpn/tunnel-terminations/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `tunnel` | [Tunnel](./vpnTunnel.md) | `CASCADE` | No | `terminations` |
| `outside_ip` | [IPAddress](./../ipam/ipamIpaddress.md) | `CASCADE` | No | — |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `termination` | (polymorphic: Device, VirtualMachine) | `termination_type` + `termination_id` |

## Notes

- **Python source:** `vpn/models/tunnels.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates `tunnel` set, termination object (Interface/VMInterface).
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** References from tunnel endpoint.
2. Change log + event.
