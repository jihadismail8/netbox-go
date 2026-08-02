# DCIM Models — Business Logic Reference

> Extracted from Python NetBox `dcim/models/` for porting to Go.
> **Source:** `netbox/netbox/dcim/models/{sites,racks,devices,device_components,cables,power,modules}.py`

---

## Model Inheritance Hierarchy

```
NetBoxFeatureSet (abstract)
  ├─ Provides: BookmarksMixin, ChangeLoggingMixin, CloningMixin, CustomFieldsMixin,
  │             CustomLinksMixin, CustomValidationMixin, ExportTemplatesMixin,
  │             JournalingMixin, NotificationsMixin, TagsMixin, EventRulesMixin
  ├─ get_absolute_url(), docs_url property
  │
  ├─ NetBoxModel(NetBoxFeatureSet)
  │    ├─ RestrictedQuerySet manager
  │    └─ clean() validates GenericForeignKey fields (ct_field + fk_field both set or both null)
  │
  ├─ PrimaryModel(NetBoxModel)
  │    ├─ description: varchar(200), blank
  │    └─ comments: text, blank
  │
  ├─ NestedGroupModel(NetBoxFeatureSet, MPTTModel)
  │    ├─ parent: TreeForeignKey(self, CASCADE, null/blank)
  │    ├─ name: varchar(100)
  │    ├─ slug: slug(100)
  │    ├─ description, comments
  │    ├─ clean(): forbids assigning self/descendant as parent
  │    └─ Meta: ordering=['name'], MPTT
  │
  └─ OrganizationalModel(PrimaryModel)
       ├─ name, slug, description, comments
       └─ Meta: ordering=['name']
```

---

## Site

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | unique=True |
| slug | SlugField(100) | unique=True |
| status | CharField(20) | default='active', choices=SiteStatusChoices |
| region | FK → Region | null/blank, SET_NULL |
| group | FK → SiteGroup | null/blank, SET_NULL |
| tenant | FK → Tenant | null/blank, SET_NULL |
| facility | CharField(50) | blank |
| asn | BigIntegerField | null/blank, validators ≥0 |
| time_zone | CharField(100) | blank (TZ identifiers) |
| description | CharField(200) | blank |
| physical_address | CharField(200) | blank |
| shipping_address | CharField(200) | blank |
| latitude | DecimalField | null/blank, max_digits=8, decimal_places=6, min=-90, max=90 |
| longitude | DecimalField | null/blank, max_digits=9, decimal_places=6, min=-180, max=180 |
| comments | TextField | blank |

**Meta:** `ordering = ('name',)`, unique constraints on name and slug.

**clean() rules:**
- None beyond field validation (no cross-field validation).

**Properties:**
- `docs_url` → docs path

---

## Region (NestedGroupModel)

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | unique=True |
| slug | SlugField(100) | unique=True |
| parent | TreeForeignKey → self | null/blank, SET_NULL |
| description | CharField(200) | blank |

**Meta:** `ordering = ('name',)`, MPTT tree structure.

**clean() rules:**
- **Cannot assign self as parent** — ValidationError if `parent == self`.
- **Cannot assign descendant as parent** — Prevents creating tree cycles.

---

## SiteGroup (NestedGroupModel)

Same structure and rules as Region. Tree hierarchy of site groups.

---

## Location (NestedGroupModel)

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | |
| slug | SlugField(100) | |
| site | FK → Site | **required**, CASCADE |
| parent | TreeForeignKey → self | null/blank, SET_NULL |
| status | CharField(20) | default='active', choices=LocationStatusChoices |
| tenant | FK → Tenant | null/blank, SET_NULL |
| facility | CharField(50) | blank |
| description | CharField(200) | blank |
| comments | TextField | blank |

**Meta:** `ordering = ('site__name', 'name')`

**Constraints:**
- `unique_together = [['site', 'name'], ['site', 'slug']]`

**clean() rules:**
- **Parent must belong to same site** — ValidationError if `parent.site != self.site`.
- **NestedGroupModel.clean()** — Cannot assign self/descendant as parent.

---

