# ObjectType

> Module: `core` | Table: `core_objecttype` | Python class: `ObjectType` | File: `core/models/contenttypes.py`

**Inheritance:** Django's ContentType (extended)

**REST URL:** N/A (internal)

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `app_label` | CharField(100) | Required |
| `model` | CharField(100) | Required |
| `name` | CharField(200) | blank=True |
| `description` | CharField(200) | blank=True |
| `release` | FK to Release | SET_NULL, null=True |

## Notes

- **Python source:** `core/models/contenttypes.py`
- Extends Django's ContentType with description and release tracking

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** System-managed (Django ContentType)

### No direct CRUD

ObjectType records are managed by Django's ContentType framework and reflect the installed models. Not directly creatable, updatable, or deletable via the API.
