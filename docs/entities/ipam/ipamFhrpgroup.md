# FHRPGroup

> Module: `ipam` | Table: `ipam_fhrpgroup` | Python class: `FHRPGroup` | File: `ipam/models/fhrp.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/fhrp-groups/`

## Implementation Status

- [ ] Go model (`internal/model/ipamFhrpgroup.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] Proto generated code
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `fhrp_groups` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `group_id` | PositiveIntegerField | Required (1-65535) |
| `protocol` | CharField(50) | choices=FHRPGroupProtocolChoices; required |
| `auth_type` | CharField(50) | choices=FHRPGroupAuthTypeChoices; null=True |
| `auth_key` | CharField(255) | blank=True |

## Referenced By

- [FHRPGroupAssignment](./ipamFhrpgroupassignment.md) via `group` (FK)
- [IPAddress](./ipamIpaddress.md) via `assigned_object` (GenericFK)

## Notes

- **Python source:** `ipam/models/fhrp.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates protocol, group_id, auth_type.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** FHRPGroupAssignments.
2. Change log + event.