## Rack (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | |
| facility_id | CharField(50) | blank, null |
| site | FK → Site | **required**, CASCADE |
| location | FK → Location | null/blank, SET_NULL |
| tenant | FK → Tenant | null/blank, SET_NULL |
| status | CharField(50) | default='active', choices=RackStatusChoices |
| role | FK → RackRole | null/blank, SET_NULL |
| serial | CharField(50) | blank |
| asset_tag | CharField(50) | blank, null, unique |
| type | CharField(100) | blank, choices=RackTypeChoices |
| width | PositiveSmallIntegerField | choices=[19,21,23], default=19 |
| u_height | PositiveSmallIntegerField | default=42, validators min=1 max=100 |
| desc_units | BooleanField | default=False |
| outer_width | PositiveSmallIntegerField | null/blank |
| outer_depth | PositiveSmallIntegerField | null/blank |
| outer_unit | CharField(30) | blank |
| weight | DecimalField | null/blank, max_digits=8, decimal_places=3 |
| max_weight | PositiveIntegerField | null/blank |
| weight_unit | CharField(50) | choices, blank |
| mounting_depth | PositiveIntegerField | null/blank |
| description, comments | | |

**Meta:** `ordering = ('site', 'location', 'name')`

**Constraints:**
- `unique_together = [['site', 'location', 'name'], ['site', 'location', 'facility_id']]`

**clean() rules:**
1. **Location must belong to same site** — ValidationError if `location.site != self.site`.
2. **Outer dimensions require unit** — If any of `outer_width`/`outer_depth` set, `outer_unit` must be set.
3. **Weight requires unit** — If `weight` or `max_weight` set, `weight_unit` must be set.
4. **u_height validation** — Must be ≥ 1.

**Properties:**
- `units` → list of rack unit positions (U-number, device if occupied)
- `elevation` → reverse-ordered unit list for display
- `total_weight` → sum of device weights

---

## RackReservation (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| rack | FK → Rack | **required**, CASCADE |
| units | ArrayField(IntegerField) | **required** |
| tenant | FK → Tenant | null/blank, SET_NULL |
| user | FK → User | **required**, CASCADE |
| description | CharField(200) | **required** |

**Meta:** `ordering = ('rack', 'units')`

**clean() rules:**
- **Units must be within rack's U-height** — Each unit must be between 1 and `rack.u_height`. ValidationError if any unit is out of range.

---

## RackRole, DeviceRole, Manufacturer, Platform (OrganizationalModel)

| Model | Key Fields | Constraints |
|-------|-----------|-------------|
| RackRole | name, slug, color (default '9e9e9e'), description | unique name+slug |
| DeviceRole | name, slug, color, vm_role (bool, default False), description | unique name+slug |
| Manufacturer | name, slug, description | unique name+slug |
| Platform | name, slug, manufacturer (FK, null), napalm_driver, napalm_args (JSON) | unique name+slug |

**clean() for Platform:**
- `napalm_args` must be valid JSON dict if provided.

---

## DeviceType (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| manufacturer | FK → Manufacturer | **required**, CASCADE |
| model | CharField(100) | |
| slug | SlugField(100) | |
| part_number | CharField(50) | blank |
| u_height | PositiveSmallIntegerField | default=1, min=0 |
| is_full_depth | BooleanField | default=True |
| subdevice_role | CharField(50) | blank, choices=[parent, child] |
| airflow | CharField(50) | blank, choices=DeviceAirflowChoices |
| front_image | ImageField | null/blank |
| rear_image | ImageField | null/blank |
| weight | DecimalField | null/blank |
| weight_unit | CharField(50) | blank |
| description, comments | | |

**Meta:** `ordering = ('manufacturer__name', 'model')`

**Constraints:**
- `unique_together = [['manufacturer', 'model'], ['manufacturer', 'slug']]`

**clean() rules:**
1. **Subdevice parent/child with U-height** — If `subdevice_role` is set (parent or child), `u_height` must be 0.
2. **Weight requires unit** — If `weight` set, `weight_unit` must be set.

---

