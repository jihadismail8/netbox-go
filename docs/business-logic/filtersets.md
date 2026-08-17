# NetBox FilterSets — partial upstream discovery reference

> Extracted from selected Python NetBox `filtersets.py` files. This is neither a complete baseline filter catalogue nor Go implementation guidance. The accepted Capability Profile defines the only public allowlist; recheck every entry against the pinned source and oracle.

---

## Dynamic Lookup Expression System

NetBox's `BaseFilterSet` auto-generates lookup variants for every declared filter at class creation time. For supported types whose `lookup_expr` is in `STANDARD_LOOKUPS = ('exact', 'iexact', 'in', 'contains')` and that do NOT use a custom `method=`, new filters are synthesized.

### Char-Based Lookup Map (CharFilter, MultiValueCharFilter, etc.)

| Suffix   | Generated Lookup        | Negation? | Example                                   |
| -------- | ----------------------- | --------- | ----------------------------------------- |
| `n`      | `exact`                 | yes       | `name__n=john` -> name != john            |
| `ic`     | `icontains`             | no        | `name__ic=john` -> name ILIKE %john%      |
| `nic`    | `icontains`             | yes       | `name__nic=john` -> NOT ILIKE %john%      |
| `iew`    | `iendswith`             | no        | `name__iew=son`                           |
| `niew`   | `iendswith`             | yes       |                                           |
| `isw`    | `istartswith`           | no        | `name__isw=john`                          |
| `nisw`   | `istartswith`           | yes       |                                           |
| `ie`     | `iexact`                | no        | `name__ie=John` -> case-insensitive exact |
| `nie`    | `iexact`                | yes       |                                           |
| `empty`  | `empty` (BooleanFilter) | no        | `name__empty=true` -> name IS NULL OR ''  |
| `regex`  | `regex`                 | no        |                                           |
| `iregex` | `iregex`                | no        |                                           |

### Numeric/Date Lookup Map (NumberFilter, DateTimeFilter, etc.)

| Suffix | Generated Lookup  | Example                   |
| ------ | ----------------- | ------------------------- |
| `n`    | `exact` (negated) | `asn__n=65001`            |
| `lte`  | `lte`             | `created__lte=2024-01-01` |
| `lt`   | `lt`              |                           |
| `gte`  | `gte`             |                           |
| `gt`   | `gt`              |                           |

### Accepted translation rule

Do not expose a generic suffix-to-GORM parser. For each promoted resource:

1. Declare every accepted parameter name and value grammar in profile metadata.
2. Reject every undeclared parameter before repository access.
3. Parse declared input into typed list-query fields that preserve presence, repeated-value, empty, negation, and ordering semantics.
4. Implement explicit parameterized repository predicates for those fields.
5. Prove each declared lookup against the pinned REST oracle, including invalid values and interaction with visibility, count, ordering, and pagination.

A shared lookup-syntax helper may reduce parsing duplication only after the resource allowlist is selected. It must never translate an arbitrary public key into a database field or SQL fragment.

---

## Common Inherited Filters (ChangeLoggedModelFilterSet)

| Field                | Type                              | Description           |
| -------------------- | --------------------------------- | --------------------- |
| `q`                  | CharFilter (method=filter_search) | Full-text search      |
| `created`            | MultiValueDateTimeFilter          | Creation timestamp    |
| `created__gte`       | derived                           |                       |
| `created__lte`       | derived                           |                       |
| `last_updated`       | MultiValueDateTimeFilter          | Last update timestamp |
| `last_updated__gte`  | derived                           |                       |
| `last_updated__lte`  | derived                           |                       |
| `created_by_request` | UUIDFilter (method)               | Filter by request ID  |
| `updated_by_request` | UUIDFilter (method)               | Filter by request ID  |
| `tag`                | CharFilter (method, multi)        | Filter by tag slugs   |
| `id`                 | MultiValueNumberFilter            | Filter by IDs (CSV)   |
| `ordering`           | OrderingFilter                    | Sort results          |

---

## DCIM FilterSets

### SiteFilterSet

| Field         | Type                   | Notes         |
| ------------- | ---------------------- | ------------- |
| name          | MultiValueCharFilter   |               |
| slug          | MultiValueCharFilter   |               |
| status        | MultipleChoiceFilter   |               |
| region        | MultiValueNumberFilter | Region IDs    |
| region_id     | MultiValueNumberFilter |               |
| group         | MultiValueNumberFilter | SiteGroup IDs |
| group_id      | MultiValueNumberFilter |               |
| tenant        | MultiValueNumberFilter |               |
| tenant_id     | MultiValueNumberFilter |               |
| facility      | MultiValueCharFilter   |               |
| asn           | MultiValueNumberFilter |               |
| latitude      | MultiValueNumberFilter |               |
| longitude     | MultiValueNumberFilter |               |
| contact_name  | MultiValueCharFilter   |               |
| contact_email | MultiValueCharFilter   |               |
| has_clusters  | BooleanFilter (method) |               |

### DeviceFilterSet

