# JournalEntry

> Module: `extras` | Table: `extras_journalentry` | Python class: `JournalEntry` | File: `extras/models/models.py`

**Inheritance:** `ChangeLoggedModel`

**REST URL:** `/api/extras/journal-entries/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `created_by` | User | `SET_NULL` | Yes | `journal_entries` |
| `assigned_object_type` | ContentType | `PROTECT` | No | — |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `assigned_object` | (polymorphic) | `assigned_object_type` + `assigned_object_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `assigned_object_id` | PositiveBigIntegerField | GenericFK object ID |
| `kind` | CharField(50) | choices=JournalEntryKindChoices; default=info |
| `comment` | TextField | Required |

## Notes

- **Python source:** `extras/models/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates assigned object (`assigned_object_type` + `assigned_object_id`).
2. Save.
3. Change log + event.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