## Device (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | blank, null |
| device_type | FK → DeviceType | **required**, PROTECT |
| role | FK → DeviceRole | **required**, PROTECT |
| tenant | FK → Tenant | null/blank, SET_NULL |
| platform | FK → Platform | null/blank, SET_NULL |
| location | FK → Location | null/blank, SET_NULL |
| site | FK → Site | **required**, CASCADE |
| rack | FK → Rack | null/blank, SET_NULL |
| position | PositiveSmallIntegerField | null/blank, validators min=0 |
| face | CharField(50) | blank, choices=[front, rear] |
| status | CharField(50) | default='active', choices=DeviceStatusChoices |
| airflow | CharField(50) | blank |
| serial | CharField(50) | blank |
| asset_tag | CharField(50) | blank, null, unique |
| cluster | FK → Cluster | null/blank, SET_NULL |
| virtual_chassis | FK → VirtualChassis | null/blank, CASCADE |
| vc_position | PositiveSmallIntegerField | null/blank, validators 0-255 |
| vc_priority | PositiveSmallIntegerField | null/blank, validators 0-255 |
| primary_ip4 | FK → IPAddress | null/blank, SET_NULL |
| primary_ip6 | FK → IPAddress | null/blank, SET_NULL |
| oob_ip | FK → IPAddress | null/blank, SET_NULL |
| description, comments | | |

**Meta:** `ordering = ('name',)`, `unique_together = [['site', 'tenant', 'name']]`

**Constraints:**
- UniqueConstraint: `name` unique per `site` + `tenant` (where name is not null)
- UniqueConstraint: `asset_tag` globally unique (where not null)

**clean() rules:**
1. **Location must belong to same site** — `location.site != self.site` → ValidationError.
2. **Rack must belong to same site** — `rack.site != self.site` → ValidationError.
3. **Position requires rack** — `position` set but `rack` is null → ValidationError.
4. **Face required for non-zero position** — `position > 0` and `face` blank → ValidationError (only if device_type.u_height > 0).
5. **Position must be valid for rack** — `position + u_height - 1 > rack.u_height` → ValidationError.
6. **Child devices can't have position** — `device_type.subdevice_role == 'child'` → position/face/rack must be null.
7. **Primary IP family check** — `primary_ip4` must be IPv4, `primary_ip6` must be IPv6, `oob_ip` can be either.
8. **Primary IP must be assigned to device** — `primary_ip4/6/oob_ip` must be assigned to this device's interface (or the device itself).

**Properties:**
- `primary_ip` → primary_ip4 or primary_ip6 (IPv4 preferred)
- `status` badge classes for UI

---

## VirtualChassis (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| name | CharField(100) | |
| domain | CharField(30) | blank |
| master | FK → Device | null/blank, SET_NULL |
| description, comments | | |

**Meta:** `ordering = ('name',)`

**clean() rules:**
- **Master must be a member** — `master` must belong to `self.virtual_chassis` (its `virtual_chassis` FK must point to this VC).

---

## Module (NetBoxModel)

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | **required**, CASCADE |
| module_bay | FK → ModuleBay | **required**, CASCADE |
| module_type | FK → ModuleType | **required**, CASCADE |
| status | CharField(20) | default='active', choices=ModuleStatusChoices |
| serial | CharField(50) | blank |
| asset_tag | CharField(50) | blank, null, unique |
| description, comments | | |

**Meta:** `ordering = ('device', 'module_bay')`, `unique_together = [['device', 'module_bay']]`

---

## ModuleBay (NetBoxModel)

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | **required**, CASCADE |
| name | CharField(100) | |
| position | CharField(30) | blank |
| label | CharField(50) | blank |
| description | CharField(200) | blank |

**Meta:** `ordering = ('device', 'position', 'name')`, `unique_together = [['device', 'name']]`

---

## ModuleType (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| manufacturer | FK → Manufacturer | **required**, CASCADE |
| model | CharField(100) | |
| slug | SlugField(100) | |
| part_number | CharField(50) | blank |
| description, comments | | |

**Meta:** `ordering = ('manufacturer__name', 'model')`, `unique_together = [['manufacturer', 'model']]`

---

## Component Models (Abstract: ComponentModel)

**Base for:** ConsolePort, ConsoleServerPort, PowerPort, PowerOutlet, Interface, FrontPort, RearPort, DeviceBay, InventoryItem, MACAddress, VirtualDeviceContext

**Fields:**
- `device` FK → Device (CASCADE) — **REQUIRED**
- `name` CharField(64, db_collation="natural_sort")
- `label` CharField(64, blank)
- `description` CharField(200, blank)
- `_site`, `_location`, `_rack` — denormalized FKs (SET_NULL) auto-set on save

**Constraints:** `UniqueConstraint(fields=('device', 'name'))`

**clean() — CRITICAL RULE:**
- **Components cannot be moved to a different device** (except `InventoryItem`). If `pk is not None` and `_original_device != device_id`, raise ValidationError on `device`.

