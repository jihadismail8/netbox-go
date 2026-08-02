# ClusterType

> Module: `virtualization` | Table: `virtualization_clustertype` | Python class: `ClusterType` | File: `virtualization/models/clusters.py`

**Inheritance:** `OrganizationalModel`

**REST URL:** `/api/virtualization/cluster-types/`

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
1. **CASCADE:** Clusters.
2. **SET_NULL:** VirtualMachines (`vm.cluster=null`).
