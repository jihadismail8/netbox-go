# NetBox Go - Entity Implementation Documentation

> **Generated/reference inventory, not implementation status.** These documents are derived from the pinned post-4.4.6 source snapshot; checklist boxes and status labels inside them do not prove that behavior is implemented or compatible. Use the canonical [status](../STATUS.md), [compatibility contract](../COMPATIBILITY.md), and [roadmap](../ROADMAP.md).

This directory contains implementation documentation for all NetBox entities, derived from the Python Django source code.

## Modules

- [DCIM](./dcim/README.md) - Data Center Infrastructure Management (40+ models)
- [IPAM](./ipam/README.md) - IP Address Management (18 models)
- [Tenancy](./tenancy/) - Tenants & Contacts (6 models)
- [Circuits](./circuits/) - Circuits & Providers (6 models)
- [Virtualization](./virtualization/) - VMs & Clusters (5 models)
- [VPN](./vpn/) - Tunnels & L2VPN (5 models)
- [Wireless](./wireless/) - Wireless LANs & Links (3 models)
- [Extras](./extras/) - Tags, Custom Fields, Config Contexts (8 models)
- [Users](./users/) - Users, Groups, Tokens (3 models)
- [Core](./core/) - Object Changes, Jobs, Data (5 models)

## Documentation Format

Each entity doc includes:
- Metadata (module, table, Python class, source file)
- Inheritance hierarchy
- REST URL
- Implementation status checklist (26 items)
- Django model fields from Python source (FK, O2O, M2M, regular, inherited)
- Constraints
- Dependencies
- Referenced-by relationships
- Implementation notes

## Source

All field data is extracted from the Python NetBox source in `/netbox/netbox/` (Django models).
