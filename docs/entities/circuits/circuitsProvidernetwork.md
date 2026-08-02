# ProviderNetwork

> Module: `circuits` | Table: `circuits_providernetwork` | Python class: `ProviderNetwork` | File: `circuits/models/providers.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/circuits/provider-networks/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsProvidernetwork.go`)
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
| `provider` | [Provider](./circuitsProvider.md) | `PROTECT` | No | `networks` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `service_id` | CharField(100) | blank=True |

## Notes

- **Python source:** `circuits/models/providers.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `provider` set.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** CircuitTerminations referencing this network.
