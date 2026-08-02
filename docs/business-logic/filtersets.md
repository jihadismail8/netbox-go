# NetBox FilterSets — Dynamic Lookup System & Filter Reference

> Extracted from Python NetBox `filtersets.py` for porting to Go.
> **Source:** `netbox/netbox/{dcim,ipam,circuits,virtualization}/filtersets.py`, `netbox/netbox/filtersets.py`

---

## Dynamic Lookup Expression System

NetBox's `BaseFilterSet` auto-generates lookup variants for every declared filter at class creation time. For supported types whose `lookup_expr` is in `STANDARD_LOOKUPS = ('exact', 'iexact', 'in', 'contains')` and that do NOT use a custom `method=`, new filters are synthesized.

### Char-Based Lookup Map (CharFilter, MultiValueCharFilter, etc.)

| Suffix | Generated Lookup | Negation? | Example |
|--------|-----------------|-----------|---------|
| `n` | `exact` | yes | `name__n=john` -> name != john |
| `ic` | `icontains` | no | `name__ic=john` -> name ILIKE %john% |
| `nic` | `icontains` | yes | `name__nic=john` -> NOT ILIKE %john% |
| `iew` | `iendswith` | no | `name__iew=son` |
| `niew` | `iendswith` | yes | |
| `isw` | `istartswith` | no | `name__isw=john` |
| `nisw` | `istartswith` | yes | |
| `ie` | `iexact` | no | `name__ie=John` -> case-insensitive exact |
| `nie` | `iexact` | yes | |
| `empty` | `empty` (BooleanFilter) | no | `name__empty=true` -> name IS NULL OR '' |
| `regex` | `regex` | no | |
| `iregex` | `iregex` | no | |

### Numeric/Date Lookup Map (NumberFilter, DateTimeFilter, etc.)

| Suffix | Generated Lookup | Example |
|--------|-----------------|---------|
| `n` | `exact` (negated) | `asn__n=65001` |
| `lte` | `lte` | `created__lte=2024-01-01` |
| `lt` | `lt` | |
| `gte` | `gte` | |
| `gt` | `gt` | |

### Go Port Strategy

Implement a generic filter parser that:
1. Splits query param keys on `__`
2. Looks up the base field name (left of `__`)
3. Maps the suffix (right of `__`) to the appropriate SQL/GORM operation
4. Handles negation by wrapping in `NOT (...)`
5. Multi-value: accept comma-separated values -> SQL `IN (...)`

---

## Common Inherited Filters (ChangeLoggedModelFilterSet)

| Field | Type | Description |
|-------|------|-------------|
| `q` | CharFilter (method=filter_search) | Full-text search |
| `created` | MultiValueDateTimeFilter | Creation timestamp |
| `created__gte` | derived | |
| `created__lte` | derived | |
| `last_updated` | MultiValueDateTimeFilter | Last update timestamp |
| `last_updated__gte` | derived | |
| `last_updated__lte` | derived | |
| `created_by_request` | UUIDFilter (method) | Filter by request ID |
| `updated_by_request` | UUIDFilter (method) | Filter by request ID |
| `tag` | CharFilter (method, multi) | Filter by tag slugs |
| `id` | MultiValueNumberFilter | Filter by IDs (CSV) |
| `ordering` | OrderingFilter | Sort results |

---

## DCIM FilterSets

### SiteFilterSet

| Field | Type | Notes |
|-------|------|-------|
| name | MultiValueCharFilter | |
| slug | MultiValueCharFilter | |
| status | MultipleChoiceFilter | |
| region | MultiValueNumberFilter | Region IDs |
| region_id | MultiValueNumberFilter | |
| group | MultiValueNumberFilter | SiteGroup IDs |
| group_id | MultiValueNumberFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| facility | MultiValueCharFilter | |
| asn | MultiValueNumberFilter | |
| latitude | MultiValueNumberFilter | |
| longitude | MultiValueNumberFilter | |
| contact_name | MultiValueCharFilter | |
| contact_email | MultiValueCharFilter | |
| has_clusters | BooleanFilter (method) | |

### DeviceFilterSet

