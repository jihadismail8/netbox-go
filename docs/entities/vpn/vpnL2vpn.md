# L2VPN

> Module: `vpn` | Table: `vpn_l2vpn` | Python class: `L2VPN` | File: `vpn/models/l2vpn.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/vpn/l2vpns/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `l2vpns` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `identifier` | CharField(100) | null=True |
| `name` | CharField(100) | Required |
| `slug` | SlugField(100) | Required |
| `type` | CharField(50) | choices=L2VPNTypeChoices; required |

## Notes

- **Python source:** `vpn/models/l2vpn.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `type`, unique identifier.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** L2VPNTerminations.
2. Change log + event.
