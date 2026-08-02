# ServiceTemplate

> Module: `ipam` | Table: `ipam_servicetemplate` | Python class: `ServiceTemplate` | File: `ipam/models/services.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/service-templates/`

## Implementation Status

- [ ] Go model (`internal/model/ipamServicetemplate.go`)
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
| `protocol` | CharField(50) | choices=ServiceProtocolChoices; required |
| `ports` | ArrayField | base_field=PositiveIntegerField; required (list of port numbers) |

## Notes

- **Python source:** `ipam/models/services.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow. No side effects.

### UPDATE
Standard flow. No side effects.

### DELETE
No downstream effects.
