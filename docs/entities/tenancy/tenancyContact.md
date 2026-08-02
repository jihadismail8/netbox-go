# Contact

> Module: `tenancy` | Table: `tenancy_contact` | Python class: `Contact` | File: `tenancy/models/contacts.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/tenancy/contacts/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyContact.go`)
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
| `group` | [ContactGroup](./tenancyContactgroup.md) | `SET_NULL` | Yes | `contacts` |
| `tenant` | [Tenant](./tenancyTenant.md) | `PROTECT` | Yes | `contacts` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `title` | CharField(100) | blank=True |
| `phone` | CharField(50) | blank=True |
| `email` | EmailField | blank=True |
| `address` | TextField | blank=True |
| `link` | URLField | blank=True |

## Notes

- **Python source:** `tenancy/models/contacts.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** ContactAssignments.
2. Change log + event.
