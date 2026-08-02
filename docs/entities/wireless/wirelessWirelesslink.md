# WirelessLink

> Module: `wireless` | Table: `wireless_wirelesslink` | Python class: `WirelessLink` | File: `wireless/models.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/wireless/wireless-links/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `interface_a` | [Interface](./../dcim/dcimInterface.md) | `CASCADE` | No | `+` |
| `interface_b` | [Interface](./../dcim/dcimInterface.md) | `CASCADE` | No | `+` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `wireless_links` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `ssid` | CharField(255) | blank=True |
| `status` | CharField(50) | choices=LinkStatusChoices; default=connected |
| `auth_type` | CharField(50) | null=True |
| `auth_cipher` | CharField(50) | null=True |
| `auth_psk` | CharField(PSK_MAX_LENGTH) | blank=True |
| `distance` | DecimalField(8,2) | null=True |
| `distance_unit` | CharField(50) | null=True |

## Notes

- **Python source:** `wireless/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), D (Cable-Connected)

### CREATE
1. `clean()` validates interface_a and interface_b are different.
2. Save.
3. Signal updates connected endpoints.

### UPDATE
Standard flow.

### DELETE
1. **Signal:** Disconnects both interface endpoints.
2. Change log + event.
