# ASN

> Module: `ipam` | Table: `ipam_asn` | Python class: `ASN` | File: `ipam/models/asns.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/ipam/asns/`

## Implementation Status

- [ ] Go model (`internal/model/ipamAsn.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] Proto generated code
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `rir` | [RIR](./ipamRir.md) | `PROTECT` | No | `asns` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `asns` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `asn` | ASNField (custom) | Required, unique (32-bit) |

## Referenced By

- [Site](./../dcim/dcimSite.md) via `asns` (M2M)
- [Provider](./../circuits/circuitsProvider.md) via `asns` (M2M)

## Notes

- **Python source:** `ipam/models/asns.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates ASN within ASNRange if range set.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Sites (M2M auto-removed).
2. Change log + event.
