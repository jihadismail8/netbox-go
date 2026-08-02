# ContactAssignment

> Module: `tenancy` | Table: `tenancy_contactassignment` | Python class: `ContactAssignment` | File: `tenancy/models/contacts.py`

**Inheritance:** `ChangeLoggedModel`

**REST URL:** `/api/tenancy/contact-assignments/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyContactassignment.go`)
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
| `contact` | [Contact](./tenancyContact.md) | `CASCADE` | No | `assignments` |
| `role` | [ContactRole](./tenancyContactrole.md) | `SET_NULL` | Yes | `assignments` |

### Generic FK (object assignment)

| Field | Related Model | Via |
|-------|---------------|-----|
| `object` | (polymorphic) | `object_type` + `object_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `priority` | CharField(50) | choices=ContactPriorityChoices; null=True |

## Notes

- **Python source:** `tenancy/models/contacts.py`
- GenericFK `object` can point to any model that inherits ContactsMixin

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
Standard flow. Links Contact to any object via GenericFK.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
