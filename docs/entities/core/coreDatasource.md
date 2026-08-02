# DataSource

> Module: `core` | Table: `core_datasource` | Python class: `DataSource` | File: `core/models/data.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/core/data-sources/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `type` | CharField(50) | Required (backend type) |
| `source_url` | CharField(200) | Required |
| `enabled` | BooleanField | default=True |
| `status` | CharField(50) | choices=DataSourceStatusChoices |
| `description` | CharField(200) | blank=True |
| `comments` | TextField | blank=True |
| `last_synced` | DateTimeField | null=True |
| `parameters` | JSONField | default=dict |

## Notes

- **Python source:** `core/models/data.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `type`, URL/path.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** DataFiles.
2. Change log + event.
