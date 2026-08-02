# ContactGroup

> Module: `tenancy` | Table: `tenancy_contactgroup` | Python class: `ContactGroup` | File: `tenancy/models/contacts.py`

**Inheritance:** `NestedGroupModel`

**REST URL:** `/api/tenancy/contact-groups/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyContactgroup.go`)
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
| `parent` | [ContactGroup](./tenancyContactgroup.md) (self) | `SET_NULL` | Yes | `children` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |

## Notes

- **Python source:** `tenancy/models/contacts.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), C (Cache Source via MPTT)

### CREATE
1. `clean()` validates parent group exists.
2. MPTT computed.
3. Save.

### UPDATE
If `parent` changed: MPTT re-balanced. Save.

### DELETE
1. **CASCADE:** Sub-groups.
2. **SET_NULL:** Contacts (`contact.group=null`).
3. MPTT re-balanced.
