# Provider

> Module: `circuits` | Table: `circuits_provider` | Python class: `Provider` | File: `circuits/models/providers.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/circuits/providers/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsProvider.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `account` | [Account](./../circuits/circuitsAccount.md) | `SET_NULL` | Yes | `providers` |

### ManyToMany Fields

| Field | Related Model | through | related_name |
|-------|---------------|---------|--------------|
| `asns` | [ASN](./../ipam/ipamAsn.md) | `(auto)` | `providers` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |

## Notes

- **Python source:** `circuits/models/providers.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Circuits and ProviderNetworks prevent deletion.
