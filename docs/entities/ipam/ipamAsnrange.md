# ASNRange

> Module: `ipam` | Table: `ipam_asnrange` | Python class: `ASNRange` | File: `ipam/models/asns.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/ipam/asn-ranges/`

## Implementation Status

- [ ] Go model (`internal/model/ipamAsnrange.go`)
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
| `rir` | [RIR](./ipamRir.md) | `PROTECT` | No | `asn_ranges` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `asn_ranges` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |
| `start` | ASNField (custom) | Required (range start) |
| `end` | ASNField (custom) | Required (range end) |

## Notes

- **Python source:** `ipam/models/asns.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** ASNs within this range.
