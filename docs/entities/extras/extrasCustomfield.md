# CustomField

> Module: `extras` | Table: `extras_customfield` | Python class: `CustomField` | File: `extras/models/customfields.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/extras/custom-fields/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `object_types` | M2M to ContentType | Models this field applies to |
| `type` | CharField(50) | choices=CustomFieldTypeChoices; required |
| `name` | CharField(50) | Required, unique |
| `label` | CharField(50) | blank=True |
| `group_name` | CharField(50) | blank=True |
| `description` | CharField(200) | blank=True |
| `required` | BooleanField | default=False |
| `unique` | BooleanField | default=False |
| `search_weight` | PositiveIntegerField | default=1000 |
| `filter_logic` | CharField(50) | default=loose |
| `default` | JSONField | null=True |
| `related_object_type` | FK to ContentType | null=True (for object-type fields) |
| `validation_minimum` | BigIntegerField | null=True |
| `validation_maximum` | BigIntegerField | null=True |
| `validation_regex` | CharField(500) | blank=True |
| `choice_set` | FK to CustomFieldChoiceSet | null=True |

## Notes

- **Python source:** `extras/models/customfields.py`
- Custom fields are stored as JSON in `custom_field_data` on each model

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates `type`, `name` (no spaces/hyphens), default value matches type.
2. Save.
3. Schema migration: custom field added to assigned content types.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** All `custom_field_data` JSON entries referencing this field are removed.
2. Change log + event.
