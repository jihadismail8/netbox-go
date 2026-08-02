# WirelessLANGroup

> Module: `wireless` | Table: `wireless_wirelesslangroup` | Python class: `WirelessLANGroup` | File: `wireless/models.py`

**Inheritance:** `NestedGroupModel`

**REST URL:** `/api/wireless/wireless-lan-groups/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |
| `parent` | self FK | SET_NULL, null=True |

## Notes

- **Python source:** `wireless/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates parent group exists.
2. MPTT computed.
3. Save.

### UPDATE
If `parent` changed: MPTT re-balanced. Save.

### DELETE
1. **CASCADE:** Sub-groups.
2. **SET_NULL:** WirelessLANs (`wirelesslan.group=null`).
