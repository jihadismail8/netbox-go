# User

> Module: `users` | Table: `users_user` | Python class: `User` | File: `users/models.py`

**Inheritance:** Django's `AbstractUser`

**REST URL:** `/api/users/users/`

## Django Model Fields

Standard Django User fields plus:

| Field | Type | Notes |
|-------|------|-------|
| `username` | CharField(150) | Required, unique |
| `first_name` | CharField(150) | blank=True |
| `last_name` | CharField(150) | blank=True |
| `email` | EmailField | blank=True |
| `is_staff` | BooleanField | default=False |
| `is_active` | BooleanField | default=True |
| `is_superuser` | BooleanField | default=False |
| `date_joined` | DateTimeField | auto_now_add |
| `config` | JSONField | default=dict (user preferences) |

## Notes

- **Python source:** `users/models.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates username, email.
2. Password hashed.
3. Save.

### UPDATE
Standard flow.

### DELETE
1. **CASCADE:** Tokens, Group memberships.
2. Change log + event.