| Field                  | Type                                | Notes |
| ---------------------- | ----------------------------------- | ----- |
| name                   | MultiValueCharFilter                |       |
| manufacturer           | MultiValueNumberFilter              |       |
| manufacturer_id        | MultiValueNumberFilter              |       |
| device_type            | MultiValueNumberFilter              |       |
| device_type_id         | MultiValueNumberFilter              |       |
| role                   | MultiValueNumberFilter              |       |
| role_id                | MultiValueNumberFilter              |       |
| tenant                 | MultiValueNumberFilter              |       |
| tenant_id              | MultiValueNumberFilter              |       |
| platform               | MultiValueNumberFilter              |       |
| platform_id            | MultiValueNumberFilter              |       |
| site                   | MultiValueNumberFilter              |       |
| site_id                | MultiValueNumberFilter              |       |
| location               | MultiValueNumberFilter              |       |
| location_id            | MultiValueNumberFilter              |       |
| rack                   | MultiValueNumberFilter              |       |
| rack_id                | MultiValueNumberFilter              |       |
| cluster                | MultiValueNumberFilter              |       |
| cluster_id             | MultiValueNumberFilter              |       |
| model                  | MultiValueCharFilter                |       |
| status                 | MultipleChoiceFilter                |       |
| serial                 | MultiValueCharFilter (method)       |       |
| asset_tag              | MultiValueCharFilter                |       |
| mac_address            | MultiValueMACAddressFilter (method) |       |
| has_primary_ip         | BooleanFilter (method)              |       |
| virtual_chassis_member | BooleanFilter (method)              |       |
| console_ports          | BooleanFilter (method)              |       |
| console_server_ports   | BooleanFilter (method)              |       |
| power_ports            | BooleanFilter (method)              |       |
| power_outlets          | BooleanFilter (method)              |       |
| interfaces             | BooleanFilter (method)              |       |
| pass_through_ports     | BooleanFilter (method)              |       |

### RackFilterSet

| Field       | Type                   | Notes |
| ----------- | ---------------------- | ----- |
| name        | MultiValueCharFilter   |       |
| facility_id | MultiValueCharFilter   |       |
| site        | MultiValueNumberFilter |       |
| site_id     | MultiValueNumberFilter |       |
| location    | MultiValueNumberFilter |       |
| location_id | MultiValueNumberFilter |       |
| tenant      | MultiValueNumberFilter |       |
| tenant_id   | MultiValueNumberFilter |       |
| status      | MultipleChoiceFilter   |       |
| role        | MultiValueNumberFilter |       |
| role_id     | MultiValueNumberFilter |       |
| serial      | MultiValueCharFilter   |       |
| asset_tag   | MultiValueCharFilter   |       |
| type        | MultiValueCharFilter   |       |
| width       | MultiValueNumberFilter |       |
| u_height    | MultiValueNumberFilter |       |
| desc_units  | BooleanFilter          |       |

---

## IPAM FilterSets

### PrefixFilterSet

| Field          | Type                            | Notes                         |
| -------------- | ------------------------------- | ----------------------------- |
| prefix         | CharFilter (method)             | IP network containment search |
| within         | CharFilter (method)             | prefix net_contained          |
| within_include | CharFilter (method)             | prefix net_contained_or_equal |
| contains       | CharFilter (method)             | prefix net_contains_or_equals |
| depth          | MultiValueNumberFilter          | Filter by _depth              |
| children       | MultiValueNumberFilter          | Filter by _children           |
| mask_length    | MultiValueNumberFilter (method) | Filter by prefix length       |
| vrf            | MultiValueNumberFilter          |                               |
| vrf_id         | MultiValueNumberFilter          |                               |
| tenant         | MultiValueNumberFilter          |                               |
| tenant_id      | MultiValueNumberFilter          |                               |
| vlan_id        | MultiValueNumberFilter          |                               |
| status         | MultipleChoiceFilter            |                               |
| role           | MultiValueNumberFilter          |                               |
| role_id        | MultiValueNumberFilter          |                               |
| is_pool        | BooleanFilter                   |                               |
| mark_utilized  | BooleanFilter                   |                               |
| present_in_vrf | CharFilter (method)             | Check import/export targets   |
| scope_type     | MultiValueCharFilter            |                               |
| scope_id       | MultiValueNumberFilter          |                               |

### IPAddressFilterSet

| Field                 | Type                            | Notes                    |
| --------------------- | ------------------------------- | ------------------------ |
| parent                | CharFilter (method)             | IP is within prefix      |
| address               | CharFilter (method)             | Exact address match      |
| mask_length           | MultiValueNumberFilter (method) | Filter by prefix length  |
| vrf                   | MultiValueNumberFilter          |                          |
| vrf_id                | MultiValueNumberFilter          |                          |
| tenant                | MultiValueNumberFilter          |                          |
| tenant_id             | MultiValueNumberFilter          |                          |
| status                | MultipleChoiceFilter            |                          |
| role                  | MultipleChoiceFilter            |                          |
| assigned_to_interface | BooleanFilter (method)          | Is assigned to interface |
| dns_name              | MultiValueCharFilter            |                          |
| present_in_vrf        | CharFilter (method)             |                          |

