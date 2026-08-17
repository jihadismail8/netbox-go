# NetBox CRUD Paradigm & Interdependency Map

> **Quarantined historical analysis, not accepted architecture or business rules.** This document mixes upstream observations with obsolete proposals for GORM hooks, generic post-hooks, Redis events, caches, and one-to-one model translation. Do not implement from those proposals. Its source baseline is the post-4.4.6 snapshot at commit `fbb948d30e79ce657fac62994a22aca72c1770a9`; every behavior must be reconciled through an accepted Capability Profile and the canonical [architecture](ARCHITECTURE.md), [coding standards](CODING_STANDARDS.md), and [compatibility contract](COMPATIBILITY.md).

> **Historical purpose:** Earlier entity pages reference these candidate patterns. Those references are discovery pointers only; this file does not define current Go behavior, accepted scope, or completion.

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [Model Hierarchy & Shared Behavior](#2-model-hierarchy--shared-behavior)
3. [The Event Pipeline (All Models)](#3-the-event-pipeline-all-models)
4. [Change Logging (All NetBoxModel)](#4-change-logging-all-netboxmodel)
5. [Denormalized Cache System](#5-denormalized-cache-system)
6. [Counter Cache Fields](#6-counter-cache-fields)
7. [Cascading Delete Rules](#7-cascading-delete-rules)
8. [Cross-Model Validation Patterns](#8-cross-model-validation-patterns)
9. [Cable & Path Tracing System](#9-cable--path-tracing-system)
10. [IPAM Hierarchy System](#10-ipam-hierarchy-system)
11. [Signal Handler Reference](#11-signal-handler-reference)
12. [Per-Entity CRUD Categories](#12-per-entity-crud-categories)

---

## 1. System Architecture

NetBox is a Django monolith with three behavioral layers on every model:

```
┌─────────────────────────────────────────────────────┐
│              HTTP Request (REST/GraphQL)             │
├─────────────────────────────────────────────────────┤
│  Serializer Validation → model.clean()              │
│  Cross-field & cross-entity validation               │
├─────────────────────────────────────────────────────┤
│  Model.save() / Model.delete()                      │
│  Denormalized cache, counters, hierarchy fields      │
├─────────────────────────────────────────────────────┤
│  Django Signals (pre_save/post_save/pre_delete/...)  │
│  Path tracing, prefix depth, IP cleanup, counters    │
├─────────────────────────────────────────────────────┤
│  ChangeLoggingMixin → ObjectChange record            │
│  EventRulesMixin → Webhooks, scripts, notifications  │
└─────────────────────────────────────────────────────┘
```

### Go Rewrite Implications

| Python Mechanism                       | Go Equivalent                                            |
| -------------------------------------- | -------------------------------------------------------- |
| `model.save()`                         | Service-layer `Create()`/`Update()` hooks                |
| `model.delete()`                       | Service-layer `Delete()` hooks                           |
| `model.clean()`                        | Service-layer `Validate()` / DAO constraints             |
| Django signals (`post_save`, etc.)     | Service-layer post-hooks or event bus                    |
| `ChangeLoggingMixin`                   | GORM hooks (`AfterCreate`, `AfterUpdate`, `AfterDelete`) |
| `EventRulesMixin`                      | Async job queue (Redis pub/sub)                          |
| Denormalized caches (`_site`, `_rack`) | Explicit cache-refresh calls in service layer            |
| `CounterCacheField`                    | Counter-refresh function after create/delete             |
| GenericForeignKey                      | `content_type_id` + `object_id` columns with join helper |

---

## 2. Model Hierarchy & Shared Behavior

```
models.Model
├── ChangeLoggedModel (ChangeLoggingMixin + CustomValidationMixin + EventRulesMixin)
│   └── snapshot() on save → _prechange_snapshot for changelog diff
├── NetBoxModel(NetBoxFeatureSet)
│   ├── clean() → validates GenericForeignKey fields (content_type + object_id both set)
│   ├── SerializedQuerySet → RestrictedQuerySet (RBAC filtering)
│   ├── Features: tags, custom_fields, journal_entries, change_logs,
│   │            custom_links, export_templates, webhooks, event_rules
│   ├── PrimaryModel → adds: description, comments
│   ├── NestedGroupModel → adds: MPTT tree (lft, rgt, tree_id, level) + parent FK
│   └── OrganizationalModel → adds: name, slug (unique)
```

### Shared CRUD behavior for ALL NetBoxModel subclasses:

| Operation  | Behavior                                                                                                                                                                                     |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **CREATE** | 1. `snapshot()` stores pre-change state (empty for create). 2. `clean()` validates GFK fields. 3. GORM save. 4. `post_save` signal fires. 5. Event queued.                                   |
| **UPDATE** | 1. `snapshot()` stores pre-change state for diff. 2. `clean()` re-validates. 3. GORM save. 4. `post_save` signal fires. 5. ChangeLoggedMixin creates `ObjectChange` record. 6. Event queued. |
| **DELETE** | 1. `snapshot()` stores pre-change state. 2. `pre_delete` signal fires (cleanup child objects). 3. GORM delete. 4. `post_delete` signal fires (counter decrements). 5. Event queued.          |

---

## 3. The Event Pipeline (All Models)

Every CRUD operation flows through a **request-scoped event queue**:

```python
# pseudocode
@contextmanager
def event_tracking(request):
    # Events are queued, not dispatched immediately
    event_queue = ContextVar('event_queue')
    yield
    # Flush after request completes
    for event in event_queue:
        process_event(event)  # → webhooks, scripts, notifications
```

### Event Types

| Constant         | Trigger                   |
| ---------------- | ------------------------- |
| `OBJECT_CREATED` | post_save (created=True)  |
| `OBJECT_UPDATED` | post_save (created=False) |
| `OBJECT_DELETED` | post_delete               |

### Go Rewrite: Use Redis pub/sub or a goroutine-based event queue flushed at end of request.

---

## 4. Change Logging (All NetBoxModel)

Every create/update/delete on a `NetBoxModel` produces an `ObjectChange` record:

| Field                 | Source                         |
| --------------------- | ------------------------------ |
| `user`                | From request context           |
| `user_name`           | Denormalized                   |
| `request_id`          | UUID per request               |
| `action`              | `create` / `update` / `delete` |
| `changed_object_type` | ContentType of the model       |
| `changed_object_id`   | PK                             |
| `object_repr`         | `str(instance)`                |
| `prechange_data`      | JSON snapshot before save      |
| `postchange_data`     | JSON snapshot after save       |

### Go Rewrite: GORM `AfterCreate`/`AfterUpdate`/`AfterDelete` hooks write to `object_changes` table.

---

## 5. Denormalized Cache System

Several models cache FK references in `_` prefixed fields for query performance:

| Cached Field  | Source                                       | Found On                                                  |
| ------------- | -------------------------------------------- | --------------------------------------------------------- |
| `_site`       | Device's `site`                              | All ComponentModels, CableTermination, CircuitTermination |
| `_location`   | Device's `location`                          | All ComponentModels, CableTermination                     |
| `_rack`       | Device's `rack`                              | All ComponentModels, CableTermination                     |
| `_region`     | Scope's region (if scope is a Site/Location) | Prefix, VLAN, Cluster                                     |
| `_site_group` | Scope's site_group (if scope is a Site)      | Prefix, VLAN, Cluster                                     |
| `_cluster`    | VirtualMachine's cluster                     | VM Interface                                              |
| `_device`     | Interface's device                           | IPAddress (primary IP), Cable path                        |
| `_tenant`     | Scoped object's tenant                       | Prefix                                                    |

### When caches update:

| Trigger                                                                  | What Updates                                                  |
| ------------------------------------------------------------------------ | ------------------------------------------------------------- |
| **Device.save()** (site/location/rack changed)                           | All child components' `_site`, `_location`, `_rack` re-cached |
| **Interface.save()** (device changed — not allowed except InventoryItem) | `_site`, `_location`, `_rack` re-cached                       |
| **VirtualMachine.save()** (cluster changed)                              | All VM interfaces' `_cluster` re-cached                       |
| **Prefix/VLAN/Cluster.save()** (scope changed)                           | `_region`, `_site_group` re-cached from new scope             |

### Go Rewrite: Service layer must call `refresh_cached_scopes(obj)` after every relevant update.

---

## 6. Counter Cache Fields

Many parent models store denormalized counts of children:

| Parent Model     | Counter Fields                                                                                                                                                                                                                                                                                                  |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **DeviceType**   | `console_port_template_count`, `console_server_port_template_count`, `power_port_template_count`, `power_outlet_template_count`, `interface_template_count`, `front_port_template_count`, `rear_port_template_count`, `device_bay_template_count`, `module_bay_template_count`, `inventory_item_template_count` |
| **Device**       | `console_port_count`, `console_server_port_count`, `power_port_count`, `power_outlet_count`, `interface_count`, `front_port_count`, `rear_port_count`, `device_bay_count`, `module_bay_count`, `inventory_item_count`                                                                                           |
| **Site**         | `rack_count`, `device_count`, `prefix_count`, `vlan_count`, `circuit_count`, `virtual_machine_count`                                                                                                                                                                                                            |
| **Rack**         | `device_count`                                                                                                                                                                                                                                                                                                  |
| **VRF**          | `prefix_count`, `ipaddress_count`                                                                                                                                                                                                                                                                               |
| **VLANGroup**    | `vlan_count`                                                                                                                                                                                                                                                                                                    |
| **Tenant**       | `circuit_count`, `site_count`, `rack_count`, `device_count`, `vrf_count`, `prefix_count`, `ipaddress_count`, `vlan_count`, `cluster_count`, `virtual_machine_count`                                                                                                                                             |
| **Manufacturer** | `device_type_count`                                                                                                                                                                                                                                                                                             |

### When counters update:

Triggered by signal on child create/delete. E.g., creating an Interface on a Device increments `Device.interface_count`.

### Go Rewrite: After create/delete of any child entity, call `update_parent_counter(parentType, parentID, delta)`.

---

## 7. Cascading Delete Rules

### CASCADE (parent deletion deletes children)

| Parent             | Children Deleted                                                                                                                                                                                                     |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Region**         | Sub-regions, Locations (CASCADE via parent FK)                                                                                                                                                                       |
| **SiteGroup**      | Sub-groups, Sites (SET_NULL for sites)                                                                                                                                                                               |
| **Site**           | Locations, Racks (CASCADE), PowerPanels, CablePaths starting here                                                                                                                                                    |
| **Location**       | Sub-locations, Racks                                                                                                                                                                                                 |
| **Rack**           | RackReservations, Devices (SET_NULL for device.rack)                                                                                                                                                                 |
| **Device**         | ALL components (ConsolePort, ConsoleServerPort, PowerPort, PowerOutlet, Interface, FrontPort, RearPort, DeviceBay, ModuleBay, InventoryItem), ConsolePort/PowerPort templates (no), cables attached to its endpoints |
| **ModuleType**     | Modules                                                                                                                                                                                                              |
| **Manufacturer**   | DeviceTypes (PROTECT), Platforms (PROTECT)                                                                                                                                                                           |
| **DeviceType**     | Component templates (ConsolePortTemplate, etc.)                                                                                                                                                                      |
| **VRF**            | Prefixes (SET_NULL), IPAddresses (SET_NULL)                                                                                                                                                                          |
| **VLANGroup**      | VLANs (SET_NULL), child VLANGroups                                                                                                                                                                                   |
| **RIR**            | Aggregates (SET_NULL)                                                                                                                                                                                                |
| **Cluster**        | VirtualMachines (SET_NULL), child clusters                                                                                                                                                                           |
| **ClusterType**    | Clusters (CASCADE), VirtualMachines (SET_NULL for cluster)                                                                                                                                                           |
| **ClusterGroup**   | Clusters (SET_NULL)                                                                                                                                                                                                  |
| **VirtualMachine** | VM Interfaces (CASCADE), Services (SET_NULL for VM)                                                                                                                                                                  |

### PROTECT (parent cannot be deleted if children exist)

| Parent           | Protected By                                       |
| ---------------- | -------------------------------------------------- |
| **Manufacturer** | DeviceType FK, Platform FK                         |
| **DeviceType**   | Device FK                                          |
| **DeviceRole**   | Device FK, VirtualMachine FK                       |
| **Platform**     | Device FK, VirtualMachine FK                       |
| **Tenant**       | Most models with `tenant` FK                       |
| **VLAN**         | Interface (assigned_vlan), Prefix (vlan) — PROTECT |
| **RIR**          | Aggregate (SET_NULL, so NOT protected)             |
| **Tag**          | All tagged items (M2M through, auto-removed)       |

### SET_NULL (parent deletion nullifies the FK)

| Parent            | Children Updated                                                     |
| ----------------- | -------------------------------------------------------------------- |
| **Region**        | Sites (`site.region = null`)                                         |
| **SiteGroup**     | Sites (`site.group = null`)                                          |
| **Rack**          | Devices (`device.rack = null`), PowerFeed (PROTECT)                  |
| **VRF**           | Prefixes (`prefix.vrf = null`), IPAddresses (`ipaddress.vrf = null`) |
| **VLANGroup**     | VLANs (`vlan.group = null`)                                          |
| **Cluster**       | VirtualMachines (`vm.cluster = null`)                                |
| **ClusterGroup**  | Clusters (`cluster.group = null`)                                    |
| **RIR**           | Aggregates (`aggregate.rir = null`)                                  |
| **Site** (cached) | Components (`_site = null`) — these are denormalized caches          |

---

## 8. Cross-Model Validation Patterns

### Status validation

All models with `status` fields validate against the model's allowed statuses:

```python
def clean(self):
    if self.status not in self.status_choices:
        raise ValidationError(...)
```

### Uniqueness constraints (beyond DB-level)

| Model         | Constraint                                                                |
| ------------- | ------------------------------------------------------------------------- |
| **Prefix**    | No duplicate (prefix, vrf, site/scope) — only if `is_pool=False`          |
| **IPAddress** | No duplicate (address, vrf, tenant) — only if `role != 'anycast'`         |
| **VLAN**      | Unique (group, vid) within group                                          |
| **VLANGroup** | Unique scope (site, location, etc.) if scope_id_vars set                  |
| **Rack**      | Unique (site, location, name), (site, location, facility_id)              |
| **Device**    | Unique (site, tenant, serial), (virtual_chassis, vc_position)             |
| **Interface** | Unique (device, name), (virtual_machine, name)                            |
| **Cable**     | Unique set of termination A+B (no duplicate cable between same endpoints) |

### Scope validation (CachedScopeMixin)

Prefix/VLAN/Cluster/Service validate that scope is one of the allowed types:

```python
def clean(self):
    if self.scope and self.scope.__class__ not in self.scope_types():
        raise ValidationError("Scope must be one of: ...")
```

---

## 9. Cable & Path Tracing System

### Cable Model

A Cable connects two termination points (each is a GenericFK to Interface, ConsolePort, etc.).

**CREATE cable:**

1. Validate termination A and B are different.
2. Validate both terminations are not already connected.
3. Save cable.
4. Signal `update_connected_endpoints()` sets `cable` FK on each termination and marks `path_endpoints`.

**DELETE cable:**

1. `pre_delete` signal fires.
2. For each termination: set `termination.cable = None` (disconnects).
3. Delete CablePath entries that traverse this cable.

### CablePath (auto-generated)

NetBox auto-generates `CablePath` records tracing connectivity:

| Field                                 | Purpose                                        |
| ------------------------------------- | ---------------------------------------------- |
| `origin_type` / `origin_id`           | GenericFK — where the path starts              |
| `destination_type` / `destination_id` | GenericFK — where it ends (null if incomplete) |
| `path`                                | Ordered array of node GenericFKs               |
| `is_active`                           | True if all cables in path are `connected`     |
| `is_split`                            | True if path forks                             |

**When a Cable is created/deleted/updated:** `update_cablepaths()` recomputes all affected paths via DFS/BFS.

### Go Rewrite: Implement path tracing as a dedicated service with a graph traversal algorithm. Cache results in a `cable_paths` table.

---

## 10. IPAM Hierarchy System

### Prefix hierarchy

Prefixes store denormalized hierarchy fields for fast lookups:

| Field       | Purpose                                           |
| ----------- | ------------------------------------------------- |
| `_depth`    | How many parent prefixes exist above this one     |
| `_children` | Count of child prefixes directly beneath this one |

**On Prefix CREATE/UPDATE/DELETE:** signal `cache_prefix_hierarchy()` recomputes `_depth` and `_children` for all prefixes in the same VRF+scope.

### IP Address to Interface assignment

When `IPAddress.assigned` is set (via `interface` or `vminterface` FK):

1. `clean()` validates the IP's `vrf` matches the interface's device/VM VRF (or both null).
2. `save()` sets `assigned_object` GFK.
3. Signal updates the interface's `ip_address_count` counter cache.

**On DELETE of IPAddress:** counter decremented. If it was a device's primary IP (`primary_ip4`/`primary_ip6`), that FK is SET_NULL.

### Go Rewrite: Implement prefix hierarchy as a service-layer function using PostgreSQL's `inet` type and recursive CTEs for depth/children calculation.

---

## 11. Signal Handler Reference

### Global signals (connected in `netbox/signals.py` and module `apps.py`)

| Signal                      | Handler                       | Effect                                            |
| --------------------------- | ----------------------------- | ------------------------------------------------- |
| `post_save` (NetBoxModel)   | `change_log_event`            | Creates `ObjectChange` record                     |
| `post_save` (NetBoxModel)   | `process_event_rules`         | Fires webhooks, scripts, notifications            |
| `post_delete` (NetBoxModel) | `change_log_event`            | Creates `ObjectChange` (action=delete)            |
| `pre_delete` (Cable)        | `nullify_connected_endpoints` | Disconnects endpoints                             |
| `post_save` (Cable)         | `update_connected_endpoints`  | Connects endpoints                                |
| `post_save` (Cable)         | `update_cablepaths`           | Recomputes CablePath graph                        |
| `post_delete` (Cable)       | `update_cablepaths`           | Recomputes CablePath graph                        |
| `post_save` (Prefix)        | `cache_prefix_hierarchy`      | Updates `_depth`, `_children`                     |
| `post_delete` (Prefix)      | `cache_prefix_hierarchy`      | Updates `_depth`, `_children`                     |
| `post_save` (IPAddress)     | `update_assigned_object`      | Sets GFK from `assigned_object_*`                 |
| `post_save` (Device)        | `update_component_caches`     | Refreshes `_site`/`_location`/`_rack` on children |
| `post_save` (Interface)     | `update_connected_endpoints`  | Re-cable path tracing                             |
| `post_save` (VMInterface)   | `update_connected_endpoints`  | Re-cable path tracing                             |
| `m2m_changed` (tags)        | `serialize_tags`              | Updates tag relationship cache                    |
| `m2m_changed` (CustomField) | `update_custom_field_data`    | Syncs custom field values                         |

---

## 12. Per-Entity CRUD Categories

Each entity falls into one or more behavioral categories. Its doc file references these:

### Category A: Simple Organizational (CRUD with no side effects)

**Models:** Manufacturer, Platform, DeviceRole, RackRole, Region, SiteGroup, Location, ClusterType, ClusterGroup, Tenant, TenantGroup, Contact, ContactGroup, ContactRole, ASNRange, RIR, Role, VLANGroup, Tag, Status, Webhook, CustomField

**CREATE:** `clean()` → save → change log → event
**UPDATE:** snapshot → `clean()` → save → change log → event
**DELETE:** change log → event → DB delete. Children with CASCADE are deleted; children with SET_NULL are nullified; PROTECT blocks.

### Category B: Denormalized Counter Source (triggers counter updates)

**Models:** DeviceType, Device, Site, Rack, VRF, VLANGroup, Tenant, Cluster, VirtualMachine, Manufacturer, VLAN

**On CREATE/DELETE of child templates/components/IPs:** Parent's counter fields increment/decrement via signal.

### Category C: Denormalized Cache Source (triggers cache refresh)

**Models:** Device (refreshes component caches), VirtualMachine (refreshes VM interface caches), Site/Location/Rack (refreshed from Device), Prefix/VLAN/Cluster (scope caches)

**On UPDATE of source:** All dependent denormalized fields refreshed.

### Category D: Cable-Connected (cable path system)

**Models:** Interface, ConsolePort, ConsoleServerPort, PowerPort, PowerOutlet, FrontPort, RearPort, Cable, CablePath, CableTermination, PowerFeed

**On CREATE/UPDATE/DELETE:** Cable path recomputation.

### Category E: IPAM Hierarchy (depth/children recalculation)

**Models:** Prefix, IPAddress, Aggregate, IPRange

**On CREATE/UPDATE/DELETE of Prefix:** Hierarchy fields recalculated.
**On IPAddress assign/unassign:** Interface counter updated.

### Category F: Scoped (GenericFK scope validation)

**Models:** Prefix, VLAN, Cluster, WirelessLAN, WirelessLink, Service

**On CREATE/UPDATE:** `clean()` validates scope type is allowed. Denormalized scope fields updated.

---

## Summary: The CRUD Flow for Every Entity

```
1. API Request arrives
2. Serializer deserializes + calls model.clean()
   ├── Field validation
   ├── Uniqueness validation
   ├── Cross-entity validation (e.g., prefix overlaps, scope types)
   └── GFK validation
3. pre_save signal (rare — mostly on_delete=CASCADE setup)
4. model.save() to DB
   ├── Denormalized cache fields written (_site, _location, _rack)
   ├── Hierarchy fields written (_depth, _children)
   └── GORM INSERT/UPDATE
5. post_save signal fires
   ├── Counter fields on parent incremented
   ├── Cable paths recomputed (if cable-connected)
   ├── Prefix hierarchy recalculated (if Prefix)
   ├── Component caches refreshed (if Device/VM)
   └── ChangeLoggedMixin: snapshot diff → ObjectChange
6. EventRulesMixin: webhook/script/notification queued
7. Response returned
8. Event queue flushed → webhooks dispatched async
```

### For DELETE:

```
1. snapshot() stores pre-delete state
2. pre_delete signal
   ├── Cable endpoints disconnected (if Cable)
   ├── PROTECT checks enforced by GORM FK constraints
   └── CASCADE children identified
3. DB DELETE
   ├── CASCADE children deleted (with their own pre/post_delete)
   ├── SET_NULL children updated
   └── PROTECT blocks if children exist
4. post_delete signal
   ├── Counter fields on parent decremented
   ├── Cable paths recomputed
   ├── Prefix hierarchy recalculated
   └── Primary IP FKs nulled (if IPAddress)
5. ChangeLoggedMixin: ObjectChange (action=delete)
6. EventRulesMixin: delete event queued
7. Response returned
8. Event queue flushed
```
