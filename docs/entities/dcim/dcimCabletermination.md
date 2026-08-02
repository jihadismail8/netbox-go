# CableTermination

> Module: `dcim` | Table: `dcim_cabletermination` | Python class: `CableTermination` | File: `dcim/models/cables.py`

**Inheritance:** `ChangeLoggedModel`

**REST URL:** N/A (internal model, accessed via Cable)

## Implementation Status

- [ ] Go model (`internal/model/dcimCabletermination.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition (`api/netbox_go/v1/dcimCabletermination.proto`)
- [ ] Proto generated code
- [ ] DAO layer (`internal/dao/dcimCabletermination.go`)
- [ ] DAO unit tests
- [ ] Cache layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] Bulk operations
- [ ] Filtering support
- [ ] Pagination support
- [ ] RBAC / permissions
- [ ] API integration tests
- [ ] Vue.js views

## Django Model Fields (from Python source)

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `cable` | [Cable](./dcimCable.md) | `CASCADE` | No | `terminations` |
| `termination_type` | ContentType | `PROTECT` | No | `+` |
| `_device` | [Device](./dcimDevice.md) | `CASCADE` | Yes | — |
| `_rack` | [Rack](./dcimRack.md) | `CASCADE` | Yes | — |
| `_location` | [Location](./dcimLocation.md) | `CASCADE` | Yes | — |
| `_site` | [Site](./dcimSite.md) | `CASCADE` | Yes | — |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `termination` | (polymorphic) | `termination_type` (ContentType) + `termination_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `cable_end` | CharField(1) | choices=CableEndChoices (A or B) |
| `termination_id` | PositiveBigIntegerField | GenericFK object ID |

## Constraints

- UniqueConstraint: `(termination_type, termination_id)` — each termination object used once

## Notes

- **Python source:** `dcim/models/cables.py`
- Maps Cable to its termination endpoints (polymorphic via GenericFK)
- Cached `_device`, `_rack`, `_location`, `_site` fields populated by `cache_related_objects()` for efficient filtering
- `save()` sets cable on the termination object; `delete()` clears it

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected)

### CREATE
1. Created/deleted when a Cable is created/deleted.
2. Cache (`_site`, `_location`, `_rack`) from termination endpoint.
3. Change log + event.

### UPDATE
Standard flow.

### DELETE
1. Cable deletion cascades to CableTermination.
2. Change log + event.