### VLANFilterSet

| Field      | Type                            | Notes |
| ---------- | ------------------------------- | ----- |
| site       | MultiValueNumberFilter          |       |
| site_id    | MultiValueNumberFilter          |       |
| group      | MultiValueNumberFilter          |       |
| group_id   | MultiValueNumberFilter          |       |
| vid        | MultiValueNumberFilter          |       |
| name       | MultiValueCharFilter            |       |
| tenant     | MultiValueNumberFilter          |       |
| tenant_id  | MultiValueNumberFilter          |       |
| status     | MultipleChoiceFilter            |       |
| role       | MultiValueNumberFilter          |       |
| role_id    | MultiValueNumberFilter          |       |
| region     | MultiValueNumberFilter (method) |       |
| region_id  | MultiValueNumberFilter (method) |       |
| site_group | MultiValueNumberFilter (method) |       |

---

## Circuits FilterSets

### CircuitFilterSet

| Field       | Type                            | Notes      |
| ----------- | ------------------------------- | ---------- |
| cid         | MultiValueCharFilter            | Circuit ID |
| provider    | MultiValueNumberFilter          |            |
| provider_id | MultiValueNumberFilter          |            |
| type        | MultiValueNumberFilter          |            |
| type_id     | MultiValueNumberFilter          |            |
| status      | MultipleChoiceFilter            |            |
| tenant      | MultiValueNumberFilter          |            |
| tenant_id   | MultiValueNumberFilter          |            |
| region      | MultiValueNumberFilter (method) |            |
| site        | MultiValueNumberFilter (method) |            |
| site_id     | MultiValueNumberFilter (method) |            |

### ProviderFilterSet

| Field         | Type                            | Notes |
| ------------- | ------------------------------- | ----- |
| name          | MultiValueCharFilter            |       |
| slug          | MultiValueCharFilter            |       |
| asn           | MultiValueNumberFilter          |       |
| account       | MultiValueCharFilter            |       |
| contact_name  | MultiValueCharFilter            |       |
| contact_email | MultiValueCharFilter            |       |
| region        | MultiValueNumberFilter (method) |       |
| site          | MultiValueNumberFilter (method) |       |
| site_group    | MultiValueNumberFilter (method) |       |

---

## Virtualization FilterSets

### VirtualMachineFilterSet

| Field            | Type                                | Notes |
| ---------------- | ----------------------------------- | ----- |
| name             | MultiValueCharFilter                |       |
| cluster          | MultiValueNumberFilter              |       |
| cluster_id       | MultiValueNumberFilter              |       |
| cluster_type     | MultiValueNumberFilter (method)     |       |
| cluster_type_id  | MultiValueNumberFilter (method)     |       |
| cluster_group    | MultiValueNumberFilter (method)     |       |
| cluster_group_id | MultiValueNumberFilter (method)     |       |
| device           | MultiValueNumberFilter (method)     |       |
| device_id        | MultiValueNumberFilter (method)     |       |
| status           | MultipleChoiceFilter                |       |
| tenant           | MultiValueNumberFilter              |       |
| tenant_id        | MultiValueNumberFilter              |       |
| platform         | MultiValueNumberFilter              |       |
| platform_id      | MultiValueNumberFilter              |       |
| role             | MultiValueNumberFilter              |       |
| role_id          | MultiValueNumberFilter              |       |
| region           | MultiValueNumberFilter (method)     |       |
| site             | MultiValueNumberFilter (method)     |       |
| site_id          | MultiValueNumberFilter (method)     |       |
| has_primary_ip   | BooleanFilter (method)              |       |
| mac_address      | MultiValueMACAddressFilter (method) |       |

### VMInterfaceFilterSet

| Field              | Type                       | Notes |
| ------------------ | -------------------------- | ----- |
| virtual_machine    | MultiValueNumberFilter     |       |
| virtual_machine_id | MultiValueNumberFilter     |       |
| name               | MultiValueCharFilter       |       |
| mac_address        | MultiValueMACAddressFilter |       |
| enabled            | BooleanFilter              |       |
| mtu                | MultiValueNumberFilter     |       |
| mode               | MultipleChoiceFilter       |       |

---

## Implementation and evidence checklist

For a promoted resource:

- [ ] Every accepted filter and ordering key exists in resource metadata.
- [ ] Unknown keys and unsupported lookup suffixes fail closed.
- [ ] Empty, repeated, comma-separated, null-like, boolean, numeric, date, and malformed values match the baseline where applicable.
- [ ] Search, relationship, network-containment, and method filters have source-linked behavior rather than guessed SQL equivalents.
- [ ] Visibility is applied before count and pagination.
- [ ] Ordering, null placement, collation-sensitive behavior, limit, and offset are exact or use the documented hybrid path.
- [ ] Repository predicates are explicit and parameterized.
- [ ] REST differential cases prove each declared key and relevant combination.
- [ ] Equivalent gRPC list semantics reach the same typed application query.
- [ ] The traceability matrix links metadata, source, Go tests, external cases, and retained evidence.
