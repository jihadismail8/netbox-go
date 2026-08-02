# Token

> Module: `users` | Table: `users_token` | Python class: `Token` | File: `users/models.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/users/tokens/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `user` | [User](./usersUser.md) | `CASCADE` | No | `tokens` |
| `write_enabled` | BooleanField | — | — | default=False |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `key` | CharField(40) | Required, unique (API token) |
| `write_enabled` | BooleanField | default=False |
| `expires` | DateTimeField | null=True |
| `last_used` | DateTimeField | null=True |

## Notes

- **Python source:** `users/models.py`
- Token-based authentication for REST API access

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
1. `clean()` validates `user` set, token generated if not provided.
2. Save.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
