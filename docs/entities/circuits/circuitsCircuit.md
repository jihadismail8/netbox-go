# Circuit

> Module: `circuits` | Table: `circuits_circuit` | Python class: `Circuit` | File: `circuits/models/circuits.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/circuits/circuits/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsCircuit.go`)
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
| `provider` | [Provider](./circuitsProvider.md) | `PROTECT` | No | `circuits` |
| `provider_account` | [ProviderAccount](./circuitsAccount.md) | `SET_NULL` | Yes | `circuits` |
| `type` | [CircuitType](./circuitsCircuittype.md) | `PROTECT` | No | `circuits` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `circuits` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `cid` | CharField(100) | Required (circuit ID) |
| `status` | CharField(50) | choices=CircuitStatusChoices; default=active |
| `install_date` | DateField | null=True |
| `termination_date` | DateField | null=True |
| `commit_rate` | PositiveIntegerField | null=True (Kbps) |

## Notes

- **Python source:** `circuits/models/circuits.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `provider`, `type` set, unique (provider, cid).
2. Save.
3. Site `circuit_count` incremented if termination has site.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** CircuitTerminations.
2. Counter decrement.
3. Change log + event.
