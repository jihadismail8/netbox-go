# Tunnel

> Module: `vpn` | Table: `vpn_tunnel` | Python class: `Tunnel` | File: `vpn/models/tunnels.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/vpn/tunnels/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `status` | CharField(50) | — | — | choices=TunnelStatusChoices; default=planned |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `tunnels` |
| `ipsec_profile` | [IPSecProfile](./vpnIpsecprofile.md) | `SET_NULL` | Yes | `tunnels` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `status` | CharField(50) | choices=TunnelStatusChoices |
| `encapsulation` | CharField(50) | choices=TunnelEncapsulationChoices; required |
| `ipsec_profile` | FK | null=True |

## Notes

- **Python source:** `vpn/models/tunnels.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `status`, tunnel type.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** TunnelTerminations.
2. Change log + event.
