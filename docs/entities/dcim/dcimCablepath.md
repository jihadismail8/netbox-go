# CablePath

> Module: `dcim` | Table: `dcim_cablepath` | Python class: `CablePath` | File: `dcim/models/cables.py`

**Inheritance:** `models.Model` (standalone, no NetBox base)

**REST URL:** N/A (internal model)

## Implementation Status

- [ ] Go model (`internal/model/dcimCablepath.go`)
- [ ] GORM mapping verified
- [ ] Column whitelist complete
- [ ] Proto definition
- [ ] Proto generated code
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] API integration tests

## Django Model Fields (from Python source)

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `origin` | (GenericFK, polymorphic) | — | — | — |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `path` | JSONField | List of (type, id) tuples representing path nodes |
| `is_active` | BooleanField | default=False |
| `is_complete` | BooleanField | default=False |
| `is_split` | BooleanField | default=False |
| `_nodes` | PathField (custom) | Flattened list of all nodes for filtering |

## Notes

- **Python source:** `dcim/models/cables.py`
- `_netbox_private = True` — not exposed via REST API directly
- Represents physical cable/wireless path from origin to destination
- `from_origin()` classmethod traces path through cables, front/rear ports, circuits
- `retrace()` retraces path when cable topology changes
- Complex path tracing logic including split paths and circuit terminations

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Auto-generated, read-only)

### No direct CRUD

CablePath records are automatically created/updated/deleted by the path tracing system whenever a Cable is created/updated/deleted. They cannot be directly created or modified via the API.
