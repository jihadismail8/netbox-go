# ConfigTemplate

> Module: `extras` | Table: `extras_configtemplate` | Python class: `ConfigTemplate` | File: `extras/models/configs.py`

**Inheritance:** `ConfigContextModel`

**REST URL:** `/api/extras/config-templates/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `template_code` | TextField | Required (Jinja2 template) |
| `environment_params` | JSONField | default=dict |
| `mime_type` | CharField(255) | blank=True |
| `file_name` | CharField(255) | blank=True |
| `is_active` | BooleanField | default=True |

## Referenced By

- [DeviceRole](./../dcim/dcimDevicerole.md) via `config_template` (FK)
- [Platform](./../dcim/dcimPlatform.md) via `config_template` (FK)
- [Device](./../dcim/dcimDevice.md) via `config_template` (FK, from RenderConfigMixin)

## Notes

- **Python source:** `extras/models/configs.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
1. `clean()` validates template syntax.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** ConfigContexts referencing this template prevent deletion.
