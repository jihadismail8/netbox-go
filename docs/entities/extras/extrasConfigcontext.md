# ConfigContext

> Module: `extras` | Table: `extras_configcontext` | Python class: `ConfigContext` | File: `extras/models/configs.py`

**Inheritance:** `ConfigContextModel`

**REST URL:** `/api/extras/config-contexts/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `weight` | PositiveIntegerField | default=1000 |
| `is_active` | BooleanField | default=True |
| `data` | JSONField | Required (the config context data) |

### ManyToMany (assignment filters)

| Field | Related Model |
|-------|---------------|
| `regions` | Region |
| `sites` | Site |
| `locations` | Location |
| `roles` | DeviceRole |
| `platforms` | Platform |
| `cluster_types` | ClusterType |
| `cluster_groups` | ClusterGroup |
| `clusters` | Cluster |
| `tenant_groups` | TenantGroup |
| `tenants` | Tenant |
| `tags` | Tag |

## Notes

- **Python source:** `extras/models/configs.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational)

### CREATE
1. `clean()` validates at least one assignment (regions, sites, roles, etc.).
2. Save.

### UPDATE
Standard flow.

### DELETE
Change log + event.
