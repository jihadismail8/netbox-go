# ContactRole

> Module: `tenancy` | Table: `tenancy_contactrole` | Python class: `ContactRole` | File: `tenancy/models/contacts.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/tenancy/contact-roles/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyContactrole.go`)
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

- **Python source:** `tenancy/models/contacts.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** ContactAssignments referencing this role.
