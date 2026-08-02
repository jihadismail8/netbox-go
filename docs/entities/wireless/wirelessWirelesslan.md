# WirelessLAN

> Module: `wireless` | Table: `wireless_wirelesslan` | Python class: `WirelessLAN` | File: `wireless/models.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/wireless/wireless-lans/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `group` | [WirelessLANGroup](./wirelessWirelesslangroup.md) | `SET_NULL` | Yes | `wireless_lans` |
| `vlan` | [VLAN](./../ipam/ipamVlan.md) | `PROTECT` | Yes | `wireless_lans` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `wireless_lans` |

### Scope (GenericFK)

| Field | Related Model | Via |
|-------|---------------|-----|
| `scope` | (polymorphic: Site, Region, etc.) | `scope_type` + `scope_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `ssid` | CharField(255) | Required |
| `status` | CharField(50) | choices=WirelessLANStatusChoices; default=active |
| `auth_type` | CharField(50) | choices=WirelessAuthTypeChoices; null=True |
| `auth_cipher` | CharField(50) | choices=WirelessAuthCipherChoices; null=True |
| `auth_psk` | CharField(PSK_MAX_LENGTH) | blank=True |

## Notes

- **Python source:** `wireless/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), F (Scoped)

### CREATE
1. `clean()` validates `status`, scope type allowed.
2. Save.

### UPDATE
Standard flow.

### DELETE
Change log + event.
