# DataFile

> Module: `core` | Table: `core_datafile` | Python class: `DataFile` | File: `core/models/data.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/core/data-files/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `source` | [DataSource](./coreDatasource.md) | `CASCADE` | No | `data_files` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `path` | CharField(1000) | Required |
| `last_updated` | DateTimeField | auto_now |
| `size` | PositiveBigIntegerField | null=True |
| `hash` | CharField(64) | blank=True |

## Notes

- **Python source:** `core/models/data.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates `source` set.
2. Save.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