---

## ConsolePort / ConsoleServerPort / PowerPort

| Model | Key Fields | Special |
|-------|-----------|---------|
| ConsolePort | type (choices), speed (choices) | |
| ConsoleServerPort | type (choices) | |
| PowerPort | type (choices), maximum_draw (W, int), allocated_draw (W, int) | |

**clean() for PowerPort:**
- `allocated_draw` cannot exceed `maximum_draw` (if both set).

---

## PowerOutlet

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | required |
| name, label | CharField | |
| type | CharField | choices |
| power_port | FK → PowerPort | null/blank, SET_NULL |
| feed_leg | CharField | choices=[A,B,C] |

**clean() rules:**
- **Power port must be on same device** — `power_port.device != self.device` → ValidationError.

---

## Interface

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | required |
| name, label | CharField | |
| type | CharField(50) | choices=InterfaceTypeChoices |
| enabled | BooleanField | default=True |
| parent | FK → self (Interface) | null/blank |
| bridge | FK → self | null/blank |
| lag | FK → self | null/blank |
| mtu | PositiveIntegerField | null/blank, validators 60-65536 |
| mac_address | MACAddress | null/blank |
| mgmt_only | BooleanField | default=False |
| description | CharField | |
| mode | CharField(50) | blank, choices=[access, tagged, tagged-all] |
| rf_role, rf_channel, rf_channel_frequency, rf_channel_width | | wireless fields |
| tx_power, wireless_role | | wireless fields |
| poe_mode, poe_type | CharField | blank, choices |
| vrf | FK → VRF | null/blank |
| untagged_vlan | FK → VLAN | null/blank |
| tagged_vlans | M2M → VLAN | blank |
| vdcs | M2M → VirtualDeviceContext | blank |

**clean() rules:**
1. **Parent interface must be on same device** — `parent.device != self.device` → ValidationError.
2. **Bridge interface must be on same device.**
3. **LAG interface must be on same device.**
4. **LAG must be type 'lag'** — If a `lag` is assigned, the LAG interface must have `type == 'lag'`.
5. **Sub-interface must be on same device as parent.**
6. **Untagged VLAN must be present in tagged_vlans** — Cannot have same VLAN in both.
7. **Mode 'tagged-all' must have no tagged_vlans** — It implies all VLANs.
8. **Alphanumeric wireless channel** — `rf_channel` must match `rf_role` constraints.
9. **MGMT only validation** — Management-only interfaces can't be assigned to non-management roles.

---

## FrontPort / RearPort

| Model | Key Fields | Special |
|-------|-----------|---------|
| RearPort | type (choices), positions (int, 1-64) | |
| FrontPort | type, rear_port (FK → RearPort, required), rear_port_position (int, 1-64) | |

**clean() for FrontPort:**
- **Rear port must be on same device** — `rear_port.device != self.device` → ValidationError.
- **Rear port position must be valid** — `rear_port_position` must be ≤ `rear_port.positions`.
- **Unique rear_port + position** — Each rear port position can only have one front port.

---

## DeviceBay

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | required (must be type=parent) |
| name | CharField | |
| installed_device | OneToOne → Device | null/blank |

**clean() rules:**
1. **Parent device must be subdevice_role=parent** — `device.subdevice_role != 'parent'` → ValidationError.
2. **Installed device must be subdevice_role=child** — `installed_device.subdevice_role != 'child'` → ValidationError.
3. **Installed device must not be installed elsewhere** — If `installed_device.installed_device_bay` exists elsewhere → ValidationError.

---

## InventoryItem (ComponentModel exception)

Unlike other components, InventoryItem **CAN be moved between devices.**

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | required |
| parent | TreeForeignKey → self | null/blank |
| name, label | CharField | |
| manufacturer | FK → Manufacturer | null/blank |
| part_id | CharField(50) | blank |
| serial | CharField(50) | blank |
| asset_tag | CharField(50) | blank, null, unique |
| discovered | BooleanField | default=False |
| description | CharField | |
| role | FK → InventoryItemRole | null/blank |

**Meta:** `ordering = ('device__id', 'parent__id', 'name')`, unique_together=[['device','parent','name']]

**clean():** No "can't move" restriction (deliberately overridden).

---

## Cable

