# Webhook

> Module: `extras` | Table: `extras_webhook` | Python class: `Webhook` | File: `extras/models/models.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/extras/webhooks/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(150) | Required, unique |
| `payload_url` | CharField(500) | Required (Jinja2 template) |
| `http_method` | CharField(30) | default=POST |
| `http_content_type` | CharField(100) | default=application/json |
| `additional_headers` | TextField | blank=True |
| `body_template` | TextField | blank=True |
| `secret` | CharField(255) | blank=True |
| `ssl_verification` | BooleanField | default=True |
| `ca_file_path` | CharField(4096) | null=True |

## Notes

- **Python source:** `extras/models/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates URL, payload format, content types.
2. Save.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** EventRules referencing this webhook.
2. Change log + event.
