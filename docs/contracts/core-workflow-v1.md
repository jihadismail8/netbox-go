# Core workflow capability contract

> Generated from `contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml`. Contract state: **pre-publication**.

| Module | Resource | REST | gRPC service | Tier | Owner |
| --- | --- | --- | --- | --- | --- |
| dcim | Site | `/api/dcim/sites/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | Manufacturer | `/api/dcim/manufacturers/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | RackRole | `/api/dcim/rack-roles/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | RackType | `/api/dcim/rack-types/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | Rack | `/api/dcim/racks/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | DeviceRole | `/api/dcim/device-roles/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | DeviceType | `/api/dcim/device-types/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | InterfaceTemplate | `/api/dcim/interface-templates/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | Device | `/api/dcim/devices/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| dcim | Interface | `/api/dcim/interfaces/` | `netbox.dcim.v1.DCIMService` | T1 | dcim-core-workflow |
| ipam | VRF | `/api/ipam/vrfs/` | `netbox.ipam.v1.IPAMService` | T1 | ipam-core-workflow |
| ipam | Prefix | `/api/ipam/prefixes/` | `netbox.ipam.v1.IPAMService` | T1 | ipam-core-workflow |
| ipam | IPAddress | `/api/ipam/ip-addresses/` | `netbox.ipam.v1.IPAMService` | T1 | ipam-core-workflow |

The identity surface is an extension with tier `not_applicable`; it cannot be counted as baseline T2 compatibility.
