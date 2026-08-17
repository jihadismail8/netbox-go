# IPAM models — partial business-logic discovery

> Historical extraction from selected Python NetBox `ipam/models/` files, not an accepted specification. It contains known relationship and inferred-rule errors. Reconcile each rule to the pinned source, upstream tests, Capability Profile, and oracle before implementation.

---

## VRF (PrimaryModel)

| Field          | Type           | Constraints                               |
| -------------- | -------------- | ----------------------------------------- |
| name           | CharField(100) | db_collation="natural_sort"               |
| rd             | CharField(21)  | unique, blank, null — Route Distinguisher |
| enforce_unique | BooleanField   | default=True                              |
| tenant         | FK → Tenant    | PROTECT, null/blank                       |

**Relationships:**

- `import_targets` M2M → RouteTarget (related_name=`importing_vrfs`)
- `export_targets` M2M → RouteTarget (related_name=`exporting_vrfs`)

**Meta:** `ordering = ('name',)`

**clean() rules:** None beyond base validation.

---

## RouteTarget (PrimaryModel)

| Field  | Type          | Constraints         |
| ------ | ------------- | ------------------- |
| name   | CharField(21) | unique              |
| tenant | FK → Tenant   | PROTECT, null/blank |

**Meta:** `ordering = ('name',)`

---

## RIR (OrganizationalModel)

| Field      | Type           | Constraints   |
| ---------- | -------------- | ------------- |
| name       | CharField(100) |               |
| slug       | SlugField(100) |               |
| is_private | BooleanField   | default=False |

**Meta:** `ordering = ('name',)`

---

## Aggregate (PrimaryModel)

| Field      | Type           | Constraints                         |
| ---------- | -------------- | ----------------------------------- |
| prefix     | IPNetworkField | **required** — IPv4 or IPv6 network |
| rir        | FK → RIR       | **required**, PROTECT               |
| tenant     | FK → Tenant    | PROTECT, null/blank                 |
| date_added | DateField      | null/blank                          |

**Meta:** `ordering = ('prefix', 'pk')`

**clean() rules:**

1. **No /0 mask** — `prefix.prefixlen == 0` → ValidationError("Cannot create aggregate with /0 mask.")
2. **No overlapping aggregates (covered by existing)** — Checks `prefix__net_contains_or_equals` in existing aggregates. If found → ValidationError("Aggregates cannot overlap.")
3. **No covering existing aggregates** — Checks `prefix__net_contained` in existing aggregates. If found → ValidationError("Prefixes cannot overlap aggregates.")

**Properties:**

- `family` → prefix version (4 or 6)
- `ipv6_full` → expanded IPv6 notation

**Methods:**

- `get_child_prefixes()` → all Prefixes within this aggregate
- `get_utilization()` → percentage of utilized child prefix space

---

## Role (OrganizationalModel)

| Field  | Type                      | Constraints  |
| ------ | ------------------------- | ------------ |
| name   | CharField(100)            |              |
| slug   | SlugField(100)            |              |
| weight | PositiveSmallIntegerField | default=1000 |

**Meta:** `ordering = ('weight', 'name')`

---

## Prefix (PrimaryModel) — COMPLEX

| Field         | Type                      | Constraints                      |
| ------------- | ------------------------- | -------------------------------- |
| prefix        | IPNetworkField            | **required** — network with mask |
| vrf           | FK → VRF                  | PROTECT, null/blank              |
| tenant        | FK → Tenant               | PROTECT, null/blank              |
| vlan          | FK → VLAN                 | PROTECT, null/blank              |
| status        | CharField(50)             | choices, default='active'        |
| role          | FK → Role                 | SET_NULL, null/blank             |
| is_pool       | BooleanField              | default=False                    |
| mark_utilized | BooleanField              | default=False                    |
| _depth        | PositiveSmallIntegerField | auto-computed, not editable      |
| _children     | PositiveBigIntegerField   | auto-computed, not editable      |

**Meta:**

- `ordering = (F('vrf').asc(nulls_first=True), 'prefix', 'pk')`
- GistIndex on `prefix` with `inet_ops` opclass

**clean() rules:** 0. **Public API prefix validation** — `prefix_validator` rejects host-bit input and returns the canonical CIDR as a suggestion before model save.

1. **No /0 mask** — `prefix.prefixlen == 0` → ValidationError.
2. **Enforce unique IP space:**
   - If VRF is null and `ENFORCE_GLOBAL_UNIQUE` config is True, OR
   - If VRF is set and `vrf.enforce_unique` is True
   - Check for duplicates (same prefix + same VRF). If found → ValidationError("Duplicate prefix found in {table}").

