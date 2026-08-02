# API Serialization Patterns & Custom Endpoints Reference

> Extracted from Python NetBox REST API for porting to Go.
> **Source:** `netbox/netbox/{dcim,ipam,circuits,virtualization,vpn}/api/views.py`, `netbox/netbox/netbox/api/`

---

## Base Architecture

### ViewSet Hierarchy

```
BaseViewSet
  ├─ Object-level permission enforcement (HTTP method -> action mapping)
  ├─ Brief mode (?brief=true)
  ├─ Dynamic prefetch/annotation resolution
  │
  ├─ NetBoxReadOnlyModelViewSet (GET, OPTIONS, HEAD only)
  │
  └─ NetBoxModelViewSet (full CRUD + bulk ops)
       ├─ BulkUpdateModelMixin (PATCH/PUT on lists)
       ├─ BulkDestroyModelMixin (DELETE on lists)
       ├─ ObjectValidationMixin (validation before create/update/delete)
       └─ CustomFieldsMixin (custom_field_data handling)
```

### HTTP Method -> Permission Action Map

| HTTP Method | Permission Action |
|-------------|-------------------|
| GET | view |
| OPTIONS | (none) |
| HEAD | view |
| POST | add |
| PUT | change |
| PATCH | change |
| DELETE | delete |

### Response Format

**List Response:**
```json
{
  "count": 150,
  "next": "https://netbox/api/dcim/sites/?limit=50&offset=50",
  "previous": null,
  "results": [ { ...objects... } ]
}
```

**Detail Response** (PrimaryModel):
```json
{
  "id": 1,
  "url": "https://netbox/api/dcim/sites/1/",
  "display_url": "https://netbox/dcim/sites/1/",
  "display": "Site Name",
  "name": "Site Name",
  "slug": "site-name",
  "...model fields...": "...",
  "description": "",
  "comments": "",
  "tags": [],
  "custom_fields": {},
  "created": "2024-01-15T10:30:00Z",
  "last_updated": "2024-01-15T10:30:00Z"
}
```

**Nested Object Format** (in responses):
```json
{
  "id": 1,
  "url": "https://netbox/api/dcim/sites/1/",
  "display": "Site Name",
  "name": "Site Name",
  "slug": "site-name"
}
```

**Choice Field Format:**
```json
{
  "value": "active",
  "label": "Active"
}
```

---

## Custom Action Endpoints

### DCIM Custom Actions

| Endpoint | HTTP | Path | Description |
|----------|------|------|-------------|
| cables/trace/ | GET | /api/dcim/cables/{id}/trace/ | Trace cable path from origin |
| connected-device/ | POST | /api/dcim/connected-device/ | Find device reachable from interface |
| device-bays/{id}/populate/ | POST | /api/dcim/device-bays/{id}/populate/ | Auto-populate device bay |
| front-ports/{id}/paths/ | GET | /api/dcim/front-ports/{id}/paths/ | Get cable paths from front port |
| interfaces/{id}/paths/ | GET | /api/dcim/interfaces/{id}/paths/ | Get cable paths from interface |
| interfaces/{id}/trace/ | GET | /api/dcim/interfaces/{id}/trace/ | Trace cable from interface |
| racks/elevation/ | GET | /api/dcim/racks/{id}/elevation/ | Get rack elevation display |
| rear-ports/{id}/paths/ | GET | /api/dcim/rear-ports/{id}/paths/ | Get cable paths from rear port |

### IPAM Custom Actions

| Endpoint | HTTP | Path | Description |
|----------|------|------|-------------|
| aggregates/{id}/available-prefixes/ | GET | /api/ipam/aggregates/{id}/available-prefixes/ | List available child prefixes |
| aggregates/{id}/available-prefixes/ | POST | /api/ipam/aggregates/{id}/available-prefixes/ | Create new prefix(s) from available space |
| prefixes/{id}/available-ips/ | GET | /api/ipam/prefixes/{id}/available-ips/ | List available IPs in prefix |
| prefixes/{id}/available-ips/ | POST | /api/ipam/prefixes/{id}/available-ips/ | Create new IP(s) from available space |
| prefixes/{id}/available-prefixes/ | GET | /api/ipam/prefixes/{id}/available-prefixes/ | List available child prefixes |
| prefixes/{id}/available-prefixes/ | POST | /api/ipam/prefixes/{id}/available-prefixes/ | Create child prefix(s) |
| ip-ranges/{id}/available-ips/ | GET | /api/ipam/ip-ranges/{id}/available-ips/ | List available IPs in range |
| ip-ranges/{id}/available-ips/ | POST | /api/ipam/ip-ranges/{id}/available-ips/ | Create IP(s) from range |

### Available IPs/Prefixes Response Format

**GET available-prefixes:**
```json
[
  {"family": 4, "prefix": "10.0.0.0/24"},
  {"family": 4, "prefix": "10.0.1.0/24"}
]
```

**POST available-prefixes** (request body):
```json
{
  "prefix_length": 24,
  "description": "New subnet",
  "quantity": 1
}
```

---

## Bulk Operations

### Bulk Create
```
POST /api/dcim/sites/
Content-Type: application/json

[
  {"name": "Site A", "slug": "site-a"},
  {"name": "Site B", "slug": "site-b"}
]
```
Response: 201 Created with array of created objects.

### Bulk Update (PATCH)
```
PATCH /api/dcim/sites/
Content-Type: application/json

{
  "ids": [1, 2, 3],
  "data": {"status": "planned"}
}
```

### Bulk Delete
```
DELETE /api/dcim/sites/
Content-Type: application/json

{
  "ids": [1, 2, 3]
}
```

---

## Count Fields (Read-Only Annotations)

| Model | Count Fields |
|-------|-------------|
| Site | circuit_count, device_count, prefix_count, rack_count, virtualmachine_count, vlan_count |
| Region | site_count, circuit_count, device_count, prefix_count, rack_count, virtualmachine_count, vlan_count |
| Manufacturer | devicetype_count, inventoryitem_count, platform_count |
| DeviceType | device_count, consoleport_count, consoleserverport_count, powerport_count, poweroutlet_count, interface_count, frontport_count, rearport_count, devicebay_count, modulebay_count, inventoryitem_count |
| DeviceRole | device_count, virtualmachine_count |
| RackRole | rack_count |
| Platform | device_count, virtualmachine_count |
| VLANGroup | vlan_count |
| VLAN | prefix_count |
| VRF | prefix_count, ipaddress_count |
| RIR | aggregate_count |
| Tenant | circuit_count, device_count, ipaddress_count, prefix_count, rack_count, site_count, virtualmachine_count, vlan_count |
| Cluster | device_count, virtualmachine_count |
| Provider | circuit_count, asccount |

---

## Standard Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| limit | int | Results per page (default 50, max 1000) |
| offset | int | Pagination offset |
| brief | bool | Return brief objects (id, url, display, name/slug only) |
| fields | csv | Comma-separated list of fields to include |
| excluding | csv | Comma-separated fields to exclude |
| created__gte | datetime | Created on or after |
| created__lte | datetime | Created on or before |
| last_updated__gte | datetime | Updated on or after |
| last_updated__lte | datetime | Updated on or before |
| created_by_request | uuid | Created by request ID |
| updated_by_request | uuid | Updated by request ID |
| q | string | Full-text search query |
| tag | csv | Filter by tags |
| id | csv | Filter by IDs |
| ordering | csv | Sort fields (prefix - for descending) |
