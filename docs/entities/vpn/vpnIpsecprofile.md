# IPSecProfile

> Module: `vpn` | Table: `vpn_ipsecprofile` | Python class: `IPSecProfile` | File: `vpn/models/crypto.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/vpn/ipsec-profiles/`

## Django Model Fields

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `ike_version` | PositiveSmallIntegerField | choices: 1 or 2 |
| `mode` | CharField(50) | choices=IPSecModeChoices |
| `phase1_encryption` | CharField(50) | choices |
| `phase1_authentication` | CharField(50) | choices |
| `phase1_group` | PositiveSmallIntegerField | choices |
| `phase1_lifetime` | PositiveIntegerField | seconds |
| `phase2_encryption` | CharField(50) | choices |
| `phase2_authentication` | CharField(50) | choices |
| `phase2_group` | PositiveSmallIntegerField | choices |
| `phase2_pfs_group` | PositiveSmallIntegerField | choices |
| `phase2_lifetime` | PositiveIntegerField | seconds |

## Notes

- **Python source:** `vpn/models/crypto.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
1. `clean()` validates encryption, authentication, DH group.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Tunnels referencing this profile prevent deletion.