**save() behavior:**

- **Canonical storage guard** — model save clears host bits (`self.prefix = self.prefix.cidr`), although the REST serializer validator rejects such input before save.
- Caches related objects for filtering

**Properties:**

- `family` → prefix version
- `mask_length` → prefix length
- `depth` → cached `_depth`
- `children` → cached `_children`

**Methods:**

- `get_parents(include_self)` → containing Prefixes via `net_contains`/`net_contains_or_equals`
- `get_children(include_self)` → covered Prefixes via `net_contained`/`net_contained_or_equal`
- `get_duplicates()` → same VRF + same prefix
- `get_child_prefixes()` → if container in global table, return all children; else same VRF children
- `get_child_ranges()` → IPRanges within this prefix
- `get_child_ips()` → IPAddresses within this prefix
- `get_available_ips()` → IPSet of available IPs (excludes network/broadcast for IPv4 unless /31-/32 or is_pool; excludes first address for IPv6 unless /127+)
- `get_first_available_ip()` → string of first available IP
- `get_utilization()` → percentage:
  - If `mark_utilized` → 100
  - If status is 'container' → based on child prefix coverage
  - Otherwise → based on child IP count / prefix size (minus 2 for network/broadcast unless is_pool or /31+)

---

## IPRange (PrimaryModel) — COMPLEX

| Field          | Type                 | Constraints                          |
| -------------- | -------------------- | ------------------------------------ |
| start_address  | IPAddressField       | **required** — IP with mask          |
| end_address    | IPAddressField       | **required** — IP with mask          |
| size           | PositiveIntegerField | auto-computed, not editable          |
| vrf            | FK → VRF             | PROTECT, null/blank                  |
| tenant         | FK → Tenant          | PROTECT, null/blank                  |
| status         | CharField(50)        | choices, default='active'            |
| role           | FK → Role            | SET_NULL, null/blank                 |
| mark_populated | BooleanField         | default=False — prevents IP creation |
| mark_utilized  | BooleanField         | default=False                        |

**Meta:** `ordering = (F('vrf').asc(nulls_first=True), 'start_address', 'pk')`

**clean() rules:**

1. **IP versions must match** — `start_address.version != end_address.version` → ValidationError.
2. **Prefix lengths must match** — `start_address.prefixlen != end_address.prefixlen` → ValidationError.
3. **End must be greater than start** — `end_address <= start_address` → ValidationError.
4. **No overlapping ranges** — Checks for ranges in same VRF where start/end fall within existing ranges. If overlap → ValidationError.
5. **Max size check** — Range cannot exceed `2^32 - 1` addresses.

**save() behavior:**

- Computes `size = end_address.ip - start_address.ip + 1`

**Properties:**

- `family` → IP version
- `range` → `netaddr.IPRange(start, end)`
- `mask_length` → prefix length from start_address
- `name` → compact string representation (e.g., `192.0.2.1-100/24`)

**Methods:**

- `get_child_ips()` → IPAddresses within range
- `get_available_ips()` → IPSet of available IPs (empty if mark_populated)
- `utilization` (cached) → percentage of used IPs

---

## IPAddress (PrimaryModel) — MOST COMPLEX

| Field                | Type                    | Constraints                                       |
| -------------------- | ----------------------- | ------------------------------------------------- |
| address              | IPAddressField          | **required** — IP with mask                       |
| vrf                  | FK → VRF                | PROTECT, null/blank                               |
| tenant               | FK → Tenant             | PROTECT, null/blank                               |
| status               | CharField(50)           | choices, default='active'                         |
| role                 | CharField(50)           | choices, blank, null                              |
| assigned_object_type | FK → ContentType        | PROTECT, null/blank                               |
| assigned_object_id   | PositiveBigIntegerField | null/blank                                        |
| assigned_object      | GenericForeignKey       | (computed from type+id)                           |
| nat_inside           | FK → self               | SET_NULL, null/blank (related_name=`nat_outside`) |
| dns_name             | CharField(255)          | blank, DNSValidator                               |

**Meta:**

- `ordering = ('address', 'pk')`
- Index on host address (Cast/Host function)
- Index on (assigned_object_type, assigned_object_id)

**clean() rules:**

1. **No /0 mask** — `address.prefixlen == 0` → ValidationError.
2. **Network/broadcast assignment check** — If `assigned_object` is set:
   - IPv4: network ID cannot be assigned unless /31 or /32 → ValidationError.
   - IPv4: broadcast address cannot be assigned unless /31 or /32 → ValidationError.
   - IPv6: network ID cannot be assigned unless /127 or /128 → ValidationError.
