# CircuitType

> Module: `circuits` | Table: `circuits_circuittype` | Python class: `CircuitType` | File: `circuits/models/circuits.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/circuits/circuit-types/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsCircuittype.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Vue.js views

## Django Model Fields

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |

## Notes

- **Python source:** `circuits/models/circuits.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Circuits referencing this type prevent deletion.