| Field | Type | Notes |
|-------|------|-------|
| name | MultiValueCharFilter | |
| manufacturer | MultiValueNumberFilter | |
| manufacturer_id | MultiValueNumberFilter | |
| device_type | MultiValueNumberFilter | |
| device_type_id | MultiValueNumberFilter | |
| role | MultiValueNumberFilter | |
| role_id | MultiValueNumberFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| platform | MultiValueNumberFilter | |
| platform_id | MultiValueNumberFilter | |
| site | MultiValueNumberFilter | |
| site_id | MultiValueNumberFilter | |
| location | MultiValueNumberFilter | |
| location_id | MultiValueNumberFilter | |
| rack | MultiValueNumberFilter | |
| rack_id | MultiValueNumberFilter | |
| cluster | MultiValueNumberFilter | |
| cluster_id | MultiValueNumberFilter | |
| model | MultiValueCharFilter | |
| status | MultipleChoiceFilter | |
| serial | MultiValueCharFilter (method) | |
| asset_tag | MultiValueCharFilter | |
| mac_address | MultiValueMACAddressFilter (method) | |
| has_primary_ip | BooleanFilter (method) | |
| virtual_chassis_member | BooleanFilter (method) | |
| console_ports | BooleanFilter (method) | |
| console_server_ports | BooleanFilter (method) | |
| power_ports | BooleanFilter (method) | |
| power_outlets | BooleanFilter (method) | |
| interfaces | BooleanFilter (method) | |
| pass_through_ports | BooleanFilter (method) | |

### RackFilterSet

| Field | Type | Notes |
|-------|------|-------|
| name | MultiValueCharFilter | |
| facility_id | MultiValueCharFilter | |
| site | MultiValueNumberFilter | |
| site_id | MultiValueNumberFilter | |
| location | MultiValueNumberFilter | |
| location_id | MultiValueNumberFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| status | MultipleChoiceFilter | |
| role | MultiValueNumberFilter | |
| role_id | MultiValueNumberFilter | |
| serial | MultiValueCharFilter | |
| asset_tag | MultiValueCharFilter | |
| type | MultiValueCharFilter | |
| width | MultiValueNumberFilter | |
| u_height | MultiValueNumberFilter | |
| desc_units | BooleanFilter | |

---

## IPAM FilterSets

### PrefixFilterSet

| Field | Type | Notes |
|-------|------|-------|
| prefix | CharFilter (method) | IP network containment search |
| within | CharFilter (method) | prefix net_contained |
| within_include | CharFilter (method) | prefix net_contained_or_equal |
| contains | CharFilter (method) | prefix net_contains_or_equals |
| depth | MultiValueNumberFilter | Filter by _depth |
| children | MultiValueNumberFilter | Filter by _children |
| mask_length | MultiValueNumberFilter (method) | Filter by prefix length |
| vrf | MultiValueNumberFilter | |
| vrf_id | MultiValueNumberFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| vlan_id | MultiValueNumberFilter | |
| status | MultipleChoiceFilter | |
| role | MultiValueNumberFilter | |
| role_id | MultiValueNumberFilter | |
| is_pool | BooleanFilter | |
| mark_utilized | BooleanFilter | |
| present_in_vrf | CharFilter (method) | Check import/export targets |
| scope_type | MultiValueCharFilter | |
| scope_id | MultiValueNumberFilter | |

### IPAddressFilterSet

| Field | Type | Notes |
|-------|------|-------|
| parent | CharFilter (method) | IP is within prefix |
| address | CharFilter (method) | Exact address match |
| mask_length | MultiValueNumberFilter (method) | Filter by prefix length |
| vrf | MultiValueNumberFilter | |
| vrf_id | MultiValueNumberFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| status | MultipleChoiceFilter | |
| role | MultipleChoiceFilter | |
| assigned_to_interface | BooleanFilter (method) | Is assigned to interface |
| dns_name | MultiValueCharFilter | |
| present_in_vrf | CharFilter (method) | |

### VLANFilterSet

| Field | Type | Notes |
|-------|------|-------|
| site | MultiValueNumberFilter | |
| site_id | MultiValueNumberFilter | |
| group | MultiValueNumberFilter | |
| group_id | MultiValueNumberFilter | |
| vid | MultiValueNumberFilter | |
| name | MultiValueCharFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| status | MultipleChoiceFilter | |
| role | MultiValueNumberFilter | |
| role_id | MultiValueNumberFilter | |
| region | MultiValueNumberFilter (method) | |
| region_id | MultiValueNumberFilter (method) | |
| site_group | MultiValueNumberFilter (method) | |

