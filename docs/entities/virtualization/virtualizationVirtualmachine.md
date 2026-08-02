# VirtualMachine

> Module: `virtualization` | Table: `virtualization_virtualmachine` | Python class: `VirtualMachine` | File: `virtualization/models/virtualmachines.py`

**Inheritance:** `ContactsMixin <- ConfigContextModel <- PrimaryModel`

**REST URL:** `/api/virtualization/virtual-machines/`

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `cluster` | [Cluster](./virtualizationCluster.md) | `SET_NULL` | Yes | `virtual_machines` |
| `device` | [Device](./../dcim/dcimDevice.md) | `SET_NULL` | Yes | `virtual_machines` |
| `role` | [DeviceRole](./../dcim/dcimDevicerole.md) | `SET_NULL` | Yes | `virtual_machines` |
| `platform` | [Platform](./../dcim/dcimPlatform.md) | `SET_NULL` | Yes | `virtual_machines` |
| `tenant` | [Tenant](./../tenancy/tenancyTenant.md) | `PROTECT` | Yes | `virtual_machines` |
| `site` | [Site](./../dcim/dcimSite.md) | `PROTECT` | Yes | `virtual_machines` |
| `primary_ip4` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` |
| `primary_ip6` | [IPAddress](./../ipam/ipamIpaddress.md) | `SET_NULL` | Yes | `+` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(64) | Required; db_collation="natural_sort" |
| `status` | CharField(50) | choices=VirtualMachineStatusChoices; default=active |
| `serial` | CharField(50) | blank=True |
| `vcpus` | PositiveSmallIntegerField | null=True |
| `memory` | PositiveIntegerField | null=True (MB) |
| `disk` | PositiveIntegerField | null=True (GB) |

## Notes

- **Python source:** `virtualization/models/virtualmachines.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source), C (Cache Source)

### CREATE
1. `clean()` validates `status`, unique (cluster, tenant, serial) if serial set.
2. Save.
3. Change log + event.

### UPDATE
1. Snapshot.
2. **CRITICAL:** If `cluster` changed — signal refreshes all VM interfaces' `_cluster` cache.
3. If `primary_ip4`/`primary_ip6` changed: validates IP assigned to this VM.
4. Save.
5. Change log + event.

### DELETE
1. **CASCADE:** VM Interfaces, Services (via GenericFK SET_NULL).
2. **Counter decrement:** Cluster, Site, Tenant.
3. **Primary IP:** Not relevant (VM being deleted).
4. Change log + event.

### Interdependencies
- **Counter fields:** None directly, but referenced by Cluster, Site, Tenant counters.
- **Cache source for:** VMInterface (`_cluster`).
