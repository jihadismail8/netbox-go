# FHRPGroupAssignment

> Module: `ipam` | Table: `ipam_fhrpgroupassignment` | Python class: `FHRPGroupAssignment` | File: `ipam/models/fhrp.py`

**Inheritance:** `ChangeLoggedModel`

**REST URL:** `/api/ipam/fhrp-group-assignments/`

## Implementation Status

- [ ] Go model (`internal/model/ipamFhrpgroupassignment.go`)
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
| `group` | [FHRPGroup](./ipamFhrpgroup.md) | `CASCADE` | No | `assignments` |

### Generic FK (interface assignment)

| Field | Related Model | Via |
|-------|---------------|-----|
| `interface` | (polymorphic: Interface, VMInterface) | `interface_type` + `interface_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `priority` | PositiveIntegerField | Required; min=0, max=255 |

## Notes

- **Python source:** `ipam/models/fhrp.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
Standard flow. Links Interface to FHRPGroup.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
