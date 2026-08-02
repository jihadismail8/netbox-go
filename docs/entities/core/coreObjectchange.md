# ObjectChange

> Module: `core` | Table: `core_objectchange` | Python class: `ObjectChange` | File: `core/models/change_logging.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/core/object-changes/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `user` | User | `SET_NULL` | Yes | `changes` |
| `changed_object_type` | ObjectType | `PROTECT` | No | — |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `changed_object` | (polymorphic) | `changed_object_type` + `changed_object_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `request_id` | UUIDField | Required |
| `action` | CharField(50) | choices=ObjectChangeActionChoices |
| `prechange_data` | JSONField | null=True |
| `postchange_data` | JSONField | null=True |
| `object_repr` | CharField(200) | Required |

## Notes

- **Python source:** `core/models/change_logging.py`
- Change logging audit trail

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** System-managed (read-only)

### No direct CRUD via API

ObjectChange records are automatically created by `ChangeLoggingMixin` on every create/update/delete of any NetBoxModel. They are read-only and cannot be created, updated, or deleted via the API.
