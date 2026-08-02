# Tag

> Module: `extras` | Table: `extras_tag` | Python class: `Tag` | File: `extras/models/tags.py`

**Inheritance:** `NestedGroupModel` (Changed in recent versions) / `NetBoxModel`

**REST URL:** `/api/extras/tags/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |
| `color` | ColorField | default=grey |
| `description` | CharField(200) | blank=True |
| `object_types` | M2M to ContentType | Restrict tag to specific content types |

## Notes

- **Python source:** `extras/models/tags.py`
- Tags are attached to models via `TaggableManager` from django-taggit

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **M2M auto-removed:** All tagged objects lose this tag.
2. Change log + event.
