# CustomFieldChoiceSet

> Module: `extras` | Table: `extras_customfieldchoiceset` | Python class: `CustomFieldChoiceSet` | File: `extras/models/customfields.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/extras/custom-field-choice-sets/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(50) | Required, unique |
| `description` | CharField(200) | blank=True |
| `base_choices` | CharField(50) | null=True (predefined choice sets) |
| `extra_choices` | JSONField | default=dict |
| `order_alphabetically` | BooleanField | default=False |

## Notes

- **Python source:** `extras/models/customfields.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates choice set format.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** CustomFields referencing this set (`custom_field.choices=null`).
