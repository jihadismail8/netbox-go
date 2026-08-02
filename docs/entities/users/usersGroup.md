# Group

> Module: `users` | Table: `auth_group` | Python class: `Group` | File: Django built-in

**Inheritance:** Django's `Group`

**REST URL:** `/api/users/groups/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(150) | Required, unique |

## Notes

- Standard Django auth Group model

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** Group memberships.
2. Change log + event.
