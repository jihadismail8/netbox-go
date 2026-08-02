# ProviderAccount

> Module: `circuits` | Table: `circuits_provideraccount` | Python class: `ProviderAccount` | File: `circuits/models/providers.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/circuits/provider-accounts/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsAccount.go`)
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
| `provider` | [Provider](./circuitsProvider.md) | `PROTECT` | No | `accounts` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `account` | CharField(100) | Required |
| `name` | CharField(100) | blank=True |

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
1. **PROTECT:** Circuits referencing this account prevent deletion.
