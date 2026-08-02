# Service

> Module: `ipam` | Table: `ipam_service` | Python class: `Service` | File: `ipam/models/services.py`

**Inheritance:** `PrimaryModel <- ServiceBase`

**REST URL:** `/api/ipam/services/`

## Implementation Status

- [ ] Go model (`internal/model/ipamService.go`)
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
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `services` |

### Generic FK (parent object)

| Field | Related Model | Via |
|-------|---------------|-----|
| `parent_object` | (polymorphic: Device, VirtualMachine) | `parent_object_type` + `parent_object_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |
| `protocol` | CharField(50) | choices=ServiceProtocolChoices; required |
| `ports` | ArrayField | base_field=PositiveIntegerField; required |

## Notes

- **Python source:** `ipam/models/services.py`
- `parent_object` is GenericFK to Device or VirtualMachine
- `ports` is PostgreSQL ArrayField storing port numbers

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple), F (Scoped via parent)

### CREATE
Standard flow. Assigned to Device or VirtualMachine via GenericFK.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
