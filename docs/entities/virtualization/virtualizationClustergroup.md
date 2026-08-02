# ClusterGroup

> Module: `virtualization` | Table: `virtualization_clustergroup` | Python class: `ClusterGroup` | File: `virtualization/models/clusters.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/virtualization/cluster-groups/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |

## Notes

- **Python source:** `virtualization/models/clusters.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **SET_NULL:** Clusters (`cluster.group=null`).