| Field | Type | Constraints |
|-------|------|-------------|
| a_terminations | GenericRelation → CableTermination | |
| b_terminations | GenericRelation → CableTermination | |
| status | CharField(50) | default='connected', choices=LinkStatusChoices |
| type | CharField(50) | blank, choices=CableTypeChoices |
| tenant | FK → Tenant | null/blank |
| label | CharField(100) | blank |
| color | CharField(6) | blank |
| length | DecimalField | null/blank |
| length_unit | CharField(50) | blank, choices |
| description, comments | | |

**Meta:** `ordering = ('pk',)`

**clean() rules:**
1. **Cannot connect a termination to itself** — A and B termination sets must not overlap.
2. **Cannot connect different types** — All A terminations must be same type; all B terminations must be same type.
3. **A and B can be different types** (e.g., interface to front port).
4. **Status validation** — Status must be one of: connected, planned, decommissioning.
5. **Length requires unit** — If `length` is set, `length_unit` must be set.

**save() behavior:**
- On save, updates all termination statuses to match cable status.
- Triggers CablePath rebuild for affected paths.

**delete() behavior:**
- Removes cable, resets termination statuses to "not connected."
- Rebuilds affected CablePaths.

---

## CableTermination

| Field | Type | Constraints |
|-------|------|-------------|
| cable | FK → Cable | CASCADE |
| termination_type | FK → ContentType | |
| termination_id | PositiveIntegerField | |
| termination = GenericForeignKey | | |
| cable_end | CharField(1) | choices=[A, B] |

**Constraints:**
- `unique_together = [['termination_type', 'termination_id', 'cable_end']]`

**clean() rules:**
- **One active cable per termination per end** — A termination can't be connected to multiple cables on the same end.

---

## CablePath

| Field | Type | Constraints |
|-------|------|-------------|
| origin_type | FK → ContentType | |
| origin_id | PositiveIntegerField | |
| origin = GenericForeignKey | | |
| destination_type | FK → ContentType | null/blank |
| destination_id | PositiveIntegerField | null/blank |
| destination = GenericForeignKey | null/blank |
| path = ArrayField(GenericForeignKey) | | serialized path |
| is_active | BooleanField | default=True |
| is_complete | BooleanField | default=True |
| is_split | BooleanField | default=False |

**Business rules:**
- CablePath objects are **auto-generated** when cables are created/updated/deleted.
- `from_origin()` classmethod traces a path from a given origin.
- `is_active` becomes False if any cable in the path is not 'connected'.

---

## PowerPanel

| Field | Type | Constraints |
|-------|------|-------------|
| site | FK → Site | **required**, CASCADE |
| location | FK → Location | null/blank, SET_NULL |
| name | CharField(100) | |
| description, comments | | |

**clean() rules:**
- **Location must belong to same site** — Same as Rack/Device.

---

## PowerFeed (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| power_panel | FK → PowerPanel | **required**, CASCADE |
| rack | FK → Rack | null/blank, SET_NULL |
| name | CharField(100) | |
| status | CharField(50) | default='active', choices=PowerFeedStatusChoices |
| type | CharField(50) | default='primary', choices=[primary, redundant] |
| supply | CharField(50) | default='ac', choices=[ac, dc] |
| phase | CharField(50) | default='single-phase', choices |
| voltage | SmallIntegerField | default=120 |
| amperage | PositiveSmallIntegerField | default=15 |
| max_utilization | PositiveSmallIntegerField | default=80, validators 1-100 |
| available_power | DecimalField | read-only computed |
| description, comments | | |

**clean() rules:**
- **Rack must belong to same site as power_panel** — `rack.site != power_panel.site` → ValidationError.

---

## MACAddress (NetBoxModel)

| Field | Type | Constraints |
|-------|------|-------------|
| mac_address | MACAddressField | unique validation |
| interface | FK → Interface | SET_NULL (via assignment) |
| description | CharField | |

---

## VirtualDeviceContext (PrimaryModel)

| Field | Type | Constraints |
|-------|------|-------------|
| device | FK → Device | **required**, CASCADE |
| name | CharField(100) | |
| identifier | PositiveIntegerField | null/blank |
| status | CharField(20) | default='active', choices |
| primary_ip4 | FK → IPAddress | null/blank |
| primary_ip6 | FK → IPAddress | null/blank |
| description, comments | | |

**Constraints:**
- `unique_together = [['device', 'name'], ['device', 'identifier']]`