---

## Circuits FilterSets

### CircuitFilterSet

| Field | Type | Notes |
|-------|------|-------|
| cid | MultiValueCharFilter | Circuit ID |
| provider | MultiValueNumberFilter | |
| provider_id | MultiValueNumberFilter | |
| type | MultiValueNumberFilter | |
| type_id | MultiValueNumberFilter | |
| status | MultipleChoiceFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| region | MultiValueNumberFilter (method) | |
| site | MultiValueNumberFilter (method) | |
| site_id | MultiValueNumberFilter (method) | |

### ProviderFilterSet

| Field | Type | Notes |
|-------|------|-------|
| name | MultiValueCharFilter | |
| slug | MultiValueCharFilter | |
| asn | MultiValueNumberFilter | |
| account | MultiValueCharFilter | |
| contact_name | MultiValueCharFilter | |
| contact_email | MultiValueCharFilter | |
| region | MultiValueNumberFilter (method) | |
| site | MultiValueNumberFilter (method) | |
| site_group | MultiValueNumberFilter (method) | |

---

## Virtualization FilterSets

### VirtualMachineFilterSet

| Field | Type | Notes |
|-------|------|-------|
| name | MultiValueCharFilter | |
| cluster | MultiValueNumberFilter | |
| cluster_id | MultiValueNumberFilter | |
| cluster_type | MultiValueNumberFilter (method) | |
| cluster_type_id | MultiValueNumberFilter (method) | |
| cluster_group | MultiValueNumberFilter (method) | |
| cluster_group_id | MultiValueNumberFilter (method) | |
| device | MultiValueNumberFilter (method) | |
| device_id | MultiValueNumberFilter (method) | |
| status | MultipleChoiceFilter | |
| tenant | MultiValueNumberFilter | |
| tenant_id | MultiValueNumberFilter | |
| platform | MultiValueNumberFilter | |
| platform_id | MultiValueNumberFilter | |
| role | MultiValueNumberFilter | |
| role_id | MultiValueNumberFilter | |
| region | MultiValueNumberFilter (method) | |
| site | MultiValueNumberFilter (method) | |
| site_id | MultiValueNumberFilter (method) | |
| has_primary_ip | BooleanFilter (method) | |
| mac_address | MultiValueMACAddressFilter (method) | |

### VMInterfaceFilterSet

| Field | Type | Notes |
|-------|------|-------|
| virtual_machine | MultiValueNumberFilter | |
| virtual_machine_id | MultiValueNumberFilter | |
| name | MultiValueCharFilter | |
| mac_address | MultiValueMACAddressFilter | |
| enabled | BooleanFilter | |
| mtu | MultiValueNumberFilter | |
| mode | MultipleChoiceFilter | |

---

## Go Port Implementation Notes

### Filter Registration Pattern

For each model, create a filter registration function that maps query param names to GORM scopes.

```go
type FilterConfig struct {
    Field       string
    Type        FilterType  // exact, icontains, in, custom
    Method      func(*gorm.DB, interface{}) *gorm.DB
    Negatable   bool
}
```

### Dynamic Lookup Generation

```go
func ParseFilter(key string, values []string, config FilterConfig) func(*gorm.DB) *gorm.DB {
    parts := strings.SplitN(key, "__", 2)
    baseField := parts[0]

    if len(parts) == 1 {
        return applyExact(config, values)  // default: exact or IN
    }

    suffix := parts[1]
    switch suffix {
    case "ic":   return applyILike(config, values, false)
    case "nic":  return applyILike(config, values, true)
    case "ie":   return applyIExact(config, values, false)
    case "nie":  return applyIExact(config, values, true)
    case "isw":  return applyIStartsWith(config, values, false)
    case "nisw": return applyIStartsWith(config, values, true)
    case "iew":  return applyIEndsWith(config, values, false)
    case "niew": return applyIEndsWith(config, values, true)
    case "n":    return applyExact(config, values)  // with negation
    case "gte":  return applyGte(config, values)
    case "lte":  return applyLte(config, values)
    case "gt":   return applyGt(config, values)
    case "lt":   return applyLt(config, values)
    case "empty": return applyEmpty(config, values)
    case "regex":  return applyRegex(config, values, false)
    case "iregex": return applyRegex(config, values, true)
    }
    return nil
}
```
