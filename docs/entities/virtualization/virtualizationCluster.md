# Cluster

> Module: `virtualization` | Table: `virtualization_cluster` | Python class: `Cluster` | File: `virtualization/models/clusters.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/virtualization/clusters/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `type` | [ClusterType](./virtualizationClustertype.md) | `PROTECT` | No | `clusters` |
| `group` | [ClusterGroup](./virtualizationClustergroup.md) | `SET_NULL` | Yes | `clusters` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `clusters` |
| `site` | [Site](./../dcim/dcimSite.md) | `PROTECT` | Yes | `clusters` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |

## Notes

- **Python source:** `virtualization/models/clusters.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), F (Scoped)

### CREATE
1. `clean()` validates `type` set, optional `group`/`site`.
2. Save.

### UPDATE
1. If `site` changed: `_region`/`_site_group` re-cached.
2. Save.

### DELETE
1. **SET_NULL:** VirtualMachines (`vm.cluster=null`).
2. **CASCADE:** child clusters.
3. Change log + event.

### Interdependencies
- **Counter fields:** `virtual_machine_count`.
