# ImageAttachment

> Module: `extras` | Table: `extras_imageattachment` | Python class: `ImageAttachment` | File: `extras/models/models.py`

**Inheritance:** `NetBoxModel`

**REST URL:** N/A (managed via parent object)

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `content_type` | ContentType | `CASCADE` | No | — |

### Generic FK

| Field | Related Model | Via |
|-------|---------------|-----|
| `parent` | (polymorphic) | `content_type` + `object_id` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `object_id` | PositiveBigIntegerField | GenericFK object ID |
| `image` | ImageField | upload_to=image-attachments |
| `name` | CharField(50) | blank=True |

## Notes

- **Python source:** `extras/models/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates `content_type` + `object_id` (parent object).
2. Image file processed, resized.
3. Save.

### UPDATE
Standard flow.

### DELETE
1. Image file deleted from storage.
2. Change log + event.