3. **Enforce unique IP space:**
   - If VRF is null and `ENFORCE_GLOBAL_UNIQUE` config is True, OR
   - If VRF is set and `vrf.enforce_unique` is True
   - Check for duplicates (same host IP in same VRF). Exception: IPs with roles in `IPADDRESS_ROLES_NONUNIQUE` (anycast, etc.) are allowed to be duplicates.
   - If found → ValidationError("Duplicate IP address found in {table}").
4. **No IP creation inside mark_populated ranges** — If address falls within an IPRange with `mark_populated=True` → ValidationError("Cannot create IP address {ip} inside range {range}.")
5. **Primary IP reassignment check** — If this IP is a primary IP for a device/VM, it cannot be reassigned to a different parent object → ValidationError("Cannot reassign IP address while it is designated as the primary IP").
6. **SLAAC status is IPv6 only** — `status == 'slaac'` and `family != 6` → ValidationError.

**save() behavior:**

- Forces `dns_name` to lowercase

**Properties:**

- `family` → IP version
- `ipv6_full` → expanded IPv6 notation
- `is_primary_ip` → True if this is primary_ip4 or primary_ip6 of assigned parent
- `is_oob_ip` → True if this is oob_ip of assigned parent

**Methods:**

- `get_duplicates()` → same VRF + same host address
- `get_next_available_ip()` → next available IP in the same subnet
- `get_related_ips()` → other IPs in same VRF within same CIDR
- `clone()` → populates with next available IP

---

## VLANGroup (PrimaryModel)

| Field      | Type                      | Constraints                                        |
| ---------- | ------------------------- | -------------------------------------------------- |
| name       | CharField(100)            |                                                    |
| slug       | SlugField(100)            |                                                    |
| scopes     | —                         | Generic scope (region, site group, site, location) |
| scope_type | FK → ContentType          | null/blank                                         |
| scope_id   | PositiveBigIntegerField   | null/blank                                         |
| min_vid    | PositiveSmallIntegerField | null/blank, validators 1-4094                      |
| max_vid    | PositiveSmallIntegerField | null/blank, validators 1-4094                      |

**Meta:** `ordering = ('name',)`

**clean() rules:**

- **max_vid must be >= min_vid** — If both set and `max_vid < min_vid` → ValidationError.

**Methods:**

- `get_next_available_vid()` → next available VLAN ID in range (checks existing VLANs)

---

## VLAN (PrimaryModel)

| Field       | Type                      | Constraints                     |
| ----------- | ------------------------- | ------------------------------- |
| site        | FK → Site                 | SET_NULL, null/blank            |
| group       | FK → VLANGroup            | SET_NULL, null/blank            |
| vid         | PositiveSmallIntegerField | **required**, validators 1-4094 |
| name        | CharField(64)             | **required**                    |
| tenant      | FK → Tenant               | SET_NULL, null/blank            |
| status      | CharField(50)             | choices, default='active'       |
| role        | FK → Role                 | SET_NULL, null/blank            |
| description | CharField(200)            | blank                           |

**Meta:** `ordering = ('site', 'group', 'vid')`

**Constraints:**

- `unique_together = [['group', 'vid'], ['group', 'name']]`

**clean() rules:**

1. **VID within group range** — If `group` is set and group has `min_vid`/`max_vid`, VID must be within range → ValidationError if out of bounds.
2. **Unique VID per group** — Enforced by unique_together, but also checked in clean() for better error messages.
3. **Unique name per group** — Enforced by unique_together.

---

## VLANTranslationPolicy (PrimaryModel)

| Field       | Type           | Constraints |
| ----------- | -------------- | ----------- |
| name        | CharField(100) |             |
| description | CharField(200) | blank       |

**Meta:** `ordering = ('name',)`

---

## VLANTranslationRule (NetBoxModel)

| Field      | Type                       | Constraints                   |
| ---------- | -------------------------- | ----------------------------- |
| policy     | FK → VLANTranslationPolicy | CASCADE, related_name=`rules` |
| local_vid  | PositiveSmallIntegerField  | validators 1-4094             |
| remote_vid | PositiveSmallIntegerField  | validators 1-4094             |

**Constraints:**

- `unique_together = [['policy', 'local_vid']]`

---

## FHRPGroup (PrimaryModel)

| Field       | Type                 | Constraints                                    |
| ----------- | -------------------- | ---------------------------------------------- |
| group_id    | PositiveIntegerField | **required**                                   |
| protocol    | CharField(50)        | choices (hsrp, vrrp2, vrrp3, glbp, carp, etc.) |
| auth_type   | CharField(50)        | blank, choices                                 |
| auth_key    | CharField(255)       | blank                                          |
| description | CharField(200)       | blank                                          |

