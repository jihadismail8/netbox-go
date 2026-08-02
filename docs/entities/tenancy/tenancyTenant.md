# Tenant

> Module: `tenancy` | Table: `tenancy_tenant` | Python class: `Tenant` | File: `tenancy/models/tenants.py`

**Inheritance:** `ContactsMixin <- PrimaryModel`

**REST URL:** `/api/tenancy/tenants/`

## Implementation Status

- [ ] Go model (`internal/model/tenancyTenant.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `group` | [TenantGroup](./tenancyTenantgroup.md) | `SET_NULL` | Yes | `tenants` |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required, unique |
| `slug` | SlugField(100) | Required, unique |

## Referenced By

Referenced by nearly all models with tenancy support (Site, Rack, Device, Cluster, Prefix, IPAddress, VLAN, etc.)

## Notes

- **Python source:** `tenancy/models/tenants.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Organizational), B (Counter Source)

### CREATE
Standard flow.

### UPDATE
Standard flow.

### DELETE
1. **PROTECT:** Most models with `tenant` FK prevent deletion (Devices, Racks, IPs, etc.).
2. Change log + event.

### Interdependencies
- **Counter fields:** `circuit_count`, `site_count`, `rack_count`, `device_count`, `vrf_count`, `prefix_count`, `ipaddress_count`, `vlan_count`, `cluster_count`, `virtual_machine_count`.