**Meta:** `ordering = ('protocol', 'group_id', 'pk')`

**Constraints:**

- `unique_together = [['protocol', 'group_id']]`

---

## FHRPGroupAssignment (NetBoxModel)

| Field                         | Type                      | Constraints                         |
| ----------------------------- | ------------------------- | ----------------------------------- |
| group                         | FK → FHRPGroup            | CASCADE, related_name=`assignments` |
| interface_type                | FK → ContentType          | CASCADE                             |
| interface_id                  | PositiveBigIntegerField   |                                     |
| interface = GenericForeignKey |                           |                                     |
| priority                      | PositiveSmallIntegerField | validators 0-255                    |
| is_primary                    | BooleanField              | default=False                       |

**Constraints:**

- `unique_together = [['group', 'interface_type', 'interface_id']]`

---

## ServiceTemplate (PrimaryModel)

| Field       | Type                     | Constraints                         |
| ----------- | ------------------------ | ----------------------------------- |
| name        | CharField(100)           |                                     |
| protocol    | CharField(50)            | choices (tcp, udp, sctp)            |
| ports       | ArrayField(IntegerField) | **required** — port numbers 1-65535 |
| description | CharField(200)           | blank                               |

**Meta:** `ordering = ('name',)`

**clean() rules:**

- Each port in `ports` must be between 1 and 65535.

---

## Service (PrimaryModel)

| Field                               | Type                     | Constraints  |
| ----------------------------------- | ------------------------ | ------------ |
| name                                | CharField(100)           |              |
| protocol                            | CharField(50)            | choices      |
| ports                               | ArrayField(IntegerField) | **required** |
| ipaddresses                         | M2M → IPAddress          | blank        |
| description                         | CharField(200)           | blank        |
| assigned_object_type                | FK → ContentType         | null/blank   |
| assigned_object_id                  | PositiveBigIntegerField  | null/blank   |
| assigned_object = GenericForeignKey |                          |              |

**Meta:** `ordering = ('protocol', 'ports', 'pk')`

**clean() rules:**

- Each port must be 1-65535.

---

## ASN (PrimaryModel)

| Field  | Type            | Constraints                          |
| ------ | --------------- | ------------------------------------ |
| asn    | BigIntegerField | **required**, unique, validators >=0 |
| rir    | FK → RIR        | **required**, PROTECT                |
| tenant | FK → Tenant     | SET_NULL, null/blank                 |

**Meta:** `ordering = ('asn',)`

---

## ASNRange (PrimaryModel)

| Field  | Type            | Constraints                  |
| ------ | --------------- | ---------------------------- |
| name   | CharField(100)  |                              |
| slug   | SlugField(100)  |                              |
| rir    | FK → RIR        | **required**, PROTECT        |
| start  | BigIntegerField | **required**, validators >=0 |
| end    | BigIntegerField | **required**, validators >=0 |
| tenant | FK → Tenant     | SET_NULL, null/blank         |

**Meta:** `ordering = ('start', 'pk')`

**clean() rules:**

- **End must be greater than start** — `end <= start` → ValidationError.

---

## Candidate IPAM rules requiring source reconciliation

### IP Address/Prefix Operations

1. **Prefix boundary and storage** — public Prefix input with host bits is rejected with a canonical suggestion; accepted Prefixes are stored in CIDR form
2. **IP family matching** — Ranges require start/end to be same family
3. **Prefix length matching** — Ranges require start/end to have same prefix length
4. **Overlap detection** — Uses PostgreSQL GIST inet operators (`net_contains`, `net_contained`, `net_host`)
5. **Available IP computation** — Complex logic excluding network/broadcast (IPv4), subnet-router anycast (IPv6)
6. **Utilization calculation** — Different for containers (child prefix space) vs regular (child IP count)

### Validation Patterns

- **Network/broadcast assignment** — Cannot assign network ID or broadcast to interfaces (IPv4 /31,/32 and IPv6 /127,/128 exempt)
- **Unique IP enforcement** — Configurable via VRF.enforce_unique or global ENFORCE_GLOBAL_UNIQUE
- **NAT roles non-unique** — SLAAC, anycast roles exempt from uniqueness check
- **Primary IP immutability** — Can't reassign IP that's a primary IP
- **SLAAC is IPv6 only**

### PostgreSQL-Specific Features

- GIST index on prefix/inet columns with inet_ops opclass
- Host() function for extracting host address
- ArrayField for ports and units
- GenericForeignKey for assigned_object patterns
